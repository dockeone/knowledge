# kube-proxy

service是一组pod的服务抽象，相当于一组pod的LB，负责将请求分发给对应的pod。service会为这个LB提供一个IP，一般称为cluster IP。
kube-proxy的作用主要是负责service的实现，具体来说，就是实现了内部从pod到service和外部的从node port向service的访问。

举个例子，现在有podA，podB，podC和serviceAB。serviceAB是podA，podB的服务抽象(service)。那么kube-proxy的作用就是可以将pod(不管是podA，podB或者podC)向serviceAB的请求，
进行转发到service所代表的一个具体pod(podA或者podB)上。请求的分配方法一般分配是采用轮询方法进行分配。

另外，kubernetes还提供了一种在node节点上暴露一个端口，从而提供从外部访问service的方式。

比如我们使用这样的一个manifest来创建service

apiVersion: v1
kind: Service
metadata:
  labels:
    name: mysql
    role: service
  name: mysql-service
spec:
  ports:
    - port: 3306  // cluster ip port
      targetPort: 3306 // pod port
      nodePort: 30964 // node port
  type: NodePort
  selector:
    mysql-service: "true"

他的含义是在node上暴露出30964端口，当访问node上的30964端口或cluster ip的3306端口时，请求会转发到后端的pod3306端口。

nodePort跟LoadBalancer其实是同一种方式。区别在于LoadBalancer比nodePort多了一步，就是可以调用cloud provider去创建LB来向节点导流。cloud provider支持了openstack、gce等系统。
nodePort的原理在于在node上开了一个端口，将向该端口的流量导入到kube-proxy，然后由kube-proxy进一步导给对应的pod。
所以service采用nodePort的方式，正确的方法是在前面有一个lb，然后lb的后端挂上所有node的对应端口。这样即使node1挂了。lb也可以把流量导给其他node的对应端口。
使用get service可以看到虽然type是NodePort，但是依然为其分配了一个clusterIP；

kuer-proxy目前有userspace和iptables两种实现方式：

1. userspace是在用户空间，通过kuber-proxy实现LB的代理服务。这个是kube-proxy的最初的版本，较为稳定，但是效率也自然不太高。为7层转发，具体实现原理是，通过iptables规则将访问node port或
clusterip:port 的请求转发到本机kuber-proxy监听的端口，然后通过不同node上kube-proxy间的转发将请求转发给对应的pod。kube-proxy自己内部实现有负载均衡的方法，并可以查询到这个service下对应
pod的地址和端口(endpoints，可以使用命令etcdctl cat /registry/services/endpoints/default/my-nginx查看my-nginx service的endpoints)，进而把数据转发给对应的pod的地址和端口。

2. 另外一种方式是iptables的方式。是纯采用iptables来实现LB。是目前一般kube默认的方式。为4层转发，具体实现原理是通过iptables规则将访问node port或clusterip:port 的请求直接转发给对应的后端
podip:port（一般是轮询）；

# userspace

这里具体举个例子，以ssh-service1为例，kube为其分配了一个clusterIP。分配clusterIP的作用还是如上文所说，是方便pod到service的数据访问。

[minion@te-yuab6awchg-0-z5nlezoa435h-kube-master-udhqnaxpu5op ~]$ kubectl get service
NAME             LABELS                                    SELECTOR              IP(S)            PORT(S)
kubernetes       component=apiserver,provider=kubernetes   <none>                10.254.0.1       443/TCP
ssh-service1     name=ssh,role=service                     ssh-service=true      10.254.132.107   2222/TCP

使用describe可以查看到详细信息。可以看到暴露出来的NodePort端口30239。

[minion@te-yuab6awchg-0-z5nlezoa435h-kube-master-udhqnaxpu5op ~]$ kubectl describe service ssh-service1 
Name:           ssh-service1
Namespace:      default
Labels:         name=ssh,role=service
Selector:       ssh-service=true
Type:           LoadBalancer
IP:         10.254.132.107
Port:           <unnamed>   2222/TCP
NodePort:       <unnamed>   30239/TCP
Endpoints:      <none>
Session Affinity:   None
No events.

nodePort的工作原理与clusterIP大致相同，是发送到node上指定端口的数据，通过iptables重定向到kube-proxy对应的端口上。然后由kube-proxy进一步把数据发送到其中的一个pod上。
该node的ip为10.0.0.5

[minion@te-yuab6awchg-0-z5nlezoa435h-kube-master-udhqnaxpu5op ~]$ sudo iptables -S -t nat
...
// 从Container访问nodePort的规则
-A KUBE-NODEPORT-CONTAINER -p tcp -m comment --comment "default/ssh-service1:" -m tcp --dport 30239 -j REDIRECT --to-ports 36463
// 从Host访问nodePort的规则
-A KUBE-NODEPORT-HOST -p tcp -m comment --comment "default/ssh-service1:" -m tcp --dport 30239 -j DNAT --to-destination 10.0.0.5:36463
// 从Container访问cluster ip和port的规则
-A KUBE-PORTALS-CONTAINER -d 10.254.132.107/32 -p tcp -m comment --comment "default/ssh-service1:" -m tcp --dport 2222 -j REDIRECT --to-ports 36463
// 从Host访问cluster ip和port的规则
-A KUBE-PORTALS-HOST -d 10.254.132.107/32 -p tcp -m comment --comment "default/ssh-service1:" -m tcp --dport 2222 -j DNAT --to-destination 10.0.0.5:36463

可以看到访问node时候的30239端口会被转发到node上的36463端口。而且在访问clusterIP 10.254.132.107的2222端口时，也会把请求转发到本地的36463端口。

36463端口实际被kube-proxy所监听，将流量进行导向到后端的pod上。

# iptables
iptables的方式则是利用了linux的iptables的nat转发进行实现。在本例中，创建了名为mysql-service的service。

apiVersion: v1
kind: Service
metadata:
  labels:
    name: mysql
    role: service
  name: mysql-service
spec:
  ports:
    - port: 3306
      targetPort: 3306
      nodePort: 30964
  type: NodePort
  selector:
    mysql-service: "true"

mysql-service对应的nodePort暴露出来的端口为30964，对应的cluster IP(10.254.162.44)的端口为3306，进一步对应于后端的pod的端口为3306。
mysql-service后端代理了两个pod，ip分别是192.168.125.129和192.168.125.131。先来看一下iptables。