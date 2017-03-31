<!-- toc -->

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