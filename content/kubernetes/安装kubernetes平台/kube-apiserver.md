## 修改配置/etc/kubernetes/apiserver

``` bash
$ cat apiserver |grep -v '^#'|grep -v '^$'
KUBE_API_ADDRESS="--insecure-bind-address=10.64.3.7"
KUBE_ETCD_SERVERS="--etcd-servers=http://10.64.3.7:2379"
KUBE_SERVICE_ADDRESSES="--service-cluster-ip-range=10.254.0.0/16"  # **指定 service 的 cluster 子网网段**
KUBE_ADMISSION_CONTROL="--admission-control=NamespaceLifecycle,NamespaceExists,LimitRanger,SecurityContextDeny,ServiceAccount,ResourceQuota"
KUBE_API_ARGS="--bind-address=10.64.3.7 --service_account_key_file=/etc/kubernetes/ssl/server.key --client-ca-file=/etc/kubernetes/ssl/ca.crt --tls-cert-file=/etc/kubernetes/ssl/server.cert --tls-private-key-file=/etc/kubernetes/ssl/server.key"
```

注意：

1. 为了以后Pod容器能访问apiserver，`KUBE_ADMISSION_CONTROL`里面必须要包含`ServiceAccount`；
1. 需要指定`--service_account_key_file`参数，值为签名`ServiceAccount`的私钥，且必须与`kube-controller-manager`的`--service_account_private_key_file`参数指向同一个文件；如果未指定该参数，创建Pod会失败：
    ``` bash
    $  kubectl create -f Pod-nginx.yaml
    Error from server: error when creating "Pod-nginx.yaml": Pods "nginx" is forbidden: no API token found for service account default/default, retry after the token is automatically created and added to the service account
    ```
1. `--bind-address`**不能设置为127.0.0.1**, 否则Pod访问apiserver对应的kubernetes service cluster ip会被refuse，因为127.0.0.1是Pod所在的容器而非apiserver；
1. 如果没有指定`--tls-cert-file`和`--tls-private-key-file`，apiserver启动时会自动创建公私钥，但不会创建签名它们的ca证书和ca的key，所以**client不能对apiserver的证书进行验证**，保存在`/var/run/kubernetes/`目录；
1. 在 `--bind-address 6443`监听安全端口；
1. 在 `--insecure-bind-address 8080` 监听非安全端口，apisrver对访问该端口的请求不做认证、授权，**后面的kubectl配置文件~/.kube/conf使用的就是这个地址和端口**；

## 修改配置/usr/lib/systemd/system/kube-apiserver.service

添加kube账户对`/var/run/kubernetes`目录的读写权限

``` bash
[Service]
PermissionsStartOnly=true
ExecStartPre=-/usr/bin/mkdir /var/run/kubernetes
ExecStartPre=/usr/bin/chown -R kube:kube /var/run/kubernetes/
```

修改systemd unit文件后，需要reload生效：

``` bash
$ systemctl daemon-reload
$
```

## 重启进程

``` bash
$ systemctl restart kube-apiserver
$ ps -e -o args|grep apiserver
/usr/bin/kube-apiserver --logtostderr=true --v=0 --etcd-servers=http://127.0.0.1:2379 --insecure-bind-address=127.0.0.1 --allow-privileged=false --service-cluster-ip-range=10.254.0.0/16 --admission-control=NamespaceLifecycle,NamespaceExists,LimitRanger,SecurityContextDeny,ServiceAccount,ResourceQuota --bind-address=10.64.3.7 --service_account_key_file=/var/run/kubernetes/serviceAccount.key
```

## 查看日志


``` bash
$ journalctl -u kube-apiserver -o export|grep MESSAGE
$
```