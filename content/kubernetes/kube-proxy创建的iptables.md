# iptables处理流程

收包：

收到的Package -> NAT PREROUTING  -> 路由判断：目标地址为其它机器 -> FILTER FORWARD -> NAT POSTROUTING
收到的Package -> NAT PREROUTING -> 路由判断：目标地址为本机 -> FILTER INPUT -> 本机进程进程

发包：

本机进程发Package -> NAT OUT -> FILTER OUTPUT -> 路由判断：目标地址为其它机器 -> NAT POSTROUTING
本机进程发Package -> NAT OUT -> FILTER OUTPUT -> 路由判断：目标地址为本机 -> 本机进程处理

# kuber-proxy创建的iptables规则

相关代码：https://github.com/kubernetes/kubernetes/blob/master/pkg/proxy/iptables/proxier.go#L992

## nat表

kube-proxy对收包(NAT PREROUTINGN)和发包(NAT OUTPUT)创建了如下iptables规则：
    -> KUBE-SERVICES:
        #1. 如果目标地址是 ClusterIP:Port，但源地址不在Cluster CIDR中，则设置SNAT的标记，继续步骤2;
        2. 如果目标地址是 ClusterIP:Port，则跳转到 KUBE-SVC-XXX Chain；
        -> KUBE-SVC-XXX:
            1. 如果Service绑定了N个Pods，则KUBE-SVC-XXX下面有N条KUBE-SEP-XXX规则，每条规则对应一个Pod IP；
            2. 每条KUBE-SEP-XXX规则指定了statistic mode random probability(如果只有一条规则，则无需指定概率)，iptabels**按概率**执行某条规则；
            -> KUBE-SEP-XXX:
                1. 如果源地址是对应的PodIP则设置SNAT标记；
                2. 设置到目标PodIP:Port的DNAT；
        3. 最后一条规则是匹配目标地址是LOCAL的NodePort情况，复用对应service的KUBE-SVC-XXX Chain；

## filter表

1. 创建Chain KUBE-FIREWALL，DROP所有在NAT阶段标记为drop的包；
1. 创建Chain KUBE-SERVICES，REJECT没有Endpoint的ClusterIP请求；


## 总结

1. kubep-proxy为**访问Service Cluster IP:Port或NodePort的请求**创建DNAT规则(有时还创建SNAT规则)，映射到PodIP:PodPort；
1. 如果`kube-proxy`指定了`--cluster-cidr`或`--masquerade-all=true`参数，则kube-proxy对访问Cluster IP的请求做SNAT；这两个参数**默认未配置**，可能会引起**Node访问Cluster IP失败**，参考[k8s安装.md](k8s-安装.md)文档中关于kubel-proxy的介绍；
    https://github.com/kubernetes/kubernetes/blob/master/pkg/proxy/iptables/proxier.go#L944
1. 访问NodePort的请求将会被做SNAT；
1. 访问ClusterIP或NodePort的请求一旦被做了SNAT，**Pod容器进程将看不到客户请求的真实IP**(访问Port IP不受影响)；
1. 如果Pod访问**自身关联的**Cluster IP时，kube-proxy会对请求除了做上面的DNAT外，还做SNAT；
    https://github.com/kubernetes/kubernetes/blob/master/pkg/proxy/iptables/proxier.go#L992
1. 访问docker0所在的Pod IP段时，由kernel直接转发；
1. 访问非自己docker0所在的Pod IP段时，由flanneld转发到其它Node上的flanned(flannel接口的路由表项指向整个Pod IP段，16位掩码)；

https://github.com/kubernetes/kubernetes/issues/10921
v1.7可能会增加一个Source IP preservation for Virtual IPs的特性，这样容器进程可以看到用户实际IP：
https://github.com/kubernetes/features/issues/27
https://github.com/kubernetes/kubernetes/pull/41162
https://kubernetes.io/docs/tasks/access-application-cluster/create-external-load-balancer/#annotation-to-modify-the-loadbalancer-behavior-for-preservation-of-source-ip

kube-proxy对经过NAT PREROUTING的package，在FILTER INPUT、FORWARD、OUTPUT中定义了如下过滤规则：

1. Chain KUBE-FIREWALL 丢弃在NAT阶段被Mark DROP 0x8000的包；
1. Chain KUBE-SERVICES 会REJECT访问不存在 endpoints 的Service Cluster IP；


docker 也会对iptables的filter和nat表进行管理，主要应用场景是hostPort端口映射的情况；详见 [docker的iptables.md](../docker/docker的iptables.md)

# filter 表
Chain INPUT (policy ACCEPT)
target     prot opt source               destination
KUBE-FIREWALL  all  --  0.0.0.0/0            0.0.0.0/0

Chain FORWARD (policy ACCEPT)
target     prot opt source               destination

Chain OUTPUT (policy ACCEPT)
target     prot opt source               destination
KUBE-SERVICES  all  --  0.0.0.0/0            0.0.0.0/0            /* kubernetes service portals */
KUBE-FIREWALL  all  --  0.0.0.0/0            0.0.0.0/0

Chain KUBE-FIREWALL (2 references)
target     prot opt source               destination
DROP       all  --  0.0.0.0/0            0.0.0.0/0            /* kubernetes firewall for dropping marked packets */ mark match 0x8000/0x8000

Chain KUBE-SERVICES (1 references)
target     prot opt source               destination
REJECT     tcp  --  0.0.0.0/0            10.254.144.118       /* default/rc-demo-svc: has no endpoints */ tcp dpt:80 reject-with icmp-port-unreachable


# nat 表

Chain PREROUTING (policy ACCEPT)
target     prot opt source               destination         
KUBE-SERVICES  all  --  0.0.0.0/0            0.0.0.0/0            /* kubernetes service portals */

Chain OUTPUT (policy ACCEPT)
target     prot opt source               destination         
KUBE-SERVICES  all  --  0.0.0.0/0            0.0.0.0/0            /* kubernetes service portals */

// POSTROUTING 对设置了0x4000/0x4000 mark的Package做SNAT，这是Package离开本机时的最后一步处理步骤
Chain POSTROUTING (policy ACCEPT)
target     prot opt source               destination         
KUBE-POSTROUTING  all  --  0.0.0.0/0            0.0.0.0/0            /* kubernetes postrouting rules */
Chain KUBE-POSTROUTING (1 references)
target     prot opt source               destination         
MASQUERADE  all  --  0.0.0.0/0            0.0.0.0/0            /* kubernetes service traffic requiring SNAT */ mark match 0x4000/0x4000

// 将Package设置0x8000标记，然后FILTER表中OUTPU链将DROP有该标记的Package，该条规则适合本机向外发Package
Chain KUBE-MARK-DROP (0 references)
target     prot opt source               destination         
MARK       all  --  0.0.0.0/0            0.0.0.0/0            MARK or 0x8000

// 将Package设置0x4000标记
Chain KUBE-MARK-MASQ (37 references)
target     prot opt source               destination         
MARK       all  --  0.0.0.0/0            0.0.0.0/0            MARK or 0x4000

---------
Chain KUBE-SERVICES (2 references)
target     prot opt source               destination
    // 如果目标地址是 ClusterIP:Port，但源地址不在Cluster CIDR中，则设置SNAT的标记, 由于MARK标记不会终止Chain匹配，所以iptables会继续执行下一条规则
KUBE-MARK-MASQ  tcp  -- !10.254.0.0/16        10.254.158.94        /* default/l5d:incoming cluster IP */ tcp dpt:4141
    // 如果目标地址是 ClusterIP:Port，则跳转到另一条 SVC Chain
KUBE-SVC-MFRAO7LAQRPHV6LR  tcp  --  0.0.0.0/0            10.254.158.94        /* default/l5d:incoming cluster IP */ tcp dpt:4141
    Chain KUBE-SVC-MFRAO7LAQRPHV6LR (2 references)
    target     prot opt source               destination
    // 跳转到另一条 SEP Chain
    KUBE-SEP-WH5Y55GJXY5XEZPN  all  --  0.0.0.0/0            0.0.0.0/0            /* default/l5d:incoming */
        Chain KUBE-SEP-WH5Y55GJXY5XEZPN (1 references)
        target     prot opt source               destination
        // 如果源地址是POD IP 172.30.60.6即pod访问自己绑定的Service情况，则设置SNAT标记；
        KUBE-MARK-MASQ  all  --  172.30.60.6          0.0.0.0/0            /* default/l5d:incoming */
        // 设置DNAT，目标地址和端口为本Service绑定的PodIP:Port
        DNAT       tcp  --  0.0.0.0/0            0.0.0.0/0            /* default/l5d:incoming */ tcp to:172.30.60.6:4141
KUBE-MARK-MASQ  tcp  -- !10.254.0.0/16        10.254.0.159         /* default/my-nginx: cluster IP */ tcp dpt:80
KUBE-SVC-BEPXDJBUHFCSYIC3  tcp  --  0.0.0.0/0            10.254.0.159         /* default/my-nginx: cluster IP */ tcp dpt:80
    Chain KUBE-SVC-BEPXDJBUHFCSYIC3 (1 references)
    target     prot opt source               destination         
    KUBE-SEP-LUW52DEBW3YDNB3U  all  --  0.0.0.0/0            0.0.0.0/0            /* default/my-nginx: */ statistic mode random probability 0.50000000000
        Chain KUBE-SEP-LUW52DEBW3YDNB3U (1 references)
        target     prot opt source               destination         
        KUBE-MARK-MASQ  all  --  172.30.60.18         0.0.0.0/0            /* default/my-nginx: */
        DNAT       tcp  --  0.0.0.0/0            0.0.0.0/0            /* default/my-nginx: */ tcp to:172.30.60.18:80

    KUBE-SEP-Q3RE55W63QJG7E2K  all  --  0.0.0.0/0            0.0.0.0/0            /* default/my-nginx: */
        Chain KUBE-SEP-Q3RE55W63QJG7E2K (1 references)
        target     prot opt source               destination         
        KUBE-MARK-MASQ  all  --  172.30.60.22         0.0.0.0/0            /* default/my-nginx: */
        DNAT       tcp  --  0.0.0.0/0            0.0.0.0/0            /* default/my-nginx: */ tcp to:172.30.60.22:80
KUBE-MARK-MASQ  tcp  -- !10.254.0.0/16        10.254.137.54        /* default/deployment-demo-svc: cluster IP */ tcp dpt:80
KUBE-SVC-NZQF2F2VOEDENRAX  tcp  --  0.0.0.0/0            10.254.137.54        /* default/deployment-demo-svc: cluster IP */ tcp dpt:80
    Chain KUBE-SVC-NZQF2F2VOEDENRAX (1 references)
    target     prot opt source               destination         
    KUBE-SEP-DFFML6INFICBZQVD  all  --  0.0.0.0/0            0.0.0.0/0            /* default/deployment-demo-svc: */ statistic mode random probability 0.25000000000
        Chain KUBE-SEP-DFFML6INFICBZQVD (1 references)
        target     prot opt source               destination         
        KUBE-MARK-MASQ  all  --  172.30.60.15         0.0.0.0/0            /* default/deployment-demo-svc: */
        DNAT       tcp  --  0.0.0.0/0            0.0.0.0/0            /* default/deployment-demo-svc: */ tcp to:172.30.60.15:80
    KUBE-SEP-S3Q56YEVF5QRC6CL  all  --  0.0.0.0/0            0.0.0.0/0            /* default/deployment-demo-svc: */ statistic mode random probability 0.33332999982
        Chain KUBE-SEP-S3Q56YEVF5QRC6CL (1 references)
        target     prot opt source               destination         
        KUBE-MARK-MASQ  all  --  172.30.60.2          0.0.0.0/0            /* default/deployment-demo-svc: */
        DNAT       tcp  --  0.0.0.0/0            0.0.0.0/0            /* default/deployment-demo-svc: */ tcp to:172.30.60.2:80
    KUBE-SEP-TTYVKBSK6QD4HBYK  all  --  0.0.0.0/0            0.0.0.0/0            /* default/deployment-demo-svc: */ statistic mode random probability 0.50000000000
        Chain KUBE-SEP-TTYVKBSK6QD4HBYK (1 references)
        target     prot opt source               destination         
        KUBE-MARK-MASQ  all  --  172.30.60.4          0.0.0.0/0            /* default/deployment-demo-svc: */
        DNAT       tcp  --  0.0.0.0/0            0.0.0.0/0            /* default/deployment-demo-svc: */ tcp to:172.30.60.4:80
    KUBE-SEP-2QZYIWISCGMAOKL7  all  --  0.0.0.0/0            0.0.0.0/0            /* default/deployment-demo-svc: */
        Chain KUBE-SEP-2QZYIWISCGMAOKL7 (1 references)
        target     prot opt source               destination         
        KUBE-MARK-MASQ  all  --  172.30.60.7          0.0.0.0/0            /* default/deployment-demo-svc: */
        DNAT       tcp  --  0.0.0.0/0            0.0.0.0/0            /* default/deployment-demo-svc: */ tcp to:172.30.60.7:80
KUBE-NODEPORTS  all  --  0.0.0.0/0            0.0.0.0/0            /* kubernetes service nodeports; NOTE: this must be the last rule in this chain */ ADDRTYPE match dst-type LOCAL
    Chain KUBE-NODEPORTS (1 references)
    target     prot opt source               destination         
    KUBE-MARK-MASQ  tcp  --  0.0.0.0/0            0.0.0.0/0            /* default/l5d:incoming */ tcp dpt:31648
    KUBE-SVC-MFRAO7LAQRPHV6LR  tcp  --  0.0.0.0/0            0.0.0.0/0            /* default/l5d:incoming */ tcp dpt:31648