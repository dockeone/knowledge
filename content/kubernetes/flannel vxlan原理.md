# flannel vxlan原理

如上图所示，我们来看看从pod1：172.16.99.8发出的数据包是如何到达pod3：172.16.57.15的（比如：在pod1的某个container中ping -c 3 172.16.57.15）。

a) 从Pod出发

由于k8s更改了docker的DOCKER_OPTS，显式指定了–bip，这个值与分配给该node上的subnet的范围是一致的。这样一来，docker引擎每次创建一个Docker container，该container被分配到的ip都在flannel subnet范围内。

当我们在Pod1下的某个容器内执行ping -c 3 172.16.57.15，数据包便开始了它在flannel network中的旅程。

Pod是Kubernetes调度的基本unit。Pod内的多个container共享一个network namespace。kubernetes在创建Pod时，首先先创建pause容器，然后再以pause的network namespace为基础，创建pod内的其他容器（–net=container:xxx），这样Pod内的所有容器便共享一个network namespace，这些容器间的访问直接通过localhost即可。比如Pod下A容器启动了一个服务，监听8080端口，那么同一个Pod下面的另外一个B容器通过访问localhost:8080即可访问到A容器下面的那个服务。

我们看一下Pod1中某Container内的路由信息：

$ docker exec ba75f81455c7 ip route
default via 172.16.99.1 dev eth0
172.16.99.0/24 dev eth0  proto kernel  scope link  src 172.16.99.8

目的地址172.16.57.15并不在直连网络中，因此数据包通过default路由出去。default路由的路由器地址是172.16.99.1，也就是上面的docker0 bridge的IP地址。相当于docker0 bridge以“三层的工作模式”直接接收到来自容器的数据包(而并非从bridge的二层端口接收)。

b) docker0与flannel.1之间的包转发

数据包到达docker0后，docker0的内核栈处理程序发现这个数据包的目的地址是172.16.57.15，并不是真的要送给自己，于是开始为该数据包找下一hop。根据master node上的路由表：

master node：

$ ip route
... ...
172.16.0.0/16 dev flannel.1  proto kernel  scope link  src 172.16.99.0
172.16.99.0/24 dev docker0  proto kernel  scope link  src 172.16.99.1
... ...

我们匹配到“172.16.0.0/16”这条路由！这是一条直连路由，数据包被直接送到flannel.1设备上。

c) flannel.1设备以及flanneld的功用

flannel.1是否会重复docker0的套路呢：包不是发给自己，转发数据包？会，也不会。

“会”是指flannel.1肯定要将包转发出去，因为毕竟包不是给自己的（包目的ip是172.16.57.15, vxlan设备ip是172.16.99.0）。
“不会”是指flannel.1不会走寻常套路去转发包，因为它是一个vxlan类型的设备，也称为vtep，virtual tunnel end point。

那么它到底是怎么处理数据包的呢？这里涉及一些Linux内核对vxlan处理的内容，详细内容可参见本文末尾的参考资料。

flannel.1收到数据包后，由于自己不是目的地，也要尝试将数据包重新发送出去。数据包沿着网络协议栈向下流动，在二层时需要封二层以太包，填写目的mac地址，这时一般应该发出arp：”who is 172.16.57.15″。但vxlan设备的特殊性就在于它并没有真正在二层发出这个arp包，因为下面的这个内核参数设置：

master node:

$ cat /proc/sys/net/ipv4/neigh/flannel.1/app_solicit
3

而是由linux kernel引发一个”L3 MISS”事件并将arp请求发到用户空间的flanned程序。

flanned程序收到”L3 MISS”内核事件以及arp请求(who is 172.16.57.15)后，并不会向外网发送arp request，而是尝试**从etcd查找该地址匹配的子网的vtep信息**。在前面章节我们曾经展示过etcd中Flannel network的配置信息：

master node:

$ etcdctl --endpoints http://127.0.0.1:{etcd listen port} ls  /coreos.com/network/subnets
/coreos.com/network/subnets/172.16.99.0-24
/coreos.com/network/subnets/172.16.57.0-24

$ curl -L http://127.0.0.1:{etcd listen port}/v2/keys/coreos.com/network/subnets/172.16.57.0-24
{"action":"get","node":{"key":"/coreos.com/network/subnets/172.16.57.0-24","value":"{\"PublicIP\":\"{minion node local ip}\",\"BackendType\":\"vxlan\",\"BackendData\":{\"VtepMAC\":\"d6:51:2e:80:5c:69\"}}","expiration":"2017-01-17T09:46:20.607339725Z","ttl":21496,"modifiedIndex":2275460,"createdIndex":2275460}}

flanneld从etcd中找到了答案：

subnet: 172.16.57.0/24
public ip: {minion node local ip}
VtepMAC: d6:51:2e:80:5c:69

我们查看minion node上的信息，发现minion node上的flannel.1 设备mac就是d6:51:2e:80:5c:69：

minion node:

$ ip -d link show

349: flannel.1: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1450 qdisc noqueue state UNKNOWN mode DEFAULT group default
    link/ether d6:51:2e:80:5c:69 brd ff:ff:ff:ff:ff:ff promiscuity 0
    vxlan id 1 local 10.46.181.146 dev eth0 port 0 0 nolearning ageing 300

接下来，flanned将查询到的信息放入master node host的arp cache表中：

master node:

$ ip n |grep 172.16.57.15
172.16.57.15 dev flannel.1 lladdr d6:51:2e:80:5c:69 REACHABLE

flanneld完成这项工作后，linux kernel就可以在arp table中找到 172.16.57.15对应的mac地址并封装二层以太包了。

不过这个封包还不能在物理网络上传输，因为它实际上只是vxlan tunnel上的packet。

d) kernel的vxlan封包

我们需要将上述的packet从master node传输到minion node，需要将上述packet再次封包。这个任务在backend为vxlan的flannel network中由linux kernel来完成。

flannel.1为vxlan设备，linux kernel可以自动识别，并将上面的packet进行vxlan封包处理。在这个封包过程中，kernel需要知道该数据包究竟发到哪个node上去。kernel需要查看node上的**fdb**(forwarding database)以获得上面对端vtep设备（已经从arp table中查到其mac地址：d6:51:2e:80:5c:69）所在的node地址。如果fdb中没有这个信息，那么kernel会向用户空间的flanned程序发起”L2 MISS”事件。flanneld收到该事件后，会查询etcd，获取该vtep设备对应的node的”Public IP“，并将信息注册到fdb中。

这样Kernel就可以顺利查询到该信息并封包了：

master node:

$ bridge fdb show dev flannel.1|grep d6:51:2e:80:5c:69
d6:51:2e:80:5c:69 dst {minion node local ip} self permanent

由于目标ip是minion node，查找路由表，包应该从master node的eth0发出，这样src ip和src mac地址也就确定了。封好的包示意图如下：

e) kernel的vxlan拆包

minion node上的eth0接收到上述vxlan包，kernel将识别出这是一个vxlan包，于是拆包后将flannel.1 packet转给minion node上的vtep（flannel.1）。minion node上的flannel.1再将这个数据包转到minion node上的docker0，继而由docker0传输到Pod3的某个容器里。

# Pod内到外部网络

我们在Pod中除了可以与pod network中的其他pod通信外，还可以访问外部网络，比如：

master node:
$ docker exec ba75f81455c7 ping -c 3 baidu.com
PING baidu.com (180.149.132.47): 56 data bytes
64 bytes from 180.149.132.47: icmp_seq=0 ttl=54 time=3.586 ms
64 bytes from 180.149.132.47: icmp_seq=1 ttl=54 time=3.752 ms
64 bytes from 180.149.132.47: icmp_seq=2 ttl=54 time=3.722 ms
--- baidu.com ping statistics ---
3 packets transmitted, 3 packets received, 0% packet loss
round-trip min/avg/max/stddev = 3.586/3.687/3.752/0.072 ms

这个通信与vxlan就没有什么关系了，主要是通过**docker引擎在iptables的POSTROUTING chain中设置的MASQUERADE规则**：

mastre node:

$ iptables -t nat -nL
... ...
Chain POSTROUTING (policy ACCEPT)
target     prot opt source               destination
MASQUERADE  all  --  172.16.99.0/24       0.0.0.0/0
... ...

docker将容器的pod network地址伪装为node ip出去，包回来时再snat回容器的pod network地址，这样网络就通了。

# ”不真实”的Service网络

每当我们在k8s cluster中创建一个service，k8s cluster就会在–service-cluster-ip-range的范围内为service分配一个cluster-ip，比如本文开始时提到的：

$ kubectl get services
NAME           CLUSTER-IP      EXTERNAL-IP   PORT(S)     AGE
index-api      192.168.3.168   <none>        30080/TCP   18d
kubernetes     192.168.3.1     <none>        443/TCP     94d
my-nginx       192.168.3.179   <nodes>       80/TCP      90d
nginx-kit      192.168.3.196   <nodes>       80/TCP      12d
rbd-rest-api   192.168.3.22    <none>        8080/TCP    60d
这个cluster-ip只是一个虚拟的ip，并不真实绑定某个物理网络设备或虚拟网络设备，仅仅存在于iptables的规则中：

Chain PREROUTING (policy ACCEPT)
target     prot opt source               destination
KUBE-SERVICES  all  --  0.0.0.0/0            0.0.0.0/0            /* kubernetes service portals */

$ iptables -t nat -nL|grep 192.168.3
Chain KUBE-SERVICES (2 references)
target     prot opt source               destination
KUBE-SVC-XGLOHA7QRQ3V22RZ  tcp  --  0.0.0.0/0            192.168.3.182        /* kube-system/kubernetes-dashboard: cluster IP */ tcp dpt:80
KUBE-SVC-NPX46M4PTMTKRN6Y  tcp  --  0.0.0.0/0            192.168.3.1          /* default/kubernetes:https cluster IP */ tcp dpt:443
KUBE-SVC-AU252PRZZQGOERSG  tcp  --  0.0.0.0/0            192.168.3.22         /* default/rbd-rest-api: cluster IP */ tcp dpt:8080
KUBE-SVC-TCOU7JCQXEZGVUNU  udp  --  0.0.0.0/0            192.168.3.10         /* kube-system/kube-dns:dns cluster IP */ udp dpt:53
KUBE-SVC-BEPXDJBUHFCSYIC3  tcp  --  0.0.0.0/0            192.168.3.179        /* default/my-nginx: cluster IP */ tcp dpt:80
KUBE-SVC-UQG6736T32JE3S7H  tcp  --  0.0.0.0/0            192.168.3.196        /* default/nginx-kit: cluster IP */ tcp dpt:80
KUBE-SVC-ERIFXISQEP7F7OF4  tcp  --  0.0.0.0/0            192.168.3.10         /* kube-system/kube-dns:dns-tcp cluster IP */ tcp dpt:53
... ...

可以看到在PREROUTING环节，k8s设置了一个target: KUBE-SERVICES。而KUBE-SERVICES下面又设置了许多target，一旦destination和dstport匹配，就会沿着chain进行处理。

比如：当我们在pod网络curl 192.168.3.22 8080时，匹配到下面的KUBE-SVC-AU252PRZZQGOERSG target：

KUBE-SVC-AU252PRZZQGOERSG  tcp  --  0.0.0.0/0            192.168.3.22         /* default/rbd-rest-api: cluster IP */ tcp dpt:8080
沿着target，我们看到”KUBE-SVC-AU252PRZZQGOERSG”对应的内容如下：

Chain KUBE-SVC-AU252PRZZQGOERSG (1 references)
target     prot opt source               destination
KUBE-SEP-I6L4LR53UYF7FORX  all  --  0.0.0.0/0            0.0.0.0/0            /* default/rbd-rest-api: */ statistic mode random probability 0.50000000000
KUBE-SEP-LBWOKUH4CUTN7XKH  all  --  0.0.0.0/0            0.0.0.0/0            /* default/rbd-rest-api: */

Chain KUBE-SEP-I6L4LR53UYF7FORX (1 references)
target     prot opt source               destination
KUBE-MARK-MASQ  all  --  172.16.99.6          0.0.0.0/0            /* default/rbd-rest-api: */
DNAT       tcp  --  0.0.0.0/0            0.0.0.0/0            /* default/rbd-rest-api: */ tcp to:172.16.99.6:8080

Chain KUBE-SEP-LBWOKUH4CUTN7XKH (1 references)
target     prot opt source               destination
KUBE-MARK-MASQ  all  --  172.16.99.7          0.0.0.0/0            /* default/rbd-rest-api: */
DNAT       tcp  --  0.0.0.0/0            0.0.0.0/0            /* default/rbd-rest-api: */ tcp to:172.16.99.7:8080

Chain KUBE-MARK-MASQ (17 references)
target     prot opt source               destination
MARK       all  --  0.0.0.0/0            0.0.0.0/0            MARK or 0x4000

请求被按5：5开的比例分发（起到负载均衡的作用）到KUBE-SEP-I6L4LR53UYF7FORX 和KUBE-SEP-LBWOKUH4CUTN7XKH，而这两个chain的处理方式都是一样的，那就是先做mark，然后做dnat，将service ip改为pod network中的Pod IP，进而请求被实际传输到某个service下面的pod中处理了。