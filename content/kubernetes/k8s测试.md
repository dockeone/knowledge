[root@tjwq01-sys-bs003007 ~]# kubectl create -f pod-nginx.yaml
pod "nginx" created
[root@tjwq01-sys-bs003007 ~]# kubectl get pods
NAME      READY     STATUS    RESTARTS   AGE
nginx     1/1       Running   0          7s
[root@tjwq01-sys-bs003007 ~]# docker ps -a
CONTAINER ID        IMAGE                                                        COMMAND                  CREATED             STATUS              PORTS               NAMES
2c79df4182b7        nginx:1.7.9                                                  "nginx -g 'daemon off"   17 seconds ago      Up 15 seconds                           k8s_nginx.7413299_nginx_default_7b5cc851-06e5-11e7-8472-8cdcd4b3be48_b01a814e
e52a54bc7693        registry.access.redhat.com/rhel7/pod-infrastructure:latest   "/pod"                   18 seconds ago      Up 16 seconds                           k8s_POD.a8590b41_nginx_default_7b5cc851-06e5-11e7-8472-8cdcd4b3be48_bde886e8
0c4caa5345f7        pstauffer/curl:latest                                        "bash"                   12 weeks ago        Created

查看kube-scheduler的日志，最后一行内容为：
MESSAGE=I1216 13:25:00.830715   10957 event.go:211] Event(api.ObjectReference{Kind:"Pod", Namespace:"default", Name:"nginx", UID:"fdbd3d8b-c34f-11e6-b4dd-8cdcd4b3be48", APIVersion:"v1", ResourceVersion:"281", FieldPath:""}): type: 'Normal' reason: 'Scheduled' Successfully assigned nginx to 127.0.0.1

稍等一会查看pod状态
[root@tjwq01-sys-bs003007 ~]# kubectl describe pod nginx
Name:           nginx                                                                                                                                 [2/1952]
Namespace:      default
Node:           127.0.0.1/127.0.0.1
Start Time:     Sun, 12 Mar 2017 13:33:54 +0800
Labels:         <none>
Status:         Running
IP:             172.30.19.2
Controllers:    <none>
Containers:
  nginx:
    Container ID:       docker://2c79df4182b7c12cb18d10cbf05f1577c78522d6b038390f80765c40e294d94a
    Image:              nginx:1.7.9
    Image ID:           docker://sha256:84581e99d807a703c9c03bd1a31cd9621815155ac72a7365fd02311264512656
    Port:               80/TCP
    State:              Running
      Started:          Sun, 12 Mar 2017 13:33:56 +0800
    Ready:              True
    Restart Count:      0
    Volume Mounts:
      /var/run/secrets/kubernetes.io/serviceaccount from default-token-icgdo (ro)
    Environment Variables:      <none>
Conditions:
  Type          Status
  Initialized   True
  Ready         True
  PodScheduled  True
Volumes:
  default-token-icgdo:
    Type:       Secret (a volume populated by a Secret)
    SecretName: default-token-icgdo
QoS Class:      BestEffort
Tolerations:    <none>
Events:
  FirstSeen     LastSeen        Count   From                    SubObjectPath           Type            Reason                  Message
  ---------     --------        -----   ----                    -------------           --------        ------                  -------
  1m            1m              1       {default-scheduler }                            Normal          Scheduled               Successfully assigned nginx to
 127.0.0.1
  1m            1m              2       {kubelet 127.0.0.1}                             Warning         MissingClusterDNS       kubelet does not have ClusterD
NS IP configured and cannot create Pod using "ClusterFirst" policy. Falling back to DNSDefault policy.
  1m            1m              1       {kubelet 127.0.0.1}     spec.containers{nginx}  Normal          Pulled                  Container image "nginx:1.7.9"
already present on machine
  1m            1m              1       {kubelet 127.0.0.1}     spec.containers{nginx}  Normal          Created                 Created container with docker
id 2c79df4182b7
  1m            1m              1       {kubelet 127.0.0.1}     spec.containers{nginx}  Normal          Started                 Started container with docker
id 2c79df4182b7

如果启用了service account，在创建pod时，自动在 /var/run/secrets/kubernetes.io/serviceaccount 位置挂载 sercret;

[root@tjwq01-sys-bs003007 ~]# docker ps
CONTAINER ID        IMAGE                                                        COMMAND                  CREATED             STATUS              PORTS               NAMES
90f7e9cd2bc1        nginx:1.7.9                                                  "nginx -g 'daemon off"   2 minutes ago       Up 2 minutes                            k8s_nginx.4580025_nginx_default_fdbd3d8b-c34f-11e6-b4dd-8cdcd4b3be48_ad0b32f1
5a0c9355bfc8        registry.access.redhat.com/rhel7/pod-infrastructure:latest   "/pod"                   6 minutes ago       Up 6 minutes                            k8s_POD.c36b0a77_nginx_default_fdbd3d8b-c34f-11e6-b4dd-8cdcd4b3be48_3ed87315
[root@tjwq01-sys-bs003007 ~]# docker images
REPOSITORY                                            TAG                 IMAGE ID            CREATED             SIZE
registry.access.redhat.com/rhel7/pod-infrastructure   latest              7d5548e9fb99        5 weeks ago         205.3 MB
docker.io/nginx                                       1.7.9               84581e99d807        22 months ago       91.64 MB

[root@tjwq01-sys-bs003007 ~]# docker exec -it 2c79df4182b7 bash
root@nginx:/# ip addr show
1: lo: <LOOPBACK,UP,LOWER_UP> mtu 65536 qdisc noqueue state UNKNOWN qlen 1
    link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00
    inet 127.0.0.1/8 scope host lo
       valid_lft forever preferred_lft forever
    inet6 ::1/128 scope host
       valid_lft forever preferred_lft forever
8: eth0@if9: <BROADCAST,MULTICAST,UP,LOWER_UP,M-DOWN> mtu 1450 qdisc noqueue state UP
    link/ether 02:42:ac:1e:13:02 brd ff:ff:ff:ff:ff:ff
    inet 172.30.19.2/24 scope global eth0
       valid_lft forever preferred_lft forever
    inet6 fe80::42:acff:fe1e:1302/64 scope link tentative dadfailed
       valid_lft forever preferred_lft forever

可见nginx容器的IP地址为flannel的一个/24段地址中的第二个；

root@nginx:/# ls -l /run/secrets/kubernetes.io/serviceaccount/
total 0
lrwxrwxrwx 1 root root 16 Mar 12 05:33 namespace -> ..data/namespace
lrwxrwxrwx 1 root root 12 Mar 12 05:33 token -> ..data/token

起另外一个测试pod:
[root@tjwq01-sys-bs003007 ~]# kubectl run busybox --image=busybox --restart=Never --tty -i  --env "POD_IP=$(kubectl get pod nginx -o go-template='{{.status.podIP}}')"
Waiting for pod default/busybox to be running, status is Pending, pod ready: false
If you don't see a command prompt, try pressing enter.
/ # ifconfig
eth0      Link encap:Ethernet  HWaddr 02:42:AC:1E:13:03
          inet addr:172.30.19.3  Bcast:0.0.0.0  Mask:255.255.255.0
          inet6 addr: fe80::42:acff:fe1e:1303/64 Scope:Link
          UP BROADCAST RUNNING MULTICAST  MTU:1450  Metric:1
          RX packets:10 errors:0 dropped:0 overruns:0 frame:0
          TX packets:3 errors:0 dropped:0 overruns:0 carrier:0
          collisions:0 txqueuelen:0
          RX bytes:816 (816.0 B)  TX bytes:258 (258.0 B)
...
/ # wget -qO- http://$POD_IP  // 可以用nginx pod 的 pod ip访问web服务；

对于docker而言，每起一个pod都会在host上创建一个名称类似veth24a8440的虚拟网卡，该网卡和对端的container eth网卡互联，都接到docker0 bridge上；

[zhangjun3@tjwq01-sys-bs003007 ~]$ ifconfig |grep veth
veth24a8440: flags=4163<UP,BROADCAST,RUNNING,MULTICAST>  mtu 1450
vethb747827: flags=4163<UP,BROADCAST,RUNNING,MULTICAST>  mtu 1450

[root@tjwq01-sys-bs003007 ~]# kubectl create -f my-nginx.yaml
deployment "my-nginx" created
[root@tjwq01-sys-bs003007 ~]# kubectl get deploy
NAME       DESIRED   CURRENT   UP-TO-DATE   AVAILABLE   AGE
my-nginx   2         2         2            2           11s
[root@tjwq01-sys-bs003007 ~]# kubectl get pods
NAME                        READY     STATUS    RESTARTS   AGE
busybox                     1/1       Running   0          6m
my-nginx-3467165555-1jw53   1/1       Running   0          4s
my-nginx-3467165555-m3c25   1/1       Running   0          4s
nginx                       1/1       Running   0          15m
[root@tjwq01-sys-bs003007 ~]# kubectl describe  deploy my-nginx
Name:                   my-nginx
Namespace:              default
CreationTimestamp:      Sun, 12 Mar 2017 13:48:57 +0800
Labels:                 run=my-nginx
Selector:               run=my-nginx
Replicas:               2 updated | 2 total | 2 available | 0 unavailable
StrategyType:           RollingUpdate
MinReadySeconds:        0
RollingUpdateStrategy:  1 max unavailable, 1 max surge
Conditions:
  Type          Status  Reason
  ----          ------  ------
  Available     True    MinimumReplicasAvailable
OldReplicaSets: <none>
NewReplicaSet:  my-nginx-3467165555 (2/2 replicas created)
Events:
  FirstSeen     LastSeen        Count   From                            SubObjectPath   Type            Reason                  Message
  ---------     --------        -----   ----                            -------------   --------        ------                  -------
  2m            2m              1       {deployment-controller }                        Normal          ScalingReplicaSet       Scaled up replica set my-nginx-3467165555 to 2

my-nginx分别分配了两个pod ip：172.30.19.4、172.30.19.5

[root@tjwq01-sys-bs003007 ~]# kubectl expose deploy my-nginx
service "my-nginx" exposed
[root@tjwq01-sys-bs003007 ~]# kubectl get service
NAME         CLUSTER-IP      EXTERNAL-IP   PORT(S)   AGE
kubernetes   10.254.0.1      <none>        443/TCP   85d
my-nginx     10.254.238.37   <none>        80/TCP    8s

可见分配的server IP 为kube-apiserver 参数service-cluster-ip-range=10.254.0.0/16值中的一部分；

[root@tjwq01-sys-bs003007 ~]# kubectl describe svc my-nginx
Name:                   my-nginx
Namespace:              default
Labels:                 run=my-nginx
Selector:               run=my-nginx
Type:                   ClusterIP
IP:                     10.254.238.37
Port:                   <unset> 80/TCP
Endpoints:              172.30.19.4:80,172.30.19.5:80
Session Affinity:       None
No events.

注意上面的Endpoints是两个my-nginx pod的pod IP地址和容器端口；Node上的kubelet-proxy进程会获取各server ip和service port对应的Endpoints对象，在Node上创建iptables规则，实现service到pod的请求转发路由表，从而实现service的智能负载均衡机制；

[root@tjwq01-sys-bs003007 ~]# kubectl run busybox --image=busybox --restart=Never --tty -i
Error from server: pods "busybox" already exists
[root@tjwq01-sys-bs003007 ~]# kubectl get pods
NAME                        READY     STATUS    RESTARTS   AGE
my-nginx-3569648077-i5w63   1/1       Running   0          9m
my-nginx-3569648077-jfff0   1/1       Running   0          9m
nginx                       1/1       Running   0          27m
[root@tjwq01-sys-bs003007 ~]# kubectl get pods -a
NAME                        READY     STATUS      RESTARTS   AGE
busybox                     0/1       Completed   0          15m
my-nginx-3569648077-i5w63   1/1       Running     0          10m
my-nginx-3569648077-jfff0   1/1       Running     0          10m
nginx                       1/1       Running     0          28m
[root@tjwq01-sys-bs003007 ~]# kubectl delete pod busybox  # kubernetes没有重新启动pod的功，故需要删除后重建pod
pod "busybox" deleted
[root@tjwq01-sys-bs003007 ~]# kubectl run busybox --image=busybox --restart=Never --tty -i --generator=run-pod/v1
Waiting for pod default/busybox to be running, status is Pending, pod ready: false

Hit enter for command prompt
/ # env|sort   # 可见kubernetes自动在新建的pod中为各service创建了环境变量；
HOME=/root
HOSTNAME=busybox
KUBERNETES_PORT=tcp://10.254.0.1:443
KUBERNETES_PORT_443_TCP=tcp://10.254.0.1:443
KUBERNETES_PORT_443_TCP_ADDR=10.254.0.1
KUBERNETES_PORT_443_TCP_PORT=443
KUBERNETES_PORT_443_TCP_PROTO=tcp
KUBERNETES_SERVICE_HOST=10.254.0.1
KUBERNETES_SERVICE_PORT=443
KUBERNETES_SERVICE_PORT_HTTPS=443
MY_NGINX_PORT=tcp://10.254.238.37:80
MY_NGINX_PORT_80_TCP=tcp://10.254.238.37:80
MY_NGINX_PORT_80_TCP_ADDR=10.254.238.37
MY_NGINX_PORT_80_TCP_PORT=80
MY_NGINX_PORT_80_TCP_PROTO=tcp
MY_NGINX_SERVICE_HOST=10.254.238.37
MY_NGINX_SERVICE_PORT=80
/ # ping 10.254.238.37
PING 10.254.238.37 (10.254.238.37): 56 data bytes
^C
--- 10.254.238.37 ping statistics ---
9 packets transmitted, 0 packets received, 100% packet loss
/ # wget -q -O - 10.254.238.37:80 // 使用服务IP可以正常访问nginx pod


测试 NodePort
[root@tjwq01-sys-bs003007 ~]# kubectl create -f my-nginx-nodeport.yaml
The Service "my-nginx-nodePort" is invalid.
metadata.name: Invalid value: "my-nginx-nodePort": must be a DNS 952 label (at most 24 characters, matching regex [a-z]([-a-z0-9]*[a-z0-9])?): e.g. "my-name"
[root@tjwq01-sys-bs003007 ~]# vim my-nginx-nodeport.yaml
[root@tjwq01-sys-bs003007 ~]# kubectl create -f my-nginx-nodeport.yaml
service "my-nginx-nodeport" created

[root@tjwq01-sys-bs003007 ~]# kubectl get svc
NAME                CLUSTER-IP      EXTERNAL-IP   PORT(S)        AGE
kubernetes          10.254.0.1      <none>        443/TCP        85d
my-nginx            10.254.238.37   <none>        80/TCP         15m // 只能在集群内部通过CLUSTER-IP访问服务；
my-nginx-nodeport   10.254.83.41    <nodes>       80:30062/TCP   1m // 注意是EXTERNAL-IP是nodes

[root@tjwq01-sys-bs003007 ~]# kubectl describe svc my-nginx-nodeport
Name:                   my-nginx-nodeport
Namespace:              default
Labels:                 run=my-nginx
Selector:               run=my-nginx
Type:                   NodePort
IP:                     10.254.83.41
Port:                   <unset> 80/TCP
NodePort:               <unset> 30062/TCP
Endpoints:              172.30.19.4:80,172.30.19.5:80
Session Affinity:       None
No events.

[root@tjwq01-sys-bs003007 ~]#  netstat -lnpt|grep kube|sort
tcp        0      0 10.64.3.7:6443          0.0.0.0:*               LISTEN      46304/kube-apiserve
tcp        0      0 127.0.0.1:10248         0.0.0.0:*               LISTEN      3025/kubelet
tcp        0      0 127.0.0.1:10249         0.0.0.0:*               LISTEN      46662/kube-proxy
tcp        0      0 127.0.0.1:10250         0.0.0.0:*               LISTEN      3025/kubelet
tcp        0      0 127.0.0.1:10251         0.0.0.0:*               LISTEN      46584/kube-schedule
tcp        0      0 127.0.0.1:10252         0.0.0.0:*               LISTEN      46424/kube-controll
tcp        0      0 127.0.0.1:10255         0.0.0.0:*               LISTEN      3025/kubelet
tcp        0      0 127.0.0.1:8080          0.0.0.0:*               LISTEN      46304/kube-apiserve
tcp6       0      0 :::30062                :::*                    LISTEN      46662/kube-proxy
tcp6       0      0 :::4194                 :::*                    LISTEN      3025/kubelet

注意NodePort 30062是kube-proxy创建的；4194是kubelet cAdvisor监听的端口；

[root@tjwq01-sys-bs003007 ~]# curl 120.92.8.114:30062 #通过node的公网IP和端口访问成功，参考下面的iptables nat规则

下面是通过NodePort访问service，进而重定向到两个pod中容器端口的iptables nat规则(同时也包含通过cluster IP访问服务，重定向到pod IP的规则，注意如果pod位于其它node上，则是通过flannel的overlay网络进行互通的)，这些规则是kube-proxy维护的(kube-proxy的proxy-mode默认值为iptables(4层代理)，如果使用userspace(七层代理，不建议)，则由node上的kube-proxy进行转发到后端pod，这会引起double connection)：

Chain PREROUTING (policy ACCEPT)
target     prot opt source               destination
KUBE-SERVICES  all  --  0.0.0.0/0            0.0.0.0/0            /* kubernetes service portals */
DOCKER     all  --  0.0.0.0/0            0.0.0.0/0            ADDRTYPE match dst-type LOCAL

Chain OUTPUT (policy ACCEPT)
target     prot opt source               destination
KUBE-SERVICES  all  --  0.0.0.0/0            0.0.0.0/0            /* kubernetes service portals */

[root@tjwq01-sys-bs003007 ~]# kubectl create -f pod-nginx.yaml
pod "nginx" created
[root@tjwq01-sys-bs003007 ~]# kubectl get pods
NAME      READY     STATUS    RESTARTS   AGE
nginx     1/1       Running   0          7s
[root@tjwq01-sys-bs003007 ~]# docker ps -a
CONTAINER ID        IMAGE                                                        COMMAND                  CREATED             STATUS              PORTS               NAMES
2c79df4182b7        nginx:1.7.9                                                  "nginx -g 'daemon off"   17 seconds ago      Up 15 seconds                           k8s_nginx.7413299_nginx_default_7b5cc851-06e5-11e7-8472-8cdcd4b3be48_b01a814e
e52a54bc7693        registry.access.redhat.com/rhel7/pod-infrastructure:latest   "/pod"                   18 seconds ago      Up 16 seconds                           k8s_POD.a8590b41_nginx_default_7b5cc851-06e5-11e7-8472-8cdcd4b3be48_bde886e8
0c4caa5345f7        pstauffer/curl:latest                                        "bash"                   12 weeks ago        Created                          

查看kube-scheduler的日志，最后一行内容为：
MESSAGE=I1216 13:25:00.830715   10957 event.go:211] Event(api.ObjectReference{Kind:"Pod", Namespace:"default", Name:"nginx", UID:"fdbd3d8b-c34f-11e6-b4dd-8cdcd4b3be48", APIVersion:"v1", ResourceVersion:"281", FieldPath:""}): type: 'Normal' reason: 'Scheduled' Successfully assigned nginx to 127.0.0.1

稍等一会查看pod状态
[root@tjwq01-sys-bs003007 ~]# kubectl describe pod nginx
Name:           nginx                                                                                                                                 [2/1952]
Namespace:      default
Node:           127.0.0.1/127.0.0.1
Start Time:     Sun, 12 Mar 2017 13:33:54 +0800
Labels:         <none>
Status:         Running
IP:             172.30.19.2
Controllers:    <none>
Containers:
  nginx:
    Container ID:       docker://2c79df4182b7c12cb18d10cbf05f1577c78522d6b038390f80765c40e294d94a
    Image:              nginx:1.7.9
    Image ID:           docker://sha256:84581e99d807a703c9c03bd1a31cd9621815155ac72a7365fd02311264512656
    Port:               80/TCP
    State:              Running
      Started:          Sun, 12 Mar 2017 13:33:56 +0800
    Ready:              True
    Restart Count:      0
    Volume Mounts:
      /var/run/secrets/kubernetes.io/serviceaccount from default-token-icgdo (ro)
    Environment Variables:      <none>
Conditions:
  Type          Status
  Initialized   True
  Ready         True
  PodScheduled  True
Volumes:
  default-token-icgdo:
    Type:       Secret (a volume populated by a Secret)
    SecretName: default-token-icgdo
QoS Class:      BestEffort
Tolerations:    <none>
Events:
  FirstSeen     LastSeen        Count   From                    SubObjectPath           Type            Reason                  Message
  ---------     --------        -----   ----                    -------------           --------        ------                  -------
  1m            1m              1       {default-scheduler }                            Normal          Scheduled               Successfully assigned nginx to
 127.0.0.1
  1m            1m              2       {kubelet 127.0.0.1}                             Warning         MissingClusterDNS       kubelet does not have ClusterD
NS IP configured and cannot create Pod using "ClusterFirst" policy. Falling back to DNSDefault policy.
  1m            1m              1       {kubelet 127.0.0.1}     spec.containers{nginx}  Normal          Pulled                  Container image "nginx:1.7.9"
already present on machine
  1m            1m              1       {kubelet 127.0.0.1}     spec.containers{nginx}  Normal          Created                 Created container with docker
id 2c79df4182b7
  1m            1m              1       {kubelet 127.0.0.1}     spec.containers{nginx}  Normal          Started                 Started container with docker
id 2c79df4182b7

如果启用了service account，在创建pod时，自动在 /var/run/secrets/kubernetes.io/serviceaccount 位置挂载 sercret;

[root@tjwq01-sys-bs003007 ~]# docker ps
CONTAINER ID        IMAGE                                                        COMMAND                  CREATED             STATUS              PORTS               NAMES
90f7e9cd2bc1        nginx:1.7.9                                                  "nginx -g 'daemon off"   2 minutes ago       Up 2 minutes                            k8s_nginx.4580025_nginx_default_fdbd3d8b-c34f-11e6-b4dd-8cdcd4b3be48_ad0b32f1
5a0c9355bfc8        registry.access.redhat.com/rhel7/pod-infrastructure:latest   "/pod"                   6 minutes ago       Up 6 minutes                            k8s_POD.c36b0a77_nginx_default_fdbd3d8b-c34f-11e6-b4dd-8cdcd4b3be48_3ed87315
[root@tjwq01-sys-bs003007 ~]# docker images
REPOSITORY                                            TAG                 IMAGE ID            CREATED             SIZE
registry.access.redhat.com/rhel7/pod-infrastructure   latest              7d5548e9fb99        5 weeks ago         205.3 MB
docker.io/nginx                                       1.7.9               84581e99d807        22 months ago       91.64 MB

[root@tjwq01-sys-bs003007 ~]# docker exec -it 2c79df4182b7 bash
root@nginx:/# ip addr show
1: lo: <LOOPBACK,UP,LOWER_UP> mtu 65536 qdisc noqueue state UNKNOWN qlen 1
    link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00
    inet 127.0.0.1/8 scope host lo
       valid_lft forever preferred_lft forever
    inet6 ::1/128 scope host
       valid_lft forever preferred_lft forever
8: eth0@if9: <BROADCAST,MULTICAST,UP,LOWER_UP,M-DOWN> mtu 1450 qdisc noqueue state UP
    link/ether 02:42:ac:1e:13:02 brd ff:ff:ff:ff:ff:ff
    inet 172.30.19.2/24 scope global eth0
       valid_lft forever preferred_lft forever
    inet6 fe80::42:acff:fe1e:1302/64 scope link tentative dadfailed
       valid_lft forever preferred_lft forever

可见nginx容器的IP地址为flannel的一个/24段地址中的第二个；

root@nginx:/# ls -l /run/secrets/kubernetes.io/serviceaccount/
total 0
lrwxrwxrwx 1 root root 16 Mar 12 05:33 namespace -> ..data/namespace
lrwxrwxrwx 1 root root 12 Mar 12 05:33 token -> ..data/token

起另外一个测试pod:
[root@tjwq01-sys-bs003007 ~]# kubectl run busybox --image=busybox --restart=Never --tty -i  --env "POD_IP=$(kubectl get pod nginx -o go-template='{{.status.podIP}}')"
Waiting for pod default/busybox to be running, status is Pending, pod ready: false
If you don't see a command prompt, try pressing enter.
/ # ifconfig
eth0      Link encap:Ethernet  HWaddr 02:42:AC:1E:13:03
          inet addr:172.30.19.3  Bcast:0.0.0.0  Mask:255.255.255.0
          inet6 addr: fe80::42:acff:fe1e:1303/64 Scope:Link
          UP BROADCAST RUNNING MULTICAST  MTU:1450  Metric:1
          RX packets:10 errors:0 dropped:0 overruns:0 frame:0
          TX packets:3 errors:0 dropped:0 overruns:0 carrier:0
          collisions:0 txqueuelen:0
          RX bytes:816 (816.0 B)  TX bytes:258 (258.0 B)
...
/ # wget -qO- http://$POD_IP  // 可以用nginx pod 的 pod ip访问web服务；

FIXME: 对于docker而言，每起一个pod都会在host上创建一个名称类似veth24a8440的虚拟网卡，该网卡和对端的container eth网卡互联，都接到docker0 bridge上；
[zhangjun3@tjwq01-sys-bs003007 ~]$ ifconfig |grep veth
veth24a8440: flags=4163<UP,BROADCAST,RUNNING,MULTICAST>  mtu 1450
vethb747827: flags=4163<UP,BROADCAST,RUNNING,MULTICAST>  mtu 1450

[root@tjwq01-sys-bs003007 ~]# kubectl create -f my-nginx.yaml
deployment "my-nginx" created
[root@tjwq01-sys-bs003007 ~]# kubectl get deploy
NAME       DESIRED   CURRENT   UP-TO-DATE   AVAILABLE   AGE
my-nginx   2         2         2            2           11s
[root@tjwq01-sys-bs003007 ~]# kubectl get pods
NAME                        READY     STATUS    RESTARTS   AGE
busybox                     1/1       Running   0          6m
my-nginx-3467165555-1jw53   1/1       Running   0          4s
my-nginx-3467165555-m3c25   1/1       Running   0          4s
nginx                       1/1       Running   0          15m
[root@tjwq01-sys-bs003007 ~]# kubectl describe  deploy my-nginx
Name:                   my-nginx
Namespace:              default
CreationTimestamp:      Sun, 12 Mar 2017 13:48:57 +0800
Labels:                 run=my-nginx
Selector:               run=my-nginx
Replicas:               2 updated | 2 total | 2 available | 0 unavailable
StrategyType:           RollingUpdate
MinReadySeconds:        0
RollingUpdateStrategy:  1 max unavailable, 1 max surge
Conditions:
  Type          Status  Reason
  ----          ------  ------
  Available     True    MinimumReplicasAvailable
OldReplicaSets: <none>
NewReplicaSet:  my-nginx-3467165555 (2/2 replicas created)
Events:
  FirstSeen     LastSeen        Count   From                            SubObjectPath   Type            Reason                  Message
  ---------     --------        -----   ----                            -------------   --------        ------                  -------
  2m            2m              1       {deployment-controller }                        Normal          ScalingReplicaSet       Scaled up replica set my-nginx-3467165555 to 2

my-nginx分别分配了两个pod ip：172.30.19.4、172.30.19.5
[root@tjwq01-sys-bs003007 ~]# kubectl expose deploy my-nginx
service "my-nginx" exposed
[root@tjwq01-sys-bs003007 ~]# kubectl get service
NAME         CLUSTER-IP      EXTERNAL-IP   PORT(S)   AGE
kubernetes   10.254.0.1      <none>        443/TCP   85d
my-nginx     10.254.238.37   <none>        80/TCP    8s

可见分配的server IP 为kube-apiserver 参数service-cluster-ip-range=10.254.0.0/16值中的一部分；

[root@tjwq01-sys-bs003007 ~]# kubectl describe svc my-nginx
Name:                   my-nginx
Namespace:              default
Labels:                 run=my-nginx
Selector:               run=my-nginx
Type:                   ClusterIP
IP:                     10.254.238.37
Port:                   <unset> 80/TCP
Endpoints:              172.30.19.4:80,172.30.19.5:80
Session Affinity:       None
No events.

注意上面的Endpoints是两个my-nginx pod的pod IP地址和容器端口；Node上的kubelet-proxy进程会获取各server ip和service port对应的Endpoints对象，在Node上创建iptables规则，实现
service到pod的请求转发路由表，从而实现service的智能负载均衡机制；

[root@tjwq01-sys-bs003007 ~]# kubectl run busybox --image=busybox --restart=Never --tty -i
Error from server: pods "busybox" already exists
[root@tjwq01-sys-bs003007 ~]# kubectl get pods
NAME                        READY     STATUS    RESTARTS   AGE
my-nginx-3569648077-i5w63   1/1       Running   0          9m
my-nginx-3569648077-jfff0   1/1       Running   0          9m
nginx                       1/1       Running   0          27m
[root@tjwq01-sys-bs003007 ~]# kubectl get pods -a
NAME                        READY     STATUS      RESTARTS   AGE
busybox                     0/1       Completed   0          15m
my-nginx-3569648077-i5w63   1/1       Running     0          10m
my-nginx-3569648077-jfff0   1/1       Running     0          10m
nginx                       1/1       Running     0          28m
[root@tjwq01-sys-bs003007 ~]# kubectl delete pod busybox  # kubernetes没有重新启动pod的功，故需要删除后重建pod
pod "busybox" deleted
[root@tjwq01-sys-bs003007 ~]# kubectl run busybox --image=busybox --restart=Never --tty -i --generator=run-pod/v1
Waiting for pod default/busybox to be running, status is Pending, pod ready: false

Hit enter for command prompt
/ # env|sort   # 可见kubernetes自动在新建的pod中为各service创建了环境变量；
HOME=/root
HOSTNAME=busybox
KUBERNETES_PORT=tcp://10.254.0.1:443
KUBERNETES_PORT_443_TCP=tcp://10.254.0.1:443
KUBERNETES_PORT_443_TCP_ADDR=10.254.0.1
KUBERNETES_PORT_443_TCP_PORT=443
KUBERNETES_PORT_443_TCP_PROTO=tcp
KUBERNETES_SERVICE_HOST=10.254.0.1
KUBERNETES_SERVICE_PORT=443
KUBERNETES_SERVICE_PORT_HTTPS=443
MY_NGINX_PORT=tcp://10.254.238.37:80
MY_NGINX_PORT_80_TCP=tcp://10.254.238.37:80
MY_NGINX_PORT_80_TCP_ADDR=10.254.238.37
MY_NGINX_PORT_80_TCP_PORT=80
MY_NGINX_PORT_80_TCP_PROTO=tcp
MY_NGINX_SERVICE_HOST=10.254.238.37
MY_NGINX_SERVICE_PORT=80
/ # ping 10.254.238.37
PING 10.254.238.37 (10.254.238.37): 56 data bytes
^C
--- 10.254.238.37 ping statistics ---
9 packets transmitted, 0 packets received, 100% packet loss
/ # wget -q -O - 10.254.238.37:80 // 使用服务IP可以正常访问nginx pod


测试 NodePort
[root@tjwq01-sys-bs003007 ~]# kubectl create -f my-nginx-nodeport.yaml
The Service "my-nginx-nodePort" is invalid.
metadata.name: Invalid value: "my-nginx-nodePort": must be a DNS 952 label (at most 24 characters, matching regex [a-z]([-a-z0-9]*[a-z0-9])?): e.g. "my-name"
[root@tjwq01-sys-bs003007 ~]# vim my-nginx-nodeport.yaml
[root@tjwq01-sys-bs003007 ~]# kubectl create -f my-nginx-nodeport.yaml
service "my-nginx-nodeport" created

[root@tjwq01-sys-bs003007 ~]# kubectl get svc
NAME                CLUSTER-IP      EXTERNAL-IP   PORT(S)        AGE
kubernetes          10.254.0.1      <none>        443/TCP        85d
my-nginx            10.254.238.37   <none>        80/TCP         15m // 只能在集群内部通过CLUSTER-IP访问服务；
my-nginx-nodeport   10.254.83.41    <nodes>       80:30062/TCP   1m // 注意是EXTERNAL-IP是nodes

[root@tjwq01-sys-bs003007 ~]# kubectl describe svc my-nginx-nodeport
Name:                   my-nginx-nodeport
Namespace:              default
Labels:                 run=my-nginx
Selector:               run=my-nginx
Type:                   NodePort
IP:                     10.254.83.41
Port:                   <unset> 80/TCP
NodePort:               <unset> 30062/TCP
Endpoints:              172.30.19.4:80,172.30.19.5:80
Session Affinity:       None
No events.

[root@tjwq01-sys-bs003007 ~]#  netstat -lnpt|grep kube|sort
tcp        0      0 10.64.3.7:6443          0.0.0.0:*               LISTEN      46304/kube-apiserve
tcp        0      0 127.0.0.1:10248         0.0.0.0:*               LISTEN      3025/kubelet
tcp        0      0 127.0.0.1:10249         0.0.0.0:*               LISTEN      46662/kube-proxy
tcp        0      0 127.0.0.1:10250         0.0.0.0:*               LISTEN      3025/kubelet
tcp        0      0 127.0.0.1:10251         0.0.0.0:*               LISTEN      46584/kube-schedule
tcp        0      0 127.0.0.1:10252         0.0.0.0:*               LISTEN      46424/kube-controll
tcp        0      0 127.0.0.1:10255         0.0.0.0:*               LISTEN      3025/kubelet
tcp        0      0 127.0.0.1:8080          0.0.0.0:*               LISTEN      46304/kube-apiserve
tcp6       0      0 :::30062                :::*                    LISTEN      46662/kube-proxy
tcp6       0      0 :::4194                 :::*                    LISTEN      3025/kubelet

调度中心master，主要有四个组件构成：

1. etcd 作为配置中心和存储服务保存了所有组件的定义以及状态，Kubernetes的多个组件之间的互相交互也主要通过etcd。
[root@tjwq01-sys-bs003007 ~]# etcdctl ls /
/kube-centos
/registry

[root@tjwq01-sys-bs003007 ~]# etcdctl ls /registry
/registry/services
/registry/serviceaccounts
/registry/events #保存所有变更事件
/registry/pods 
/registry/ranges
/registry/namespaces
/registry/minions  #保存所有node节点信息
/registry/deployments
/registry/replicasets
[root@tjwq01-sys-bs003007 ~]# etcdctl ls /registry/services
/registry/services/specs
/registry/services/endpoints
[root@tjwq01-sys-bs003007 ~]# etcdctl ls /registry/services/specs/default
/registry/services/specs/default/kubernetes   // kubernets 是kube自身的服务apiserver
/registry/services/specs/default/my-nginx
/registry/services/specs/default/my-nginx-nodeport
[root@tjwq01-sys-bs003007 ~]# etcdctl get /registry/services/specs/default/kubernetes 
{"kind":"Service","apiVersion":"v1","metadata":{"name":"kubernetes","namespace":"default","uid":"7f0e4a85-c344-11e6-8440-8cdcd4b3be48","creationTimestamp":"2016-12-16T04:02:43Z","labels":{"component":"apiserver","provider":"kubernetes"}},"spec":{"ports":[{"name":"https","protocol":"TCP","port":443,"targetPort":443}],"portalIP":"10.254.0.1","clusterIP":"10.254.0.1","type":"ClusterIP","sessionAffinity":"None"},"status":{"loadBalancer":{}}}

[root@tjwq01-sys-bs003007 ~]# etcdctl ls /registry/services/endpoints/default
/registry/services/endpoints/default/kubernetes
/registry/services/endpoints/default/my-nginx
/registry/services/endpoints/default/my-nginx-nodeport

[root@tjwq01-sys-bs003007 ~]# etcdctl get  /registry/services/endpoints/default/kubernetes
{"kind":"Endpoints","apiVersion":"v1","metadata":{"name":"kubernetes","namespace":"default","uid":"7f0e8771-c344-11e6-8440-8cdcd4b3be48","creationTimestamp":"2016-12-16T04:02:43Z"},"subsets":[{"addresses":[{"ip":"120.92.8.114"}],"ports":[{"name":"https","port":6443,"protocol":"TCP"}]}]}

[root@tjwq01-sys-bs003007 ~]# etcdctl get /registry/pods/default/nginx
{"kind":"Pod","apiVersion":"v1","metadata":{"name":"nginx","namespace":"default","selfLink":"/api/v1/namespaces/default/pods/nginx","uid":"fdbd3d8b-c34f-11e6-b4dd-8cdcd4b3be48","creationTimestamp":"2016-12-16T05:25:00Z"},"spec":{"containers":[{"name":"nginx","image":"nginx:1.7.9","ports":[{"containerPort":80,"protocol":"TCP"}],"resources":{},"terminationMessagePath":"/dev/termination-log","imagePullPolicy":"IfNotPresent"}],"restartPolicy":"Always","terminationGracePeriodSeconds":30,"dnsPolicy":"ClusterFirst","host":"127.0.0.1","nodeName":"127.0.0.1","securityContext":{}},"status":{"phase":"Running","conditions":[{"type":"Ready","status":"True","lastProbeTime":null,"lastTransitionTime":"2016-12-16T05:28:51Z"}],"hostIP":"127.0.0.1","podIP":"172.30.35.2","startTime":"2016-12-16T05:25:00Z","containerStatuses":[{"name":"nginx","state":{"running":{"startedAt":"2016-12-16T05:28:51Z"}},"lastState":{},"ready":true,"restartCount":0,"image":"nginx:1.7.9","imageID":"docker://sha256:84581e99d807a703c9c03bd1a31cd9621815155ac72a7365fd02311264512656","containerID":"docker://90f7e9cd2bc1481dd8e6cedb357b193ef91afeb3f36eb13f21afaaae61c4753f"}]}}

2. kube-apiserver 提供和外部交互的接口，提供安全机制，大多数接口都是直接读写etcd中的数据。
3. kube-scheduler 调度器，主要干一件事情：监听etcd中的pods目录变更，然后通过调度算法分配node，最后调用apiserver的bind接口将分配的node和pod进行关联（修改pod节点中的nodeName属性）。
scheduler在Kubernetes中是一个plugin，可以用其他的实现替换（比如mesos）。。有不同的算法提供，算法接口如下:

  type ScheduleAlgorithm interface {
      Schedule(api.Pod, NodeLister) (selectedMachine string, err error)
  }   
4. kube-controller-manager 承担了master的主要功能，比如和CloudProvider(IaaS)交互，管理node，pod，replication，service，namespace等。基本机制是监听etcd /registry/events下对应的事件，进行处理。具体的逻辑需要专门文章分析，此处不进行详解。

节点上的agent，主要有两个组件：

1. kubelet 主要包含容器管理，镜像管理，Volume管理等。同时kubelet也是一个rest服务，和pod相关的命令操作都是通过调用接口实现的。比如：查看pod日志，在pod上执行命令等。pod的启动以及销毁操作依然是通过监听etcd的变更进行操作的。
但kubelet不直接和etcd交互，而是通过apiserver提供的watch机制，应该是出于安全的考虑。kubelet提供插件机制，用于支持Volume和Network的扩展。
2. kube-proxy 主要用于实现Kubernetes的service机制。提供一部分SDN功能以及集群内部的智能LoadBalancer。前面我们也分析了，应用实例在多个服务器节点之间迁移的一个难题是网络和端口冲突问题。Kubernetes为每个service分配一个clusterIP（虚拟ip）。不同的service用不同的ip，所以端口也不会冲突。
Kubernetes的虚拟ip是通过iptables机制实现的。每个service定义的端口，kube-proxy都会监听一个随机端口对应，然后通过iptables nat规则做转发。比如Kubernetes上有个dns服务，clusterIP:10.254.0.10，端口:53。应用对10.254.0.10:53的请求会被转发到该node的kube-proxy监听的随机端口上，
然后再转发给对应的pod。如果该服务的pod不在当前node上，会先在kube-proxy之间进行转发。当前版本的kube-proxy是通过tcp代理实现的，性能损失比较大（具体参看后面的压测比较），1.2版本中已经计划将kube-proxy完全通过iptables实现(https://github.com/kubernetes/kubernetes/issues/3760)
3. Pods Kubernetes将应用的具体实例抽象为pod。每个pod首先会启动一个google_containers/pause docker容器，然后再启动应用真正的docker容器。这样做的目的是为了可以将多个docker容器封装到一个pod中，共享同一个pod网络地址。
4. Replication Controller 控制pod的副本数量，高可用就靠它了。
5. Services service是对一组pods的抽象，通过kube-proxy的智能LoadBalancer机制，pods的销毁迁移不会影响services的功能以及上层的调用方。Kubernetes对service的抽象可以将底层服务和上层服务的依赖关系解耦，同时实现了和Docker links类似的环境变量注入机制(https://github.com/kubernetes/kubernetes/blob/release-1.0/docs/user-guide/services.md#environment-variables)，但更灵活。如果配合dns的短域名解析机制，最终可实现完全解耦。
6. Label key-value格式的标签，主要用于筛选，比如service和后端的pod是通过label进行筛选的，是弱关联的。
7. Namespace Kubernetes中的namespace主要用来避免pod，service的名称冲突。同一个namespace内的pod，service的名称必须是唯一的。
8. Kubectl Kubernetes的命令行工具，主要是通过调用apiserver来实现管理。
9. Kube-dns dns是Kubernetes之上的应用，通过设置Pod的dns searchDomain（由kubelet启动pod时进行操作），可以实现同一个namespace中的service直接通过名称解析（这样带来的好处是开发测试正式环境可以共用同一套配置）。主要包含以下组件，这几个组件是打包到同一个pod中的。
    + etcd skydns依赖，用于存储dns数据
    + skydns 开源的dns服务
    + kube2sky 通过apiserver的接口监听kube内部变更，然后调用skydns的接口操作dns
10. Networking Kubernetes的理念里，pod之间是可以直接通讯的(http://kubernetes.io/v1.1/docs/admin/networking.html)，但实际上并没有内置解决方案，需要用户自己选择解决方案: Flannel,OpenVSwitch,Weave 等。我们测试用的是Flannel,比较简单。
11. 配置文件 Kubernetes 支持yaml和json格式的配置文件，主要用来定义pod,replication controller,service,namespace等。


参考：http://tonybai.com/2016/10/18/learn-how-to-install-kubernetes-on-ubuntu/
DOCKER     all  --  0.0.0.0/0           !127.0.0.0/8          ADDRTYPE match dst-type LOCAL

Chain KUBE-SERVICES (2 references)
target     prot opt source               destination
KUBE-SVC-NPX46M4PTMTKRN6Y  tcp  --  0.0.0.0/0            10.254.0.1           /* default/kubernetes:https cluster IP */ tcp dpt:443
KUBE-SVC-BEPXDJBUHFCSYIC3  tcp  --  0.0.0.0/0            10.254.214.196       /* default/my-nginx: cluster IP */ tcp dpt:80
KUBE-SVC-6JXEEPSEELXY3JZG  tcp  --  0.0.0.0/0            10.254.75.20         /* default/my-nginx-nodeport: cluster IP */ tcp dpt:80
KUBE-NODEPORTS  all  --  0.0.0.0/0            0.0.0.0/0            /* kubernetes service nodeports; NOTE: this must be the last rule in this chain */ ADDRTYPE
 match dst-type LOCAL

Chain KUBE-NODEPORTS (1 references)
target     prot opt source               destination
KUBE-MARK-MASQ  tcp  --  0.0.0.0/0            0.0.0.0/0            /* default/my-nginx-nodeport: */ tcp dpt:30062
KUBE-SVC-6JXEEPSEELXY3JZG  tcp  --  0.0.0.0/0            0.0.0.0/0            /* default/my-nginx-nodeport: */ tcp dpt:30062

Chain KUBE-MARK-MASQ (6 references)
target     prot opt source               destination
MARK       all  --  0.0.0.0/0            0.0.0.0/0            MARK or 0x4000

Chain KUBE-SVC-6JXEEPSEELXY3JZG (2 references)
target     prot opt source               destination
KUBE-SEP-4FG6L5UGKXBFMUQO  all  --  0.0.0.0/0            0.0.0.0/0            /* default/my-nginx-nodeport: */ statistic mode random probability 0.50000000000
KUBE-SEP-5ZT2IRSHQSOHPNW6  all  --  0.0.0.0/0            0.0.0.0/0            /* default/my-nginx-nodeport: */

Chain KUBE-SEP-4FG6L5UGKXBFMUQO (1 references)
target     prot opt source               destination
KUBE-MARK-MASQ  all  --  172.30.35.3          0.0.0.0/0            /* default/my-nginx-nodeport: */
DNAT       tcp  --  0.0.0.0/0            0.0.0.0/0            /* default/my-nginx-nodeport: */ tcp to:172.30.35.3:80

Chain KUBE-SEP-5ZT2IRSHQSOHPNW6 (1 references)
target     prot opt source               destination
KUBE-MARK-MASQ  all  --  172.30.35.4          0.0.0.0/0            /* default/my-nginx-nodeport: */
DNAT       tcp  --  0.0.0.0/0            0.0.0.0/0            /* default/my-nginx-nodeport: */ tcp to:172.30.35.4:80

所有NAT规则：
[root@tjwq01-sys-bs003007 ~]#  iptables  -t nat -S|sort
-A DOCKER -i docker0 -j RETURN
-A KUBE-MARK-MASQ -j MARK --set-xmark 0x4000/0x4000
-A KUBE-NODEPORTS -p tcp -m comment --comment "default/my-nginx-nodeport:" -m tcp --dport 30062 -j KUBE-MARK-MASQ
-A KUBE-NODEPORTS -p tcp -m comment --comment "default/my-nginx-nodeport:" -m tcp --dport 30062 -j KUBE-SVC-6JXEEPSEELXY3JZG
-A KUBE-POSTROUTING -m comment --comment "kubernetes service traffic requiring SNAT" -m mark --mark 0x4000/0x4000 -j MASQUERADE
-A KUBE-SEP-4FG6L5UGKXBFMUQO -p tcp -m comment --comment "default/my-nginx-nodeport:" -m tcp -j DNAT --to-destination 172.30.35.3:80
-A KUBE-SEP-4FG6L5UGKXBFMUQO -s 172.30.35.3/32 -m comment --comment "default/my-nginx-nodeport:" -j KUBE-MARK-MASQ
-A KUBE-SEP-5ZT2IRSHQSOHPNW6 -p tcp -m comment --comment "default/my-nginx-nodeport:" -m tcp -j DNAT --to-destination 172.30.35.4:80
-A KUBE-SEP-5ZT2IRSHQSOHPNW6 -s 172.30.35.4/32 -m comment --comment "default/my-nginx-nodeport:" -j KUBE-MARK-MASQ
-A KUBE-SEP-6UWOCEJGDDLPG44I -p tcp -m comment --comment "default/my-nginx:" -m tcp -j DNAT --to-destination 172.30.35.4:80
-A KUBE-SEP-6UWOCEJGDDLPG44I -s 172.30.35.4/32 -m comment --comment "default/my-nginx:" -j KUBE-MARK-MASQ
-A KUBE-SEP-MD7CSUJEHAB42YAZ -p tcp -m comment --comment "default/kubernetes:https" -m tcp -j DNAT --to-destination 120.92.8.114:6443
-A KUBE-SEP-MD7CSUJEHAB42YAZ -s 120.92.8.114/32 -m comment --comment "default/kubernetes:https" -j KUBE-MARK-MASQ
-A KUBE-SEP-S3PQ6OUSIWQAQBLR -p tcp -m comment --comment "default/my-nginx:" -m tcp -j DNAT --to-destination 172.30.35.3:80
-A KUBE-SEP-S3PQ6OUSIWQAQBLR -s 172.30.35.3/32 -m comment --comment "default/my-nginx:" -j KUBE-MARK-MASQ
-A KUBE-SERVICES -d 10.254.0.1/32 -p tcp -m comment --comment "default/kubernetes:https cluster IP" -m tcp --dport 443 -j KUBE-SVC-NPX46M4PTMTKRN6Y
-A KUBE-SERVICES -d 10.254.214.196/32 -p tcp -m comment --comment "default/my-nginx: cluster IP" -m tcp --dport 80 -j KUBE-SVC-BEPXDJBUHFCSYIC3
-A KUBE-SERVICES -d 10.254.75.20/32 -p tcp -m comment --comment "default/my-nginx-nodeport: cluster IP" -m tcp --dport 80 -j KUBE-SVC-6JXEEPSEELXY3JZG
-A KUBE-SERVICES -m comment --comment "kubernetes service nodeports; NOTE: this must be the last rule in this chain" -m addrtype --dst-type LOCAL -j KUBE-NODE
PORTS
-A KUBE-SVC-6JXEEPSEELXY3JZG -m comment --comment "default/my-nginx-nodeport:" -j KUBE-SEP-5ZT2IRSHQSOHPNW6
-A KUBE-SVC-6JXEEPSEELXY3JZG -m comment --comment "default/my-nginx-nodeport:" -m statistic --mode random --probability 0.50000000000 -j KUBE-SEP-4FG6L5UGKXBF
MUQO
-A KUBE-SVC-BEPXDJBUHFCSYIC3 -m comment --comment "default/my-nginx:" -j KUBE-SEP-6UWOCEJGDDLPG44I
-A KUBE-SVC-BEPXDJBUHFCSYIC3 -m comment --comment "default/my-nginx:" -m statistic --mode random --probability 0.50000000000 -j KUBE-SEP-S3PQ6OUSIWQAQBLR
-A KUBE-SVC-NPX46M4PTMTKRN6Y -m comment --comment "default/kubernetes:https" -j KUBE-SEP-MD7CSUJEHAB42YAZ
-A OUTPUT ! -d 127.0.0.0/8 -m addrtype --dst-type LOCAL -j DOCKER
-A OUTPUT -m comment --comment "kubernetes service portals" -j KUBE-SERVICES
-A POSTROUTING -m comment --comment "kubernetes postrouting rules" -j KUBE-POSTROUTING
-A POSTROUTING -s 172.30.35.0/24 ! -o docker0 -j MASQUERADE
-A PREROUTING -m addrtype --dst-type LOCAL -j DOCKER
-A PREROUTING -m comment --comment "kubernetes service portals" -j KUBE-SERVICES


调度中心master，主要有四个组件构成：

1. etcd 作为配置中心和存储服务保存了所有组件的定义以及状态，Kubernetes的多个组件之间的互相交互也主要通过etcd。
[root@tjwq01-sys-bs003007 ~]# etcdctl ls /
/kube-centos
/registry

[root@tjwq01-sys-bs003007 ~]# etcdctl ls /registry
/registry/services
/registry/serviceaccounts
/registry/events #保存所有变更事件
/registry/pods 
/registry/ranges
/registry/namespaces
/registry/minions  #保存所有node节点信息
/registry/deployments
/registry/replicasets
[root@tjwq01-sys-bs003007 ~]# etcdctl ls /registry/services
/registry/services/specs
/registry/services/endpoints
[root@tjwq01-sys-bs003007 ~]# etcdctl ls /registry/services/specs/default
/registry/services/specs/default/kubernetes   // kubernets 是kube自身的服务apiserver
/registry/services/specs/default/my-nginx
/registry/services/specs/default/my-nginx-nodeport
[root@tjwq01-sys-bs003007 ~]# etcdctl get /registry/services/specs/default/kubernetes 
{"kind":"Service","apiVersion":"v1","metadata":{"name":"kubernetes","namespace":"default","uid":"7f0e4a85-c344-11e6-8440-8cdcd4b3be48","creationTimestamp":"2016-12-16T04:02:43Z","labels":{"component":"apiserver","provider":"kubernetes"}},"spec":{"ports":[{"name":"https","protocol":"TCP","port":443,"targetPort":443}],"portalIP":"10.254.0.1","clusterIP":"10.254.0.1","type":"ClusterIP","sessionAffinity":"None"},"status":{"loadBalancer":{}}}

[root@tjwq01-sys-bs003007 ~]# etcdctl ls /registry/services/endpoints/default
/registry/services/endpoints/default/kubernetes
/registry/services/endpoints/default/my-nginx
/registry/services/endpoints/default/my-nginx-nodeport

[root@tjwq01-sys-bs003007 ~]# etcdctl get  /registry/services/endpoints/default/kubernetes
{"kind":"Endpoints","apiVersion":"v1","metadata":{"name":"kubernetes","namespace":"default","uid":"7f0e8771-c344-11e6-8440-8cdcd4b3be48","creationTimestamp":"2016-12-16T04:02:43Z"},"subsets":[{"addresses":[{"ip":"120.92.8.114"}],"ports":[{"name":"https","port":6443,"protocol":"TCP"}]}]}

[root@tjwq01-sys-bs003007 ~]# etcdctl get /registry/pods/default/nginx
{"kind":"Pod","apiVersion":"v1","metadata":{"name":"nginx","namespace":"default","selfLink":"/api/v1/namespaces/default/pods/nginx","uid":"fdbd3d8b-c34f-11e6-b4dd-8cdcd4b3be48","creationTimestamp":"2016-12-16T05:25:00Z"},"spec":{"containers":[{"name":"nginx","image":"nginx:1.7.9","ports":[{"containerPort":80,"protocol":"TCP"}],"resources":{},"terminationMessagePath":"/dev/termination-log","imagePullPolicy":"IfNotPresent"}],"restartPolicy":"Always","terminationGracePeriodSeconds":30,"dnsPolicy":"ClusterFirst","host":"127.0.0.1","nodeName":"127.0.0.1","securityContext":{}},"status":{"phase":"Running","conditions":[{"type":"Ready","status":"True","lastProbeTime":null,"lastTransitionTime":"2016-12-16T05:28:51Z"}],"hostIP":"127.0.0.1","podIP":"172.30.35.2","startTime":"2016-12-16T05:25:00Z","containerStatuses":[{"name":"nginx","state":{"running":{"startedAt":"2016-12-16T05:28:51Z"}},"lastState":{},"ready":true,"restartCount":0,"image":"nginx:1.7.9","imageID":"docker://sha256:84581e99d807a703c9c03bd1a31cd9621815155ac72a7365fd02311264512656","containerID":"docker://90f7e9cd2bc1481dd8e6cedb357b193ef91afeb3f36eb13f21afaaae61c4753f"}]}}

2. kube-apiserver 提供和外部交互的接口，提供安全机制，大多数接口都是直接读写etcd中的数据。
3. kube-scheduler 调度器，主要干一件事情：监听etcd中的pods目录变更，然后通过调度算法分配node，最后调用apiserver的bind接口将分配的node和pod进行关联（修改pod节点中的nodeName属性）。
scheduler在Kubernetes中是一个plugin，可以用其他的实现替换（比如mesos）。。有不同的算法提供，算法接口如下:

  type ScheduleAlgorithm interface {
      Schedule(api.Pod, NodeLister) (selectedMachine string, err error)
  }   
4. kube-controller-manager 承担了master的主要功能，比如和CloudProvider(IaaS)交互，管理node，pod，replication，service，namespace等。基本机制是监听etcd /registry/events下对应的事件，进行处理。具体的逻辑需要专门文章分析，此处不进行详解。

节点上的agent，主要有两个组件：

1. kubelet 主要包含容器管理，镜像管理，Volume管理等。同时kubelet也是一个rest服务，和pod相关的命令操作都是通过调用接口实现的。比如：查看pod日志，在pod上执行命令等。pod的启动以及销毁操作依然是通过监听etcd的变更进行操作的。
但kubelet不直接和etcd交互，而是通过apiserver提供的watch机制，应该是出于安全的考虑。kubelet提供插件机制，用于支持Volume和Network的扩展。
2. kube-proxy 主要用于实现Kubernetes的service机制。提供一部分SDN功能以及集群内部的智能LoadBalancer。前面我们也分析了，应用实例在多个服务器节点之间迁移的一个难题是网络和端口冲突问题。Kubernetes为每个service分配一个clusterIP（虚拟ip）。不同的service用不同的ip，所以端口也不会冲突。
Kubernetes的虚拟ip是通过iptables机制实现的。每个service定义的端口，kube-proxy都会监听一个随机端口对应，然后通过iptables nat规则做转发。比如Kubernetes上有个dns服务，clusterIP:10.254.0.10，端口:53。应用对10.254.0.10:53的请求会被转发到该node的kube-proxy监听的随机端口上，
然后再转发给对应的pod。如果该服务的pod不在当前node上，会先在kube-proxy之间进行转发。当前版本的kube-proxy是通过tcp代理实现的，性能损失比较大（具体参看后面的压测比较），1.2版本中已经计划将kube-proxy完全通过iptables实现(https://github.com/kubernetes/kubernetes/issues/3760)
3. Pods Kubernetes将应用的具体实例抽象为pod。每个pod首先会启动一个google_containers/pause docker容器，然后再启动应用真正的docker容器。这样做的目的是为了可以将多个docker容器封装到一个pod中，共享同一个pod网络地址。
4. Replication Controller 控制pod的副本数量，高可用就靠它了。
5. Services service是对一组pods的抽象，通过kube-proxy的智能LoadBalancer机制，pods的销毁迁移不会影响services的功能以及上层的调用方。Kubernetes对service的抽象可以将底层服务和上层服务的依赖关系解耦，同时实现了和Docker links类似的环境变量注入机制(https://github.com/kubernetes/kubernetes/blob/release-1.0/docs/user-guide/services.md#environment-variables)，但更灵活。如果配合dns的短域名解析机制，最终可实现完全解耦。
6. Label key-value格式的标签，主要用于筛选，比如service和后端的pod是通过label进行筛选的，是弱关联的。
7. Namespace Kubernetes中的namespace主要用来避免pod，service的名称冲突。同一个namespace内的pod，service的名称必须是唯一的。
8. Kubectl Kubernetes的命令行工具，主要是通过调用apiserver来实现管理。
9. Kube-dns dns是Kubernetes之上的应用，通过设置Pod的dns searchDomain（由kubelet启动pod时进行操作），可以实现同一个namespace中的service直接通过名称解析（这样带来的好处是开发测试正式环境可以共用同一套配置）。主要包含以下组件，这几个组件是打包到同一个pod中的。
    + etcd skydns依赖，用于存储dns数据
    + skydns 开源的dns服务
    + kube2sky 通过apiserver的接口监听kube内部变更，然后调用skydns的接口操作dns
10. Networking Kubernetes的理念里，pod之间是可以直接通讯的(http://kubernetes.io/v1.1/docs/admin/networking.html)，但实际上并没有内置解决方案，需要用户自己选择解决方案: Flannel,OpenVSwitch,Weave 等。我们测试用的是Flannel,比较简单。
11. 配置文件 Kubernetes 支持yaml和json格式的配置文件，主要用来定义pod,replication controller,service,namespace等。


参考：http://tonybai.com/2016/10/18/learn-how-to-install-kubernetes-on-ubuntu/


4. kubelet运行一段时间后，挂载大量的tmpfs，如：
[root@tjwq01-sys-bs003007 ~]# mount|grep tmpfs|tail -3
tmpfs            95G  8.0K   95G   1% /var/lib/kubelet/pods/49e40c6e-c5a5-11e6-866d-8cdcd4b3be48/volumes/kubernetes.io~secret/default-token-icgdo
tmpfs            95G  8.0K   95G   1% /var/lib/kubelet/pods/6d69821c-c5a5-11e6-866d-8cdcd4b3be48/volumes/kubernetes.io~secret/default-token-icgdo
tmpfs            95G  8.0K   95G   1% /var/lib/kubelet/pods/6ecffb3d-c5a5-11e6-866d-8cdcd4b3be48/volumes/kubernetes.io~secret/default-token-icgdo
[root@tjwq01-sys-bs003007 ~]#

最终导致node的内存被耗尽，kubectl describe nodes的输出：
Events:
  FirstSeen     LastSeen        Count   From                    SubObjectPath   Type            Reason                  Message
  ---------     --------        -----   ----                    -------------   --------        ------                  -------
  4m            4m              1       {kubelet 127.0.0.1}                     Normal          Starting                Starting kubelet.
  4m            4m              2       {kubelet 127.0.0.1}                     Normal          NodeHasSufficientDisk   Node 127.0.0.1 status is now: NodeHasSufficientDisk
  4m            4m              1       {kubelet 127.0.0.1}                     Normal          NodeHasSufficientMemory Node 127.0.0.1 status is now: NodeHasSufficientMemory
  4m

原因：
1. Centos 7.2 YUM源里的kubenetes是1.2的版本，有bug，升级到1.3.6，https://github.com/kubernetes/kubernetes/issues/22911；
2. 查看是否有jobs在创建大量的pods，可能是由于创建的pod短时间内执行失败结束，导致job创建大量的pods；jobs结束时才删除创建的pods：
    1. kubectl get pods -a
    2. kubectl get jobs





# 集群外机器访问k8s的cluster ip
[root@tjwq01-sys-bs003007 linkerd]# cat /etc/kubernetes/proxy
KUBE_PROXY_ARGS="--bind-address=10.64.3.7 --cluster-cidr=10.254.0.0/16"
[root@tjwq01-sys-bs003007 linkerd]# systemctl restart kube-proxy.service
[root@tjwq01-sys-bs003007 linkerd]# ps -elf|grep kube-proxy
4 S root     37303     1  2  80   0 - 10469 futex_ 05:37 ?        00:00:00 /root/local/bin/kube-proxy --logtostderr=true --v=0 --master=http://10.64.3.7:8080 --bind-address=10.64.3.7 --cluster-cidr=10.254.0.0/16
0 S root     37396 40428  0  80   0 - 34400 pipe_w 05:37 pts/3    00:00:00 grep --color=auto kube-proxy

kube-proxy必须指定 --cluster-cidr 值，才会Bridge off-cluster traffic into services by masquerading，也就是在node上可以直接向cluster-ip发请求；

[root@tjwq01-sys-power003008 ~]# diff  none_cluster_cidr has_cluster_cidr  # 可见当kube-proxy指定了 --cluster-cide时，会创建很多规则：
81a82,96
> -A KUBE-SERVICES ! -s 10.254.0.0/16 -d 10.254.0.1/32 -p tcp -m comment --comment "default/kubernetes:https cluster IP" -m tcp --dport 443 -j KUBE-MARK-MASQ
> -A KUBE-SERVICES ! -s 10.254.0.0/16 -d 10.254.0.159/32 -p tcp -m comment --comment "default/my-nginx: cluster IP" -m tcp --dport 80 -j KUBE-MARK-MASQ
> -A KUBE-SERVICES ! -s 10.254.0.0/16 -d 10.254.0.2/32 -p tcp -m comment --comment "kube-system/kube-dns:dns-tcp cluster IP" -m tcp --dport 53 -j KUBE-MARK-MASQ
> -A KUBE-SERVICES ! -s 10.254.0.0/16 -d 10.254.0.2/32 -p udp -m comment --comment "kube-system/kube-dns:dns cluster IP" -m udp --dport 53 -j KUBE-MARK-MASQ
> -A KUBE-SERVICES ! -s 10.254.0.0/16 -d 10.254.119.50/32 -p tcp -m comment --comment "kube-system/kibana-logging: cluster IP" -m tcp --dport 5601 -j KUBE-MARK-MASQ
> -A KUBE-SERVICES ! -s 10.254.0.0/16 -d 10.254.120.25/32 -p tcp -m comment --comment "kube-system/kubernetes-dashboard: cluster IP" -m tcp --dport 80 -j KUBE-MARK-MASQ
> -A KUBE-SERVICES ! -s 10.254.0.0/16 -d 10.254.12.82/32 -p tcp -m comment --comment "kube-system/elasticsearch-logging: cluster IP" -m tcp --dport 9200 -j KUBE-MARK-MASQ
> -A KUBE-SERVICES ! -s 10.254.0.0/16 -d 10.254.137.54/32 -p tcp -m comment --comment "default/deployment-demo-svc: cluster IP" -m tcp --dport 80 -j KUBE-MARK-MASQ
> -A KUBE-SERVICES ! -s 10.254.0.0/16 -d 10.254.140.27/32 -p tcp -m comment --comment "kube-system/monitoring-influxdb: cluster IP" -m tcp --dport 8086 -j KUBE-MARK-MASQ
> -A KUBE-SERVICES ! -s 10.254.0.0/16 -d 10.254.144.118/32 -p tcp -m comment --comment "default/rc-demo-svc: cluster IP" -m tcp --dport 80 -j KUBE-MARK-MASQ
> -A KUBE-SERVICES ! -s 10.254.0.0/16 -d 10.254.158.94/32 -p tcp -m comment --comment "default/l5d:admin cluster IP" -m tcp --dport 9990 -j KUBE-MARK-MASQ
> -A KUBE-SERVICES ! -s 10.254.0.0/16 -d 10.254.158.94/32 -p tcp -m comment --comment "default/l5d:incoming cluster IP" -m tcp --dport 4141 -j KUBE-MARK-MASQ
> -A KUBE-SERVICES ! -s 10.254.0.0/16 -d 10.254.158.94/32 -p tcp -m comment --comment "default/l5d:outgoing cluster IP" -m tcp --dport 4140 -j KUBE-MARK-MASQ
> -A KUBE-SERVICES ! -s 10.254.0.0/16 -d 10.254.170.7/32 -p tcp -m comment --comment "kube-system/monitoring-grafana: cluster IP" -m tcp --dport 80 -j KUBE-MARK-MASQ
> -A KUBE-SERVICES ! -s 10.254.0.0/16 -d 10.254.206.144/32 -p tcp -m comment --comment "kube-system/heapster: cluster IP" -m tcp --dport 80 -j KUBE-MARK-MASQ