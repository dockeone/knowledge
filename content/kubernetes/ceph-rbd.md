1. pod直接挂载RBD image；
1. 创建PV和PVC，pod使用PVC挂载RBD Image；
1. 创建RBD StorageClass、PVC，pod使用PVC挂载RBD Image；

# 一、pod挂载已创建的RBD Image

1. 创建rbd image
[zhangjun3@bjzjm-sys-power005010 ~]$ rbd create foo -s 1024
[zhangjun3@bjzjm-sys-power005010 ~]$ rbd list
foo
2. 默认map会失败，原因是CentoOS kernel不支持全部特性；
[zhangjun3@bjzjm-sys-power005010 ~]$ sudo rbd map foo
rbd: sysfs write failed
RBD image feature set mismatch. You can disable features unsupported by the kernel with "rbd feature disable".
In some cases useful info is found in syslog - try "dmesg | tail" or so.
rbd: map failed: (6) No such device or address
[zhangjun3@bjzjm-sys-power005010 ~]$ rbd info foo
rbd image 'foo':
        size 1024 MB in 256 objects
        order 22 (4096 kB objects)
        block_name_prefix: rbd_data.101a74b0dc51
        format: 2
        features: layering, exclusive-lock, object-map, fast-diff, deep-flatten
        flags:

3. 解决方法是disable部分rbd特性，如果没有disable不支持的feature，k8s在创建pod时会一直处于ContainerCreating状态：
[zhangjun3@bjzjm-sys-power005010 ~]$ rbd feature disable foo fast-diff object-map  exclusive-lock, deep-flatten
[zhangjun3@tjwq01-sys-power003009 ceph]$ sudo rbd map foo
/dev/rbd0
[zhangjun3@tjwq01-sys-power003009 ceph]$ sudo mkfs.ext4 /dev/rbd0 # 格式化

## ceph的默认配置文件启用了auth，k8s需要通过Secret的形式来访问ceph rbd
[zhangjun3@tjwq01-sys-power003008 ceph-cluster]$ cat ceph.conf
[global]
fsid = 0dca8efc-5444-4fa0-88a8-2c0751b47d28
mon_initial_members = tjwq01-sys-power003009
mon_host = 10.64.3.9
auth_cluster_required = cephx
auth_service_required = cephx
auth_client_required = cephx

osd pool default size = 2

参考文件：
https://github.com/kubernetes/kubernetes/tree/master/examples/volumes/rbd

[zhangjun3@tjwq01-sys-power003009 ceph]$ ceph auth get-key /etc/ceph/client.admin
AQD+EsRYOM2kARAAt6LxrFX+7u7GFBfY3WF6aQ==
[zhangjun3@tjwq01-sys-power003009 ceph]$ echo 'AQD+EsRYOM2kARAAt6LxrFX+7u7GFBfY3WF6aQ=='|base64
QVFEK0VzUllPTTJrQVJBQXQ2THhyRlgrN3U3R0ZCZlkzV0Y2YVE9PQo=

[root@tjwq01-sys-bs003007 ~]# cat ceph-secret.yaml
apiVersion: v1
kind: Secret
metadata:
  name: ceph-secret
data:
  key: QVFEK0VzUllPTTJrQVJBQXQ2THhyRlgrN3U3R0ZCZlkzV0Y2YVE9PQo=

注意，如果是通过命令行创建secret-admin，则**不需要**对key进行base64编码：
kubectl create secret generic ceph-secret-admin --from-literal=key='AQD+EsRYOM2kARAAt6LxrFX+7u7GFBfY3WF6aQ==' --namespace=kube-system --type=kubernetes.io/rbd

[root@tjwq01-sys-bs003007 ~]# kubectl create -f ceph-secret.yaml
secret "ceph-secret" created
[root@tjwq01-sys-bs003007 ~]# kubectl get secret
NAME                  TYPE                                  DATA      AGE
ceph-secret           Opaque                                1         10s
default-token-icgdo   kubernetes.io/service-account-token   2         85d

[root@tjwq01-sys-bs003007 ~]# yum install ceph-common # 使用ceph rbd的node上需要安装ceph-common 
[root@tjwq01-sys-bs003007 ~]# cat rbd-with-secret.json # 注意是只读挂载
{
    "apiVersion": "v1",
    "kind": "Pod",
    "metadata": {
        "name": "rbd2"
    },
    "spec": {
        "containers": [
            {
                "name": "rbd-rw",
                "image": "nginx:1.7.9",
                "volumeMounts": [
                    {
                        "mountPath": "/mnt/rbd",
                        "name": "rbdpd"
                    }
                ]
            }
        ],
        "volumes": [
            {
                "name": "rbdpd",
                "rbd": {
                    "monitors": [
                                "10.64.3.9:6789"
                     ],
                    "pool": "rbd",
                    "image": "foo",
                    "user": "admin",
                    "secretRef": {
                          "name": "ceph-secret"
                     },
                    "fsType": "ext4",
                    "readOnly": true
                }
            }
        ]
    }
}

[root@tjwq01-sys-bs003007 ~]# kubectl create -f rbd-with-secret.json
pod "rbd2" created
[root@tjwq01-sys-bs003007 ~]# kubectl get pods
NAME      READY     STATUS              RESTARTS   AGE
rbd2      0/1       ContainerCreating   0          2s
[root@tjwq01-sys-bs003007 ~]# kubectl describe pods rbd2
Name:           rbd2
Namespace:      default
Node:           127.0.0.1/127.0.0.1
Start Time:     Sun, 12 Mar 2017 15:05:58 +0800
Labels:         <none>
Status:         Running
IP:             172.30.19.2
Controllers:    <none>
Containers:
  rbd-rw:
    Container ID:       docker://6f0c86aadd0ac9065f432b186e8d605535fc03e4e196607b2e14f8dda3dbb759
    Image:              nginx:1.7.9
    Image ID:           docker://sha256:84581e99d807a703c9c03bd1a31cd9621815155ac72a7365fd02311264512656
    Port:
    State:              Running
      Started:          Sun, 12 Mar 2017 15:06:00 +0800
    Ready:              True
    Restart Count:      0
    Volume Mounts:
      /mnt/rbd from rbdpd (rw)
      /var/run/secrets/kubernetes.io/serviceaccount from default-token-icgdo (ro)
    Environment Variables:      <none>
Conditions:
  Type          Status
  Initialized   True
  Ready         True
  PodScheduled  True
Volumes:
  rbdpd:
    Type:               RBD (a Rados Block Device mount on the host that shares a pod's lifetime)
    CephMonitors:       [10.64.3.9:6789]
    RBDImage:           foo
    FSType:             ext4
    RBDPool:            rbd
    RadosUser:          admin
    Keyring:            /etc/ceph/keyring
    SecretRef:          &{ceph-secret}
    ReadOnly:           true
  default-token-icgdo:
    Type:       Secret (a volume populated by a Secret)
    SecretName: default-token-icgdo
QoS Class:      BestEffort
Tolerations:    <none>
Events:
  FirstSeen     LastSeen        Count   From                    SubObjectPath           Type            Reason                  Message
  ---------     --------        -----   ----                    -------------           --------        ------                  -------
  26s           26s             1       {default-scheduler }                            Normal          Scheduled               Successfully assigned rbd2 to
127.0.0.1
  24s           24s             2       {kubelet 127.0.0.1}                             Warning         MissingClusterDNS       kubelet does not have ClusterD
NS IP configured and cannot create Pod using "ClusterFirst" policy. Falling back to DNSDefault policy.
  24s           24s             1       {kubelet 127.0.0.1}     spec.containers{rbd-rw} Normal          Pulled                  Container image "nginx:1.7.9"
already present on machine
  24s           24s             1       {kubelet 127.0.0.1}     spec.containers{rbd-rw} Normal          Created                 Created container with docker
id 6f0c86aadd0a
  24s           24s             1       {kubelet 127.0.0.1}     spec.containers{rbd-rw} Normal          Started                 Started container with docker
id 6f0c86aadd0a

[root@tjwq01-sys-bs003007 ~]# docker ps
CONTAINER ID        IMAGE                                                        COMMAND                  CREATED             STATUS              PORTS
        NAMES
6f0c86aadd0a        nginx:1.7.9                                                  "nginx -g 'daemon off"   48 seconds ago      Up 47 seconds
        k8s_rbd-rw.b46d3044_rbd2_default_57f4e073-06f2-11e7-8472-8cdcd4b3be48_c2ba02b6
082d220ae15d        registry.access.redhat.com/rhel7/pod-infrastructure:latest   "/pod"                   48 seconds ago      Up 48 seconds
        k8s_POD.ae8ee9ac_rbd2_default_57f4e073-06f2-11e7-8472-8cdcd4b3be48_a17502b9
[root@tjwq01-sys-bs003007 ~]# docker exec -i -t 6f0c86aadd0a bash
root@rbd2:/# mount|grep rbd
/dev/rbd0 on /mnt/rbd type ext4 (ro,relatime,stripe=1024,data=ordered)
root@rbd2:/# echo 'hello, rbd!' >/mnt/rbd/hello
bash: /mnt/rbd/hello: Read-only file system

[root@tjwq01-sys-bs003007 ~]# docker inspect 6f0c86aadd0a
        "Mounts": [
            {
                "Source": "/var/lib/kubelet/pods/57f4e073-06f2-11e7-8472-8cdcd4b3be48/containers/rbd-rw/c2ba02b6",
                "Destination": "/dev/termination-log",
                "Mode": "",
                "RW": true,
                "Propagation": "rslave"
            },
            {
                "Name": "3841d5179dc2c569a7fdb7e46b29575f439a1ef818c0154839d5418048997f03",
                "Source": "/var/lib/docker/volumes/3841d5179dc2c569a7fdb7e46b29575f439a1ef818c0154839d5418048997f03/_data",
                "Destination": "/var/cache/nginx",
                "Driver": "local",
                "Mode": "",
                "RW": true,
                "Propagation": ""
            },
            {
                "Source": "/var/lib/kubelet/pods/57f4e073-06f2-11e7-8472-8cdcd4b3be48/volumes/kubernetes.io~rbd/rbdpd",
                "Destination": "/mnt/rbd",
                "Mode": "",
                "RW": true,  // 只读挂载
                "Propagation": "rslave"
            },
            {
                "Source": "/var/lib/kubelet/pods/57f4e073-06f2-11e7-8472-8cdcd4b3be48/volumes/kubernetes.io~secret/default-token-icgdo",
                "Destination": "/var/run/secrets/kubernetes.io/serviceaccount",
                "Mode": "ro",
                "RW": false,
                "Propagation": "rslave"
            },
            {
                "Source": "/var/lib/kubelet/pods/57f4e073-06f2-11e7-8472-8cdcd4b3be48/etc-hosts",
                "Destination": "/etc/hosts",
                "Mode": "",
                "RW": true,
                "Propagation": "rslave"
            }
        ],

[root@tjwq01-sys-bs003007 ~]# kubectl delete pods rbd2 # 删除旧的pod
pod "rbd2" deleted
[root@tjwq01-sys-bs003007 ~]# kubectl apply -f  rbd-with-secret.json  # 新建新的pod
pod "rbd2" created
[root@tjwq01-sys-bs003007 ~]# kubectl exec rbd2 -i -t bash
root@rbd2:/# mount|grep rbd
/dev/rbd0 on /mnt/rbd type ext4 (rw,relatime,stripe=1024,data=ordered)
root@rbd2:/# echo 'hello, ceph!' >/mnt/rbd/hello
root@rbd2:/#


# 二、创建PV和PVC，pod引用PVC；
## 创建 RBD image
[zhangjun3@tjwq01-sys-power003008 ceph-cluster]$ ssh bjzjm-sys-power005010.bjzjm01.ksyun.com
Last login: Sun Mar 12 13:43:40 from 10.160.109.152
[zhangjun3@bjzjm-sys-power005010 ~]$ rbd create ceph-image -s 128
[zhangjun3@bjzjm-sys-power005010 ~]$ rbd info rbd/ceph-image
rbd image 'ceph-image':
        size 128 MB in 32 objects
        order 22 (4096 kB objects)
        block_name_prefix: rbd_data.10542ae8944a
        format: 2
        features: layering, exclusive-lock, object-map, fast-diff, deep-flatten
        flags:

注意：要disable centos kernel不支持的feature，否则后续创建pod挂载该image时会出错；
一劳永逸的方法是在各个cluster node的/etc/ceph/ceph.conf中加上这样一行配置：
	rbd_default_features = 1 #仅是layering对应的bit码所对应的整数值

[zhangjun3@bjzjm-sys-power005010 ~]$ rbd feature disable ceph-image exclusive-lock object-map fast-diff deep-flatten
https://arpnetworks.com/blog/2016/08/26/fixing-ceph-rbd-map-failed-6-no-such-device-or-address.html

## 创建PV
[root@tjwq01-sys-bs003007 ~]# cat ceph-pv.yaml
apiVersion: v1
kind: PersistentVolume
metadata:
  name: ceph-pv
spec:
  capacity:
    storage: 1Gi
  accessModes:
    - ReadWriteOnce
  rbd:
    monitors:
      - 10.64.3.9:6789
    pool: rbd
    image: ceph-image
    user: admin
    secretRef:
      name: ceph-secret
    fsType: ext4
    readOnly: false
  persistentVolumeReclaimPolicy: Recycle
[root@tjwq01-sys-bs003007 ~]# kubectl create -f ceph-pv.yaml
persistentvolume "ceph-pv" created
[root@tjwq01-sys-bs003007 ~]# kubectl get pv
NAME      CAPACITY   ACCESSMODES   RECLAIMPOLICY   STATUS      CLAIM     REASON    AGE
ceph-pv   1Gi        RWO           Recycle         Available                       6s
[root@tjwq01-sys-bs003007 ~]# kubectl describe pv ceph-pv
Name:           ceph-pv
Labels:         <none>
StorageClass:
Status:         Available
Claim:
Reclaim Policy: Recycle
Access Modes:   RWO
Capacity:       1Gi
Message:
Source:
    Type:               RBD (a Rados Block Device mount on the host that shares a pod's lifetime)
    CephMonitors:       [10.64.3.9:6789]
    RBDImage:           ceph-image
    FSType:             ext4
    RBDPool:            rbd
    RadosUser:          admin
    Keyring:            /etc/ceph/keyring
    SecretRef:          &{ceph-secret}
    ReadOnly:           false
No events.

## 创建PVC
[root@tjwq01-sys-bs003007 ~]# cat ceph-pvc.yaml
kind: PersistentVolumeClaim
apiVersion: v1
metadata:
  name: ceph-claim
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 1Gi
      
[root@tjwq01-sys-bs003007 ~]# kubectl create -f ceph-pvc.yaml
persistentvolumeclaim "ceph-claim" created
[root@tjwq01-sys-bs003007 ~]# kubectl get pvc
NAME         STATUS    VOLUME    CAPACITY   ACCESSMODES   AGE
ceph-claim   Bound     ceph-pv   1Gi        RWO           6s

## 创建使用PVC的pod
[root@tjwq01-sys-bs003007 ~]# cat ceph-pod1.yaml
apiVersion: v1
kind: Pod
metadata:
  name: ceph-pod1
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
      claimName: ceph-claim

[root@tjwq01-sys-bs003007 ~]# kubectl create -f ceph-pod1.yaml
pod "ceph-pod1" created
[root@tjwq01-sys-bs003007 ~]# kubectl get pod
NAME        READY     STATUS              RESTARTS   AGE
ceph-pod1   0/1       ContainerCreating   0          5s
rbd2        1/1       Running             0          16h

如果ceph-pod1一直处于ContainerCreating状态，可以查看kubelet的日志：
[root@tjwq01-sys-bs003007 ~]# journalctl -u kubelet -o export|grep MESSAGE
MESSAGE=E0313 12:46:25.811750    3025 nestedpendingoperations.go:233] Operation for "\"kubernetes.io/rbd/6916a8b6-07a7-11e7-8472-8cdcd4b3be48-ceph-vol1\" (\"6
916a8b6-07a7-11e7-8472-8cdcd4b3be48\")" failed. No retries permitted until 2017-03-13 12:48:25.811707944 +0800 CST (durationBeforeRetry 2m0s). Error: MountVol
ume.SetUp failed for volume "kubernetes.io/rbd/6916a8b6-07a7-11e7-8472-8cdcd4b3be48-ceph-vol1" (spec.Name: "ceph-pv") pod "6916a8b6-07a7-11e7-8472-8cdcd4b3be4
8" (UID: "6916a8b6-07a7-11e7-8472-8cdcd4b3be48") with: rbd: map failed exit status 6 2017-03-13 12:46:25.756304 7fafcda547c0 -1 did not load config file, usin
g default settings.
MESSAGE=rbd: sysfs write failed
MESSAGE=rbd: map failed: (6) No such device or address

这是由于使用的RBD启用了centos kernel不支持的feature所致，解决方法是disable相应的feature(参考上文)；
https://arpnetworks.com/blog/2016/08/26/fixing-ceph-rbd-map-failed-6-no-such-device-or-address.html


Mount过程中kubelet的日志，k8s通过fsck发现这个image是一个空image，没有fs在里面，于是默认采用ext4为其格式化，成功后，再行挂载：
MESSAGE=I0313 12:50:27.097306    3025 reconciler.go:254] MountVolume operation started for volume "kubernetes.io/rbd/6916a8b6-07a7-11e7-8472-8cdcd4b3be48-ceph
-vol1" (spec.Name: "ceph-pv") to pod "6916a8b6-07a7-11e7-8472-8cdcd4b3be48" (UID: "6916a8b6-07a7-11e7-8472-8cdcd4b3be48").
MESSAGE=I0313 12:50:28.808730    3025 mount_linux.go:267] `fsck` error fsck from util-linux 2.23.2
MESSAGE=fsck.ext2: Bad magic number in super-block while trying to open /dev/rbd1
MESSAGE=/dev/rbd1:
MESSAGE=The superblock could not be read or does not describe a correct ext2
MESSAGE=filesystem.  If the device is valid and it really contains an ext2
MESSAGE=filesystem (and not swap or ufs or something else), then the superblock
MESSAGE=is corrupt, and you might try running e2fsck with an alternate superblock:
MESSAGE=e2fsck -b 8193 <device>
MESSAGE=E0313 12:50:28.814556    3025 mount_linux.go:110] Mount failed: exit status 32
MESSAGE=Mounting arguments: /dev/rbd1 /var/lib/kubelet/plugins/kubernetes.io/rbd/rbd/rbd-image-ceph-image ext4 [defaults]
MESSAGE=Output: mount: wrong fs type, bad option, bad superblock on /dev/rbd1,
MESSAGE=missing codepage or helper program, or other error
MESSAGE=In some cases useful info is found in syslog - try
MESSAGE=dmesg | tail or so.
MESSAGE=I0313 12:50:28.829946    3025 mount_linux.go:287] Disk "/dev/rbd1" appears to be unformatted, attempting to format as type: "ext4" with options: [-E l
azy_itable_init=0,lazy_journal_init=0 -F /dev/rbd1]
MESSAGE=I0313 12:50:28.983704    3025 mount_linux.go:292] Disk successfully formatted (mkfs): ext4 - /dev/rbd1 /var/lib/kubelet/plugins/kubernetes.io/rbd/rbd/
rbd-image-ceph-image

[root@tjwq01-sys-bs003007 ~]# kubectl describe pod ceph-pod1                                                                                          [8/1881]
Name:           ceph-pod1
Namespace:      default
Node:           127.0.0.1/127.0.0.1
Start Time:     Mon, 13 Mar 2017 12:53:28 +0800
Labels:         <none>
Status:         Running
IP:             172.30.19.3
Controllers:    <none>
Containers:
  ceph-busybox1:
    Container ID:       docker://b82c01fb9d88a20118f3ef5f95192f4ec57f353098ee4b6d8c2d3317fe7d3613
    Image:              nginx:1.7.9
    Image ID:           docker://sha256:84581e99d807a703c9c03bd1a31cd9621815155ac72a7365fd02311264512656
    Port:
    State:              Running
      Started:          Mon, 13 Mar 2017 12:53:30 +0800
    Ready:              True
    Restart Count:      0
    Volume Mounts:
      /usr/share/busybox from ceph-vol1 (rw)
      /var/run/secrets/kubernetes.io/serviceaccount from default-token-icgdo (ro)
    Environment Variables:      <none>
Conditions:
  Type          Status
  Initialized   True
  Ready         True
  PodScheduled  True
Volumes:
  ceph-vol1:
    Type:       PersistentVolumeClaim (a reference to a PersistentVolumeClaim in the same namespace)
    ClaimName:  ceph-claim
    ReadOnly:   false
  default-token-icgdo:
    Type:       Secret (a volume populated by a Secret)
    SecretName: default-token-icgdo
QoS Class:      BestEffort
Tolerations:    <none>
Events:
  FirstSeen     LastSeen        Count   From                    SubObjectPath                   Type            Reason                  Message
  ---------     --------        -----   ----                    -------------                   --------        ------                  -------
  4m            4m              1       {default-scheduler }                                    Normal          Scheduled               Successfully assigned ceph-pod1 to 127.0.0.1
  4m            4m              2       {kubelet 127.0.0.1}                                     Warning         MissingClusterDNS       kubelet does not have ClusterDNS IP configured and cannot create Pod using "ClusterFirst" policy. Falling back to DNSDefault policy.
  4m            4m              1       {kubelet 127.0.0.1}     spec.containers{ceph-busybox1}  Normal          Pulled                  Container image "nginx:1.7.9" already present on machine
  4m            4m              1       {kubelet 127.0.0.1}     spec.containers{ceph-busybox1}  Normal          Created                 Created container with docker id b82c01fb9d88
  4m            4m              1       {kubelet 127.0.0.1}     spec.containers{ceph-busybox1}  Normal          Started                 Started container with docker id b82c01fb9d88

## 向PV里面写些数据
[root@tjwq01-sys-bs003007 ~]# kubectl exec ceph-pod1 -i -t bash
root@ceph-pod1:/# df -h|grep rbd
/dev/rbd1                                                                                         120M  1.6M  110M   2% /usr/share/busybox
root@ceph-pod1:/# echo 'hello, k8s pv&pvc' > /usr/share/busybox/hello

## 删除使用PV的Pod
[root@tjwq01-sys-bs003007 ~]# kubectl delete pod ceph-pod1
pod "ceph-pod1" deleted
[root@tjwq01-sys-bs003007 ~]# kubectl get pod
NAME      READY     STATUS    RESTARTS   AGE
rbd2      1/1       Running   0          17h
[root@tjwq01-sys-bs003007 ~]# kubectl get pv,pvc
NAME         CAPACITY   ACCESSMODES   RECLAIMPOLICY   STATUS    CLAIM                REASON    AGE
pv/ceph-pv   1Gi        RWO           Recycle         Bound     default/ceph-claim             1h

NAME             STATUS    VOLUME    CAPACITY   ACCESSMODES   AGE
pvc/ceph-claim   Bound     ceph-pv   1Gi        RWO           1h

pod的删除并没有影响到pv和pvc object，它们依旧存在；

## 新建一个pod
[root@tjwq01-sys-bs003007 ~]# diff ceph-pod1.yaml ceph-pod2.yaml
4c4
<   name: ceph-pod1
---
>   name: ceph-pod2
7c7
<   - name: ceph-busybox1
---
>   - name: ceph-busybox2
10c10
<     - name: ceph-vol1
---
>     - name: ceph-vol2
14c14
<   - name: ceph-vol1
---
>   - name: ceph-vol2
[root@tjwq01-sys-bs003007 ~]# mount|grep rbd
/dev/rbd0 on /var/lib/kubelet/plugins/kubernetes.io/rbd/rbd/rbd-image-foo type ext4 (rw,relatime,stripe=1024,data=ordered)
/dev/rbd0 on /var/lib/kubelet/pods/f33658f3-071c-11e7-8472-8cdcd4b3be48/volumes/kubernetes.io~rbd/rbdpd type ext4 (rw,relatime,stripe=1024,data=ordered)
[root@tjwq01-sys-bs003007 ~]# kubectl create -f ceph-pod2.yaml
pod "ceph-pod2" created
[root@tjwq01-sys-bs003007 ~]# kubectl get pod
NAME        READY     STATUS              RESTARTS   AGE
ceph-pod2   0/1       ContainerCreating   0          2s
rbd2        1/1       Running             0          17h
[root@tjwq01-sys-bs003007 ~]# kubectl get pod
NAME        READY     STATUS    RESTARTS   AGE
ceph-pod2   1/1       Running   0          3s
rbd2        1/1       Running   0          17h

## 数据完好无损的被ceph-pod2读取到了
[root@tjwq01-sys-bs003007 ~]# kubectl exec ceph-pod2 -i -t bash
root@ceph-pod2:/# cat /usr/share/busybox/hello
hello, k8s pv&pvc

参考：
http://tonybai.com/2016/11/07/integrate-kubernetes-with-ceph-rbd/


# 三、创建RBD StorageClass，pod使用PVC Template动态创建PV；

## 创建ceph rbd StorageClass
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

## 创建PersistentVolumeClaim

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