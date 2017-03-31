## 修改配置/etc/kubernetes/controller

``` bash
$ cat controller-manager |grep -v '^#'|grep -v '^$'
KUBE_CONTROLLER_MANAGER_ARGS="--address=127.0.0.1 --service_account_private_key_file=/etc/kubernetes/ssl/server.key --root-ca-file=/etc/kubernetes/ssl/ca.crt --cluster-cidr 10.254.0.0/16"
```

注意：

1. `--service_account_private_key_file`需要与apiserver的`service_account_key_file`参数一致；
1. `--root-ca-file` 用于对apiserver提供的证书进行校验，**指定该参数后才在各Pod的ServiceAccount挂载目录中包含ca文件**；
1. `--cluster-cidr` 指定与apiserver配置一致的 service cluster ip网段 ，kubelet会读取该参数设置`configureCBR0`，如果未指定，kubelet启动时提示(hairpin的含义见后文kublet一节)：

    ``` bash
    $ journalctl -u kubelet |tail -1000|grep 'Hai'
    Mar 29 01:43:09 tjwq01-sys-bs003007.tjwq01.ksyun.com kubelet[3766]: W0329 01:43:09.473594    3766 kubelet.go:584] Hairpin mode set to "promiscuous-bridge" but configureCBR0 is false, falling back to "hairpin-veth"
    ```

## 重启进程

``` bash
$ systemctl restart kube-controller-manager
$ ps -e -o ppid,pid,args -H |grep kube-con
1 30898   /root/local/bin/kube-controller-manager --logtostderr=true --v=0 --master=http://10.64.3.7:8080 --address=127.0.0.1 --service_account_private_key_file=/etc/kubernetes/ssl/server.key --root-ca-file=/etc/kubernetes/ssl/ca.crt --cluster-cidr 10.254.0.0/16
$ netstat -lnpt|grep kube-controll
tcp        0      0 127.0.0.1:10252         0.0.0.0:*               LISTEN      28636/kube-controll
```

## 查看日志

``` bash
$ journalctl -u kube-controller-manager -o export|grep MESSAGE
$
```

