<!-- toc -->

## 配置flanneld的配置文件/etc/sysconfig/flanneld

``` bash
$ cat /etc/sysconfig/flanneld|grep -v '^#'
FLANNEL_ETCD_ENDPOINTS="http://10.64.3.7:2379"
FLANNEL_ETCD_PREFIX="/kube-centos/network"
FLANNEL_OPTIONS="-iface=eth0"  # 可选，最好指定为内网接口，否则对于默认路由为外网的IDC机器，flanneld使用公网的接口通信；
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
[root@tjwq01-sys-bs003007 ~]# ip -d link show flannel.1  # vlxlan类型网络设备，和-iface关联；
472: flannel.1: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1450 qdisc noqueue state UNKNOWN mode DEFAULT
    link/ether 86:4b:9c:65:f2:12 brd ff:ff:ff:ff:ff:ff promiscuity 0
    vxlan id 1 local 10.64.3.7 dev eth0 srcport 0 0 dstport 8472 nolearning ageing 300 addrgenmode eui64
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
1. flanneld启动后从etcd中读取配置，并请求获取一个subnet lease(租约)，有效期是24hrs，并且监视etcd的数据更新；
1. 配置信息写入/run/flannel/subnet.env文件；
1. 各个node上的网络设备列表新增一个名为flannel.1的类型为**vxlan**的网络设备，和`-face`关联；

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

## 查看某一网段的具体配置

``` bash
$ etcdctl --endpoints 10.64.3.7:2379 ls  /kube-centos/network/subnets/172.30.19.0-24
{"PublicIP":"10.64.3.7","BackendType":"vxlan","BackendData":{"VtepMAC":"d6:51:2e:80:5c:69"}}
```


## 删除已分配的Pod网段(适合重启flanneld的情形)

``` bash
$ etcdctl --endpoints 10.64.3.7:2379 rmdir  /kube-centos/network/subnets/172.30.38.0-24
$
```