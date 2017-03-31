# 系统环境

$ uname -a
Linux tjwq01-sys-power003008.tjwq01.ksyun.com 3.10.0-327.el7.x86_64 #1 SMP Thu Nov 19 22:10:57 UTC 2015 x86_64 x86_64 x86_64 GNU/Linux

$ docker version
Client:
 Version:      17.03.0-ce
 API version:  1.26
 Go version:   go1.7.5
 Git commit:   3a232c8
 Built:        Tue Feb 28 08:10:07 2017
 OS/Arch:      linux/amd64

Server:
 Version:      17.03.0-ce
 API version:  1.26 (minimum version 1.12)
 Go version:   go1.7.5
 Git commit:   3a232c8
 Built:        Tue Feb 28 08:10:07 2017
 OS/Arch:      linux/amd64
 Experimental: false


# dockerd的--iptables默认参数为true，自动添加相关iptables

例如，启动一个hostPort 9128映射的容器，则dockerd自动启动一个docker-proxy进程监听主机的9128端口，创建DNAT规则并映射到容器172.30.60.6:9128：

$ docker ps|grep 9128
e3a4ea29397d        digitalocean/ceph_exporter                                   "/bin/ceph_exporter"     3 days ago          Up 3 minutes        10.64.3.7:9128->9128/tcp   thirsty_pasteur

$ ps -e -o pid,ppid,args -H|grep 9128
41438     1   /usr/bin/dockerd --bip=172.30.83.1/24 --ip-masq=true --mtu=1450
4 S root     39004 11087  0  80   0 - 45464 futex_ 15:12 ?        00:00:00 /usr/bin/docker-proxy -proto tcp -host-ip 10.64.3.7 -host-port 9128 -container-ip 172.30.60.6 -container-port 9128

$ netstat -lnpt|grep 9128
tcp        0      0 10.64.3.7:9128          0.0.0.0:*               LISTEN      39004/docker-proxy

# 添加的iptables如下

$ iptables -nL -t nat
Chain PREROUTING (policy ACCEPT)
target     prot opt source               destination
DOCKER     all  --  0.0.0.0/0            0.0.0.0/0            ADDRTYPE match dst-type LOCAL

Chain INPUT (policy ACCEPT)
target     prot opt source               destination

Chain OUTPUT (policy ACCEPT)
target     prot opt source               destination
DOCKER     all  --  0.0.0.0/0           !127.0.0.0/8          ADDRTYPE match dst-type LOCAL

Chain POSTROUTING (policy ACCEPT)
target     prot opt source               destination

Chain DOCKER (2 references)
target     prot opt source               destination
RETURN     all  --  0.0.0.0/0            0.0.0.0/0
DNAT       tcp  --  0.0.0.0/0            10.64.3.7            tcp dpt:9128 to:172.30.60.6:9128

$ iptables -nL
Chain INPUT (policy ACCEPT)
target     prot opt source               destination

Chain FORWARD (policy DROP)
target     prot opt source               destination
DOCKER-ISOLATION  all  --  0.0.0.0/0            0.0.0.0/0
DOCKER     all  --  0.0.0.0/0            0.0.0.0/0

Chain OUTPUT (policy ACCEPT)
target     prot opt source               destination

Chain DOCKER (1 references)
target     prot opt source               destination
ACCEPT     tcp  --  0.0.0.0/0            172.30.60.6          tcp dpt:9128

Chain DOCKER-ISOLATION (1 references)
target     prot opt source               destination
RETURN     all  --  0.0.0.0/0            0.0.0.0/0

# 创建的iptables如下

nat表：

1. 只对访问nodeIP:hostPort的包做DNAT，映射到容器IP(Pod IP):Port，后续的filter表允许这个包FORWARD；

filter表：

1. docker只将hostPort映射的 容器IP(Pod IP):Port 添加到Chain Docker中；
1. docker v1.3版本后，`--ip-forward`参数默认为true，将filter表中Chain FORWARD的默认polcy设置为**DROP**；
1. 如果收到不匹配Chain Docker的包(例如其它Node发过来的访问Pod容器的包)，将被Chain FORWARD丢弃。
1. **如果和kubernetes一起使用，可能需要手动将 Chain FORWARD的默认策略设置为ACCEPT**；