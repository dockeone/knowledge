# Pod

Pod由Pause特殊容器和其它紧密关联的业务容器组成：
1. Pause是业务无关的容器，代表了整个Pod的状态；
2. Pod里的业务容器共享Pause容器的IP、挂载的Volume, 简化了业务容器间的通信和文件共享问题；

Pod有两种类型：
1. 普通pod：在etcd中创建pod定义后，master调度(绑定)到某个node上；
2. 静态pod：保存在node的配置文件中，只在对应node上执行；不能通过apiserver管理，kubelet不对它进行剪卡检查，不能与RC、Deployment等关联；

当pod中的一个容器停止运行了后，master会重启该pod中的所有容器；如果所在node down机，master将pod调度到其它node执行；

kube 为每个pod分配一个IP地址，称为Pod IP，pod里的容器共享整个IP地址，在创建cluster时可以指定分配给pod的IP地址段。

通过下面两种方式可以实现pod之间使用pod ip相互通信：
1. Overlay netowrk：使用vxlan等技术对pod网络流量进行封装: 可以使用CNI网络接口技术；也可以使用其他的add-on方案如calico、fannel、ovs、romana、weave等；
2. 非overlay netowrk：配置网络设备来感知pod网络地址段，直接路由: GCE、AWS等；

Pod IP和容器端口(containerPort)组成了Endpoint资源对象，它代表着Pod里的一个服务对外通信的地址。Pod可以通过暴露多个容器端口的形式生成多个Endpoint对象。
kube-proxy从apiserver获取Service的定义，然后将Server的Cluster IP和Pode的Endpoint绑定起来，实现负载均衡；

Pod中的每一个容器都可以对它使用的资源做限制，如CPU和内存，CPU的资源单位为Core数量，以千分之一的CPU配额为最小单位，用m表示，如100m表示应0.1个CPU；
内存配置的单位是字节数。设置资源配额的参数：
1. Requests：该容器使用的最小资源量，Node必须满足后才能被绑定执行；
2. Limits：容器允许使用的最大资源量，当超过时会被k8s kill并重启；

# Label

Label可以附加到各种资源对象上，如Node、Pod、Service、RC等，用来实现多维度的资源管理，如分配、调度、配置和部署等；
通过Label Selector查询和筛选拥有这些Label的资源对象；当前有两种Label Selector表达式：
1. 基于等式的如 name=redis-slave;
2. 基于集合的，如 name in (redis-master, redis-slave)

# Replication Controller 

RC是定义了一个期望的场景，即声明某种Pod的副本数量在任意时刻都是某个预期值：
1. Pod期待的副本数目(replicas);
2. 用于筛选目标Pod的Label Selector；
3. 用于创建新Pod的Pod模板(template);

当向apiserver提交了RC定义后，controller-manager会定期巡检当前存储的Pod，确保Pod的数量符合预期值；
删除RC并不会影响该RC创建的Pod，通过设置replicas为0，可以实现删除所有Pod；

执行kubectl scale rc可以修改RC控制的Pod的数量，实现动态缩放(Scaling);
执行kubectl rolling-update可以对RC里面的Pod进行滚动升级(Rolling Update)；

由于Replication Controller与源码中的模块Replication Controller同名，kubernetes 1.2将它升级为Replica Set，即下一代的RC，
它与当前RC的唯一区别是：Replica Set支持基于集合的Label selector(set-based selector), 而RC只支持基于等式的Label selector；
selector:
    matchLabels:
        tier: fronted
    matchExpressions:
        - {key: tier, operator: In, values: [fronted]}

kubectl命令行适用于RC的绝大部分命令也适用于Replica Set；但是当前很少单独使用Replica Set，它主要被更高层的Deployment资源对象使用，使用Deployment自动创建和维护Replica Set; Replica Set和Deployment是k8s自动伸缩的基础；

# Deployment

Deployment是RC的升级版本，与RC的最大不同是可随时知道当前pod部署进度, 因为创建一个Pod会经历创建、调度、绑定、在目标Node上启动容器等过程需要一定时间；

Deployment的应用场景：
1. 创建一个Deployment对象来生成对应的Replica Set，完成Pod副本的创建；
2. 检查Deployment的状态来看部署动作是否完成(Pod的数量是否达到预期值)；
3. 更新Deployment定义以创建新的Pod；
4. 如果当前Deploment不稳定，可以回滚到以前的版本；
5. 在部署过程中，挂起或恢复一个Deployment；

Deployment、Replicat Set、pod 三者的名称关系，可以很容易看出三者之间的关系，在排查错误时很有用：
1. 创建一个名为tomcat-deploy的Deployment；
2. 自动创建名称类似于tomcat-deploy-11112222的Replica Set;
3. 自动创建名称类似于tomcat-deploy-11112222-zhrsc的Pods；

# Horizontal Pod AutoScaler (HPA)

kubectl scale命令可以实现Pod的缩放，由于需要手动执行，故不自动化和智能化；
HPA(Horizontal Pod Autoscaler)作用对象是Deployment或Replica Set，根据当前所有Pod的负载情况，来确定是否需要调整目标Pod的副本数，实现Pod的自动扩容或缩容；
当前参考的负载：
1. CPU利用率百分比(CPUUtilizationPercentage)，目前需要Heapster扩展组件来采集这个值；
2. 应用自己定义的度量指标，如TPS、QPS；

HPA对象的命令是 kubectl autoscale; 更新Deployment后执行kubectl rollout 命令来查看滚动更新状态和进度；

# Service

Service是一个微服务，k8s为它分配一个地址，用户通过这个地址来访问背后一组由Pod副本组成的集群实例；Service和后端的Pod副本集群是通过Label Selector实现无缝
对接的；

kube为每一个service分配一个全局虚拟的IP地址(不可路由, 在创建cluster时可以指定service使用的IP地址段)，称为Cluster IP。虽然Pod的Endpoint会随着Pod的销毁和重建发生改变，但是在整个Service生命周期内，它的Cluster IP不会变化；所以在k8s系统中，服务发现是通过将Service Name和它的Cluster IP做下DNS域名映射实现的。

各Node上的kube-proxy进程是一个智能软件负载均衡器，把对Service的请求转发都后端的某个Pod实例上(将访问Cluster IP的流量转为Endpoint)，在内部实现服务的负载均衡和会话保持机制；

k8s的nodePort的原理是在集群中的每个node上开了一个端口，将访问该端口的流量导入到该node上的kube-proxy，然后再由kube-proxy进一步将流量转发给该对应该
nodeport的service的alive的pod上。

service的定义文件中，port参数指定service 虚拟端口，targetPort指定用来提供服务的容器暴露的端口。如果没有指定targetPort，默认与port相同；

service可以有多组port，需要给它们分别命名，对应不同的Endpoints;
apiVersion: v1
kind: Service
metadata:
    name: tomcat-service
spec:
    ports:
    - port: 8080
    name: service-port
    - port: 8005
    name: shutdown-port
    selector:
        tier: frontend
然后每个Pod的容器在启动时，自动注入这些环境变量：
TOMCAT_SERVICE_SERVICE_HOST=169.4169.41.218
TOMCAT_SERVICE_SERVICE_PORT_SERVICE_PORT=8080
TOMCAT_SERVICE_SERVICE_PORT_SHUTDOWN_PORT=8005

# K8s的服务发现机制

对于K8s来说，服务发现指的是通过Service Name找到对应的Cluster IP，k8s提供了两种方式：
1. pod容器启动时，自动注入当前系统中Servers相关的环境变量，应用程序获取这些环境变量的值来获取服务的IP和端口；k8s不会给已运行的Pod容器注入新创建的Service
环境变量。
2. 通过DNS插件的形式，将服务名和Endopoint映射起来，可以解决方式1的问题；

# 外部系统访问Service

kubernetes有三种类型IP：
1. Node IP：每个节点物理网卡的IP；
2. Pod IP：每个Pod的IP，是从Docker的docker0网桥的IP地址段中分配的，是一个虚拟的二层网络，k8s要求不通Node上的Pod可以通过Pod IP通信，所以一个Pod容器可以
通过该虚拟的二层网络访问另一个Pod（同一个Node或不同Node）的容器，真实的流量这是通过NodeIP进行的；
3. Cluster IP: 仅为Service分配的虚拟IP，属于k8s内部的地址，外界ping不通，只能结合Service Port组成一个具体的通信端口，单独的Cluster IP不具备TCP/IP通信的基础，只在k8s内部使用，集群之外的节点如果要访问这个通信端口，需要做一些额外的工作；

使用NodePort可以解决外部访问Service的问题，它的实现方式是k8s在集群里的每个Node上为需要外部访问的Service开启一个监听端口(kube-proxy进程监听)，外部系统访问任意一个Node IP + Node Port时即可访问该服务；

Service NodePort解决了，但是没有解决负载均衡的问题，需要引入外部LB(如HAProxy、Nginx)来将用户请求转发到后面某个Node的Node Port上；

# Volume

Volume是Pod中能被多个容器访问的共享目录，与Pod的生命周期相同(而不是容器)，容器终止或重启时，Volume中的数据不会丢失；
Volume的类型：
1. emptyDir: 在Pod分配到Node时创建，内容为空，无需指定宿主机上对应的目录文件，Pod移除时内容自动清空；
2. hostPath: 将宿主机上的文件或目录挂载到Pod上；
3. gcePersistentDisk;
4. 容器配置文件集中化定义与管理的ConfigMap资源对象；
5. 用于为Pod提供加密信息的Secret，通过tmpfs实现，这种类型的volume不会持久化；
4. 其它： iscsi、flocker、glusterfs、rbd、secret等；

# Persistent Volume

PV可以理解为k8s集群中某个网络存储中的一块存储，与Volume类似，区别如下：
1. PV只能是网络存储，不属于任何Node，但是可以在每个Node上访问；
2. PV不是在Node上定义的，是独立于Node定义的；
3. 目前支持 GCE、NFS、RBD等；

Pod需要通过定义一个PersistentVolumeClaim(PVC)对象来申请符合条件的PV，然后在Pod的定义文件中引用该PVC即可；

# Namespace

Namespace 用于实现多租户的资源隔离，便于不通的分组再共享使用整个集群资源的同时还能被分别管理；同时还能限定每个namespace的资源配额。
k8s在启动时自动创建一个default namespace，一旦创建了namespace，在创建资源对象时可以指定所属的Namespace；

# Annotation

Annotation与Label类似，都是K/V键值对的形式定义，但是Annotation定义的是用户附件的任意信息，不用于Label Selector；
k8s的模块自身通过Annotation的形式给资源对象标记一些特殊信息如镜像Hash值、docker registry的地址等；



docker、kubelet、kube-proxy需要在容器外运行；etcd、kube-apiserver、kube-controller-manager、kube-sceduler建议在容器内运行，有两种方式获取相关的image：
1. 从GCR上下载；hyperkube 二进制程序包含了所有master运行所需的二进制；
2. 自己构建；

向kubeconfig文件添加认证信息的方式如下：
1. 不需要认证的情况(只使用firewall策略)：
    kubectl config set-cluster $CLUSTER_NAME --server=http://$MASTER_IP --insecure-skip-tls-verify=true
2. 否则，使用下面命令设置apiserver、client certs、用户认证信息：
    1. apiserver 
    kubectl config set-cluster $CLUSTER_NAME --certificate-authority=$CA_CERT --embed-certs=true --server=https://$MASTER_IP
    2. 客户端的证书和认证token
    kubectl config set-credentials $USER --client-certificate=$CLI_CERT --client-key=$CLI_KEY --embed-certs=true --token=$token
3. 设置缺省的context：
    kubectl config set-context $CONTEXT_NAME --cluster=$CLUSTER_NAME --user=$USER
    kubectl config use-context $CONTEXT_NAME

给kubelet、kube-proxy设置kubeconfig文件：
1. 可以使用上面配置的admin的认证信息；
2. 所有kubelet、kube-proxy、admin各使用一个token和kubeconfgi文件(共3个token和3份配置文件)， GCE使用这种方式；
3. 所有kubelet、kube-proxy、admin各自使用自己的认证信息；


Docker:
1. 如果以前安装和运行了docker，需要删除docker创建的bridge和iptables规则：
    iptables -t nat -F
    ip link set docker0 down
    ip link delete docker0
2. 给docker设置--bridge=cbr0参数；
3. --iptables=fase， 关闭docker设置host-port的iptables规则；kube-proxy会管理iptables；
4. --ip-masq=false: 关闭本机pod间相互通信，否则docker会创建NAT规则将源IP为PodIP的重写为NodeIP；


kubelet(在所有主机上安装，master和node)：
1. 如果是apiserver使用了HTTPS，则：
    --api-servers=https://$MASTER_IP
    --kubeconfig=/var/lib/kubelet/kubeconfig
2. 否则使用HTTP：
    --api-servers=http://$MASTER_IP

--config=/etc/kubernetes/manifests
--cluster-dns= to the address of the DNS server you will setup (see Starting Cluster Services.)
--cluster-domain= to the dns domain prefix to use for cluster DNS addresses.
--docker-root=
--root-dir=
--configure-cbr0= (described above)
--register-node (described in Node documentation.)


kube-proxy(node上安装，master上可选):
1. If following the HTTPS security approach:
    --master=https://$MASTER_IP
    --kubeconfig=/var/lib/kube-proxy/kubeconfig

2. Otherwise, if taking the firewall-based security approach
    --master=http://$MASTER_IP


Networking:
1. 每个node需要分配它自己的CIDR range NODE_X_POD_CIDR；
2. 需要在node上创建cbr0，它也需要从CIDR里分配一个地址，一般是第一个IP NODE_X_BRIDGE_ADDR；

1. 自动配置cbr0和CIDR：
--configure-cbr0=true，kubelet会等待node controller设置的Node.Spec.PodCIDR参数来配置cbr0；
2. 手动配置：
    1. Set --configure-cbr0=false on kubelet and restart.
    2. Create a bridge
        ip link add name cbr0 type bridge.
    3. Set appropriate MTU. NOTE: the actual value of MTU will depend on your network environment
        ip link set dev cbr0 mtu 1460
    4. Add the node’s network to the bridge (docker will go on other side of bridge).
        ip addr add $NODE_X_BRIDGE_ADDR dev cbr0
    5. Turn it on
        ip link set dev cbr0 up
        
如果关闭了docker的ip masquerading，则需要手动创建masque规则，将发送到cluster外的流量中的源IP为PodIP修改为Node IP，用于实现pods间通信：
iptables -t nat -A POSTROUTING ! -d ${CLUSTER_SUBNET} -m addrtype ! --dst-type LOCAL -j MASQUERADE
注： POSTROUTING表示对IP报文的源IP做修改，PREROUTING对IP报文的目的IP做修改；

Master部分的组件可以使用kubernetes管理和配置：
1. 使用Pod spec(yaml或json)来配置，而不是主机的init脚本或systemd unit；
2. 它们由kubernetes来保证运行；