<!-- toc -->

wget https://github.com/kubernetes/heapster/archive/v1.3.0.zip
unzip v1.3.0.zip
mv v1.3.0.zip heapster-1.3.0

cd /root/heapster-1.3.0/deploy/kube-config/influxdb

$ diff grafana-deployment.yaml.orig grafana-deployment.yaml
16c16
<         image: gcr.io/google_containers/heapster-grafana-amd64:v4.0.2
---
>         image: lvanneo/heapster-grafana-amd64:v4.0.2
40,41c40,41
<           # value: /api/v1/proxy/namespaces/kube-system/services/monitoring-grafana/
<           value: /
---
>           value: /api/v1/proxy/namespaces/kube-system/services/monitoring-grafana/
>           #value: /
注意：如果使用apiserver proxy或者kubectl proxy访问 grafana dashboard，则必须将 GF_SERVER_ROOT_URL 设置为 /api/v1/proxy/namespaces/kube-system/services/monitoring-grafana/，否则后续打开页面：
https://10.64.3.7:8080/api/v1/proxy/namespaces/kube-system/services/monitoring-grafana/
时会找不到http://10.64.3.7:8086/api/v1/proxy/namespaces/kube-system/services/monitoring-grafana/api/dashboards/home
页面；

$ diff heapster-deployment.yaml.orig heapster-deployment.yaml
16c16
<         image: gcr.io/google_containers/heapster-amd64:v1.3.0-beta.1
---
>         image: lvanneo/heapster-amd64:v1.3.0-beta.1

$ diff influxdb-deployment.yaml.orig influxdb-deployment.yaml
16c16
<         image: gcr.io/google_containers/heapster-influxdb-amd64:v1.1.1
---
>         image: lvanneo/heapster-influxdb-amd64:v1.1.1

$ kubectl create -f  .
deployment "monitoring-grafana" created
service "monitoring-grafana" created
deployment "heapster" created
service "heapster" created
deployment "monitoring-influxdb" created
service "monitoring-influxdb" created

$ kubectl get deployments -n kube-system
NAME                   DESIRED   CURRENT   UP-TO-DATE   AVAILABLE   AGE
heapster               1         1         1            1           10m
kube-dns               1         1         1            1           12d
kubernetes-dashboard   1         1         1            1           1d
monitoring-grafana     1         1         1            1           10m
monitoring-influxdb    1         1         1            1           10m

$ kubectl get pods -n kube-system
NAME                                    READY     STATUS    RESTARTS   AGE
heapster-3273315324-tmxbg               1/1       Running   0          11m
kube-dns-2200864676-p2zn6               4/4       Running   44         12d
kubernetes-dashboard-1047103861-rt8qm   1/1       Running   1          1d
monitoring-grafana-2255110352-94lpn     1/1       Running   0          11m
monitoring-influxdb-884893134-3vb6n     1/1       Running   0          11m

但是查看前端kubernetes dashboard，仍然没有图表数据。

查看heapster-3273315324-tmxbg pod的日志，发现连接 kublet 127.0.0.1:10255端口失败：
$ kubectl logs -f heapster-3273315324-tmxbg -n kube-system
I0325 15:30:16.974214       1 heapster.go:71] /heapster --source=kubernetes:https://kubernetes.default --sink=influxdb:http://monitoring-influxdb:8086
I0325 15:30:16.974316       1 heapster.go:72] Heapster version v1.3.0-beta.1
I0325 15:30:16.974551       1 configs.go:61] Using Kubernetes client with master "https://kubernetes.default" and version v1
I0325 15:30:16.974569       1 configs.go:62] Using kubelet port 10255
E0325 15:32:24.295433       1 influxdb.go:238] issues while creating an InfluxDB sink: failed to ping InfluxDB server at "monitoring-influxdb:8086" - Get http://monitoring-influxdb:8086/ping: dial tcp 10.254.170.251:8086: getsockopt: connection timed out, will retry on use
I0325 15:32:24.295479       1 influxdb.go:252] created influxdb sink with options: host:monitoring-influxdb:8086 user:root db:k8s
I0325 15:32:24.295522       1 heapster.go:193] Starting with InfluxDB Sink
I0325 15:32:24.295535       1 heapster.go:193] Starting with Metric Sink
I0325 15:32:24.310216       1 heapster.go:105] Starting heapster on port 8082
E0325 15:33:05.008436       1 kubelet.go:231] error while getting containers from Kubelet: failed to get all container stats from Kubelet URL "http://127.0.0.1:10255/stats/container/": Post http://127.0.0.1:10255/stats/container/: dial tcp 127.0.0.1:10255: getsockopt: connection refused
I0325 15:33:05.015628       1 influxdb.go:215] Created database "k8s" on influxDB server at "monitoring-influxdb:8086"
E0325 15:34:05.004126       1 kubelet.go:231] error while getting containers from Kubelet: failed to get all container stats from Kubelet URL "http://127.0.0.1:10255/stats/container/": Post http://127.0.0.1:10255/stats/container/: dial tcp 127.0.0.1:10255: getsockopt: connection refused

之所以出现这种问题，是由于/etc/kubernetes/kublet中配置的address是127.0.01：
$ kubectl get nodes
NAME        STATUS    AGE
127.0.0.1   Ready     99d

$ ps -elf|grep kubelet
4 S root     30071     1  5  80   0 - 364317 futex_ 23:27 ?       00:01:03 /usr/bin/kubelet --logtostderr=true --v=0 --api-servers=http://127.0.0.1:8080 --address=127.0.0.1 --hostname-override=127.0.0.1 --allow-privileged=false --pod-infra-container-image=registry.access.redhat.com/rhel7/pod-infrastructure:latest --tls-cert-file=/etc/kubernetes/ssl/kubecfg.crt --tls-private-key-file=/etc/kubernetes/ssl/kubecfg.key --cluster_dns=10.254.0.2 --cluster_domain=cluster.local

heapster 从apiserver 获取运行kubelet的node列表，然后从这些node poll状态数据(kublet 内置的cAdvisor搜集的数据），heapster容器获取了
名为127.0.0.1的node后，就向它发起poll请求，但实际上请求的是运行heapster的容器，不会到达node，所以被拒绝;
解决方法是 修改kublet的参数，将address修改为本机非127.0.0.1的IP地址，重启kublet；

https://github.com/kubernetes/heapster/issues/1183
Heapster has got the list of nodes from Kubernetes and is now trying to pull stats from the kublete process on each node (which has a built in cAdvisor collecting stats on the node). In this case there's only one node and it's known by 127.0.0.1 to kubernetes. And there's the problem. The Heapster container is trying to reach the node at 127.0.0.1 which is itself and of course finding no kublete process to interrogate within the Heapster container.

10255是kubelet的非安全端口：
$ kubelet --help 2>&1|grep 10255
      --read-only-port int32                                    The read-only port for the Kubelet to serve on with no authentication/authorization (set to 0 to disable) (default 10255)

如果定义service时指定了 kubernetes.io/cluster-service: 'true' annotation，则说明提供的是集群服务，可以用 kubectl cluster-info 查看访问链接；
$ grep cluster-service *
grafana-service.yaml:    kubernetes.io/cluster-service: 'true'
heapster-service.yaml:    kubernetes.io/cluster-service: 'true'
influxdb-service.yaml:    kubernetes.io/cluster-service: 'true'

$ kubectl cluster-info
Kubernetes master is running at http://10.64.3.7:8080
Heapster is running at http://10.64.3.7:8080/api/v1/proxy/namespaces/kube-system/services/heapster
KubeDNS is running at http://10.64.3.7:8080/api/v1/proxy/namespaces/kube-system/services/kube-dns
kubernetes-dashboard is running at http://10.64.3.7:8080/api/v1/proxy/namespaces/kube-system/services/kubernetes-dashboard
monitoring-grafana is running at http://10.64.3.7:8080/api/v1/proxy/namespaces/kube-system/services/monitoring-grafana
monitoring-influxdb is running at http://10.64.3.7:8080/api/v1/proxy/namespaces/kube-system/services/monitoring-influxdb

打开 http://10.64.3.7:8080/api/v1/proxy/namespaces/kube-system/services/monitoring-grafana 即可访问grafana；

或者通过kubectl proxy来访问cluster-info：
$ kubectl proxy --address='10.64.3.7' --port=8086 --accept-hosts='^*$'
Starting to serve on 10.64.3.7:8086

http://10.64.3.7:8086/api/v1/proxy/namespaces/kube-system/services/monitoring-grafana/