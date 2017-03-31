1. linkered 可以在每个pod中作为sidercar container部署，也可以通过DaemonSet的方式在每个Node上部署；


# 在每个pod中作为sidercar container部署linkered

1. 创建configmap：

$ cat linkerd-config.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: linkerd-config
data:
  config.yaml: |-
    admin:
      port: 9990

    namers:
    - kind: io.l5d.k8s
      experimental: true
      host: 127.0.0.1
      port: 8001

    routers:
    - protocol: http   # linkerd listen的IP:Port
      servers:
      - port: 8080
        ip: 0.0.0.0
      dtab: |
        /iface      => /#/io.l5d.k8s/default;  # 使用k8s作为service discovery的后端
        /svc        => /iface/http;  # 访问名为http的service port

$ kubectl create -f linkerd-config.yaml

1. 创建service：

$ cat linkerd-svc.yaml # 删除了官方文档中的 loadBalance 类型字段
kind: Service
apiVersion: v1
metadata:
  name: hello
spec:
  selector:
    app: hello
  ports:
  - name: ext
    port: 80
    targetPort: 8080  # 访问linkerd
  - name: http
    port: 8081        # 定义名为http的service port，直接访问hello-world
    targetPort: 80
  - name: admin       # 访问linkerd的admin页面
    port: 9990

$ kubectl create -f linkerd-svc.yaml

1. 创建RC

$ cat linkerd-rc.yaml
kind: ReplicationController
apiVersion: v1
metadata:
  name: hello
spec:
  replicas: 1
  selector:
    app: hello
  template:
    metadata:
      labels:
        app: hello
    spec:
      dnsPolicy: ClusterFirst
      volumes:
      - name: linkerd-config
        configMap:
          name: "linkerd-config"
      containers:
      - name: hello
        image: dockercloud/hello-world:latest
        ports:
        - name: http
          containerPort: 80

      - name: linkerd
        image: buoyantio/linkerd:latest
        args:
        - "/io.buoyant/linkerd/config/config.yaml"
        ports:
        - name: ext
          containerPort: 8080
        - name: admin
          containerPort: 9990
        volumeMounts:
        - name: "linkerd-config"
          mountPath: "/io.buoyant/linkerd/config"
          readOnly: true

      - name: kubectl
        image: buoyantio/kubectl:1.2.3
        args:
        - "proxy"
        - "-p"
        - "8001"

$ kubectl create -f linkerd-rc.yaml

$ kubectl get svc
NAME                  CLUSTER-IP       EXTERNAL-IP   PORT(S)                    AGE
deployment-demo-svc   10.254.137.54    <none>        80/TCP                     1d
hello                 10.254.70.169    <none>        80/TCP,8081/TCP,9990/TCP   2h
kubernetes            10.254.0.1       <none>        443/TCP                    101d
my-nginx              10.254.0.159     <none>        80/TCP                     14d
rc-demo-svc           10.254.144.118   <none>        80/TCP                     1d
statefulset-nginx     None             <none>        80/TCP                     21h
world-v1              None             <none>        7778/TCP                   2h

$ kubectl describe svc hello
Name:                   hello
Namespace:              default
Labels:                 <none>
Selector:               app=hello
Type:                   ClusterIP
IP:                     10.254.70.169
Port:                   **ext     80/TCP**
Endpoints:              172.30.19.21:8080,172.30.19.25:8080,172.30.19.26:8080
Port:                   **http    8081/TCP**
Endpoints:              172.30.19.21:80,172.30.19.25:80,172.30.19.26:80
Port:                   **admin   9990/TCP**
Endpoints:              172.30.19.21:9990,172.30.19.25:9990,172.30.19.26:9990
Session Affinity:       None
No events.

$ kubectl describe svc my-nginx  # my-nginx的80端口没有命名，故不在bound-names.json的结果中
Name:                   my-nginx
Namespace:              default
Labels:                 run=my-nginx
Selector:               run=my-nginx
Type:                   ClusterIP
IP:                     10.254.0.159
Port:                  **<unset> 80/TCP**
Endpoints:              172.30.19.2:80,172.30.19.6:80
Session Affinity:       None
No events.


使用io.l5d.k8s namerd自动发现kubernetes中所有**命名端口的服务**
$ curl -s http://10.254.70.169:9990/bound-names.json | jq '[.]'
[
  [
    "/#/io.l5d.k8s/default/admin/hello",
    "/#/io.l5d.k8s/default/http/hello",
    "/#/io.l5d.k8s/default/ext/hello",
    "/#/io.l5d.k8s/default/http/world-v1",
    "/#/io.l5d.k8s/default/web/statefulset-nginx",
    "/#/io.l5d.k8s/default/https/kubernetes"
  ]
]

各项的格式："/#/io.l5d.k8s/<namespace>/<portName>/<svcName>"

浏览器访问admin服务：
    http://10.64.3.7:8080/api/v1/proxy/namespaces/default/services/hello:9990/

在node上发起web请求，观察admin页面的统计：
    $ while true ;do http_proxy=http://10.254.70.169:80 curl -s http://hello >/dev/null;done
    $ while true ;do http_proxy=http://10.254.70.169:80 curl -s http://world-v1 >/dev/null;done

或者：
    $ while true ;do curl -sI  -H 'Host: hello' http://10.254.70.169:80; done
    $ while true ;do curl -sI  -H 'Host: world-v1' http://10.254.70.169:80;done

之所以指定Host: hello，是因为linkerd自动发现了名为hello的service

$  kubectl scale rc hello --replicas=3
replicationcontroller "hello" scaled

$ kubectl get pods|grep hello
hello-bk1x2                        3/3       Running   0          18s
hello-kx8c2                        3/3       Running   0          1h
hello-m4c68                        3/3       Running   0          18s


访问admin接口：
$ curl http://10.254.70.169:9990/admin
/
/admin/announcer
/admin/contention
/admin/files/
/admin/lint
/admin/lint.json
/admin/metrics.json
/admin/metrics/usage
/admin/ping
/admin/pprof/contention
/admin/pprof/heap
/admin/pprof/profile
/admin/registry.json
/admin/server_info
/admin/shutdown
/admin/threads
/admin/threads.json
/admin/tracing
/bound-names.json
/config.json
/delegator
/delegator.json
/favicon.ico
/files/
/help
/logging
/logging.json




使用io.l5d.k8s namerd自动发现kubernetes中所有**命名端口的服务**
$ curl -s http://10.254.70.169:9990/bound-names.json | jq '[.]'
[
  [
    "/#/io.l5d.k8s/default/admin/hello",
    "/#/io.l5d.k8s/default/http/hello",
    "/#/io.l5d.k8s/default/ext/hello",
    "/#/io.l5d.k8s/default/http/world-v1",
    "/#/io.l5d.k8s/default/web/statefulset-nginx",
    "/#/io.l5d.k8s/default/https/kubernetes"
  ]
]

各项的格式："/#/io.l5d.k8s/<namespace>/<portName>/<svcName>"


 