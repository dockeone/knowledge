# User accounts vs service accounts

This is a Cluster Administrator guide to service accounts. It assumes knowledge of the User Guide to Service Accounts.
Support for authorization and user accounts is planned but incomplete. Sometimes incomplete features are referred to in order to better describe service accounts.
User accounts vs service accounts

Kubernetes distinguished between the concept of a user account and a service accounts for a number of reasons:

1. User accounts are for humans. Service accounts are for processes, which run in pods.
2. User accounts are intended to be global. Names must be unique across all namespaces of a cluster, future user resource will not be namespaced. Service accounts are namespaced.
3. Typically, a cluster’s User accounts might be synced from a corporate database, where new user account creation requires special privileges and is tied to complex business processes. Service account creation is intended to be more lightweight, allowing cluster users to create service accounts for specific tasks (i.e. principle of least privilege).
4. Auditing considerations for humans and service accounts may differ.
5. A config bundle for a complex system may include definition of various service accounts for components of that system. Because service accounts can be created ad-hoc and have namespaced names, such config is portable.

# Service account automation

Three separate components cooperate to implement the automation around service accounts:

1. A Service account admission controller
2. A Token controller
3. A Service account controller

# Service Account Admission Controller(apiserver部分)
The modification of pods is implemented via a plugin called an Admission Controller. It is part of the apiserver. It acts synchronously to modify pods as they 
are created or updated. When this plugin is active (and it is by default on most distributions), then it does the following when a pod is created or modified:

1. If the pod does not have a ServiceAccount set, it sets the ServiceAccount to default.
2. It ensures that the ServiceAccount referenced by the pod exists, and otherwise rejects it.
4. If the pod does not contain any ImagePullSecrets, then ImagePullSecrets of the ServiceAccount are added to the pod.
5. It adds a volume to the pod which contains a token for API access.
6. It adds a volumeSource to each container of the pod mounted at /var/run/secrets/kubernetes.io/serviceaccount.

# Token Controller(controller-manager部分)
TokenController runs as part of controller-manager. It acts asynchronously. It:

1. observes serviceAccount creation and creates a corresponding Secret to allow API access.
2. observes serviceAccount deletion and deletes all corresponding ServiceAccountToken Secrets
3. observes secret addition, and ensures the referenced ServiceAccount exists, and adds a token to the secret if needed
4. observes secret deletion and removes a reference from the corresponding ServiceAccount if needed

# To create additional API tokens
A controller loop ensures a secret with an API token exists for each service account. To create additional API tokens for a service account, create a secret of type 
ServiceAccountToken with an annotation referencing the service account, and the controller will update it with a generated token:
secret.json:
{
    "kind": "Secret",
    "apiVersion": "v1",
    "metadata": {
        "name": "mysecretname",
        "annotations": {
            "kubernetes.io/service-account.name": "myserviceaccount"
        }
    },
    "type": "kubernetes.io/service-account-token"
}

kubectl create -f ./secret.json
kubectl describe secret mysecretname

# To delete/invalidate a service account token
kubectl delete secret mysecretname

#Service Account Controller
Service Account Controller manages ServiceAccount inside namespaces, and ensures a ServiceAccount named “default” exists in every active namespace.

[root@tjwq01-sys-bs003007 ~]# kubectl describe pod nginx
Name:           nginx
Namespace:      default
Node:           127.0.0.1/127.0.0.1
Start Time:     Fri, 16 Dec 2016 20:57:35 +0800
Labels:         <none>
Status:         Running
IP:             172.30.35.2
Controllers:    <none>
Containers:
  nginx:
    Container ID:       docker://3cfe53a58cff9c64b88ac7500f29040ef175d69b4a7c606004ee284e0001ad76
    Image:              nginx:1.7.9
    Image ID:           docker://sha256:84581e99d807a703c9c03bd1a31cd9621815155ac72a7365fd02311264512656
    Port:               80/TCP
    QoS Tier:
      cpu:              BestEffort
      memory:           BestEffort
    State:              Running
      Started:          Fri, 16 Dec 2016 20:57:36 +0800
    Ready:              True
    Restart Count:      0
    Environment Variables:
Conditions:
  Type          Status
  Ready         True
Volumes:
  default-token-icgdo:
    Type:       Secret (a volume populated by a Secret)
    SecretName: default-token-icgdo
Events:
  FirstSeen     LastSeen        Count   From                    SubobjectPath           Type            Reason                  Message
  ---------     --------        -----   ----                    -------------           --------        ------                  -------
  2m            2m              1       {default-scheduler }                            Normal          Scheduled               Successfully assigned nginx to
 127.0.0.1
  2m            2m              2       {kubelet 127.0.0.1}                             Warning         MissingClusterDNS       kubelet does not have ClusterD
NS IP configured and cannot create Pod using "ClusterFirst" policy. Falling back to DNSDefault policy.
  2m            2m              1       {kubelet 127.0.0.1}     spec.containers{nginx}  Normal          Pulled                  Container image "nginx:1.7.9"
already present on machine
  2m            2m              1       {kubelet 127.0.0.1}     spec.containers{nginx}  Normal          Created                 Created container with docker
id 3cfe53a58cff
  2m            2m              1       {kubelet 127.0.0.1}     spec.containers{nginx}  Normal          Started                 Started container with docker
id 3cfe53a58cff

如果启用了service account，在创建pod时，controller会为pod自动创建sercret，并在容器的/var/run/secrets/kubernetes.io/serviceaccount位置挂载sercret;

[root@tjwq01-sys-bs003007 ~]# kubectl get serviceaccount
NAME      SECRETS   AGE
default   1         29m

[root@tjwq01-sys-bs003007 ~]# kubectl describe serviceaccount/default
Name:           default
Namespace:      default
Labels:         <none>

Image pull secrets:     <none>

Mountable secrets:      default-token-icgdo

Tokens:                 default-token-icgdo

[root@tjwq01-sys-bs003007 ~]# kubectl describe secret/default-token-icgdo
Name:           default-token-icgdo
Namespace:      default
Labels:         <none>
Annotations:    kubernetes.io/service-account.name=default,kubernetes.io/service-account.uid=59832726-c38c-11e6-9237-8cdcd4b3be48

Type:   kubernetes.io/service-account-token

Data
====
namespace:      7 bytes
token:          eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJrdWJlcm5ldGVzL3NlcnZpY2VhY2NvdW50Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9uYW1lc3BhY2UiOiJkZWZhdWx0Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9zZWNyZXQubmFtZSI6ImRlZmF1bHQtdG9rZW4taWNnZG8iLCJrdWJlcm5ldGVzLmlvL3NlcnZpY2VhY2NvdW50L3NlcnZpY2UtYWNjb3VudC5uYW1lIjoiZGVmYXVsdCIsImt1YmVybmV0ZXMuaW8vc2VydmljZWFjY291bnQvc2VydmljZS1hY2NvdW50LnVpZCI6IjU5ODMyNzI2LWMzOGMtMTFlNi05MjM3LThjZGNkNGIzYmU0OCIsInN1YiI6InN5c3RlbTpzZXJ2aWNlYWNjb3VudDpkZWZhdWx0OmRlZmF1bHQifQ.yTPBAOTNWwP2eQ2evV5R6gcCKbPk6zX4eURVzk5Z7-TZAT6ZCZqQY0x_-va6sbahnsLEon7DofFOpKdhVWiDB90P8DU_9nMdr1vprtmG7tCRmi38wZi1rtUeXuqU19CAdBDP5NMPhXFMQR3JigAh-KKfCELTMtlsm4lxevKwmiBlPMTdHQtvHlkb7i9so7D_YGy0TA3uaKoRk30oD2b46oiIdurXrf9KidtaoF8RDp9vrtYZU71vYK3vCpDD8lpwyfNL_LLq0_aSjK3mWOvf9HxQ3oF5ryHq9D6m7pDlScePGWeXuKab_Yak_vJJSQM-j6zEpWck17UO9OK1aKIs2Q

[root@tjwq01-sys-bs003007 ~]# docker exec -it 3cfe53a58cff bash
root@nginx:~# mount|grep serviceaccount
tmpfs on /run/secrets/kubernetes.io/serviceaccount type tmpfs (ro,relatime)

root@nginx:~# ls -l /run/secrets/kubernetes.io/serviceaccount
total 8
-r--r--r-- 1 root root   7 Dec 16 12:57 namespace
-r--r--r-- 1 root root 846 Dec 16 12:57 token

注意：如果为apiserver指定了ca.crt根证书，则上面的secret/default-token-icgdo包含另一个data: ca.crt，并mount到pod容器的/var/run/secrets/kubernetes.io/serviceaccount 位置；


查看apiserver的cluster ip
[root@tjwq01-sys-bs003007 ~]# kubectl get services
NAME         CLUSTER-IP   EXTERNAL-IP   PORT(S)   AGE
kubernetes   10.254.0.1   <none>        443/TCP   37m
[root@tjwq01-sys-bs003007 ~]# kubectl describe svc/kubernetes
Name:                   kubernetes
Namespace:              default
Labels:                 component=apiserver,provider=kubernetes
Selector:               <none>
Type:                   ClusterIP
IP:                     10.254.0.1
Port:                   https   443/TCP
Endpoints:              120.92.8.114:6443   
Session Affinity:       None
No events.

注意：上面endpoint的IP 120.92.8.114实际上是错误的，因为apiserver --bind-address的地址是127.0.0.1，apiserver也是在这个端口上监听6443;
这时如果telnet apiserver的cluster ip和端口即 10.254.0.1 443，会提示refuse；

解决方法：修改apiserver --bind-address的地址为非127.0.0.1，重启apiserver；

[root@tjwq01-sys-bs003007 ~]# kubectl get endpoints
NAME         ENDPOINTS                       AGE
kubernetes   120.92.8.114:6443               1h
my-nginx     172.30.35.3:80,172.30.35.4:80   8m

[root@tjwq01-sys-bs003007 ~]# kubectl describe endpoints kubernetes
Name:           kubernetes
Namespace:      default
Labels:         <none>
Subsets:
  Addresses:            120.92.8.114
  NotReadyAddresses:    <none>
  Ports:
    Name        Port    Protocol
    ----        ----    --------
    https       6443    TCP

No events.

[root@tjwq01-sys-bs003007 ~]# kubectl describe endpoints kubernetes
apiVersion: v1
kind: Endpoints
metadata:
  creationTimestamp: 2016-12-16T12:34:54Z
  name: kubernetes
  namespace: default
  resourceVersion: "7"
  selfLink: /api/v1/namespaces/default/endpoints/kubernetes
  uid: 0c34e6c8-c38c-11e6-a508-8cdcd4b3be48
subsets:
- addresses:
  - ip: 120.92.8.114
  ports:
  - name: https
    port: 6443
    protocol: TCP

如果将上面的 ip 120.92.8.114修改为127.0.0.1, 则保存失败，提示：
endpoints "kubernetes" was not valid:
 * subsets[0].addresses[0].ip: Invalid value: "127.0.0.1": may not be in the loopback range (127.0.0.0/8)
 * subsets[0].addresses[0].ip: Invalid value: "127.0.0.1": may not be in the loopback range (127.0.0.0/8)

解决办法：
将apiserver的--bind-address=设置为非127.0.0.1的本机IP如 10.64.3.7，重启apiserver即可(endpoints会自动改变)；


在pod内外向apiserver cluster ip发起请求测试：
1. 先获取service account token:
[root@tjwq01-sys-bs003007 ~]# kubectl run busybox --image=busybox --restart=Never --tty -i --generator=run-pod/v1
Waiting for pod default/busybox to be running, status is Pending, pod ready: false


Hit enter for command prompt

/ #
/ # TOKEN="$(cat /var/run/secrets/kubernetes.io/serviceaccount/token)"
/ # echo \'$TOKEN\'
'eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJrdWJlcm5ldGVzL3NlcnZpY2VhY2NvdW50Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9uYW1lc3BhY2UiOiJkZWZhdWx0Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9zZWNyZXQubmFtZSI6ImRlZmF1bHQtdG9rZW4taWNnZG8iLCJrdWJlcm5ldGVzLmlvL3NlcnZpY2VhY2NvdW50L3NlcnZpY2UtYWNjb3VudC5uYW1lIjoiZGVmYXVsdCIsImt1YmVybmV0ZXMuaW8vc2VydmljZWFjY291bnQvc2VydmljZS1hY2NvdW50LnVpZCI6IjU5ODMyNzI2LWMzOGMtMTFlNi05MjM3LThjZGNkNGIzYmU0OCIsInN1YiI6InN5c3RlbTpzZXJ2aWNlYWNjb3VudDpkZWZhdWx0OmRlZmF1bHQifQ.yTPBAOTNWwP2eQ2evV5R6gcCKbPk6zX4eURVzk5Z7-TZAT6ZCZqQY0x_-va6sbahnsLEon7DofFOpKdhVWiDB90P8DU_9nMdr1vprtmG7tCRmi38wZi1rtUeXuqU19CAdBDP5NMPhXFMQR3JigAh-KKfCELTMtlsm4lxevKwmiBlPMTdHQtvHlkb7i9so7D_YGy0TA3uaKoRk30oD2b46oiIdurXrf9KidtaoF8RDp9vrtYZU71vYK3vCpDD8lpwyfNL_LLq0_aSjK3mWOvf9HxQ3oF5ryHq9D6m7pDlScePGWeXuKab_Yak_vJJSQM-j6zEpWck17UO9OK1aKIs2Q'
/ # 

2. 在pod内外向apiserver cluster ip发起请求测试：
[root@tjwq01-sys-bs003007 ~]# TOKEN="eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJrdWJlcm5ldGVzL3NlcnZpY2VhY2NvdW50Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9uYW1lc3BhY2UiOiJkZWZhdWx0Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9zZWNyZXQubmFtZSI6ImRlZmF1bHQtdG9rZW4taWNnZG8iLCJrdWJlcm5ldGVzLmlvL3NlcnZpY2VhY2NvdW50L3NlcnZpY2UtYWNjb3VudC5uYW1lIjoiZGVmYXVsdCIsImt1YmVybmV0ZXMuaW8vc2VydmljZWFjY291bnQvc2VydmljZS1hY2NvdW50LnVpZCI6IjU5ODMyNzI2LWMzOGMtMTFlNi05MjM3LThjZGNkNGIzYmU0OCIsInN1YiI6InN5c3RlbTpzZXJ2aWNlYWNjb3VudDpkZWZhdWx0OmRlZmF1bHQifQ.yTPBAOTNWwP2eQ2evV5R6gcCKbPk6zX4eURVzk5Z7-TZAT6ZCZqQY0x_-va6sbahnsLEon7DofFOpKdhVWiDB90P8DU_9nMdr1vprtmG7tCRmi38wZi1rtUeXuqU19CAdBDP5NMPhXFMQR3JigAh-KKfCELTMtlsm4lxevKwmiBlPMTdHQtvHlkb7i9so7D_YGy0TA3uaKoRk30oD2b46oiIdurXrf9KidtaoF8RDp9vrtYZU71vYK3vCpDD8lpwyfNL_LLq0_aSjK3mWOvf9HxQ3oF5ryHq9D6m7pDlScePGWeXuKab_Yak_vJJSQM-j6zEpWck17UO9OK1aKIs2Q"
[root@tjwq01-sys-bs003007 ~]#
[root@tjwq01-sys-bs003007 ~]#
[root@tjwq01-sys-bs003007 ~]# curl -k   https://10.254.0.1:443/version -H "Authorization: Bearer $TOKEN"
{
  "major": "1",
  "minor": "2",
  "gitVersion": "v1.2.0",
  "gitCommit": "ec7364b6e3b155e78086018aa644057edbe196e5",
  "gitTreeState": "clean"
}

3. 如果apiserver指定了ca证书，则自动挂载的目录下会有ca.crt文件，pod的容器可以使用该证书对apiserver进行验证:
[root@tjwq01-sys-bs003007 ~]# curl --cacert /var/run/secrets/kubernetes.io/serviceaccount/ca.crt  https://10.254.0.1:443/version -H "Authorization: Bearer $TOKEN"