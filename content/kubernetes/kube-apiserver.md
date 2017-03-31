
kubelet可以从它监听的目录获取pod配置，也可以从apiserver获取pod配置；

apiserver是唯一连接etcd的组件；

注意通过容器运行etcd时，指定了网络模式为host，这样可以直接通过本地IP访问；

kubelet启动时，自动向apiserver注册它所在的node；

在一个完整的cluster环境中，scheduler决定pod应该在哪个node上运行，但是可以给pod指定nodeName来手动指定运行它的node；
这个原理是：
1. kubelet wathc api server的变化；
2. kubectl客户端在api server上创建一个pod，并指定在nodeName上运行；
3. nodeName上的kubelet发现这个新pode，开始运行它；

如果集群中没有scheduler，且kubctl创建容器时没有指定nodeName，则pod将一直处于Pending状态，不会被任何node的kubelet执行；

