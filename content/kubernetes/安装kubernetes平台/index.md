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

[etcd](etcd.md)
[kube-apiserver](kube-apiserver.md)
[kubectl](kubectl.md)
[kube-controller-manager](kube-controller-manager.md)
[kube-scheduler](kube-scheduler.md)
[flanneld](flanneld.md)
[docker](docker.md)
[kube-proxy](kube-proxy.md)
[kubelet](kubelet.md)
[kube-dns](kube-dns.md)
[dashboard](dashboard.md)
[采集容器性能指标](heapster.md)
[日志搜集、分析和可视化](logging-EFk.md)
[搭建是有ceph registry](../../docker/registry-ceph.md)