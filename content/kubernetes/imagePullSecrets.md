# 从私有registry拉取镜像

registry有两方面的安全策略：

1. registry启用TLS，docker客户端对registry的证书进行验证，使用众所周知的CA和/etc/docker/certs.d/{registryip:port}}/ca.crt；
1. registry可选的对docker cli client的证书进行验证；
1. regsitry可选的使用HTTP Basic或其它认证方式；

k8s访问私有registry的方法：

1. 利用Node上的配置(CA证书和认证信息)方位registry；
1. 创建registry的secret（类型kubernetes.io/dockercfg），然后创建pod时，为imagePullSecrets参数指定该secret；如果有多个私有库，则需要创建多个secret，然后pod中同时引用这些secrets；
1. 用包含所有认证信息的yaml文件创建一个secret(类型kubernetes.io/dockerconfigjson)，为imagePullSecrets指定该secret;
1. 利用API创建kubernetes.io/dockerconfigjson的secret

注意：如果更新了secret，则k8s自动更新使用了该secret的pods。

## 一、利用Node上的配置

确认registry上有镜像 zhangjun3/nginx:1.7.9 镜像
$ curl -u foo:foo123  --cacert /etc/docker/certs.d/10.64.3.7\:8000/ca.crt https://10.64.3.7:8000/v2/_catalog
{"repositories":["library/redis","zhangjun3/busybox","zhangjun3/nginx","zhangjun3/pause","zhangjun3/pause2"]}
$ curl -u foo:foo123  --cacert /etc/docker/certs.d/10.64.3.7\:8000/ca.crt https://10.64.3.7:8000/v2/zhangjun3/nginx/tags/list
{"name":"zhangjun3/nginx","tags":["1.7.9"]}

移除node上的registry ca证书和认证信息
$ mv /etc/docker/certs.d/10.64.3.7\:8000/ca.crt{,.bak}
$ mv ~/.docker/config.json{,.bak}

$ cat nginx-registry.yaml
apiVersion: v1
kind: Pod
metadata:
  name: nginx-registry
spec:
  containers:
  - name: nginx
    image: 10.64.3.7:8000/zhangjun3/nginx:1.7.9
    ports:
    - containerPort: 80

### 创建pod时pull image失败，提示证书验证失败

$ kubectl create -f nginx-registry.yaml
pod "nginx-registry" created
$ kubectl get pod nginx-registry
NAME             READY     STATUS         RESTARTS   AGE
nginx-registry   0/1       ErrImagePull   0          13s
$ kubectl describe pod nginx-registry |tail -10
  FirstSeen     LastSeen        Count   From                    SubObjectPath           Type            Reason          Message
  ---------     --------        -----   ----                    -------------           --------        ------          -------
  1m            1m              1       {default-scheduler }                            Normal          Scheduled       Successfully assigned nginx-registry to 127.0.0.1
  1m            31s             3       {kubelet 127.0.0.1}     spec.containers{nginx}  Normal          Pulling         pulling image "10.64.3.7:8000/zhangjun3/nginx:1.7.9"
  1m            31s             3       {kubelet 127.0.0.1}     spec.containers{nginx}  Warning         Failed          Failed to pull image "10.64.3.7:8000/zhangjun3/nginx:1.7.9": image pull failed for 10.64.3.7:8000/zhangjun3/nginx:1.7.9, this may be because there are no credentials on this request.  details: (Error response from daemon: {"message":"Get https://10.64.3.7:8000/v1/_ping: x509: certificate signed by unknown authority"})
  1m            31s             3       {kubelet 127.0.0.1}                             Warning         FailedSync      Error syncing pod, skipping: failed to "StartContainer" for "nginx" with ErrImagePull: "image pull failed for 10.64.3.7:8000/zhangjun3/nginx:1.7.9, this may be because there are no credentials on this request.  details: (Error response from daemon: {\"message\":\"Get https://10.64.3.7:8000/v1/_ping: x509: certificate signed by unknown authority\"})"

  1m    1s      4       {kubelet 127.0.0.1}     spec.containers{nginx}  Normal  BackOff         Back-off pulling image "10.64.3.7:8000/zhangjun3/nginx:1.7.9"
  1m    1s      4       {kubelet 127.0.0.1}                             Warning FailedSync      Error syncing pod, skipping: failed to "StartContainer" for "nginx" with ImagePullBackOff: "Back-off pulling image \"10.64.3.7:8000/zhangjun3/nginx:1.7.9\""

### 恢复 registry 的证书，但不提供basic认证信息，提示未找到image

$ mv /etc/docker/certs.d/10.64.3.7\:8000/ca.crt{.bak,}
$ kubectl delete pod nginx-registry
$ kubectl create -f nginx-registry.yaml
pod "nginx-registry" created
$ kubectl get pod nginx-registry
NAME             READY     STATUS         RESTARTS   AGE
nginx-registry   0/1       ErrImagePull   0          13s
$ kubectl describe pod nginx-registry |tail -10
  FirstSeen     LastSeen        Count   From                    SubObjectPath           Type            Reason          Message
  ---------     --------        -----   ----                    -------------           --------        ------          -------
  7m            7m              1       {default-scheduler }                            Normal          Scheduled       Successfully assigned nginx-registry to 127.0.0.1
  6m            6m              1       {kubelet 127.0.0.1}     spec.containers{nginx}  Normal          Pulling         pulling image "10.64.3.7:8000/zhangjun3/nginx:1.7.9"
  6m            6m              1       {kubelet 127.0.0.1}     spec.containers{nginx}  Warning         Failed          Failed to pull image "10.64.3.7:8000/zhangjun3/nginx:1.7.9": image pull failed for 10.64.3.7:8000/zhangjun3/nginx:1.7.9, this may be because there are no credentials on this request.  details: (Error: image zhangjun3/nginx:1.7.9 not found)
  6m            6m              1       {kubelet 127.0.0.1}                             Warning         FailedSync      Error syncing pod, skipping: failed to "StartContainer" for "nginx" with ErrImagePull: "image pull failed for 10.64.3.7:8000/zhangjun3/nginx:1.7.9, this may be because there are no credentials on this request.  details: (Error: image zhangjun3/nginx:1.7.9 not found)"

  7m    2m      24      {kubelet 127.0.0.1}     spec.containers{nginx}  Normal  BackOff         Back-off pulling image "10.64.3.7:8000/zhangjun3/nginx:1.7.9"
  7m    2m      24      {kubelet 127.0.0.1}                             Warning FailedSync      Error syncing pod, skipping: failed to "StartContainer" for "nginx" with ImagePullBackOff: "Back-off pulling image \"10.64.3.7:8000/zhangjun3/nginx:1.7.9\""

### 恢复basic认证信息，稍等一会后，pull image成功

$ mv ~/.docker/config.json{.bak,}
$ kubectl get pod nginx-registry
NAME             READY     STATUS    RESTARTS   AGE
nginx-registry   1/1       Running   0          10m
$ kubectl describe pod nginx-registry |tail -10
  10m           10m             1       {kubelet 127.0.0.1}     spec.containers{nginx}  Normal          Pulling         pulling image "10.64.3.7:8000/zhangjun3/nginx:1.7.9"
  10m           10m             1       {kubelet 127.0.0.1}     spec.containers{nginx}  Warning         Failed          Failed to pull image "10.64.3.7:8000/zhangjun3/nginx:1.7.9": image pull failed for 10.64.3.7:8000/zhangjun3/nginx:1.7.9, this may be because there are no credentials on this request.  details: (Error: image zhangjun3/nginx:1.7.9 not found)
  10m           10m             1       {kubelet 127.0.0.1}                             Warning         FailedSync      Error syncing pod, skipping: failed to "StartContainer" for "nginx" with ErrImagePull: "image pull failed for 10.64.3.7:8000/zhangjun3/nginx:1.7.9, this may be because there are no credentials on this request.  details: (Error: image zhangjun3/nginx:1.7.9 not found)"

  10m   5m      24      {kubelet 127.0.0.1}     spec.containers{nginx}  Normal  BackOff         Back-off pulling image "10.64.3.7:8000/zhangjun3/nginx:1.7.9"
  10m   5m      24      {kubelet 127.0.0.1}                             Warning FailedSync      Error syncing pod, skipping: failed to "StartContainer" for "nginx" with ImagePullBackOff: "Back-off pulling image \"10.64.3.7:8000/zhangjun3/nginx:1.7.9\""

  5m    5m      1       {kubelet 127.0.0.1}     spec.containers{nginx}  Normal  Pulled  Container image "10.64.3.7:8000/zhangjun3/nginx:1.7.9" already present on machine
  5m    5m      1       {kubelet 127.0.0.1}     spec.containers{nginx}  Normal  Created Created container with docker id c0d45d8da711
  5m    5m      1       {kubelet 127.0.0.1}     spec.containers{nginx}  Normal  Started Started container with docker id c0d45d8da711

## 二、创建registry的secret

[secret的类型](secret的类型.md)

$ kubectl create secret docker-registry registry-1 --docker-server=10.64.3.7:8000 --docker-username=foo --docker-password=foo123 --docker-email=foo@foo
secret "registry-1" created

$ kubectl get secret registry-1
NAME         TYPE                      DATA      AGE
registry-1   kubernetes.io/dockercfg   1         10s

$ kubectl get secret registry-1  -o yaml
apiVersion: v1
data:
  .dockercfg: eyJodHRwczovL2luZGV4LmRvY2tlci5pby92MS8iOnsidXNlcm5hbWUiOiJmb28iLCJwYXNzd29yZCI6ImZvbzEyMyIsImVtYWlsIjoiZm9vQGZvbyIsImF1dGgiOiJabTl2T21admJ6RXlNdz09In19
kind: Secret
metadata:
  creationTimestamp: 2017-03-24T08:42:05Z
  name: registry-1
  namespace: default
  resourceVersion: "1878368"
  selfLink: /api/v1/namespaces/default/secrets/registry-1
  uid: c260e37e-106d-11e7-8101-8cdcd4b3be48
type: kubernetes.io/dockercfg

删除旧的pod和已经pull到本地的image
$ kubectl delete pod nginx-registry
pod "nginx-registry" deleted
$  docker rmi 10.64.3.7:8000/zhangjun3/nginx:1.7.9
Untagged: 10.64.3.7:8000/zhangjun3/nginx:1.7.9
Untagged: 10.64.3.7:8000/zhangjun3/nginx@sha256:ba6612116006c334fcad0a1beb33e688129f52de9d4bc730bb3a0011fd3ab625
删除旧的认证文件
$ rm ~/.docker/config.json

$ cat nginx-registry-2.yaml
apiVersion: v1
kind: Pod
metadata:
  name: nginx-registry-2
spec:
  containers:
  - name: nginx
    image: 10.64.3.7:8000/zhangjun3/nginx:1.7.9
    ports:
    - containerPort: 80
  imagePullSecrets:
  - name: registry-1
$ kubectl create -f nginx-registry-2.yaml
pod "nginx-registry-2" created
$ kubectl get pod nginx-registry-2
NAME               READY     STATUS    RESTARTS   AGE
nginx-registry-2   1/1       Running   0          5m
$ kubectl describe pod nginx-registry-2 |tail -4
  5m    4m      2       {kubelet 127.0.0.1}     spec.containers{nginx}  Normal  Pulling pulling image "10.64.3.7:8000/zhangjun3/nginx:1.7.9"
  4m    4m      1       {kubelet 127.0.0.1}     spec.containers{nginx}  Normal  Pulled  Successfully pulled image "10.64.3.7:8000/zhangjun3/nginx:1.7.9"
  4m    4m      1       {kubelet 127.0.0.1}     spec.containers{nginx}  Normal  Created Created container with docker id 8c6f736e5b78
  4m    4m      1       {kubelet 127.0.0.1}     spec.containers{nginx}  Normal  Started Started container with docker id 8c6f736e5b78

## 三、用包含所有认证信息的yaml文件创建一个secret

$ kubectl delete pod nginx-registry-2
$ docker rmi 10.64.3.7:8000/zhangjun3/nginx:1.7.9
Untagged: 10.64.3.7:8000/zhangjun3/nginx:1.7.9
Untagged: 10.64.3.7:8000/zhangjun3/nginx@sha256:ba6612116006c334fcad0a1beb33e688129f52de9d4bc730bb3a0011fd3ab625
$ rm ~/.docker/config.json
$ docker login 10.64.3.7:8000
Username: foo
Password:
Login Succeeded
$ base64 -w 0 ~/.docker/config.json
ewoJImF1dGhzIjogewoJCSIxMC42NC4zLjc6ODAwMCI6IHsKCQkJImF1dGgiOiAiWm05dk9tWnZiekV5TXc9PSIKCQl9Cgl9Cn0=

$ cat secret-dockerconfigjson.yaml
apiVersion: v1
kind: Secret
metadata:
  name: registry-2
  namespace: default
data:
    .dockerconfigjson: ewoJImF1dGhzIjogewoJCSIxMC42NC4zLjc6ODAwMCI6IHsKCQkJImF1dGgiOiAiWm05dk9tWnZiekV5TXc9PSIKCQl9Cgl9Cn0=
type: kubernetes.io/dockerconfigjson

删除登陆信息
$ rm ~/.docker/config.json

$ cat nginx-registry-3.yaml
apiVersion: v1
kind: Pod
metadata:
  name: nginx-registry-3
spec:
  containers:
  - name: nginx
    image: 10.64.3.7:8000/zhangjun3/nginx:1.7.9
    imagePullPolicy: Always
    ports:
    - containerPort: 80
  imagePullSecrets:
  - name: registry-2
$ kubectl create -f nginx-registry-3.yaml
pod "nginx-registry-3" created

$ kubectl get pod nginx-registry-3
NAME               READY     STATUS    RESTARTS   AGE
nginx-registry-3   1/1       Running   0          1m

$  kubectl describe pod nginx-registry-3  |tail -5
  20s           20s             1       {default-scheduler }                            Normal          Scheduled       Successfully assigned nginx-registry-3 to 127.0.0.1
  19s           19s             1       {kubelet 127.0.0.1}     spec.containers{nginx}  Normal          Pulling         pulling image "10.64.3.7:8000/zhangjun3/nginx:1.7.9"
  19s           19s             1       {kubelet 127.0.0.1}     spec.containers{nginx}  Normal          Pulled          Successfully pulled image "10.64.3.7:8000/zhangjun3/nginx:1.7.9"
  19s           19s             1       {kubelet 127.0.0.1}     spec.containers{nginx}  Normal          Created         Created container with docker id 40825beb5bb1
  19s           19s             1       {kubelet 127.0.0.1}     spec.containers{nginx}  Normal          Started         Started container with docker id 40825beb5bb1

## 利用API创建kubernetes.io/dockerconfigjson类型的secret

$ curl -v -H "Content-type: application/json"  -X POST -d ' {
  "apiVersion": "v1",
  "kind": "Secret",
  "metadata": {
    "name": "registry",
    "namespace": "default"
  },
  "data": {
    ".dockerconfigjson": "{cat ~/.docker/config.json |base64 -w 0}"
  },
  "type": "kubernetes.io/dockerconfigjson"
}' http://10.57.136.60:8080/api/v1/namespaces/default/

然后pod的imagePullSecrets使用该secret；

参考：http://tonybai.com/2016/11/16/how-to-pull-images-from-private-registry-on-kubernetes-cluster/