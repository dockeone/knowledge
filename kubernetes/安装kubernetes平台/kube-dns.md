<!-- toc -->

$ kubectl get svc --all-namespaces
NAMESPACE     NAME         CLUSTER-IP   EXTERNAL-IP   PORT(S)         AGE
default       kubernetes   10.254.0.1   <none>        443/TCP         86d

$ kubectl create -f skydns-svc.yaml
service "kube-dns" created

$ kubectl get svc --namespace=kube-system
NAME       CLUSTER-IP   EXTERNAL-IP   PORT(S)         AGE
kube-dns   10.254.0.2   <none>        53/UDP,53/TCP   35s
$ kubectl describe svc --namespace=kube-system
Name:                   kube-dns
Namespace:              kube-system
Labels:                 k8s-app=kube-dns
                        kubernetes.io/cluster-service=true
                        kubernetes.io/name=KubeDNS
Selector:               k8s-app=kube-dns
Type:                   ClusterIP
IP:                     10.254.0.2
Port:                   dns     53/UDP
Endpoints:              <none>
Port:                   dns-tcp 53/TCP
Endpoints:              <none>
Session Affinity:       None
No events.

$ kubectl create -f skydns-rc.yaml
deployment "kube-dns" created
$ kubectl get deployment --namespace=kube-system
NAME       DESIRED   CURRENT   UP-TO-DATE   AVAILABLE   AGE
kube-dns   1         1         1            1           8s
$ kubectl get pods --namespace=kube-system
NAME                        READY     STATUS    RESTARTS   AGE
kube-dns-2200864676-cs7ff   4/4       Running   0          11s
$ kubectl describe deployment kube-dns --namespace=kube-system
Name:                   kube-dns
Namespace:              kube-system
CreationTimestamp:      Mon, 13 Mar 2017 17:26:31 +0800
Labels:                 k8s-app=kube-dns
                        kubernetes.io/cluster-service=true
Selector:               k8s-app=kube-dns
Replicas:               1 updated | 1 total | 1 available | 0 unavailable
StrategyType:           RollingUpdate
MinReadySeconds:        0
RollingUpdateStrategy:  0 max unavailable, 10% max surge
Conditions:
  Type          Status  Reason
  ----          ------  ------
  Available     True    MinimumReplicasAvailable
OldReplicaSets: <none>
NewReplicaSet:  kube-dns-2200864676 (1/1 replicas created)
Events:
  FirstSeen     LastSeen        Count   From                            SubObjectPath   Type            Reason                  Message
  ---------     --------        -----   ----                            -------------   --------        ------                  -------
  1m            1m              1       {deployment-controller }                        Normal          ScalingReplicaSet       Scaled up replica set kube-dns-2200864676 to 1
Name:           kube-dns-2200864676-cs7ff
Namespace:      kube-system
Node:           127.0.0.1/127.0.0.1
Start Time:     Mon, 13 Mar 2017 17:26:31 +0800
Labels:         k8s-app=kube-dns
                pod-template-hash=2200864676
Status:         Running
IP:             172.30.19.2
Controllers:    ReplicaSet/kube-dns-2200864676
Containers:
  kubedns:
    Container ID:       docker://52fbd84d4b15c137580e2908d034f13d88649c77d4b9ed3d11b75f91a72cd268
    Image:              ist0ne/kubedns-amd64:1.9
    Image ID:           docker-pullable://docker.io/ist0ne/kubedns-amd64@sha256:c12a28611a6883a2879b8f8dae6a7b088082d40262a116be51c8ee0b69cf91e0
    Ports:              10053/UDP, 10053/TCP, 10055/TCP
    Args:
      --domain=cluster.local.
      --dns-port=10053
      --config-map=kube-dns
      --v=0
    Limits:
      memory:   170Mi
    Requests:
      cpu:              100m
      memory:           70Mi
    State:              Running
      Started:          Mon, 13 Mar 2017 17:26:32 +0800
    Ready:              True
    Restart Count:      0
    Liveness:           http-get http://:8080/healthz-kubedns delay=60s timeout=5s period=10s #success=1 #failure=5
    Readiness:          http-get http://:8080/readiness delay=3s timeout=5s period=10s #success=1 #failure=3
    Volume Mounts:
      /var/run/secrets/kubernetes.io/serviceaccount from default-token-q95f8 (ro)
    Environment Variables:
      PROMETHEUS_PORT:  10055
  dnsmasq:
    Container ID:       docker://f568ccdc833b896a294f078d651af340dbc04bf6068d4a40984e514c72843c67
    Image:              ist0ne/kube-dnsmasq-amd64:1.4
    Image ID:           docker-pullable://docker.io/ist0ne/kube-dnsmasq-amd64@sha256:e49f231477ba296992515e5824c3015692227abb8f5c6fa08557d4d83abe8058
    Ports:              53/UDP, 53/TCP
    Args:
      --cache-size=1000
            --no-resolv
      --server=127.0.0.1#10053
      --log-facility=-
    Requests:
      cpu:              150m
      memory:           10Mi
    State:              Running
      Started:          Mon, 13 Mar 2017 17:26:32 +0800
    Ready:              True
    Restart Count:      0
    Liveness:           http-get http://:8080/healthz-dnsmasq delay=60s timeout=5s period=10s #success=1 #failure=5
    Volume Mounts:
      /var/run/secrets/kubernetes.io/serviceaccount from default-token-q95f8 (ro)
    Environment Variables:      <none>
  dnsmasq-metrics:
    Container ID:       docker://3398103e3a76ae7846a57979eb8ecf4a6bdf88355d15160690de5e74ec4aaa1d
    Image:              ist0ne/dnsmasq-metrics-amd64:1.0
    Image ID:           docker-pullable://docker.io/ist0ne/dnsmasq-metrics-amd64@sha256:5ae7e3a3a2ac3f08352da3f25b9fc12482ef9798f52b0cc847c3d1258beaa7bb
    Port:               10054/TCP
    Args:
      --v=2
      --logtostderr
    Requests:
      memory:           10Mi
    State:              Running
      Started:          Mon, 13 Mar 2017 17:26:33 +0800
    Ready:              True
    Restart Count:      0
    Liveness:           http-get http://:10054/metrics delay=60s timeout=5s period=10s #success=1 #failure=5
    Volume Mounts:
      /var/run/secrets/kubernetes.io/serviceaccount from default-token-q95f8 (ro)
    Environment Variables:      <none>
  healthz:
    Container ID:       docker://6d20d58bad7640157af0b98b00688e09afff02e316368374d23605a13416cb60
    Image:              ist0ne/exechealthz-amd64:1.2
    Image ID:           docker-pullable://docker.io/ist0ne/exechealthz-amd64@sha256:67e6a74ee4242c4891c4be79da87b2e21bf7fb3645ad10fef3022d441aea463d
    Port:               8080/TCP
    Args:
      --cmd=nslookup kubernetes.default.svc.cluster.local 127.0.0.1 >/dev/null
      --url=/healthz-dnsmasq
      --cmd=nslookup kubernetes.default.svc.cluster.local 127.0.0.1:10053 >/dev/null
      --url=/healthz-kubedns
            --port=8080
      --quiet
    Limits:
      memory:   50Mi
    Requests:
      cpu:              10m
      memory:           50Mi
    State:              Running
      Started:          Mon, 13 Mar 2017 17:26:33 +0800
    Ready:              True
    Restart Count:      0
    Volume Mounts:
      /var/run/secrets/kubernetes.io/serviceaccount from default-token-q95f8 (ro)
    Environment Variables:      <none>
Conditions:
  Type          Status
  Initialized   True
  Ready         True
  PodScheduled  True
Volumes:
  default-token-q95f8:
    Type:       Secret (a volume populated by a Secret)
    SecretName: default-token-q95f8
QoS Class:      Burstable
Tolerations:    CriticalAddonsOnly=:Exists
Events:
  FirstSeen     LastSeen        Count   From                    SubObjectPath                           Type            Reason          Message
  ---------     --------        -----   ----                    -------------                           --------        ------          -------
  1m            1m              1       {default-scheduler }                                            Normal          Scheduled       Successfully assigned kube-dns-2200864676-cs7ff to 127.0.0.1
  1m            1m              1       {kubelet 127.0.0.1}     spec.containers{kubedns}                Normal          Pulled          Container image "ist0ne/kubedns-amd64:1.9" already present on machine
  1m            1m              1       {kubelet 127.0.0.1}     spec.containers{kubedns}                Normal          Created         Created container with docker id 52fbd84d4b15
  1m            1m              1       {kubelet 127.0.0.1}     spec.containers{kubedns}                Normal          Started         Started container with docker id 52fbd84d4b15
  1m            1m              1       {kubelet 127.0.0.1}     spec.containers{dnsmasq}                Normal          Pulled          Container image "ist0ne/kube-dnsmasq-amd64:1.4" already present on machine
  1m            1m              1       {kubelet 127.0.0.1}     spec.containers{dnsmasq}                Normal          Created         Created container with docker id f568ccdc833b
  1m            1m              1       {kubelet 127.0.0.1}     spec.containers{dnsmasq}                Normal          Started         Started container with docker id f568ccdc833b
    1m            1m              1       {kubelet 127.0.0.1}     spec.containers{dnsmasq}                Normal          Started         Started container with docker id f568ccdc833b
  1m            1m              1       {kubelet 127.0.0.1}     spec.containers{dnsmasq-metrics}        Normal          Pulled          Container image "ist0ne/dnsmasq-metrics-amd64:1.0" already present on machine
  1m            1m              1       {kubelet 127.0.0.1}     spec.containers{dnsmasq-metrics}        Normal          Created         Created container with docker id 3398103e3a76
  1m            1m              1       {kubelet 127.0.0.1}     spec.containers{dnsmasq-metrics}        Normal          Started         Started container with docker id 3398103e3a76
  1m            1m              1       {kubelet 127.0.0.1}     spec.containers{healthz}                Normal          Pulled          Container image "ist0ne/exechealthz-amd64:1.2" already present on machine
  1m            1m              1       {kubelet 127.0.0.1}     spec.containers{healthz}                Normal          Created         Created container with docker id 6d20d58bad76
  1m            1m              1       {kubelet 127.0.0.1}     spec.containers{healthz}                Normal          Started         Started container with docker id 6d20d58bad76
  41s           1s              10      {kubelet 127.0.0.1}     spec.containers{kubedns}                Warning         Unhealthy       Liveness probe failed: HTTP probe failed with statuscode: 503


# kubedns的确没有起来

$ docker ps
CONTAINER ID        IMAGE                                                        COMMAND                  CREATED             STATUS              PORTS               NAMES
6d20d58bad76        ist0ne/exechealthz-amd64:1.2                                 "/exechealthz '--cmd="   29 minutes ago      Up 29 minutes                           k8s_healthz.3bde37ea_kube-dns-2200864676-cs7ff_kube-system_24fa1efd-07cf-11e7-8472-8cdcd4b3be48_19a06fd4
3398103e3a76        ist0ne/dnsmasq-metrics-amd64:1.0                             "/dnsmasq-metrics --v"   29 minutes ago      Up 29 minutes                           k8s_dnsmasq-metrics.367b171c_kube-dns-2200864676-cs7ff_kube-system_24fa1efd-07cf-11e7-8472-8cdcd4b3be48_43c77b69
0e889eb3e212        registry.access.redhat.com/rhel7/pod-infrastructure:latest   "/pod"                   29 minutes ago      Up 29 minutes                           k8s_POD.3cb1f050_kube-dns-2200864676-cs7ff_kube-system_24fa1efd-07cf-11e7-8472-8cdcd4b3be48_c621f5dd
e1c36d118af1        nginx:1.7.9                                                  "nginx -g 'daemon off"   4 hours ago         Up 4 hours                              k8s_ceph-busybox2.c343866_ceph-pod2_default_28f1ba80-07b1-11e7-8472-8cdcd4b3be48_61c51adb
a483550bfae7        registry.access.redhat.com/rhel7/pod-infrastructure:latest   "/pod"                   4 hours ago         Up 4 hours                              k8s_POD.ae8ee9ac_ceph-pod2_default_28f1ba80-07b1-11e7-8472-8cdcd4b3be48_df3e6dd3

# 原因是 kubedns通过443端口连接apiserver的secure port，但是serviceaccount没有ca文件，kubedns不能对apiserver的公钥进行验证

$ kubectl logs kube-dns-2200864676-cs7ff kubedns --namespace=kube-system|tail -4
E0313 09:52:30.631501       1 reflector.go:199] pkg/dns/dns.go:148: Failed to list *api.Service: Get https://10.254.0.1:443/api/v1/services?resourceVersion=0: x509: failed to load system roots and no roots provided
E0313 09:52:31.623087       1 reflector.go:199] pkg/dns/dns.go:145: Failed to list *api.Endpoints: Get https://10.254.0.1:443/api/v1/endpoints?resourceVersion=0: x509: failed to load system roots and no roots provided
E0313 09:52:31.629979       1 reflector.go:199] pkg/dns/config/sync.go:114: Failed to list *api.ConfigMap: Get https://10.254.0.1:443/api/v1/namespaces/kube-system/configmaps?fieldSelector=metadata.name%3Dkube-dns&resourceVersion=0: x509: failed to load system roots and no roots provided
E0313 09:52:31.640905       1 reflector.go:199] pkg/dns/dns.go:148: Failed to list *api.Service: Get https://10.254.0.1:443/api/v1/services?resourceVersion=0: x509: failed to load system roots and no roots provided


# 查看 kube-dns 的命令行参数

$ kubectl exec --namespace=kube-system  -i -t kube-dns-2200864676-cs7ff -c kubedns -- /kube-dns --help
error: error executing remote command: error executing command in container: container not found ("kubedns")

$ docker run -it ist0ne/kubedns-amd64 kube-dns --help
Usage of /kube-dns:
      --alsologtostderr                  log to standard error as well as files
      --config-map string                config-map name. If empty, then the config-map will not used. Cannot be  used in conjunction with federations flag. config-map contains dynamically adjustable configuration.
      --config-map-namespace string      namespace for the config-map (default "kube-system")
      --dns-bind-address string          address on which to serve DNS requests. (default "0.0.0.0")
      --dns-port int                     port on which to serve DNS requests. (default 53)
      --domain string                    domain under which to create names (default "cluster.local.")
      --healthz-port int                 port on which to serve a kube-dns HTTP readiness probe. (default 8081)
      --kube-master-url string           URL to reach kubernetes master. Env variables in this flag will be expanded.
      --kubecfg-file string              Location of kubecfg file for access to kubernetes master service; --kube-master-url overrides the URL part of this; if neither this nor --kube-master-url are provided, defaults to service account tokens
      --log-backtrace-at traceLocation   when logging hits line file:N, emit a stack trace (default :0)
      --log-dir string                   If non-empty, write log files in this directory
      --log-flush-frequency duration     Maximum number of seconds between log flushes (default 5s)
      --logtostderr                      log to standard error instead of files (default true)
      --stderrthreshold severity         logs at or above this threshold go to stderr (default 2)
  -v, --v Level                          log level for V logs
      --version version[=true]           Print version information and quit
      --vmodule moduleSpec               comma-separated list of pattern=N settings for file-filtered logging


# 解决办法是：为apiserver、controller、kubelet重新生成ca证书、公私钥

$ curl -L -O https://storage.googleapis.com/kubernetes-release/easy-rsa/easy-rsa.tar.gz
$ mv easy-rsa.tar.gz ~/kube/
$ ls ~/kube/easy-rsa.tar.gz
/root/kube/easy-rsa.tar.gz
$ wget https://raw.githubusercontent.com/GoogleCloudPlatform/kubernetes/v0.21.1/cluster/saltbase/salt/generate-cert/make-ca-cert.sh
$ chmod 775 make-ca-cert.sh
$ export CERT_DIR=/etc/kubernetes/ssl

10.64.3.7为apiserver主机IP，后续为使用server.cert证书的IP和DNS列表，客户端在收到server发送的证书后，会检查server的ip或DNS是否位于指定的列表中，如果不在则拒绝server的证书并报错。
$ bash make-ca-cert.sh 10.64.3.7 IP:10.64.3.7,IP:10.0.0.1,DNS:kubernetes,DNS:kubernetes.default,DNS:kubernetes.default.svc,DNS:kubernetes.default.svc.cluster.local
$ ls /etc/kubernetes/ssl/
ca.crt  kubecfg.crt  kubecfg.key  server.cert  server.key

因为kubedns使用kubernetes服务名访问apiserver，所以必须修改上面的 10.0.0.1为 kubernetes service的cluster ip 10.254.0.1：

$ kubectl get services --all-namespaces
NAMESPACE     NAME         CLUSTER-IP   EXTERNAL-IP   PORT(S)         AGE
default       kubernetes   10.254.0.1   <none>        443/TCP         87d

否则kubedns容器启动后报错如下：

$ kubectl logs kube-dns-2200864676-bpgm2 -c kubedns --namespace=kube-system |tail -1
E0313 14:46:52.433533       1 reflector.go:199] pkg/dns/config/sync.go:114: Failed to list *api.ConfigMap: Get https://10.254.0.1:443/api/v1/namespaces/kube-system/configmaps?fieldSelector=metadata.name%3Dkube-dns&resourceVersion=0: x509: certificate is valid for 10.64.3.7, 10.64.3.7, 10.0.0.1, not 10.254.0.1

$ bash make-ca-cert.sh 10.64.3.7 IP:10.64.3.7,IP:10.254.0.1,DNS:kubernetes,DNS:kubernetes.default,DNS:kubernetes.default.svc,DNS:kubernetes.default.svc.cluster.local
$ ls -l /etc/kubernetes/ssl
total 28
-rw------- 1 root root 1208 Mar 13 22:49 ca.crt
-rw------- 1 root root 4458 Mar 13 22:49 kubecfg.crt
-rw------- 1 root root 1704 Mar 13 22:49 kubecfg.key
-rw------- 1 root root 4852 Mar 13 22:49 server.cert
-rw------- 1 root root 1704 Mar 13 22:49 server.key


配置apiserver的参数：
KUBE_API_ARGS="--bind-address=10.64.3.7 --service_account_key_file=/etc/kubernetes/ssl/server.key --client-ca-file=/etc/kubernetes/ssl/ca.crt --tls-cert-file=/etc/kubernetes/ssl/server.cert --tls-private-key-file=/etc/kubernetes/ssl/server.key"
重启apiserver

配置controller-manager的参数:
KUBE_CONTROLLER_MANAGER_ARGS="--address=127.0.0.1 --service_account_private_key_file=/etc/kubernetes/ssl/server.key --root-ca-file=/etc/kubernetes/ssl/ca.crt"
注意：
1. apiserver的--service_account_key_file和controller-manager的--service_account_private_key_file文件必须一致，否则controller-manager生成的serviceaccount token将不能被apiserver验证通过；
2. controller-manager必须指定--root-ca-file参数，该参数值为签名apiserver公私钥的ca证书，这样在为pod挂载serviceaccount时才会包含ca文件(仅指定
--service_account_private_key_file参数并不会给pod生成和挂载ca文件)；
重启controller-manager

注意：
1. 如果修改了 apiserver 或 controller-manager的 service_account_xxx 参数、--root-ca-file参数，则需要重启这两个进程，同时删除各命名空间名为default的serviceaccount关联的secrets, k8s也会自动生成它们：
https://github.com/kubernetes/kubernetes/issues/10265

$ kubectl get secrets  --all-namespaces
NAMESPACE     NAME                  TYPE                                  DATA
default       default-token-6oahq   kubernetes.io/service-account-token   2
kube-system   default-token-ur28n   kubernetes.io/service-account-token   2

$ kubectl delete secret --namespace=default default-token-6oahq
$ kubectl delete secret --namespace=kube-system  default-token-ur28n

配置kubelet的参数：
KUBELET_ARGS="--tls-cert-file=/etc/kubernetes/ssl/kubecfg.crt --tls-private-key-file=/etc/kubernetes/ssl/kubecfg.key --cluster_dns=10.254.0.2 --cluster_domain=cluster.local"
注意：
1. kubelet不需要对apiserver的公钥进行验证，因为没有指定ca的参数；
2. --cluster_dns参数指定 kubedns service的cluster ip； 必须与kube-system namespace中的kube-dns service cluster ip相同(参考：skydns-svc.yaml)；
3. 必须指定--cluster_domain参数(与skydns-rc.yaml中的--domain参数相同)后，kubelet才在新建pod时才会修改其/etc/resolv.conf文件中的搜索域(只指定--cluster_dns参数，不会在pod中添加skydny配置)；
重启kubelet

## 创建kube-system namespace和kube-dns service
$ diff skydns-svc.yaml.in skydns-svc.yaml
31c13
<   clusterIP: {{ pillar['dns_server'] }}
---
>   clusterIP: 10.254.0.2
$ cat skydns-svc.yaml
apiVersion: v1
kind: Service
metadata:
  name: kube-dns
  namespace: kube-system
  labels:
    k8s-app: kube-dns
    kubernetes.io/cluster-service: "true"
    kubernetes.io/name: "KubeDNS"
spec:
  selector:
    k8s-app: kube-dns
  clusterIP: 10.254.0.2
  ports:
  - name: dns
    port: 53
    protocol: UDP
  - name: dns-tcp
    port: 53
    protocol: TCP
$ kubectl create -f skydns-svc.yaml
上面命令会创建：
1. kube-dns 命名空间；
2. kube-dns命令空间对应的 default serviceaccount和token；

$ kubectl get namespaces
NAME          STATUS    AGE
default       Active    87d
kube-system   Active    1d
$ kubectl get serviceaccount --all-namespaces
NAMESPACE     NAME      SECRETS   AGE
default       default   1         35m
kube-system   default   1         34m
$ kubectl describe serviceaccount default  --namespace=kube-system
Name:           default
Namespace:      kube-system
Labels:         <none>

Image pull secrets:     <none>

Mountable secrets:      default-token-wqs8d

Tokens:                 default-token-wqs8d
$ kubectl get secrets  --namespace=kube-system
NAME                  TYPE                                  DATA      AGE
default-token-wqs8d   kubernetes.io/service-account-token   3         36m
$ kubectl describe secrets  --namespace=kube-system
Name:           default-token-wqs8d
Namespace:      kube-system
Labels:         <none>
Annotations:    kubernetes.io/service-account.name=default
                kubernetes.io/service-account.uid=ec67efc7-07fc-11e7-8101-8cdcd4b3be48

Type:   kubernetes.io/service-account-token

Data
====
namespace:      11 bytes
token:          eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJrdWJlcm5ldGVzL3NlcnZpY2VhY2NvdW50Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9uYW1lc3BhY2UiOiJrdWJlLXN5c3RlbSIsImt1YmVybmV0ZXMuaW8vc2VydmljZWFjY291bnQvc2VjcmV0Lm5hbWUiOiJkZWZhdWx0LXRva2VuLXdxczhkIiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9zZXJ2aWNlLWFjY291bnQubmFtZSI6ImRlZmF1bHQiLCJrdWJlcm5ldGVzLmlvL3NlcnZpY2VhY2NvdW50L3NlcnZpY2UtYWNjb3VudC51aWQiOiJlYzY3ZWZjNy0wN2ZjLTExZTctODEwMS04Y2RjZDRiM2JlNDgiLCJzdWIiOiJzeXN0ZW06c2VydmljZWFjY291bnQ6a3ViZS1zeXN0ZW06ZGVmYXVsdCJ9.NVpIxl8zGBiCJQhKwpvFrxKgMRuM5deTapDi5rwThKVZOBppHJQq72PT0B26LcLOIZ3hECysJpg6E6w7HLzxouycgy9fbrRK6iBBUr8nxw6tV-YARN2hhxZnEpU-j-o6nhxXTqEikd64gKSifZKD7hNSz4c7b2EWwMFU3jiQQNhjA-Ap4b1MNhDcRAliRz7pqyd4Ljvfdvr2ZdXwn_qj7FQaFAHyoIJwjIEndyjcmFTjabfY2PvOv0WjkZ3cFW4dP73RkMlmL5WYbAVga4E2jC2cnu_6odxiRs2eMGoAsOZ9Z92aJmBtAa8QVS3PDa1GnhC853KEq_3ytV3YebJv9w
ca.crt:         1208 bytes


$ diff skydns-rc.yaml.in skydns-rc.yaml|grep -v '#'
51c27
<         image: gcr.io/google_containers/kubedns-amd64:1.9
---
>         image: ist0ne/kubedns-amd64:1.9
53,56d28
74c46
<             port: 8081
---
>             port: 8080
76,77d47
81c51
<         - --domain={{ pillar['dns_domain'] }}.
---
>         - --domain=cluster.local.
87d56
<         {{ pillar['federations_domain_map'] }}
102c71
<         image: gcr.io/google_containers/kube-dnsmasq-amd64:1.4
---
>         image: ist0ne/kube-dnsmasq-amd64:1.4
124d92
130c98
<         image: gcr.io/google_containers/dnsmasq-metrics-amd64:1.0
---
>         image: ist0ne/dnsmasq-metrics-amd64:1.0
151c119
<         image: gcr.io/google_containers/exechealthz-amd64:1.2
---
>         image: ist0ne/exechealthz-amd64:1.2
157,160d124
163c127
<         - --cmd=nslookup kubernetes.default.svc.{{ pillar['dns_domain'] }} 127.0.0.1 >/dev/null
---
>         - --cmd=nslookup kubernetes.default.svc.cluster.local 127.0.0.1 >/dev/null
165c129
<         - --cmd=nslookup kubernetes.default.svc.{{ pillar['dns_domain'] }} 127.0.0.1:10053 >/dev/null
---
>         - --cmd=nslookup kubernetes.default.svc.cluster.local 127.0.0.1:10053 >/dev/null

注意：
1. 由于gcr.io被墙，所以需要替换为 docker 的registery 仓库地址，这里使用ist0ne账户下的同名、同版本image；
2. 为了加快镜像的下载速度，使用daocloud提供的加速器：
$ diff /etc/sysconfig/docker.orig /etc/sysconfig/docker
13c13
< #ADD_REGISTRY='--add-registry registry.access.redhat.com'
---
> ADD_REGISTRY='--add-registry 7e297742.m.daocloud.io --add-registry registry.access.redhat.com'
3. /healthz-dnsmasq对应的端口应该是8080，而非8081，因为healthz指定的是8080端口； FIXME!!!
4. --domain=cluster.local. 注意最后面的点号，需要与kubelet的--cluster_domain参数一致；

## 创建kube-dns deployment

$ kubectl create -f skydns-rc.yaml
$ kubectl get deployment --all-namespaces
NAMESPACE     NAME       DESIRED   CURRENT   UP-TO-DATE   AVAILABLE   AGE
kube-system   kube-dns   1         1         1            1           52m
$ kubectl describe deployment kube-dns --namespace=kube-system
Name:                   kube-dns
Namespace:              kube-system
CreationTimestamp:      Mon, 13 Mar 2017 22:55:05 +0800
Labels:                 k8s-app=kube-dns
                        kubernetes.io/cluster-service=true
Selector:               k8s-app=kube-dns
Replicas:               1 updated | 1 total | 1 available | 0 unavailable
StrategyType:           RollingUpdate
MinReadySeconds:        0
RollingUpdateStrategy:  0 max unavailable, 10% max surge
Conditions:
  Type          Status  Reason
  ----          ------  ------
  Available     True    MinimumReplicasAvailable
OldReplicaSets: <none>
NewReplicaSet:  kube-dns-2200864676 (1/1 replicas created)
Events:
  FirstSeen     LastSeen        Count   From                            SubObjectPath   Type            Reason                  Message
  ---------     --------        -----   ----                            -------------   --------        ------                  -------
  52m           52m             1       {deployment-controller }                        Normal          ScalingReplicaSet       Scaled up replica set kube-dns-2200864676 to 1
  $ kubectl  get pods --namespace=kube-system
NAME                        READY     STATUS    RESTARTS   AGE
kube-dns-2200864676-p2zn6   4/4       Running   0          52m
Name:           kube-dns-2200864676-p2zn6
Namespace:      kube-system
Node:           127.0.0.1/127.0.0.1
Start Time:     Mon, 13 Mar 2017 22:55:05 +0800
Labels:         k8s-app=kube-dns
                pod-template-hash=2200864676
Status:         Running
IP:             172.30.19.2
Controllers:    ReplicaSet/kube-dns-2200864676
Containers:
  kubedns:
    Container ID:       docker://3756f735b64d5437e846cfcd32c95166ef35ff81b2db9074c47507ad9e8e3871
    Image:              ist0ne/kubedns-amd64:1.9
    Image ID:           docker-pullable://docker.io/ist0ne/kubedns-amd64@sha256:c12a28611a6883a2879b8f8dae6a7b088082d40262a116be51c8ee0b69cf91e0
    Ports:              10053/UDP, 10053/TCP, 10055/TCP
    Args:
      --domain=cluster.local.
      --dns-port=10053
      --config-map=kube-dns
      --v=0
    Limits:
      memory:   170Mi
    Requests:
      cpu:              100m
      memory:           70Mi
    State:              Running
      Started:          Mon, 13 Mar 2017 22:55:06 +0800
    Ready:              True
    Restart Count:      0
    Liveness:           http-get http://:8080/healthz-kubedns delay=60s timeout=5s period=10s #success=1 #failure=5
    Readiness:          http-get http://:8080/readiness delay=3s timeout=5s period=10s #success=1 #failure=3
    Volume Mounts:
      /var/run/secrets/kubernetes.io/serviceaccount from default-token-wqs8d (ro)
    Environment Variables:
      PROMETHEUS_PORT:  10055
  dnsmasq:
    Container ID:       docker://be71b9e8c5815c33581b984de68e7d36ef72190997b14f4038ac8ca94d652336
    Image:              ist0ne/kube-dnsmasq-amd64:1.4
    Image ID:           docker-pullable://docker.io/ist0ne/kube-dnsmasq-amd64@sha256:e49f231477ba296992515e5824c3015692227abb8f5c6fa08557d4d83abe8058
    Ports:              53/UDP, 53/TCP
    Args:
          --cache-size=1000
      --no-resolv
      --server=127.0.0.1#10053
      --log-facility=-
    Requests:
      cpu:              150m
      memory:           10Mi
    State:              Running
      Started:          Mon, 13 Mar 2017 22:55:06 +0800
    Ready:              True
    Restart Count:      0
    Liveness:           http-get http://:8080/healthz-dnsmasq delay=60s timeout=5s period=10s #success=1 #failure=5
    Volume Mounts:
      /var/run/secrets/kubernetes.io/serviceaccount from default-token-wqs8d (ro)
    Environment Variables:      <none>
  dnsmasq-metrics:
    Container ID:       docker://33f352447749533f391a11532f93c246a158a340b0849d5aa9619a9f4222dc1d
    Image:              ist0ne/dnsmasq-metrics-amd64:1.0
    Image ID:           docker-pullable://docker.io/ist0ne/dnsmasq-metrics-amd64@sha256:5ae7e3a3a2ac3f08352da3f25b9fc12482ef9798f52b0cc847c3d1258beaa7bb
    Port:               10054/TCP
    Args:
      --v=2
      --logtostderr
    Requests:
      memory:           10Mi
    State:              Running
      Started:          Mon, 13 Mar 2017 22:55:07 +0800
    Ready:              True
    Restart Count:      0
    Liveness:           http-get http://:10054/metrics delay=60s timeout=5s period=10s #success=1 #failure=5
    Volume Mounts:
      /var/run/secrets/kubernetes.io/serviceaccount from default-token-wqs8d (ro)
    Environment Variables:      <none>
  healthz:
    Container ID:       docker://ff7aa1d8491bdb35a943ddbb3b284e58de9605e985583754136d8c8cf2e02f77
    Image:              ist0ne/exechealthz-amd64:1.2
    Image ID:           docker-pullable://docker.io/ist0ne/exechealthz-amd64@sha256:67e6a74ee4242c4891c4be79da87b2e21bf7fb3645ad10fef3022d441aea463d
    Port:               8080/TCP
    Args:
      --cmd=nslookup kubernetes.default.svc.cluster.local 127.0.0.1 >/dev/null
      --url=/healthz-dnsmasq
      --cmd=nslookup kubernetes.default.svc.cluster.local 127.0.0.1:10053 >/dev/null
      --url=/healthz-kubedns
      --port=8080
      --quiet
    Limits:
      memory:   50Mi
    Requests:
      cpu:              10m
      memory:           50Mi
    State:              Running
      Started:          Mon, 13 Mar 2017 22:55:06 +0800
    Ready:              True
    Restart Count:      0
    Volume Mounts:
      /var/run/secrets/kubernetes.io/serviceaccount from default-token-wqs8d (ro)
    Environment Variables:      <none>
Conditions:
  Type          Status
  Initialized   True
  Ready         True
  PodScheduled  True
Volumes:
  default-token-wqs8d:
    Type:       Secret (a volume populated by a Secret)
    SecretName: default-token-wqs8d
QoS Class:      Burstable
Tolerations:    CriticalAddonsOnly=:Exists
Events:
  FirstSeen     LastSeen        Count   From                    SubObjectPath                           Type            Reason          Message
  ---------     --------        -----   ----                    -------------                           --------        ------          -------
  53m           53m             1       {default-scheduler }                                            Normal          Scheduled       Successfully assigned kube-dns-2200864676-p2zn6 to 127.0.0.1
  53m           53m             1       {kubelet 127.0.0.1}     spec.containers{kubedns}                Normal          Started         Started container with docker id 3756f735b64d
  53m           53m             1       {kubelet 127.0.0.1}     spec.containers{healthz}                Normal          Created         Created container with docker id ff7aa1d8491b
  53m           53m             1       {kubelet 127.0.0.1}     spec.containers{healthz}                Normal          Started         Started container with docker id ff7aa1d8491b
  53m           53m             1       {kubelet 127.0.0.1}     spec.containers{kubedns}                Normal          Pulled          Container image "ist0ne/kubedns-amd64:1.9" already present on machine
  53m           53m             1       {kubelet 127.0.0.1}     spec.containers{kubedns}                Normal          Created         Created container with docker id 3756f735b64d
  53m           53m             1       {kubelet 127.0.0.1}     spec.containers{healthz}                Normal          Pulled          Container image "ist0ne/exechealthz-amd64:1.2" already present on machine
    53m           53m             1       {kubelet 127.0.0.1}     spec.containers{dnsmasq}                Normal          Pulled          Container image "ist0ne/kube-dnsmasq-amd64:1.4" already present on machine
  53m           53m             1       {kubelet 127.0.0.1}     spec.containers{dnsmasq}                Normal          Created         Created container with docker id be71b9e8c581
  53m           53m             1       {kubelet 127.0.0.1}     spec.containers{dnsmasq}                Normal          Started         Started container with docker id be71b9e8c581
  53m           53m             1       {kubelet 127.0.0.1}     spec.containers{dnsmasq-metrics}        Normal          Pulled          Container image "ist0ne/dnsmasq-metrics-amd64:1.0" already present on machine
  53m           53m             1       {kubelet 127.0.0.1}     spec.containers{dnsmasq-metrics}        Normal          Created         Created container with docker id 33f352447749
  53m           53m             1       {kubelet 127.0.0.1}     spec.containers{dnsmasq-metrics}        Normal          Started         Started container with docker id 33f352447749

$ kubectl logs kube-dns-2200864676-p2zn6 kubedns --namespace=kube-system
I0313 14:55:06.794846       1 dns.go:42] version: v1.6.0-alpha.0.680+3872cb93abf948-dirty
I0313 14:55:06.795276       1 server.go:107] Using https://10.254.0.1:443 for kubernetes master, kubernetes API: <nil>
I0313 14:55:06.795830       1 server.go:68] Using configuration read from ConfigMap: kube-system:kube-dns
I0313 14:55:06.795883       1 server.go:113] FLAG: --alsologtostderr="false"
I0313 14:55:06.795902       1 server.go:113] FLAG: --config-map="kube-dns"
I0313 14:55:06.795912       1 server.go:113] FLAG: --config-map-namespace="kube-system"
I0313 14:55:06.795919       1 server.go:113] FLAG: --dns-bind-address="0.0.0.0"
I0313 14:55:06.795925       1 server.go:113] FLAG: --dns-port="10053"
I0313 14:55:06.795934       1 server.go:113] FLAG: --domain="cluster.local."
I0313 14:55:06.795944       1 server.go:113] FLAG: --federations=""
I0313 14:55:06.795953       1 server.go:113] FLAG: --healthz-port="8081"
I0313 14:55:06.795960       1 server.go:113] FLAG: --kube-master-url=""
I0313 14:55:06.795976       1 server.go:113] FLAG: --kubecfg-file=""
I0313 14:55:06.795980       1 server.go:113] FLAG: --log-backtrace-at=":0"
I0313 14:55:06.795985       1 server.go:113] FLAG: --log-dir=""
I0313 14:55:06.795990       1 server.go:113] FLAG: --log-flush-frequency="5s"
I0313 14:55:06.795996       1 server.go:113] FLAG: --logtostderr="true"
I0313 14:55:06.796001       1 server.go:113] FLAG: --stderrthreshold="2"
I0313 14:55:06.796005       1 server.go:113] FLAG: --v="0"
I0313 14:55:06.796009       1 server.go:113] FLAG: --version="false"
I0313 14:55:06.796015       1 server.go:113] FLAG: --vmodule=""
I0313 14:55:06.796076       1 server.go:155] Starting SkyDNS server (0.0.0.0:10053)
I0313 14:55:06.796405       1 server.go:165] Skydns metrics enabled (/metrics:10055)
I0313 14:55:06.796661       1 logs.go:41] skydns: ready for queries on cluster.local. for tcp://0.0.0.0:10053 [rcache 0]
I0313 14:55:06.796685       1 logs.go:41] skydns: ready for queries on cluster.local. for udp://0.0.0.0:10053 [rcache 0]
E0313 14:55:06.826343       1 sync.go:105] Error getting ConfigMap kube-system:kube-dns err: configmaps "kube-dns" not found
E0313 14:55:06.826359       1 dns.go:190] Error getting initial ConfigMap: configmaps "kube-dns" not found, starting with default values
I0313 14:55:06.827616       1 server.go:126] Setting up Healthz Handler (/readiness)
I0313 14:55:06.827631       1 server.go:131] Setting up cache handler (/cache)
I0313 14:55:06.827637       1 server.go:120] Status HTTP port 8081

root@tjwq01-sys-bs003007 dns]# curl 172.30.19.2:8080/healthz-dnsmasq   # 172.30.19.2 为 kube-dns pod的pod ip
ok$ curl 172.30.19.2:8081/healthz-dnsmasq
404 page not found

#测试 kubedns
1. 新建一个Deployment
$ kubectl create -f my-nginx.yaml
2. export该Deployment, 生成服务：
$ kubectl expose deploy my-nginx
$ kubectl get services --all-namespaces
NAMESPACE     NAME         CLUSTER-IP     EXTERNAL-IP   PORT(S)         AGE
default       kubernetes   10.254.0.1     <none>        443/TCP         87d
default       my-nginx     10.254.0.159   <none>        80/TCP          58m
kube-system   kube-dns     10.254.0.2     <none>        53/UDP,53/TCP   1h
3. 创建另一个pod，查看/etc/resolv.conf是否正确，是否能够解析 my-nginx:
$ kubectl create -f pod-nginx.yaml
$ kubectl exec -i -t nginx bash


root@nginx:/# cat /etc/resolv.conf  # 可见，合并了kubedns和系统dns的配置(https://github.com/kubernetes/kubernetes/issues/41328)
search default.svc.cluster.local svc.cluster.local cluster.local tjwq01.ksyun.com
nameserver 10.254.0.2
nameserver 10.64.116.10
nameserver 10.64.116.11
options ndots:5
root@nginx:/# host my-nginx
bash: host: command not found

root@nginx:/# ping my-nginx   # 可以直接解析当前defult 命名空间的service
PING my-nginx.default.svc.cluster.local (10.254.0.159): 48 data bytes
^C--- my-nginx.default.svc.cluster.local ping statistics ---
2 packets transmitted, 0 packets received, 100% packet loss

root@nginx:/# ping kubernetes
PING kubernetes.default.svc.cluster.local (10.254.0.1): 48 data bytes
^C--- kubernetes.default.svc.cluster.local ping statistics ---
1 packets transmitted, 0 packets received, 100% packet loss
root@nginx:/#

root@nginx:/# ping kube-dns  # 不能解析其它命名空间的service
ping: unknown host

root@nginx:/# ping kube-dns.kube-system.svc.cluster.local #指定完整的域名后，可以解析
PING kube-dns.kube-system.svc.cluster.local (10.254.0.2): 48 data bytes
^C--- kube-dns.kube-system.svc.cluster.local ping statistics ---
2 packets transmitted, 0 packets received, 100% packet loss


