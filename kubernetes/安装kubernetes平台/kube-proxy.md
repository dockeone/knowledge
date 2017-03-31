<!-- toc -->

## 修改配置文件/etc/kubernetes/proxy

``` bash
$ grep -v '^#' proxy |grep -v '^$'
KUBE_PROXY_ARGS="--bind-address=10.64.3.7 --cluster-cidr=10.254.0.0/16"
```

注意：

1. `--proxy-mode`默认值为iptables，故使用iptables而非userspace机制，kube-proxy创建的iptables分析见文档 [kube-proxy和iptables.md](kube-proxy和iptables.md)
1. kube-proxy使用`--cluster-cidr`来判断集群内部和外部流量；
1. 必须指定`--cluster-cidr`或`--masquerade-all=true`参数，kube-proxy才会创建访问Cluster IP的SNAT规则，否则Node访问Service Cluster IP可能会失败(分析如下)；
1. dockerd也会创建一些必须的iptables规则，不能关闭dockerd的--iptables选项；
1. **Service Cluster IP是虚拟IP，只存在于kube-proxy创建的iptables规则中**；

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
