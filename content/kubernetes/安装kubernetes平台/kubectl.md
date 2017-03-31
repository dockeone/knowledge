## 创建kubectl配置文件

``` bash
$ ls ~/.kube/
$ kubectl config set-cluster default-cluster --server=http://10.64.3.7:8080
cluster "default-cluster" set.
$ kubectl config set-context default-context --cluster=default-cluster --user=default-admin
context "default-context" set.
$ kubectl config use-context default-context
switched to context "default-context".
$ cat ~/.kube/config
apiVersion: v1
clusters:
- cluster:
    server: http://10.64.3.7:8080
  name: default-cluster
contexts:
- context:
    cluster: default-cluster
    user: default-admin
  name: default-context
current-context: default-context
kind: Config
preferences: {}
users: []
```

注意：

1. 用node ip指定apiserver；
1. 访问的是apiserver的不需要认证和授权的**非安全端口**；