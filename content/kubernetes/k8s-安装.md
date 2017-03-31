---
title: 安装kubernetes平台
description: 安装kubernetes的master和node各组件，遇到的问题和解决办法
date: 2017-03-29T17:52:18-04:00
categories: ["容器", "技术"]
tags: ["kubernetes"]
toc: true
author: "张俊"
isCJKLanguage: true
---

# 安装相关RPMs

配置docker repo

``` bash
$ sudo yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo
$
```

启用repo

``` bash
$ sudo yum-config-manager --enable docker-ce-edge
$
```

删除旧的docker package

``` bash
$ sudo rpm  -e --nodeps docker docker-selinux container-selinux
$
```

安装docker-ce

``` bash
$ sudo yum install docker-ce
$ docker --version
Docker version 17.03.0-ce, build 3a232c8
```

安装其它RPMs

``` bash
$ yum -y install --enablerepo=virt7-docker-common-release kubernetes etcd flannel
$
```

实际安装了：etcd、flannel、kubernetes、kubernetes-client、kubernetes-master、 kubernetes-node

# k8s配置文件

目录:

``` shell
$ ls /etc/kubernetes/
apiserver  config  controller-manager  kubelet  proxy  scheduler  ssl
```

systemd启动各进程时，会将config文件和进程对应的配置文件(如kubelet进程引用kubelet文件)内容作为环境变量，如

``` text
[Service]
WorkingDirectory=/var/lib/kubelet
EnvironmentFile=-/etc/kubernetes/config
EnvironmentFile=-/etc/kubernetes/kubelet
```

``` bash
$ grep -v '^#' config |grep -v '^$'
KUBE_LOGTOSTDERR="--logtostderr=true"
KUBE_LOG_LEVEL="--v=0"
KUBE_ALLOW_PRIV="--allow-privileged=false"
KUBE_MASTER="--master=http://10.64.3.7:8080"
```

`/etc/kubernetes/ssl`目录保存apiserver和其它client使用的秘钥文件，制作方式参考[k8s证书.md](ks8证书.md);

``` bash
$ ls /etc/kubernetes/ssl/
ca.crt  kubecfg.crt  kubecfg.key  server.cert  server.key
```

# 起etcd服务

修改配置文件

``` shell
$ cat /etc/etcd/etcd.conf |grep -v '^#'
ETCD_NAME=default
ETCD_DATA_DIR="/var/lib/etcd/default.etcd"
ETCD_LISTEN_CLIENT_URLS="http://10.64.3.7:2379"
ETCD_ADVERTISE_CLIENT_URLS="http://10.64.3.7:2379"
```

重启服务

``` shell
$ systemctl restart etcd
$
```

向etcd写入flanneld读取的**Pod网络和子网信息**：

``` bash
$ etcdctl mkdir /kube-centos/network
$ etcdctl mk /kube-centos/network/config "{ \"Network\": \"172.30.0.0/16\", \"SubnetLen\": 24, \"Backend\": { \"Type\": \"vxlan\" } }"
$ netstat -lnpt|grep etcd
tcp        0      0 10.64.3.7:2379          0.0.0.0:*               LISTEN      42452/etcd
tcp        0      0 127.0.0.1:2380          0.0.0.0:*               LISTEN      42452/etcd
tcp        0      0 127.0.0.1:7001          0.0.0.0:*               LISTEN      42452/etcd
```

etcd是kubernetes**唯一有状态的服务**；缺省情况下kubernetes对象保存在`/registry`目录下，可以通过apiserver的`--etcd-prefix`参数进行配置；

apiserver是**唯一**连接etcd的kubernetes组件，其它组件都需要通过apiserver的接口来获取集群状态；


# 起apiserver

## 修改配置/etc/kubernetes/apiserver:

``` bash
$ cat apiserver |grep -v '^#'|grep -v '^$'
KUBE_API_ADDRESS="--insecure-bind-address=10.64.3.7"
KUBE_ETCD_SERVERS="--etcd-servers=http://10.64.3.7:2379"
KUBE_SERVICE_ADDRESSES="--service-cluster-ip-range=10.254.0.0/16"  # **指定 service 的 cluster 子网网段**
KUBE_ADMISSION_CONTROL="--admission-control=NamespaceLifecycle,NamespaceExists,LimitRanger,SecurityContextDeny,ServiceAccount,ResourceQuota"
KUBE_API_ARGS="--bind-address=10.64.3.7 --service_account_key_file=/etc/kubernetes/ssl/server.key --client-ca-file=/etc/kubernetes/ssl/ca.crt --tls-cert-file=/etc/kubernetes/ssl/server.cert --tls-private-key-file=/etc/kubernetes/ssl/server.key"
```

注意：

1. 为了以后Pod容器能访问apiserver，`KUBE_ADMISSION_CONTROL`里面必须要包含`ServiceAccount`；
1. 需要指定`--service_account_key_file`参数，值为签名`ServiceAccount`的私钥，且必须与`kube-controller-manager`的`--service_account_private_key_file`参数指向同一个文件；如果未指定该参数，创建Pod会失败：
    ``` bash
    $  kubectl create -f Pod-nginx.yaml
    Error from server: error when creating "Pod-nginx.yaml": Pods "nginx" is forbidden: no API token found for service account default/default, retry after the token is automatically created and added to the service account
    ```
1. `--bind-address`**不能设置为127.0.0.1**, 否则Pod访问apiserver对应的kubernetes service cluster ip会被refuse，因为127.0.0.1是Pod所在的容器而非apiserver；
1. 如果没有指定`--tls-cert-file`和`--tls-private-key-file`，apiserver启动时会自动创建公私钥，但不会创建签名它们的ca证书和ca的key，所以**client不能对apiserver的证书进行验证**，保存在`/var/run/kubernetes/`目录；
1. 在 `--bind-address 6443`监听安全端口；
1. 在 `--insecure-bind-address 8080` 监听非安全端口，apisrver对访问该端口的请求不做认证、授权，**后面的kubectl配置文件~/.kube/conf使用的就是这个地址和端口**；

## 修改配置/usr/lib/systemd/system/kube-apiserver.service

添加kube账户对`/var/run/kubernetes`目录的读写权限

``` bash
[Service]
PermissionsStartOnly=true
ExecStartPre=-/usr/bin/mkdir /var/run/kubernetes
ExecStartPre=/usr/bin/chown -R kube:kube /var/run/kubernetes/
```

修改systemd unit文件后，需要reload生效：

``` bash
$ systemctl daemon-reload
$
```

## 重启进程

``` bash
$ systemctl restart kube-apiserver
$ ps -e -o args|grep apiserver
/usr/bin/kube-apiserver --logtostderr=true --v=0 --etcd-servers=http://127.0.0.1:2379 --insecure-bind-address=127.0.0.1 --allow-privileged=false --service-cluster-ip-range=10.254.0.0/16 --admission-control=NamespaceLifecycle,NamespaceExists,LimitRanger,SecurityContextDeny,ServiceAccount,ResourceQuota --bind-address=10.64.3.7 --service_account_key_file=/var/run/kubernetes/serviceAccount.key
```

## 查看日志


``` bash
$ journalctl -u kube-apiserver -o export|grep MESSAGE
$
```

## 创建kubelet配置文件

``` bash
$ ls ~/.kube/
$ kubectl config set-cluster default-cluster --server=http://10.64.3.7:8080
cluster "default-cluster" set.
$ kubectl config set-context default-context --cluster=default-cluster --user=default-admin
context "default-context" set.
$ kubectl config use-context default-context
switched to context "default-context".
$ cat ~/.kube/config
apiVersion: v1
clusters:
- cluster:
    server: http://10.64.3.7:8080
  name: default-cluster
contexts:
- context:
    cluster: default-cluster
    user: default-admin
  name: default-context
current-context: default-context
kind: Config
preferences: {}
users: []
```

注意：

1. 用node ip指定apiserver；
1. 访问的是apiserver的不需要认证和授权的**非安全端口**；

# 起 kube-controller-manager

## 修改配置/etc/kubernetes/controller

``` bash
$ cat controller-manager |grep -v '^#'|grep -v '^$'
KUBE_CONTROLLER_MANAGER_ARGS="--address=127.0.0.1 --service_account_private_key_file=/etc/kubernetes/ssl/server.key --root-ca-file=/etc/kubernetes/ssl/ca.crt --cluster-cidr 10.254.0.0/16"
```

注意：

1. `--service_account_private_key_file`需要与apiserver的`service_account_key_file`参数一致；
1. `--root-ca-file` 用于对apiserver提供的证书进行校验，**指定该参数后才在各Pod的ServiceAccount挂载目录中包含ca文件**；
1. `--cluster-cidr` 指定与apiserver配置一致的 service cluster ip网段 ，kubelet会读取该参数设置`configureCBR0`，如果未指定，kubelet启动时提示(hairpin的含义见后文kublet一节)：
    ``` bash
    $ journalctl -u kubelet |tail -1000|grep 'Hai'
    Mar 29 01:43:09 tjwq01-sys-bs003007.tjwq01.ksyun.com kubelet[3766]: W0329 01:43:09.473594    3766 kubelet.go:584] Hairpin mode set to "promiscuous-bridge" but configureCBR0 is false, falling back to "hairpin-veth"
    ```

## 重启进程

``` bash
$ systemctl restart kube-controller-manager
$ ps -e -o ppid,pid,args -H |grep kube-con
1 30898   /root/local/bin/kube-controller-manager --logtostderr=true --v=0 --master=http://10.64.3.7:8080 --address=127.0.0.1 --service_account_private_key_file=/etc/kubernetes/ssl/server.key --root-ca-file=/etc/kubernetes/ssl/ca.crt --cluster-cidr 10.254.0.0/16
$ netstat -lnpt|grep kube-controll
tcp        0      0 127.0.0.1:10252         0.0.0.0:*               LISTEN      28636/kube-controll
```

## 查看日志

``` bash
$ journalctl -u kube-controller-manager -o export|grep MESSAGE
$
```

# 起kube-scheduler：

## 修改配置/etc/kubernetes/scheduler

``` bash
$ grep -v '^#' scheduler |grep -v '^$'
KUBE_SCHEDULER_ARGS="--address=127.0.0.1"
```

## 重启进程

``` bash
$ systemctl start kube-scheduler
$ ps -e -o ppid,pid,args -H |grep kube-sch
1 42555   /root/local/bin/kube-scheduler --logtostderr=true --v=0 --master=http://10.64.3.7:8080 --address=127.0.0.1
$ netstat  -lnpt|grep kube-schedule
tcp        0      0 127.0.0.1:10251         0.0.0.0:*               LISTEN      42555/kube-schedule
```

## 查看日志

``` bash
$ journalctl -u kube-scheduler -o export|grep MESSAGE
$ journalctl -u kube-scheduler -o export|grep MESSAGE
MESSAGE_ID=39f53479d3a045ac8e11786248231fbf
MESSAGE=Started Kubernetes Scheduler Plugin.
MESSAGE_ID=7d4958e842da4a758f6c1cdc7b36dcc5
MESSAGE=Starting Kubernetes Scheduler Plugin...
MESSAGE=I0312 12:21:27.192438   46584 leaderelection.go:188] sucessfully acquired lease kube-system/kube-scheduler
MESSAGE=I0312 12:21:27.192534   46584 event.go:217] Event(api.ObjectReference{Kind:"Endpoints", Namespace:"kube-system", Name:"kube-scheduler", UID:"5c296e1f-06db-11e7-8472-8cdcd4b3be48", APIVersion:"v1", ResourceVersion:"682325", FieldPath:""}): type: 'Normal' reason: 'LeaderElection' tjwq01-sys-bs003007.tjwq01.ksyun.com became leader
```

# 起flanneld

## 配置flanneld的配置文件/etc/sysconfig/flanneld

``` bash
$ cat /etc/sysconfig/flanneld|grep -v '^#'
FLANNEL_ETCD_ENDPOINTS="http://10.64.3.7:2379"
FLANNEL_ETCD_PREFIX="/kube-centos/network"
FLANNEL_OPTIONS="-iface=eth4"  # 可选，最好指定为内网接口，否则对于默认路由为外网的IDC机器，flanneld使用公网的接口通信；
```

## 重启服务

``` bash
$ systemctl restart flanneld
$ ps -e -o args|grep flann
/usr/bin/flanneld -etcd-endpoints=http://127.0.0.1:2379 -etcd-prefix=/kube-centos/network
$ ip a|grep -w inet
    inet 127.0.0.1/8 scope host lo
    inet 10.64.3.7/26 brd 10.64.3.63 scope global eth0
    inet 120.92.8.114/28 brd 120.92.8.127 scope global eth2
    inet 172.30.19.0/16 scope global flannel.1  // 可见配置了分配的/24网段的地址
$ ifconfig flannel.1
flannel.1: flags=4163<UP,BROADCAST,RUNNING,MULTICAST>  mtu 1450
        inet 172.30.19.0  netmask 255.255.0.0  broadcast 0.0.0.0
        inet6 fe80::b44d:4ff:fe7b:6d56  prefixlen 64  scopeid 0x20<link>
        ether b6:4d:04:7b:6d:56  txqueuelen 0  (Ethernet)
        RX packets 0  bytes 0 (0.0 B)
        RX errors 0  dropped 0  overruns 0  frame 0
        TX packets 0  bytes 0 (0.0 B)
        TX errors 0  dropped 8 overruns 0  carrier 0  collisions 0
$ ip route show
default via 120.92.8.126 dev eth2
10.0.0.0/8 via 10.64.3.62 dev eth0
10.64.3.0/26 dev eth0  proto kernel  scope link  src 10.64.3.7
120.92.8.112/28 dev eth2  proto kernel  scope link  src 120.92.8.114
169.254.0.0/16 dev eth0  scope link  metric 1002
169.254.0.0/16 dev eth2  scope link  metric 1004
172.16.0.0/12 via 10.64.3.62 dev eth0
172.30.0.0/16 dev flannel.1
172.30.60.0/24 dev docker0  proto kernel  scope link  src 172.30.60.1
```

注意：

1. **v0.7.0**以前的flanneld版本**有bug**：创建的`flannel.1`接口子网掩码是16(正确的是32)，可能导致和其它主机通信时会被当做**网络地址**而失败；
1. 访问本Node所分配的Pod明细网段(24位掩码)的请求，被kernele**转发**到docker0网桥；
1. 访问其它Pod网段的请求，被flanneld通过flannel.1接口转发到其它Node的flanneld进程；


## flannel和docker0 IP的关系

安装flannel时，自动给docker service指定一个flannel创建的环境变量文件：

``` bash
$ cat /usr/lib/systemd/system/docker.service.d/flannel.conf
[Service]
EnvironmentFile=-/run/flannel/docker
```

每次起flanneld服务时，`mk-docker-opts.sh`脚本将docker网络相关的环境变量写入`/run/flannel/docker` 文件：

``` bash
$ cat /usr/lib/systemd/system/flanneld.service |grep mk-docker-opts.sh
ExecStartPost=/usr/libexec/flannel/mk-docker-opts.sh -k DOCKER_NETWORK_OPTIONS -d /run/flannel/docker

$ cat /run/flannel/docker
DOCKER_OPT_BIP="--bip=172.30.72.1/24"
DOCKER_OPT_IPMASQ="--ip-masq=true"
DOCKER_OPT_MTU="--mtu=1450"
DOCKER_NETWORK_OPTIONS=" --bip=172.30.72.1/24 --ip-masq=true --mtu=1450 "
```

这样，**后续首次**起dockerd进程时，它会读取该文件中的环境变量，配置docker0网桥的IP地址为bip。

## docker0先于flanneld首次启动导致docker0 bip错误分配的问题

docker0启动后，自动创建的`docker0`网桥使用的是 `192.168.x.x` 网段，后续即使修改为上面的bip网段地址，重启后还是恢复到192.168。

解决方法是：
在 docker.service 的 dockerd 命令后面添加 `$DOCKER_NETWORK_OPTIONS`，这样相当于在dockerd命令行上指定了`--bip`参数：

``` bash
$ grep dockerd /usr/lib/systemd/system/docker.service
ExecStart=/usr/bin/dockerd $DOCKER_NETWORK_OPTIONS
$ systemctl daemon-reload
$ systemctl restart docker0
```

## 查看Pod网段详情

``` bash
$ etcdctl --endpoints 10.64.3.7:2379 get  /kube-centos/network/config
{ "Network": "172.30.0.0/16", "SubnetLen": 24, "Backend": { "Type": "vxlan" } }
```

## 查看已分配的Pod网段

``` bash
$ etcdctl --endpoints 10.64.3.7:2379 ls  /kube-centos/network/subnets
/kube-centos/network/subnets/172.30.19.0-24
/kube-centos/network/subnets/172.30.83.0-24
/kube-centos/network/subnets/172.30.60.0-24
```

## 删除已分配的Pod网段(适合重启flanneld的情形)

``` bash
$ etcdctl --endpoints 10.64.3.7:2379 rmdir  /kube-centos/network/subnets/172.30.38.0-24
$
```

# 起docker

dockerd启动时会读取`flanneld`启动时创建的环境变量文件`/run/flannel/docker`，然后设置网络参数；
为了加快pull image的速度，可以使用`registry-mirror`，增加`max-concurrent-downloads`的值(默认为3)；

``` bash
$ cat /etc/docker/daemon.json
{
  "registry-mirrors": ["https://docker.mirrors.ustc.edu.cn", "hub-mirror.c.163.com"],
  "max-concurrent-downloads": 10
}
$ systemctl start docker
```

注意：

1. docker 1.13版本后，`ip-forward`参数默认为`ture`(自动配置`net.ipv4.ip_forward = 1`)，dockerd**将iptables FORWARD的默认策略设置为DROP**，从而导致Node ping其它Node的Pod IP失败；

  原因分析：
        1. kube-proxy只会创建ClusterIP和nodePort的DNAT规则(也可能创建SNAT规则，参考[kue-proxy和iptables.md](kue-proxy和iptables.md)), 一般不创建FORWARD规则；参考[docker创建的iptables.md](../docker/docker创建的iptables.md)
        1. dockerd一般只会为nodePort映射的的容器创建FORWARD规则；
        1. pod IP不属于第一种情况，一般也不属于第二种情况，所以会被docker创建的FORWARD默认policy drop；

  解决办法：
        1. `$ sudo iptables -P FORWARD ACCEPT`;
        1. 或者在 /etc/docker/daemon.json中设置`"ip-forward": false`;
        https://github.com/docker/docker/pull/28257
        https://github.com/kubernetes/kubernetes/issues/40182

1. dockerd默认启用`--iptabes`、`--ip-masq`参数，用于创建**hostPort端口映射的iptables NAT规则**(将访问host:port的请求DNAT到container:port)，如果需要运行有**hostPort端口映射的容器**(如手动docker run -p80:80或pod的spec里面指定hostPort)，则**不能关闭**这两个选项，否则访问会被Refuse。其它情况下，kube-proxy会生成对访问cluster ip的NAT规则；
        原因分析：
        1. docker不再生成访问hostPort的DNAT规则，目的地址hostIP:hostPort的包转发到容器后，容器会拒绝接收；；

1. 如果指定的私有registry需要登录验证(HTTPS证书、basic账号密码)，则需要放置CA证书和生成认证信息：
  1. 将registry的CA证书放置到 `/etc/docker/certs.d/{registryIP:Port}}/ca.crt`；
  1. 执行 `docker login` 命令，docker自动将认证信息保存到`~/.docker/config.json`：
  ``` bash
        $ cat ~/.docker/config.json
        {
                "auths": {
                        "10.64.3.7:8000": {
                                "auth": "Zm9vMjpmb28y"
                        }
                }
        }
  ```

# 起 kube-proxy

## 修改配置文件/etc/kubernetes/proxy

``` bash
$ grep -v '^#' proxy |grep -v '^$'
KUBE_PROXY_ARGS="--bind-address=10.64.3.7 --cluster-cidr=10.254.0.0/16"
```

注意：

1. `--proxy-mode`默认值为iptables，故使用iptables而非userspace机制，kube-proxy创建的iptables分析见文档 [kube-proxy和iptables.md](kube-proxy和iptables.md)
1. kube-proxy使用`--cluster-cidr`来判断集群内部和外部流量；
1. 必须指定`--cluster-cidr`或`--masquerade-all=true`参数，kube-proxy才会创建访问Cluster IP的SNAT规则，否则Node访问Service Cluster IP可能会失败(分析如下)；

https://github.com/kubernetes/kubernetes/blob/master/pkg/proxy/iptables/proxier.go#L944
https://github.com/kubernetes/kubernetes/blob/master/pkg/proxy/iptables/proxier.go#L992

## 未指定--cluster-cidr或--masquerade-all=true 参数引起Node访问Cluster IP失败的情况分析

系统环境：

Node A：

``` bash
$ ifconfig |grep flags -A1
docker0: flags=4163<UP,BROADCAST,RUNNING,MULTICAST>  mtu 1450
        inet 172.30.83.1  netmask 255.255.255.0  broadcast 0.0.0.0
eth4: flags=4163<UP,BROADCAST,RUNNING,MULTICAST>  mtu 1500
        inet 10.64.3.8  netmask 255.255.255.192  broadcast 10.64.3.63
eth5: flags=4163<UP,BROADCAST,RUNNING,MULTICAST>  mtu 1500
        inet 120.92.8.115  netmask 255.255.255.240  broadcast 120.92.8.127
flannel.1: flags=4163<UP,BROADCAST,RUNNING,MULTICAST>  mtu 1450
        inet 172.30.83.0  netmask 255.255.255.255  broadcast 0.0.0.0
$ ip route show
default via 120.92.8.126 dev eth5  proto static  metric 100
10.0.0.0/8 via 10.64.3.62 dev eth4  proto static  metric 100
10.32.0.0/16 via 10.64.3.62 dev eth4
10.64.3.0/26 dev eth4  proto kernel  scope link  src 10.64.3.8  metric 100
120.92.8.112/28 dev eth5  proto kernel  scope link  src 120.92.8.115  metric 100
172.16.0.0/12 via 10.64.3.62 dev eth4  proto static  metric 100
172.30.0.0/16 dev flannel.1
172.30.83.0/24 dev docker0  proto kernel  scope link  src 172.30.83.1

$ ip route get  10.254.23.166
10.254.23.166 via 10.64.3.62 dev eth4  src 10.64.3.8
```

Node B:

``` bash
$ ifconfig |grep flag -A1|less
docker0: flags=4163<UP,BROADCAST,RUNNING,MULTICAST>  mtu 1450
        inet 172.30.60.1  netmask 255.255.255.0  broadcast 0.0.0.0
eth0: flags=4163<UP,BROADCAST,RUNNING,MULTICAST>  mtu 1500
        inet 10.64.3.7  netmask 255.255.255.192  broadcast 10.64.3.63
eth2: flags=4163<UP,BROADCAST,RUNNING,MULTICAST>  mtu 1500
        inet 120.92.8.114  netmask 255.255.255.240  broadcast 120.92.8.127
flannel.1: flags=4163<UP,BROADCAST,RUNNING,MULTICAST>  mtu 1450
        inet 172.30.60.0  netmask 255.255.255.255  broadcast 0.0.0.0
$ ip route show
default via 120.92.8.126 dev eth2
10.0.0.0/8 via 10.64.3.62 dev eth0
10.64.3.0/26 dev eth0  proto kernel  scope link  src 10.64.3.7
120.92.8.112/28 dev eth2  proto kernel  scope link  src 120.92.8.114
169.254.0.0/16 dev eth0  scope link  metric 1002
169.254.0.0/16 dev eth2  scope link  metric 1004
172.16.0.0/12 via 10.64.3.62 dev eth0
172.30.0.0/16 dev flannel.1
172.30.60.0/24 dev docker0  proto kernel  scope link  src 172.30.60.1

```

现在分析从Node A访问Cluster IP 10.254.23.166，假设该IP绑定的是Node B上的Pod IP的路径；

1. Node A访问Cluster IP时:
    原始包：目标地址ClusterIP 源地址10.64.3.8
    目标地址会被DNAT到Node B上Pod IP：172.30.60.6，按照路由表包被转发到flannel.1接口：目的地址172.30.60.6 源地址10.64.3.8；
1. Node B的flanneld在它的flannel.1接口上收到该包后，按照路由表，将包转发到docker0, 进而转给Pod中的容器处理，容器处理后发送的包目的地址10.64.3.8 源地址172.30.60.6；由于目的地址10.64.3.8为Node B IP，所以 Node B直接通过eth0接口将包发给Node A的eth4接口；
1. 对于Node A而言收发包接口不一致(发flannel.1，收eth4)，**linux的rp_filter机制会拒绝这种情况**；
   rp_fliter值默认为1即strict模式，内核收到包后会做下反向的路由查询，看查询出的接口是否和接收到的接口一致，如果不一致则丢弃；
   http://blog.clanzx.net/2013/08/22/rp_filter.html

### 解决方法：

1. 在 Node A和B 上分别设置：`sysctl -w net.ipv4.conf.all.rp_filter=2`;
1. 或者为kube-proxy指定`--cluster-cidr`参数，这样它为Node创建如下的SNAT规则，该规则只匹配**源IP非Cluster IP的情况**：
  `-A KUBE-SERVICES ! -s 10.254.0.0/16 -d 10.254.0.1/32 -p tcp -m comment --comment "default/kubernetes:https cluster IP" -m tcp --dport 443 -j KUBE-MARK-MASQ`
1. 或者为kube-proxy指定`--masquerade-all=true`参数，这样它为Node创建如下的SNAT规则，注意该规则**不区分源IP**，所以是masquerade-all：
  `-A KUBE-SERVICES -s 0.0.0.0/0 -d 10.254.0.1/32 -p tcp -m comment --comment "default/kubernetes:https cluster IP" -m tcp --dport 443 -j KUBE-MARK-MASQ`
1. 一旦指定了`--cluster-cidr`或`--masquerade-all=true`参数，由于会对访问ClusterIP的请求做SNAT(同时也会对NodePort做SNAT)，故**Pod容器进程将看不到客户请求的真实IP**(访问Port IP不受影响)；

https://github.com/kubernetes/kubernetes/issues/24224
https://github.com/kubernetes/kubernetes/pull/24429

kube-proxy创建的iptables参考[kube-proxy创建的iptabes.md](kube-proxy创建的iptabes.md)。

docker创建的iptables参考[docker/docker创建的iptables.md](../docker/docker创建的iptables.md)。

## 重启进程

``` bash
$ systemctl start kube-proxy
$ ps -e -o ppid,pid,args -H |grep kube-proxy
  1 37303   /root/local/bin/kube-proxy --logtostderr=true --v=0 --master=http://10.64.3.7:8080 --bind-address=10.64.3.7 --cluster-cidr=10.254.0.0/16
$ netstat -lnpt|grep kube-proxy
tcp        0      0 127.0.0.1:10249         0.0.0.0:*               LISTEN      37303/kube-proxy
```

## 查看日志

``` bash
$ journalctl -u kube-proxy -o export|grep MESSAGE
MESSAGE_ID=39f53479d3a045ac8e11786248231fbf
MESSAGE=Started Kubernetes Kube-Proxy Server.
MESSAGE_ID=7d4958e842da4a758f6c1cdc7b36dcc5
MESSAGE=Starting Kubernetes Kube-Proxy Server...
MESSAGE=E0312 12:22:32.783492   46662 server.go:421] Can't get Node "tjwq01-sys-bs003007.tjwq01.ksyun.com", assuming iptables proxy, err: nodes "tjwq01-sys-bs003007.tjwq01.ksyun.com" not found
MESSAGE=I0312 12:22:32.785017   46662 server.go:215] Using iptables Proxier.
MESSAGE=W0312 12:22:32.786431   46662 server.go:468] Failed to retrieve node info: nodes "tjwq01-sys-bs003007.tjwq01.ksyun.com" not found
MESSAGE=W0312 12:22:32.786512   46662 proxier.go:249] invalid nodeIP, initialize kube-proxy with 127.0.0.1 as nodeIP
MESSAGE=W0312 12:22:32.786521   46662 proxier.go:254] clusterCIDR not specified, unable to distinguish between internal and external traffic
MESSAGE=I0312 12:22:32.786541   46662 server.go:227] Tearing down userspace rules.
MESSAGE=I0312 12:22:32.799072   46662 conntrack.go:81] Set sysctl 'net/netfilter/nf_conntrack_max' to 786432
MESSAGE=I0312 12:22:32.799474   46662 conntrack.go:66] Setting conntrack hashsize to 196608
MESSAGE=I0312 12:22:32.800668   46662 conntrack.go:81] Set sysctl 'net/netfilter/nf_conntrack_tcp_timeout_established' to 86400
MESSAGE=I0312 12:22:32.800694   46662 conntrack.go:81] Set sysctl 'net/netfilter/nf_conntrack_tcp_timeout_close_wait' to 3600
```

注意：

1. 如果service较多，kube-proxy会添加很多iptables规则，所以kub-proxy启动时，增加了部分net_conntrace_*的值；


# 起kubelet

## 修改配置文件

``` bash
$ grep -v '^#' kubelet |grep -v '^$'
KUBELET_ADDRESS="--address=10.64.3.7"
KUBELET_HOSTNAME="--hostname-override=10.64.3.7"
KUBELET_API_SERVER="--api-servers=http://10.64.3.7:8080"
KUBELET_Pod_INFRA_CONTAINER="--Pod-infra-container-image=registry.access.redhat.com/rhel7/Pod-infrastructure:latest"
KUBELET_ARGS="--tls-cert-file=/etc/kubernetes/ssl/kubecfg.crt --tls-private-key-file=/etc/kubernetes/ssl/kubecfg.key --cluster_dns=10.254.0.2 --cluster_domain=cluster.local"
```

注意：

1. `--Pod-infra-container-image` 默认为`gcr.io`的image，由于该网址**被墙**，所以需要指定其他的infra-container;
1. `--tls-cert-file`、`--tls-private-key-file` 参数指定和apiserver通信的公私钥，apiserver使用它的`--client-ca-file`做验证；
1. 如果启用了kubeDNS addons，则需要**同时**指定`--cluster_dns=<kubedns cluster ip>` `--cluster_domain=cluster.local`；
1. 没有对apiserver的key做验证；
1. `--hairpin-mode`默认值为promiscuous-bridge，指定kubelet如何设置hairpin NAT规则，适用于**Pod访问和自身绑定的service情况**。

linux bridge默认会drop**从一个口output又input的Package**；

解决方法：

1. 将连接虚拟bridge的各veth设置hairpin参数：`for intf in $(ip link list | grep veth | cut -f2 -d:); do brctl hairpin cbr0 $intf on; done` ;
1. 或者将各虚拟veth接口设置hairpin参数： `for intf in /sys/devices/virtual/net/cbr0/brif/*; do echo 1 > $intf/hairpin_mode; done` 
1. 或者将**bridge设置为promiscuousm模式**，如 `ip link set cbr0 promisc on`；
  https://github.com/kubernetes/kubernetes/issues/13375

## 重启进程

``` bash
$ systemctl start kubelet
$ ps -e -o ppid,pid,args -H |grep kubelet
1 34842   /root/local/bin/kubelet --logtostderr=true --v=0 --api-servers=http://10.64.3.7:8080 --address=10.64.3.7 --hostname-override=10.64.3.7 --allow-privileged=false --Pod-infra-container-image=registry.access.redhat.com/rhel7/Pod-infrastructure:latest --tls-cert-file=/etc/kubernetes/ssl/kubecfg.crt --tls-private-key-file=/etc/kubernetes/ssl/kubecfg.key --cluster_dns=10.254.0.2 --cluster_domain=cluster.local --hairpin-mode promiscuous-bridge
```

## 查看日志

``` bash
$ journalctl -u kubelet -o export|grep MESSAGE
MESSAGE_ID=39f53479d3a045ac8e11786248231fbf
MESSAGE=Started Kubernetes Kubelet Server.
MESSAGE_ID=7d4958e842da4a758f6c1cdc7b36dcc5
MESSAGE=Starting Kubernetes Kubelet Server...
MESSAGE=I0312 11:31:03.177438   29352 docker.go:330] Start docker client with request timeout=2m0s
MESSAGE=W0312 11:31:03.184112   29352 server.go:488] Could not load kubeconfig file /var/lib/kubelet/kubeconfig: stat /var/lib/kubelet/kubeconfig: no such fil
e or directory. Trying auth path instead.
```

如果日志中包含：

    Mar 29 05:28:13 tjwq01-sys-bs003007.tjwq01.ksyun.com kubelet[31983]: I0329 05:28:13.048187   31983 kubelet.go:477] Hairpin mode set to "hairpin-veth"
    Mar 29 05:33:23 tjwq01-sys-bs003007.tjwq01.ksyun.com kubelet[34842]: W0329 05:33:23.230452   34842 kubelet_network.go:69] Hairpin mode set to "promiscuous-bridge" but kubenet is not enabled, falling back to "hairpin-veth"
    Mar 29 05:33:23 tjwq01-sys-bs003007.tjwq01.ksyun.com kubelet[34842]: I0329 05:33:23.230484   34842 kubelet.go:477] Hairpin mode set to "hairpin-veth"

这是由于当前使用的是 flanneld而非kubenet！

``` bash
$ journalctl -u kube-controller-manager -o export|grep MESSAGE // 在12:41时，新增如下内容
MESSAGE=I1216 12:28:56.902327   10829 controller.go:211] Starting Daemon Sets controller manager
MESSAGE=I1216 12:28:56.902439   10829 controllermanager.go:320] Starting ReplicaSet controller
MESSAGE=W1216 12:41:12.210177   10829 nodecontroller.go:671] Missing timestamp for Node 10.64.3.7 . Assuming now as a timestamp.
MESSAGE=I1216 12:41:12.210367   10829 event.go:211] Event(api.ObjectReference{Kind:"Node", Namespace:"", Name:"10.64.3.7", UID:"10.64.3.7", APIVersion:"", ResourceVersion:"", FieldPath:""}): type: 'Normal' reason: 'RegisteredNode' Node 10.64.3.7  event: Registered Node 10.64.3.7  in NodeController
```

kubelet自动向apiserver注册，故可以看到node信息：

``` bash
$ kubectl get nodes
NAME        STATUS    AGE
10.64.3.7   Ready     6m
```