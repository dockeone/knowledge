# 修改配置文件

``` shell
$ cat /etc/etcd/etcd.conf |grep -v '^#'
ETCD_NAME=default
ETCD_DATA_DIR="/var/lib/etcd/default.etcd"
ETCD_LISTEN_CLIENT_URLS="http://10.64.3.7:2379"
ETCD_ADVERTISE_CLIENT_URLS="http://10.64.3.7:2379"
```

# 重启服务

``` shell
$ systemctl restart etcd
$
```

# 向etcd写入flanneld读取的Pod网络和子网信息

``` bash
$ etcdctl mkdir /kube-centos/network
$ etcdctl mk /kube-centos/network/config "{ \"Network\": \"172.30.0.0/16\", \"SubnetLen\": 24, \"Backend\": { \"Type\": \"vxlan\" } }"
$ netstat -lnpt|grep etcd
tcp        0      0 10.64.3.7:2379          0.0.0.0:*               LISTEN      42452/etcd
tcp        0      0 127.0.0.1:2380          0.0.0.0:*               LISTEN      42452/etcd
tcp        0      0 127.0.0.1:7001          0.0.0.0:*               LISTEN      42452/etcd
```

etcd是kubernetes**唯一有状态的服务**；缺省情况下kubernetes对象保存在`/registry`目录下，可以通过apiserver的`--etcd-prefix`参数进行配置；

apiserver是**唯一**连接etcd的kubernetes组件，其它组件都需要通过apiserver的接口来获取集群状态；
