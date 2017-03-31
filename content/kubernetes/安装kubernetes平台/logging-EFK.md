http://tonybai.com/2017/03/03/implement-kubernetes-cluster-level-logging-with-fluentd-and-elasticsearch-stack/

$ pwd
/root/kubernetes/cluster/addons/fluentd-elasticsearch

修改image文件，使用docker hub上的image：
$ diff es-controller.yaml.orig es-controller.yaml
9a10
>     addonmanager.kubernetes.io/mode: Reconcile
23c24
<       - image: gcr.io/google_containers/elasticsearch:v2.4.1
---
>       - image: bigwhite/elasticsearch:v2.4.1-1
40a42,46
>         env:
>         - name: "NAMESPACE"
>           valueFrom:
>             fieldRef:
>               fieldPath: metadata.namespace

$  diff kibana-controller.yaml.orig kibana-controller.yaml
8a9
>     addonmanager.kubernetes.io/mode: Reconcile
21c22
<         image: gcr.io/google_containers/kibana:v4.6.1
---
>         image: bigwhite/kibana:v4.6.1-1

创建 fluentd 失败：
$ kubectl create -f fluentd-es-ds.yaml --record
error: error validating "fluentd-es-ds.yaml": error validating data: found invalid field tolerations for v1.PodSpec; if you choose to ignore these errors, turn validation off with --validate=false

这是由于spec.tolerations is the new feature of v1.6.x, not for v1.5.x. you should remove it for v1.5.x.
解决方法是 注释掉 tolerations相关的内容，然后重新create；

创建fluentd daemonset成功，但是并没有创建pods
$ kubectl get daemonset -n kube-system
NAME               DESIRED   CURRENT   READY     NODE-SELECTOR                              AGE
fluentd-es-v1.22   0         0         0         beta.kubernetes.io/fluentd-ds-ready=true   1m

这是由于该daemonset只在设置了 beta.kubernetes.io/fluentd-ds-ready=true 的node上执行；
$ kubectl describe daemonset -n kube-system
Name:           fluentd-es-v1.22
Image(s):       bigwhite/fluentd-elasticsearch:1.22
Selector:       k8s-app=fluentd-es,kubernetes.io/cluster-service=true,version=v1.22
Node-Selector:  beta.kubernetes.io/fluentd-ds-ready=true
Labels:         addonmanager.kubernetes.io/mode=Reconcile
                k8s-app=fluentd-es
                kubernetes.io/cluster-service=true
                version=v1.22
Desired Number of Nodes Scheduled: 0
Current Number of Nodes Scheduled: 0
Number of Nodes Misscheduled: 0
Pods Status:    0 Running / 0 Waiting / 0 Succeeded / 0 Failed
No events.

解决办法是给所有node打上beta.kubernetes.io/fluentd-ds-ready=true的标签
$ kubectl label nodes 10.64.3.7 beta.kubernetes.io/fluentd-ds-ready=true
node "10.64.3.7" labeled
$ kubectl get  nodes 10.64.3.7 -o yaml
apiVersion: v1
kind: Node
metadata:
  annotations:
    volumes.kubernetes.io/controller-managed-attach-detach: "true"
  creationTimestamp: 2017-03-25T15:51:13Z
  labels:
    beta.kubernetes.io/arch: amd64
    beta.kubernetes.io/fluentd-ds-ready: "true"
    beta.kubernetes.io/os: linux
    kubernetes.io/hostname: 10.64.3.7
...

daemonset正确地在node上启动了pod
$ kubectl get daemonset -n kube-system
NAME               DESIRED   CURRENT   READY     NODE-SELECTOR                              AGE
fluentd-es-v1.22   1         1         1         beta.kubernetes.io/fluentd-ds-ready=true   10m
$ kubectl describe daemonset -n kube-system
Name:           fluentd-es-v1.22
Image(s):       bigwhite/fluentd-elasticsearch:1.22
Selector:       k8s-app=fluentd-es,kubernetes.io/cluster-service=true,version=v1.22
Node-Selector:  beta.kubernetes.io/fluentd-ds-ready=true
Labels:         addonmanager.kubernetes.io/mode=Reconcile
                k8s-app=fluentd-es
                kubernetes.io/cluster-service=true
                version=v1.22
Desired Number of Nodes Scheduled: 1
Current Number of Nodes Scheduled: 1
Number of Nodes Misscheduled: 0
Pods Status:    1 Running / 0 Waiting / 0 Succeeded / 0 Failed
Events:
  FirstSeen     LastSeen        Count   From            SubObjectPath   Type            Reason                  Message
  ---------     --------        -----   ----            -------------   --------        ------                  -------
  1m            1m              1       {daemon-set }                   Normal          SuccessfulCreate        Created pod: fluentd-es-v1.22-36j61

$ kubectl logs  fluentd-es-v1.22-36j61 -n kube-system
/opt/td-agent/embedded/lib/ruby/gems/2.1.0/gems/json-1.8.1/lib/json/version.rb:3: warning: already initialized constant JSON::VERSION
/opt/td-agent/embedded/lib/ruby/2.1.0/json/version.rb:3: warning: previous definition of VERSION was here
/opt/td-agent/embedded/lib/ruby/gems/2.1.0/gems/json-1.8.1/lib/json/version.rb:4: warning: already initialized constant JSON::VERSION_ARRAY
/opt/td-agent/embedded/lib/ruby/2.1.0/json/version.rb:4: warning: previous definition of VERSION_ARRAY was here
/opt/td-agent/embedded/lib/ruby/gems/2.1.0/gems/json-1.8.1/lib/json/version.rb:5: warning: already initialized constant JSON::VERSION_MAJOR
/opt/td-agent/embedded/lib/ruby/2.1.0/json/version.rb:5: warning: previous definition of VERSION_MAJOR was here

创建elasticsearch
$ kubectl create -f es-controller.yaml
replicationcontroller "elasticsearch-logging-v1" created
$ kubectl create -f es-service.yaml
service "elasticsearch-logging" created
$ kubectl get pods -n kube-system
NAME                                    READY     STATUS    RESTARTS   AGE
elasticsearch-logging-v1-cm1xp          1/1       Running   0          10s
elasticsearch-logging-v1-drc8d          1/1       Running   0          10s

$  kubectl logs  elasticsearch-logging-v1-cm1xp -n kube-system -f
[2017-03-26 15:00:33,559][INFO ][node                     ] [elasticsearch-logging-v1-cm1xp] version[2.4.1], pid[33], build[c67dc32/2016-09-27T18:57:55Z]
[2017-03-26 15:00:33,560][INFO ][node                     ] [elasticsearch-logging-v1-cm1xp] initializing ...
[2017-03-26 15:00:34,665][INFO ][plugins                  ] [elasticsearch-logging-v1-cm1xp] modules [reindex, lang-expression, lang-groovy], plugins [], sites []
[2017-03-26 15:00:34,763][INFO ][env                      ] [elasticsearch-logging-v1-cm1xp] using [1] data paths, mounts [[/data (/dev/sda2)]], net usable_space [338.9gb], net total_space [549.5gb], spins? [possibly], types [ext4]
[2017-03-26 15:00:34,763][INFO ][env                      ] [elasticsearch-logging-v1-cm1xp] heap size [989.8mb], compressed ordinary object pointers [true]
[2017-03-26 15:00:40,078][INFO ][node                     ] [elasticsearch-logging-v1-cm1xp] initialized
[2017-03-26 15:00:40,078][INFO ][node                     ] [elasticsearch-logging-v1-cm1xp] starting ...
[2017-03-26 15:00:40,259][INFO ][transport                ] [elasticsearch-logging-v1-cm1xp] publish_address {172.30.19.11:9300}, bound_addresses {[::]:9300}
[2017-03-26 15:00:40,263][INFO ][discovery                ] [elasticsearch-logging-v1-cm1xp] kubernetes-logging/S7rl46BRRV-12ypRiunovg
[2017-03-26 15:00:43,345][INFO ][cluster.service          ] [elasticsearch-logging-v1-cm1xp] detected_master {elasticsearch-logging-v1-drc8d}{I5XXU5FHTOGO4gx7y_Wrvg}{172.30.19.10}{172.30.19.10:9300}{master=true}, added {{elasticsearch-logging-v1-drc8d}{I5XXU5FHTOGO4gx7y_Wrvg}{172.30.19.10}{172.30.19.10:9300}{master=true},}, reason: zen-disco-receive(from master [{elasticsearch-logging-v1-drc8d}{I5XXU5FHTOGO4gx7y_Wrvg}{172.30.19.10}{172.30.19.10:9300}{master=true}])
[2017-03-26 15:00:43,455][INFO ][http                     ] [elasticsearch-logging-v1-cm1xp] publish_address {172.30.19.11:9200}, bound_addresses {[::]:9200}
[2017-03-26 15:00:43,455][INFO ][node                     ] [elasticsearch-logging-v1-cm1xp] started

创建 kibana
$ kubectl create -f kibana-controller.yaml
deployment "kibana-logging" created
$ kubectl create -f kibana-service.yaml
service "kibana-logging" created
$ kubectl get pods -n kube-system |grep kibana
kibana-logging-561879838-t9h4v          1/1       Running   0          6s


等待一段时间(10~20分钟左右)，kibana启动完成：
$ kubectl logs  kibana-logging-561879838-t9h4v -n kube-system -f
ELASTICSEARCH_URL=http://elasticsearch-logging:9200
server.basePath: /api/v1/proxy/namespaces/kube-system/services/kibana-logging
{"type":"log","@timestamp":"2017-03-26T15:03:07Z","tags":["info","optimize"],"pid":7,"message":"Optimizing and caching bundles for kibana and statusPage. This
 may take a few minutes"}
{"type":"log","@timestamp":"2017-03-26T15:17:51Z","tags":["info","optimize"],"pid":7,"message":"Optimization of bundles for kibana and statusPage complete in
884.11 seconds"}
{"type":"log","@timestamp":"2017-03-26T15:17:52Z","tags":["status","plugin:kibana@1.0.0","info"],"pid":7,"state":"green","message":"Status changed from uninit
ialized to green - Ready","prevState":"uninitialized","prevMsg":"uninitialized"}
{"type":"log","@timestamp":"2017-03-26T15:17:53Z","tags":["status","plugin:elasticsearch@1.0.0","info"],"pid":7,"state":"yellow","message":"Status changed fro
m uninitialized to yellow - Waiting for Elasticsearch","prevState":"uninitialized","prevMsg":"uninitialized"}


$ kubectl cluster-info
Kubernetes master is running at http://10.64.3.7:8080
Elasticsearch is running at http://10.64.3.7:8080/api/v1/proxy/namespaces/kube-system/services/elasticsearch-logging
Heapster is running at http://10.64.3.7:8080/api/v1/proxy/namespaces/kube-system/services/heapster
Kibana is running at http://10.64.3.7:8080/api/v1/proxy/namespaces/kube-system/services/kibana-logging
KubeDNS is running at http://10.64.3.7:8080/api/v1/proxy/namespaces/kube-system/services/kube-dns
kubernetes-dashboard is running at http://10.64.3.7:8080/api/v1/proxy/namespaces/kube-system/services/kubernetes-dashboard
monitoring-grafana is running at http://10.64.3.7:8080/api/v1/proxy/namespaces/kube-system/services/monitoring-grafana
monitoring-influxdb is running at http://10.64.3.7:8080/api/v1/proxy/namespaces/kube-system/services/monitoring-influxdb

然后访问：http://10.64.3.7:8080/api/v1/proxy/namespaces/kube-system/services/kibana-logging/app/kibana

在 Settings -》Indices 页面创建一个index（相当于mysql中的一个database）：
取消“Index contains time-based events”，pattern中使用默认的 logstash-*，然后点击“Create”即可创建一个Index。

点击页面上的”Setting” -> “Status”，可以查看当前elasticsearch logging的整体状态：
创建Index后，可以在Discover下看到ElasticSearch logging中汇聚的日志；
另外ElasticSearch logging默认挂载的volume是emptyDir，实验用可以。但要部署在生产环境，必须换成Persistent Volume，比如：CephRBD。