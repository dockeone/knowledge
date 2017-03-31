1. statefulset中的pod必须使用StorageClass机制的PersistentVolume Provisioner，k8s根据定义中的 volumeClaimTemplates来动态生成PVC和PV；
1. statefulset当前需要创建一个Headless Service来创建Pod的网络标识(the network identity of the Pods)，即通过$(podname).$(service name).$(namespace)来找到各Pod，需要在StatefulSet之前创建好；
1. 删除statefulset时不会自动删除PVC和PV；删除PVC后，才会真正删除对应的PV；
1. 如果一个PV是PVC动态provisioned(如statefulset为每个Pod创建的PVC)，则该PV将一直和对应的PVC绑定；所以对于statefulset pod的PV，如果后续该Pod被调度到其它Node，会一直使用该PVC和PV； 

If a PV was dynamically provisioned for a new PVC, the loop will always bind that PV to the PVC. Otherwise, the user will always get at least what they asked for, but the volume may be in excess of what was requested. Once bound;


# 创建ceph rbd StorageClass

$ ceph auth get-key /etc/ceph/client.admin
AQD+EsRYOM2kARAAt6LxrFX+7u7GFBfY3WF6aQ==
$ echo 'AQD+EsRYOM2kARAAt6LxrFX+7u7GFBfY3WF6aQ=='|base64
QVFEK0VzUllPTTJrQVJBQXQ2THhyRlgrN3U3R0ZCZlkzV0Y2YVE9PQo=

$  cat ceph-secret-admin.yal
apiVersion: v1
kind: Secret
metadata:
  name: ceph-secret-admin
data:
  key: QVFEK0VzUllPTTJrQVJBQXQ2THhyRlgrN3U3R0ZCZlkzV0Y2YVE9PQo=

$ kubectl create -f ceph-secret-admin.yal

注意，如果是通过命令行创建secret-admin，则**不需要**对key进行base64编码：
kubectl create secret generic ceph-secret-admin --from-literal=key='AQD+EsRYOM2kARAAt6LxrFX+7u7GFBfY3WF6aQ==' --namespace=kube-system --type=kubernetes.io/rbd

$ cat rbd-storage-class.yaml
apiVersion: storage.k8s.io/v1beta1  # 注意，k8s 1.5.5版本对应的是 v1beta1，最新的1.6版本是 v1
kind: StorageClass
metadata:
   name: slow
provisioner: kubernetes.io/rbd
parameters:
    monitors: 10.64.3.9:6789
    adminId: admin
    adminSecretName: ceph-secret-admin
    adminSecretNamespace: "default"
    pool: rbd
    userId: admin
    userSecretName: ceph-secret-admin

$ kubectl get storageclass
NAME      TYPE
slow      kubernetes.io/rbd

# 创建PersistentVolumeClaim

$ cat statefulset-claim.json # 对应**测试的1.5.4版本**
{
  "kind": "PersistentVolumeClaim",
  "apiVersion": "v1",
  "metadata": {
    "name": "statefulset-claim",
    "annotations": {
    "volume.beta.kubernetes.io/storage-class": "slow"
    }
  },
  "spec": {
    "accessModes": [
      "ReadWriteOnce"
    ],
    "resources": {
      "requests": {
        "storage": "1Gi"
      }
    }
  }
}


$ cat statefulset-claim.json # 对应的是**最新1.6版本**
{
  "kind": "PersistentVolumeClaim",
  "apiVersion": "v1",
  "metadata": {
    "name": "statefulset-claim"
  },
  "spec": {
    "accessModes": [
      "ReadWriteOnce"
    ],
    "resources": {
      "requests": {
        "storage": "1Gi"
      }
    },
    "storageClassName": "slow"
  }
}

$ kubectl create -f statefulset-claim.json
persistentvolumeclaim "statefulset-claim" created

$ kubectl get persistentvolumeclaim
NAME                STATUS    VOLUME    CAPACITY   ACCESSMODES   AGE
ceph-claim          Bound     ceph-pv   1Gi        RWO           14d
statefulset-claim   Pending                                      20s

$ kubectl get persistentvolumeclaim
NAME                STATUS    VOLUME                                     CAPACITY   ACCESSMODES   AGE
ceph-claim          Bound     ceph-pv                                    1Gi        RWO           14d
statefulset-claim   Bound     pvc-9ba09509-12db-11e7-b7fe-8cdcd4b3be48   1Gi        RWO           6m

可以，已经自动创建了PV并绑定；

$ kubectl get pv
NAME                                       CAPACITY   ACCESSMODES   RECLAIMPOLICY   STATUS    CLAIM                       REASON    AGE
ceph-pv                                    1Gi        RWO           Recycle         Bound     default/ceph-claim                    14d
pvc-9ba09509-12db-11e7-b7fe-8cdcd4b3be48   1Gi        RWO           Delete          Bound     default/statefulset-claim             5m

$ kubectl describe pv pvc-9ba09509-12db-11e7-b7fe-8cdcd4b3be48
Name:           pvc-9ba09509-12db-11e7-b7fe-8cdcd4b3be48
Labels:         <none>
StorageClass:   slow
Status:         Bound
Claim:          default/statefulset-claim
Reclaim Policy: Delete
Access Modes:   RWO
Capacity:       1Gi
Message:
Source:
    Type:               RBD (a Rados Block Device mount on the host that shares a pod's lifetime)
    CephMonitors:       [10.64.3.9:6789]
    RBDImage:           kubernetes-dynamic-pvc-ef65887b-12db-11e7-b0ce-8cdcd4b3be48
    FSType:
    RBDPool:            rbd
    RadosUser:          admin
    Keyring:            /etc/ceph/keyring
    SecretRef:          &{ceph-secret-admin}
    ReadOnly:           false
No events.

## 创建使用PVC的pod
[root@tjwq01-sys-bs003007 ~]# cat ceph-rbd.yaml
apiVersion: v1
kind: Pod
metadata:
  name: ceph-rbd
spec:
  containers:
  - name: ceph-busybox1
    image: nginx:1.7.9
    volumeMounts:
    - name: ceph-vol1
      mountPath: /usr/share/busybox
      readOnly: false
  volumes:
  - name: ceph-vol1
    persistentVolumeClaim:
      claimName: statefulset-claim


# 测试 statefulset

$ cat statefulset-nginx.json
---
apiVersion: v1
kind: Service
metadata:
  name: statefulset-nginx
  labels:
    app: statefulset-nginx
spec:
  ports:
  - port: 80
    name: web
  clusterIP: None
  selector:
    app: statefulset-nginx
---
apiVersion: apps/v1beta1
kind: StatefulSet
metadata:
  name: web
spec:
  serviceName: "statefulset-nginx"
  replicas: 3
  template:
    metadata:
      labels:
        app: statefulset-nginx
    spec:
      terminationGracePeriodSeconds: 10
      containers:
      - name: nginx
        image: nginx:1.7.9
        ports:
        - containerPort: 80
          name: web
        volumeMounts:
        - name: www
          mountPath: /usr/share/nginx/html
  volumeClaimTemplates:
  - metadata:
      name: www
      annotations:
        volume.beta.kubernetes.io/storage-class: slow # 测试的是1.5.4版本，在annotations里面指定storage class
    spec:
      accessModes: [ "ReadWriteOnce" ]
      resources:
        requests:
          storage: 1Gi

注意：
1. 上面测试的是1.5.4版本，在annotations里面指定storage class；
2. 如果是新版如1.6，则 volumeClaimTemplates 的配置如下：
  volumeClaimTemplates:
  - metadata:
      name: www
    spec:
      accessModes: [ "ReadWriteOnce" ]
      resources:
        requests:
          storage: 1Gi
      storageClassName: "slow"

正确地创建了三个pods：
$ kubectl get statefulset
NAME      DESIRED   CURRENT   AGE
web       3         3         11s

## statefulset的特点
+ 按顺序创建各Pod，pod名称(hostname)最后一位为顺序号，从0开始；
+ 自动为每个Pod按照volumeClaimTemplates创建一个PVC和PV；
+ 为各pod的hostname创建了DNS A记录，在pod中ping $(podname).$(service name).$(namespace) 是通的；

$ kubectl describe statefulset web
Name:                   web
Namespace:              default
Image(s):               nginx:1.7.9
Selector:               app=statefulset-nginx
Labels:                 app=statefulset-nginx
Replicas:               3 current / 3 desired
Annotations:            <none>
CreationTimestamp:      Mon, 27 Mar 2017 20:20:37 +0800
Pods Status:            3 Running / 0 Waiting / 0 Succeeded / 0 Failed
No volumes.
Events:
  FirstSeen     LastSeen        Count   From            SubObjectPath   Type            Reason                  Message
  ---------     --------        -----   ----            -------------   --------        ------                  -------
  3m            3m              1       {statefulset }                  Normal          SuccessfulCreate        pvc: www-web-0
  3m            3m              1       {statefulset }                  Normal          SuccessfulCreate        pet: web-0
  3m            3m              1       {statefulset }                  Normal          SuccessfulCreate        pvc: www-web-1
  3m            3m              1       {statefulset }                  Normal          SuccessfulCreate        pvc: www-web-2
  3m            3m              1       {statefulset }                  Normal          SuccessfulCreate        pet: web-1
  3m            3m              1       {statefulset }                  Normal          SuccessfulCreate        pet: web-2

$ kubectl get pods -o wide
NAME                               READY     STATUS    RESTARTS   AGE       IP             NODE
deployment-demo-2940201158-1n7cc   1/1       Running   0          3h        172.30.19.15   10.64.3.7
deployment-demo-2940201158-4tvjn   1/1       Running   0          3h        172.30.19.14   10.64.3.7
deployment-demo-2940201158-7npf4   1/1       Running   0          3h        172.30.19.16   10.64.3.7
deployment-demo-2940201158-nh7pv   1/1       Running   0          3h        172.30.19.18   10.64.3.7
my-nginx-3467165555-0wnw3          1/1       Running   0          1d        172.30.19.6    10.64.3.7
my-nginx-3467165555-t3g84          1/1       Running   0          1d        172.30.19.2    10.64.3.7
web-0                              1/1       Running   0          15m       172.30.19.17   10.64.3.7
web-1                              1/1       Running   0          15m       172.30.19.19   10.64.3.7
web-2                              1/1       Running   0          15m       172.30.19.20   10.64.3.7

由于创建的是 Headless Service，所以ClusterIP为空：
$ kubectl get services
NAME                  CLUSTER-IP       EXTERNAL-IP   PORT(S)   AGE
deployment-demo-svc   10.254.137.54    <none>        80/TCP    4h
kubernetes            10.254.0.1       <none>        443/TCP   100d
my-nginx              10.254.0.159     <none>        80/TCP    13d
rc-demo-svc           10.254.144.118   <none>        80/TCP    4h
statefulset-nginx     None             <none>        80/TCP    24m

为各pod的hostname创建了DNS A记录，在pod中ping $(podname).$(service name).$(namespace) 是通的：
$ kubectl exec web-0 -it  bash
root@web-0:/# ping web-1
ping: unknown host
root@web-0:/# ping web-2
ping: unknown host
root@web-0:/# cat /etc/resolv.conf
search default.svc.cluster.local svc.cluster.local cluster.local tjwq01.ksyun.com
nameserver 10.254.0.2
nameserver 10.64.116.10
nameserver 10.64.116.11
options ndots:5
root@web-0:/# ping web-1.statefulset-nginx
PING web-1.statefulset-nginx.default.svc.cluster.local (172.30.19.19): 48 data bytes
56 bytes from 172.30.19.19: icmp_seq=0 ttl=64 time=0.129 ms
^C--- web-1.statefulset-nginx.default.svc.cluster.local ping statistics ---
1 packets transmitted, 1 packets received, 0% packet loss
round-trip min/avg/max/stddev = 0.129/0.129/0.129/0.000 ms
root@web-0:/# ping web-2.statefulset-nginx
PING web-2.statefulset-nginx.default.svc.cluster.local (172.30.19.20): 48 data bytes
56 bytes from 172.30.19.20: icmp_seq=0 ttl=64 time=0.140 ms
56 bytes from 172.30.19.20: icmp_seq=1 ttl=64 time=0.064 ms
^C--- web-2.statefulset-nginx.default.svc.cluster.local ping statistics ---
2 packets transmitted, 2 packets received, 0% packet loss
round-trip min/avg/max/stddev = 0.064/0.102/0.140/0.038 ms
root@web-0:/#


## 如果删除了statefulset，k8s不会自动删除PVC和PV，需要手动删除PVC：
$ kubectl delete statefulset web
statefulset "web" deleted
$ kubectl get pvc
NAME                STATUS    VOLUME                                     CAPACITY   ACCESSMODES   AGE
www-web-0           Pending                                                                       9m
www-web-1           Pending                                                                       9m
www-web-2           Pending                                                                       9m

$ kubectl delete pvc www-web-0 www-web-1 www-web-2 # 手动删除pvc
persistentvolumeclaim "www-web-0" deleted
persistentvolumeclaim "www-web-1" deleted
persistentvolumeclaim "www-web-2" deleted

# 手动删除PVC后，statefulset自动创建同名的PVC并创建新的PV，然后自动挂在到Pod
$ kubectl delete pvc www-web-0
persistentvolumeclaim "www-web-0" deleted

$ kubectl describe statefulset
Name:                   web
Namespace:              default
Image(s):               nginx:1.7.9
Selector:               app=statefulset-nginx
Labels:                 app=statefulset-nginx
Replicas:               3 current / 3 desired
Annotations:            <none>
CreationTimestamp:      Mon, 27 Mar 2017 20:20:37 +0800
Pods Status:            3 Running / 0 Waiting / 0 Succeeded / 0 Failed
No volumes.
Events:
  FirstSeen     LastSeen        Count   From            SubObjectPath   Type            Reason                  Message
  ---------     --------        -----   ----            -------------   --------        ------                  -------
  2h            10s             2       {statefulset }                  Normal          SuccessfulCreate        pvc: www-web-0

$  kubectl get pvc
NAME                STATUS    VOLUME                                     CAPACITY   ACCESSMODES   AGE
www-web-0           Bound     pvc-2bfab38d-12ff-11e7-b7fe-8cdcd4b3be48   1Gi        RWO           34s
www-web-1           Bound     pvc-c8bd936c-12e7-11e7-b7fe-8cdcd4b3be48   1Gi        RWO           2h
www-web-2           Bound     pvc-c8be2d92-12e7-11e7-b7fe-8cdcd4b3be48   1Gi        RWO           2h

$ kubectl get pv  # 注意以前的 www-web-0 PV处于Failed状态，新建了一个www-web-0 PV
NAME                                       CAPACITY   ACCESSMODES   RECLAIMPOLICY   STATUS    CLAIM                       REASON    AGE
pvc-2bfab38d-12ff-11e7-b7fe-8cdcd4b3be48   1Gi        RWO           Delete          Bound     default/www-web-0                     50s
pvc-c8bcdf06-12e7-11e7-b7fe-8cdcd4b3be48   1Gi        RWO           Delete          Failed    default/www-web-0                     2h
pvc-c8bd936c-12e7-11e7-b7fe-8cdcd4b3be48   1Gi        RWO           Delete          Bound     default/www-web-1                     2h
pvc-c8be2d92-12e7-11e7-b7fe-8cdcd4b3be48   1Gi        RWO           Delete          Bound     default/www-web-2                     2h

$ kubectl describe pv pvc-c8bcdf06-12e7-11e7-b7fe-8cdcd4b3be48
Name:           pvc-c8bcdf06-12e7-11e7-b7fe-8cdcd4b3be48
Labels:         <none>
StorageClass:   slow
Status:         Failed
Claim:          default/www-web-0
Reclaim Policy: Delete
Access Modes:   RWO
Capacity:       1Gi
Message:        rbd kubernetes-dynamic-pvc-c8bd9287-12e7-11e7-b0ce-8cdcd4b3be48 is still being used
Source:
    Type:               RBD (a Rados Block Device mount on the host that shares a pod's lifetime)
    CephMonitors:       [10.64.3.9:6789]
    RBDImage:           kubernetes-dynamic-pvc-c8bd9287-12e7-11e7-b0ce-8cdcd4b3be48
    FSType:
    RBDPool:            rbd
    RadosUser:          admin
    Keyring:            /etc/ceph/keyring
    SecretRef:          &{ceph-secret-admin}
    ReadOnly:           false
Events:
  FirstSeen     LastSeen        Count   From                            SubObjectPath   Type            Reason                  Message
  ---------     --------        -----   ----                            -------------   --------        ------                  -------
  10m           10m             1       {persistentvolume-controller }                  Warning         VolumeFailedDelete      rbd kubernetes-dynamic-pvc-c8bd9287-12e7-11e7-b0ce-8cdcd4b3be48 is still being used

由于删除PVC时，PV正在被pod使用，所以k8s自动删除PVC关联的PV失败，现在Pod已经mount了新的PV，故可以手动删除PV：
$ kubectl delete  pv pvc-c8bcdf06-12e7-11e7-b7fe-8cdcd4b3be48
persistentvolume "pvc-c8bcdf06-12e7-11e7-b7fe-8cdcd4b3be48" deleted
$ kubectl get pv
NAME                                       CAPACITY   ACCESSMODES   RECLAIMPOLICY   STATUS    CLAIM                REASON    AGE
ceph-pv                                    1Gi        RWO           Recycle         Bound     default/ceph-claim             14d
pvc-2bfab38d-12ff-11e7-b7fe-8cdcd4b3be48   1Gi        RWO           Delete          Bound     default/www-web-0              14m
pvc-c8bd936c-12e7-11e7-b7fe-8cdcd4b3be48   1Gi        RWO           Delete          Bound     default/www-web-1              3h
pvc-c8be2d92-12e7-11e7-b7fe-8cdcd4b3be48   1Gi        RWO           Delete          Bound     default/www-web-2              3h

# 手动删除PV后，PVC状态异常，但是Pod仍然可以使用PV

[root@tjwq01-sys-bs003007 ~]# kubectl delete pv pvc-2bfab38d-12ff-11e7-b7fe-8cdcd4b3be48
persistentvolume "pvc-2bfab38d-12ff-11e7-b7fe-8cdcd4b3be48" deleted
[root@tjwq01-sys-bs003007 ~]# kubectl get pv
NAME                                       CAPACITY   ACCESSMODES   RECLAIMPOLICY   STATUS    CLAIM                REASON    AGE
ceph-pv                                    1Gi        RWO           Recycle         Bound     default/ceph-claim             14d
pvc-c8bd936c-12e7-11e7-b7fe-8cdcd4b3be48   1Gi        RWO           Delete          Bound     default/www-web-1              3h
pvc-c8be2d92-12e7-11e7-b7fe-8cdcd4b3be48   1Gi        RWO           Delete          Bound     default/www-web-2              3h
[root@tjwq01-sys-bs003007 ~]# kubectl get pvc
NAME         STATUS    VOLUME                                     CAPACITY   ACCESSMODES   AGE
ceph-claim   Bound     ceph-pv                                    1Gi        RWO           14d
www-web-0    Lost      pvc-2bfab38d-12ff-11e7-b7fe-8cdcd4b3be48   0                        17m
www-web-1    Bound     pvc-c8bd936c-12e7-11e7-b7fe-8cdcd4b3be48   1Gi        RWO           3h
www-web-2    Bound     pvc-c8be2d92-12e7-11e7-b7fe-8cdcd4b3be48   1Gi        RWO           3h

Pod仍然挂载删除的PV，读写正常：
[root@tjwq01-sys-bs003007 ~]# kubectl get pod web-0 -o yaml|grep uid
  uid: c8bd136f-12e7-11e7-b7fe-8cdcd4b3be48
[root@tjwq01-sys-bs003007 ~]# mount|grep 'c8bd136f-12e7-11e7-b7fe-8cdcd4b3be48'
tmpfs on /var/lib/kubelet/pods/c8bd136f-12e7-11e7-b7fe-8cdcd4b3be48/volumes/kubernetes.io~secret/default-token-8khbh type tmpfs (rw,relatime)
/dev/rbd0 on /var/lib/kubelet/pods/c8bd136f-12e7-11e7-b7fe-8cdcd4b3be48/volumes/kubernetes.io~rbd/pvc-c8bcdf06-12e7-11e7-b7fe-8cdcd4b3be48 type ext4 (rw,relatime,stripe=1024,data=ordered)
[root@tjwq01-sys-bs003007 ~]# kubectl exec  web-0 -it bash
root@web-0:/# ls /usr/share/nginx/html/
lost+found
root@web-0:/# mount|grep html
/dev/rbd0 on /usr/share/nginx/html type ext4 (rw,relatime,stripe=1024,data=ordered)
root@web-0:/# echo test >/usr/share/nginx/html/test
root@web-0:/# cat /usr/share/nginx/html/test
test

删除PV对应的PVC，k8s自动创建新的同名PVC，挂载以前的PV(PV的名字发生了改变)：
[root@tjwq01-sys-bs003007 ~]# kubectl delete pvc www-web-0
persistentvolumeclaim "www-web-0" deleted
[root@tjwq01-sys-bs003007 ~]# kubectl get pvc
NAME         STATUS    VOLUME                                     CAPACITY   ACCESSMODES   AGE
ceph-claim   Bound     ceph-pv                                    1Gi        RWO           14d
www-web-0    Bound     pvc-864a9f72-1302-11e7-b7fe-8cdcd4b3be48   1Gi        RWO           6s
www-web-1    Bound     pvc-c8bd936c-12e7-11e7-b7fe-8cdcd4b3be48   1Gi        RWO           3h
www-web-2    Bound     pvc-c8be2d92-12e7-11e7-b7fe-8cdcd4b3be48   1Gi        RWO           3h
[root@tjwq01-sys-bs003007 ~]# kubectl exec  web-0 -it bash
root@web-0:/# cat /usr/share/nginx/html/
lost+found/ test
root@web-0:/# cat /usr/share/nginx/html/test
test
root@web-0:/# exit


[root@tjwq01-sys-bs003007 ~]# kubectl get statefulset
NAME      DESIRED   CURRENT   AGE
web       3         3         3h
[root@tjwq01-sys-bs003007 ~]# kubectl delete statefulset web # 删除statefulset和对应的pods
statefulset "web" deleted
[root@tjwq01-sys-bs003007 ~]# kubectl get pvc  # 但是PVC和PV都还存在；
NAME         STATUS    VOLUME                                     CAPACITY   ACCESSMODES   AGE
ceph-claim   Bound     ceph-pv                                    1Gi        RWO           14d
www-web-0    Bound     pvc-864a9f72-1302-11e7-b7fe-8cdcd4b3be48   1Gi        RWO           10m
www-web-1    Bound     pvc-c8bd936c-12e7-11e7-b7fe-8cdcd4b3be48   1Gi        RWO           3h
www-web-2    Bound     pvc-c8be2d92-12e7-11e7-b7fe-8cdcd4b3be48   1Gi        RWO           3h

[root@tjwq01-sys-bs003007 ~]# mount|grep rbd  # Node自动umount挂载的RBD
[root@tjwq01-sys-bs003007 ~]#
[root@tjwq01-sys-bs003007 ~]# rbd list  # 自动创建的RBD都还存在
kubernetes-dynamic-pvc-2bfbb97c-12ff-11e7-b0ce-8cdcd4b3be48
kubernetes-dynamic-pvc-864b8d14-1302-11e7-b0ce-8cdcd4b3be48
kubernetes-dynamic-pvc-c8bd9287-12e7-11e7-b0ce-8cdcd4b3be48
kubernetes-dynamic-pvc-c8be598f-12e7-11e7-b0ce-8cdcd4b3be48
kubernetes-dynamic-pvc-c8bf0747-12e7-11e7-b0ce-8cdcd4b3be48
kubernetes-dynamic-pvc-ef65887b-12db-11e7-b0ce-8cdcd4b3be48
ceph-image
foo
prometheus
zabbix_database
[root@tjwq01-sys-bs003007 ~]# kubectl get pvc
NAME         STATUS    VOLUME                                     CAPACITY   ACCESSMODES   AGE
ceph-claim   Bound     ceph-pv                                    1Gi        RWO           14d
www-web-0    Bound     pvc-864a9f72-1302-11e7-b7fe-8cdcd4b3be48   1Gi        RWO           13m
www-web-1    Bound     pvc-c8bd936c-12e7-11e7-b7fe-8cdcd4b3be48   1Gi        RWO           3h
www-web-2    Bound     pvc-c8be2d92-12e7-11e7-b7fe-8cdcd4b3be48   1Gi        RWO           3h

**删除PVC后，对应的PV才被真正删除！**
[root@tjwq01-sys-bs003007 ~]# kubectl delete pvc www-web-0
persistentvolumeclaim "www-web-0" deleted
[root@tjwq01-sys-bs003007 ~]# kubectl get pv
NAME                                       CAPACITY   ACCESSMODES   RECLAIMPOLICY   STATUS    CLAIM                REASON    AGE
ceph-pv                                    1Gi        RWO           Recycle         Bound     default/ceph-claim             14d
pvc-c8bd936c-12e7-11e7-b7fe-8cdcd4b3be48   1Gi        RWO           Delete          Bound     default/www-web-1              3h
pvc-c8be2d92-12e7-11e7-b7fe-8cdcd4b3be48   1Gi        RWO           Delete          Bound     default/www-web-2              3h
[root@tjwq01-sys-bs003007 ~]# rbd list
2017-03-27 23:46:03.645217 7fac9a94d700  0 -- :/1040524 >> 10.64.3.86:6789/0 pipe(0x7fac91529000 sd=3 :0 s=1 pgs=0 cs=0 l=1 c=0x7fac91451600).fault
kubernetes-dynamic-pvc-2bfbb97c-12ff-11e7-b0ce-8cdcd4b3be48
kubernetes-dynamic-pvc-c8bd9287-12e7-11e7-b0ce-8cdcd4b3be48
kubernetes-dynamic-pvc-c8be598f-12e7-11e7-b0ce-8cdcd4b3be48
kubernetes-dynamic-pvc-c8bf0747-12e7-11e7-b0ce-8cdcd4b3be48
kubernetes-dynamic-pvc-ef65887b-12db-11e7-b0ce-8cdcd4b3be48
ceph-image
foo
prometheus
zabbix_database

## 重新创建statefulset，以前的PVC和PV被重用
[root@tjwq01-sys-bs003007 ~]# kubectl create -f statefulset-nginx.json
statefulset "web" created
Error from server (AlreadyExists): error when creating "statefulset-nginx.json": services "statefulset-nginx" already exists
[root@tjwq01-sys-bs003007 ~]# kubectl get pvc
NAME         STATUS    VOLUME                                     CAPACITY   ACCESSMODES   AGE
ceph-claim   Bound     ceph-pv                                    1Gi        RWO           14d
www-web-0    Bound     pvc-bb0f7f2e-1304-11e7-b7fe-8cdcd4b3be48   1Gi        RWO           3s
www-web-1    Bound     pvc-c8bd936c-12e7-11e7-b7fe-8cdcd4b3be48   1Gi        RWO           3h
www-web-2    Bound     pvc-c8be2d92-12e7-11e7-b7fe-8cdcd4b3be48   1Gi        RWO           3h
[root@tjwq01-sys-bs003007 ~]# kubectl get pv
NAME                                       CAPACITY   ACCESSMODES   RECLAIMPOLICY   STATUS    CLAIM                REASON    AGE
ceph-pv                                    1Gi        RWO           Recycle         Bound     default/ceph-claim             14d
pvc-bb0f7f2e-1304-11e7-b7fe-8cdcd4b3be48   1Gi        RWO           Delete          Bound     default/www-web-0              16s
pvc-c8bd936c-12e7-11e7-b7fe-8cdcd4b3be48   1Gi        RWO           Delete          Bound     default/www-web-1              3h
pvc-c8be2d92-12e7-11e7-b7fe-8cdcd4b3be48   1Gi        RWO           Delete          Bound     default/www-web-2              3h