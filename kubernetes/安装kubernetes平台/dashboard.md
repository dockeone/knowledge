<!-- toc -->

# 安装dashboard

$ cd /root/kubernetes/cluster/addons/dashboard
$ ls
dashboard-controller.yaml  dashboard-controller.yaml.orig  dashboard-service.yaml  MAINTAINERS.md  README.md

由于gcr.io被墙，所以使用docker hub上的镜像
$ diff dashboard-controller.yaml.orig dashboard-controller.yaml
24c24
<         image: gcr.io/google_containers/kubernetes-dashboard-amd64:v1.5.0
---
>         image: ist0ne/kubernetes-dashboard-amd64:v1.5.0

$ kubectl create -f dashboard-service.yaml
$ kubectl create -f dashboard-controller.yaml


$ kubectl get services --namespace kube-system
NAME                   CLUSTER-IP      EXTERNAL-IP   PORT(S)         AGE
kube-dns               10.254.0.2      <none>        53/UDP,53/TCP   10d
kubernetes-dashboard   10.254.120.25   <none>        80/TCP          27m

1. 可以修改 dashboard-service.yaml 指定 nodePort，这样外界就可以使用 nodeIP:nodePort 的方式来访问 dashboard；
1. 也可以使用 kubectl proxy 来提供对 apiserver、pod、node的服务的访问：
1. 还可以通过直接访问apiserver的方式安全端口的形式来访问dashboard等service；

## kubectl proxy方式

$ kubectl proxy --address='10.64.3.7' --port=8086
Starting to serve on 10.64.3.7:8086

浏览器打开 kubectl proxy 监听的地址，提示失败
http://10.64.3.7:8086/ui
<h3>Unauthorized</h3>

需要修改kubectl proxy的命令行参数，添加 accept-hosts 参数：
$ kubectl proxy --address='10.64.3.7' --port=8086 --accept-hosts='^*$'

浏览器打开 kubectl proxy监听的地址和端口，查看ui界面：
http://10.64.3.7:8086/ui

跳转到：
http://10.64.3.7:8086/api/v1/proxy/namespaces/kube-system/services/kubernetes-dashboard/#/workload?namespace=default

## apiserver方式

打开浏览器访问：

https://{master node public ip}:6443

这时浏览器会提示你：证书问题。忽略之（由于apiserver采用的是自签署的私有证书，浏览器端无法验证apiserver的server.crt），继续访问，浏览器弹出登录对话框，让你输入用户名和密码，这里我们输入apiserver —basic-auth-file中的用户名和密码，就可以成功登录apiserver，并在浏览器页面看到如下内容：

{
  "paths": [
    "/api",
    "/api/v1",
    "/apis",
    "/apis/apps",
    "/apis/apps/v1alpha1",
    "/apis/autoscaling",
    "/apis/autoscaling/v1",
    "/apis/batch",
    "/apis/batch/v1",
    "/apis/batch/v2alpha1",
    "/apis/extensions",
    "/apis/extensions/v1beta1",
    "/apis/policy",
    "/apis/policy/v1alpha1",
    "/apis/rbac.authorization.k8s.io",
    "/apis/rbac.authorization.k8s.io/v1alpha1",
    "/healthz",
    "/healthz/ping",
    "/logs/",
    "/metrics",
    "/swaggerapi/",
    "/ui/",
    "/version"
  ]
}

接下来，我们访问下面地址：

https://{master node public ip}:6443/ui

页面跳转到：

https://101.201.78.51:6443/api/v1/proxy/namespaces/kube-system/services/kubernetes-dashboard/

现在dashboard已经可以使用，但它还缺少metric和类仪表盘图形展示功能，这两个功能需要额外安装Heapster才能实现。