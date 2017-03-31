[root@tjwq01-sys-bs003007 linkerd]# cat linkerd-metric.yaml

# runs linkerd in a daemonset, in linker-to-linker mode
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: l5d-config
data:
  config.yaml: |-
    admin:
      port: 9990

    namers:
    - kind: io.l5d.k8s
      experimental: true
      host: localhost
      port: 8001

    telemetry:
    - kind: io.l5d.prometheus
    - kind: io.l5d.recentRequests
      sampleRate: 0.25
    - kind: io.l5d.tracelog
      sampleRate: 0.2
      level: ALL

    usage:
      orgId: linkerd-examples-daemonset

    routers:
    - protocol: http
      label: outgoing
      dtab: |
        /srv        => /#/io.l5d.k8s/default/http;
        /host       => /srv;
        /svc        => /host;
        /host/world => /srv/world-v1;
      interpreter:
        kind: default
        transformers:
        - kind: io.l5d.k8s.daemonset
          namespace: default
          port: incoming
          service: l5d
    servers:
      - port: 4140
        ip: 0.0.0.0
      responseClassifier:
        kind: io.l5d.retryableRead5XX

    - protocol: http
      label: incoming
      dtab: |
        /srv        => /#/io.l5d.k8s/default/http;
        /host       => /srv;
        /svc        => /host;
        /host/world => /srv/world-v1;
      interpreter:
        kind: default
        transformers:
        - kind: io.l5d.k8s.localnode
      servers:
      - port: 4141
        ip: 0.0.0.0
---
apiVersion: extensions/v1beta1
kind: DaemonSet
metadata:
  labels:
    app: l5d
  name: l5d
spec:
  template:
    metadata:
      labels:
        app: l5d
    spec:
      volumes:
      - name: l5d-config
        configMap:
          name: "l5d-config"
      containers:
      - name: l5d
        image: buoyantio/linkerd:0.9.1
        env:
        - name: POD_IP
        valueFrom:
            fieldRef:
              fieldPath: status.podIP
        args:
        - /io.buoyant/linkerd/config/config.yaml
        ports:
        - name: outgoing
          containerPort: 4140
          hostPort: 4140          # 注意是hostPort，这意味着可以使用 nodeIP:hostPort的方式访问该pod服务；
        - name: incoming
          containerPort: 4141
        - name: admin
          containerPort: 9990
        volumeMounts:
        - name: "l5d-config"
          mountPath: "/io.buoyant/linkerd/config"
          readOnly: true

      - name: kubectl
        image: buoyantio/kubectl:v1.4.0
        args:
        - "proxy"
        - "-p"
        - "8001"
---
apiVersion: v1
kind: Service
metadata:
  name: l5d
spec:
  selector:
    app: l5d
  ports:
  - name: outgoing
    port: 4140
  - name: incoming
    port: 4141
  - name: admin
    port: 9990


[root@tjwq01-sys-bs003007 linkerd]# cat linkerd-hello-world.yaml
---
apiVersion: v1
kind: ReplicationController
metadata:
  name: hello
spec:
  replicas: 3
  selector:
    app: hello
  template:
    metadata:
      labels:
        app: hello
    spec:
      dnsPolicy: ClusterFirst
      containers:
      - name: service
        image: buoyantio/helloworld:0.1.2
        env:
        - name: NODE_NAME
          valueFrom:
            fieldRef:
              fieldPath: spec.nodeName   # 获取nodeName
        - name: POD_IP
          valueFrom:
            fieldRef:
              fieldPath: status.podIP
        - name: http_proxy
          value: $(NODE_NAME):4140  # 访问node上监听hostPort 4140的linkerd
        args:
        - "-addr=:7777"
        - "-text=Hello"
        - "-target=world"
        ports:
        - name: service
          containerPort: 7777
---
apiVersion: v1
kind: Service
metadata:
  name: hello
spec:
  selector:
    app: hello
  clusterIP: None
  ports:
  - name: http
    port: 7777
---
apiVersion: v1
kind: ReplicationController
metadata:
  name: world-v1
spec:
  replicas: 3
  selector:
    app: world-v1
  template:
    metadata:
      labels:
        app: world-v1
    spec:
      dnsPolicy: ClusterFirst
      containers:
      - name: service
        image: buoyantio/helloworld:0.1.2
        env:
        - name: POD_IP
          valueFrom:
            fieldRef:
              fieldPath: status.podIP
        - name: TARGET_WORLDg
          value: world
        args:
        - "-addr=:7778"
        ports:
        - name: service
          containerPort: 7778
---
apiVersion: v1
kind: Service
metadata:
  name: world-v1
spec:
  selector:
    app: world-v1
  clusterIP: None
  ports:
  - name: http
    port: 7778

在部署时一台机器的pod一直处于pending状态：
$ kubectl describe pods hello-g4z29 |tail -5
Events:
  FirstSeen     LastSeen        Count   From                    SubObjectPath   Type            Reason                  Message
  ---------     --------        -----   ----                    -------------   --------        ------                  -------
  2m            2m              1       {default-scheduler }                    Normal          Scheduled               Successfully assigned hello-g4z29 to 10.64.3.7
  2m            2m              1       {kubelet 10.64.3.7}                     Warning         FailedValidation        Error validating pod hello-g4z29.default from api, ignoring: spec.containers[0].env[0].valueFrom.fieldRef.fieldPath: Unsupported value: "spec.nodeName": supported values: metadata.name, metadata.namespace, status.podIP

这是由于kuelet的版本太低所致，上面的fieldPath: spec.nodeName 是 Kublet 1.4以后才支持的Downward API.:
$ /usr/bin/kubelet --version
Kubernetes v1.3.0
升级kublet版本到 1.5.5 后，重启kubelet进程解决；


分别
[root@tjwq01-sys-bs003007 linkerd]# http_proxy=10.64.3.7:4140 curl -s http://hello
Hello (172.30.60.22) world (172.30.83.7)!![root@tjwq01-sys-bs003007 linkerd]# http_proxy=10.64.3.7:4140 curl -s http://hello
Hello (172.30.83.4) world (172.30.60.21)!![root@tjwq01-sys-bs003007 linkerd]# http_proxy=10.64.3.7:4140 curl -s http://hello
Hello (172.30.83.4) world (172.30.83.7)!![root@tjwq01-sys-bs003007 linkerd]# http_proxy=10.64.3.7:4140 curl -s http://hello
Hello (172.30.60.22) world (172.30.60.21)!![root@tjwq01-sys-bs003007 linkerd]# http_proxy=10.64.3.7:4140 curl -s http://hello
Hello (172.30.60.22) world (172.30.60.21)!![root@tjwq01-sys-bs003007 linkerd]# http_proxy=10.64.3.7:4140 curl -s http://hello
Hello (172.30.83.5) world (172.30.60.21)!![root@tjwq01-sys-bs003007 linkerd]# http_proxy=10.64.3.7:4140 curl -s http://hello
Hello (172.30.60.22) world (172.30.83.6)!![root@tjwq01-sys-bs003007 linkerd]# http_proxy=10.64.3.8:4140 curl -s http://hello
Hello (172.30.83.5) world (172.30.83.7)!![root@tjwq01-sys-bs003007 linkerd]# http_proxy=10.64.3.8:4140 curl -s http://hello
Hello (172.30.60.22) world (172.30.60.21)!![root@tjwq01-sys-bs003007 linkerd]# http_proxy=10.64.3.8:4140 curl -s http://hello
Hello (172.30.60.22) world (172.30.83.7)!![root@tjwq01-sys-bs003007 linkerd]# http_proxy=10.64.3.8:4140 curl -s http://hello
Hello (172.30.83.4) world (172.30.83.7)!![root@tjwq01-sys-bs003007 linkerd]#



[root@tjwq01-sys-bs003007 linkerd]# http_proxy=10.64.3.7:4140 curl -s http://hello
Hello (172.30.60.22) world (172.30.60.21)!![root@tjwq01-sys-bs003007 linkerd]# http_proxy=10.64.3.7:4140 curl -s http://hello
Hello (172.30.83.4) world (172.30.83.6)!![root@tjwq01-sys-bs003007 linkerd]#
[root@tjwq01-sys-bs003007 linkerd]#
[root@tjwq01-sys-bs003007 linkerd]# http_proxy=10.64.3.7:4140 curl -s http://hello
Hello (172.30.60.22) world (172.30.60.21)!![root@tjwq01-sys-bs003007 linkerd]#