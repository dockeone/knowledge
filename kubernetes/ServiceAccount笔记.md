# ServiceAccount

Service Account概念的引入是基于这样的使用场景：运行在pod里的进程需要调用Kubernetes API以及非Kubernetes API的服务（如image repository/被mount到pod上的NFS volumes中的file等），使用Service Account来为pod提供id。

Service Account和User account可能会带来一定程度上的混淆：
1. user account通常是为human设计的，而service account则是为跑在pod里的process。
2. user account是global的，即跨namespace使用；而service account是namespaced的，即仅在所属的namespace下使用。
3. 创建一个新的user account通常需要较高的特权并且需要经过比较复杂的business process（即对于集群的访问权限的创建），而service account则不然。

要使用ServiceAccount，在启apiservera的时候指定--service_account_key_file，admission_control包含ServiceAccount:
$ kube-apiserver ... --service_account_key_file=/tmp/kube-serviceaccount.key --admission_control=...ServiceAccount,...

启动Controller Manager增加service_account_private_key_file，使用的文件和apiserver必须一致:
$ kube-controller-manager ... --service_account_private_key_file=/tmp/kube-serviceaccount.key...

controller-manager为每个namespace默认创建一个ServiceAccount:
$ kubectl get serviceaccount --all-namespaces
NAMESPACE       NAME      SECRETS
default         default   1
kube-system     default   3


下面是启用了ServiceAccount的admission-control后说执行的逻辑：
https://github.com/fabric8io/kansible/blob/master/vendor/k8s.io/kubernetes/plugin/pkg/admission/serviceaccount/admission.go

// EnforceMountableSecretsAnnotation is a default annotation that indicates that a service account should enforce mountable secrets.
// The value must be true to have this annotation take effect
const EnforceMountableSecretsAnnotation = "kubernetes.io/enforce-mountable-secrets"

如果定义service account时指定kubernetes.io/enforce-mountable-secrets annotation 为true，则k8s会检查引用的secret是否和对应的serviceaccount关联；

// NewServiceAccount returns an admission.Interface implementation which limits admission of Pod CREATE requests based on the pod's ServiceAccount:
// 1. If the pod does not specify a ServiceAccount, it sets the pod's ServiceAccount to "default"
// 2. It ensures the ServiceAccount referenced by the pod exists
// 3. If LimitSecretReferences is true, it rejects the pod if the pod references Secret objects which the pod's ServiceAccount does not reference
// 4. If the pod does not contain any ImagePullSecrets, the ImagePullSecrets of the service account are added.
// 5. If MountServiceAccountToken is true, it adds a VolumeMount with the pod's ServiceAccount's api token secret to containers

// TODO: enable this once we've swept secret usage to account for adding secret references to service accounts
		LimitSecretReferences: false,
LimitSecretReferences是全局性的，优先级高于各service account的设置，表示需要检查引用的secret是否和对应的serviceaccount关联；

// Auto mount service account API token secrets
		MountServiceAccountToken: true,
MountServiceAccountToken 表示是否自动mount serviceaccount自动生成的API token secrets；

// Reject pod creation until a service account token is available
		RequireAPIToken: true,
RequireAPIToken：如果pod依赖的service account不可用，则阻止创建pod；

// enforceMountableSecrets indicates whether mountable secrets should be enforced for a particular service account
// A global setting of true will override any flag set on the individual service account


kubectl get  serviceaccount  build-robot 和 kubectl get  serviceaccount  build-robot -o yaml 输出的secret是在创建serviceaccount时指定
的secret+自动创建的 API Token Secret；

如果在创建serviceaccount后，再创建和它关联的secret，则不在上面两个命令的输出中显示，需要使用kubectl describe serviceaccount  build-robot来获取serviceaccount关联的所有secret；

kubectl get serviceaccount 的secrets字段显示的是可以被Mount到Pod容器的Secret列表，其中一半包含一个自动创建的kubernetes.io/service-account-token类型的API Token Secret，其它为Opaque类型的自定义Secret；注意：即使关联了多个kubernetes.io/service-account-token类型Secret，secrets字段也只会列出
会自动mount到pod容器的、自动创建的API Token Secret。

kubectl describe serviceaccount 的Mountable secrets字段内容和kubectl get serviceaccount 的secrets字段内容一致，但是Tokens字段的内容只包含
kubernetes.io/service-account-token类型的API Token Secrets；

如果创建Pod时没有指定imagePullSecret参数，则k8s使用serviceaccount中的imagePullSecret参数来和registry做验证；

如果更新了apiserver或controller-manager的-service_account_key，则需要删除各serviceaccount关联的API Token Secret，k8s会用新
的key重新生成和关联一个，然后需要重新启动(FIXME!!!???)已经mount旧API Token Secret的Pod；

## kube-controller-manager负责为每个namespace创建一个名为default的serviceaccount, 同时自动关联一个secret token，使用该serviceaccount时会自动将该secrets关联到pod容器中；

[root@tjwq01-sys-bs003007 ~]# kubectl get serviceaccount
NAME          SECRETS   AGE
default       1         17h

[root@tjwq01-sys-bs003007 ~]# kubectl get serviceaccount default  -o yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  creationTimestamp: 2017-03-13T14:54:05Z
  name: default
  namespace: default
  resourceVersion: "857134"
  selfLink: /api/v1/namespaces/default/serviceaccounts/default
  uid: e766c8ce-07fc-11e7-8101-8cdcd4b3be48
secrets:
- name: default-token-8khbh  # 自动关联的sercret，包含namespace、ca、token三个data；

[root@tjwq01-sys-bs003007 ~]# kubectl describe serviceaccount  default  # 这个serviceaccount的token是可以mount的；
Name:           default
Namespace:      default
Labels:         <none>

Image pull secrets:     <none>

Mountable secrets:      default-token-8khbh

Tokens:                 default-token-8khbh

[root@tjwq01-sys-bs003007 ~]# kubectl get secrets
NAME                      TYPE                                  DATA      AGE
default-token-8khbh       kubernetes.io/service-account-token   3         17h

[root@tjwq01-sys-bs003007 ~]# kubectl get secrets default-token-8khbh  -o yaml
apiVersion: v1
data:  # 和serviceaccount自动关联的secret token，包含3个固定的data；
  ca.crt: ...
  namespace: ZGVmYXVsdA==
  token: ...
kind: Secret
metadata:
  annotations:
    kubernetes.io/service-account.name: default
    kubernetes.io/service-account.uid: e766c8ce-07fc-11e7-8101-8cdcd4b3be48
  creationTimestamp: 2017-03-13T14:54:05Z
  name: default-token-8khbh
  namespace: default
  resourceVersion: "857133"
  selfLink: /api/v1/namespaces/default/secrets/default-token-8khbh
  uid: e7683e66-07fc-11e7-8101-8cdcd4b3be48
type: kubernetes.io/service-account-token


# 手动创建serviceaccount

[root@tjwq01-sys-bs003007 ~]# cat serviceaccount.yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: build-robot

[root@tjwq01-sys-bs003007 ~]# kubectl create -f serviceaccount.yaml
serviceaccount "build-robot" created

[root@tjwq01-sys-bs003007 ~]# kubectl get serviceaccount build-robot -o yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  creationTimestamp: 2017-03-14T08:25:34Z
  name: build-robot  # 这是serviceaccount关联的数据；
  namespace: default
  resourceVersion: "926583"
  selfLink: /api/v1/namespaces/default/serviceaccounts/build-robot
  uid: cb6472b6-088f-11e7-8101-8cdcd4b3be48
secrets:
- name: build-robot-token-5p1jw     # 自动为该serviceaccount生产和关联token secrets

[root@tjwq01-sys-bs003007 ~]# kubectl describe serviceaccount  build-robot
Name:           build-robot
Namespace:      default
Labels:         <none>

Image pull secrets:     <none>

Mountable secrets:      build-robot-token-5p1jw

Tokens:                 build-robot-token-5p1jw

[root@tjwq01-sys-bs003007 ~]# kubectl get secrets
NAME                      TYPE                                  DATA      AGE
build-robot-token-5p1jw   kubernetes.io/service-account-token   3         1h
default-token-8khbh       kubernetes.io/service-account-token   3         18h

[root@tjwq01-sys-bs003007 ~]# kubectl get secrets build-robot-token-5p1jw -o yaml
apiVersion: v1
data:   # 使用该serviceaccount时会自动挂载到pod容器中的数据；
  ca.crt: ***
  namespace: ZGVmYXVsdA==
  token: ***
kind: Secret
metadata:
  annotations:
    kubernetes.io/service-account.name: build-robot
    kubernetes.io/service-account.uid: cb6472b6-088f-11e7-8101-8cdcd4b3be48
  creationTimestamp: 2017-03-14T08:25:34Z
  name: build-robot-token-5p1jw
  namespace: default
  resourceVersion: "926582"
  selfLink: /api/v1/namespaces/default/secrets/build-robot-token-5p1jw
  uid: cb65acc6-088f-11e7-8101-8cdcd4b3be48
type: kubernetes.io/service-account-token

k8s将ServiceAccount的metadata保存到secret中，如果删除了ServiceAccount包含的secret则k8s会自动创建它们；
也可以手动创建secret，然后添加到ServiceAccount。

# serviceaccount会自动关联kubernetes.io/service-account-token类型的Secret

为名为build-root的service account创建secret，注意annotations和type，
无需为kubernetes.io/service-account-token类型的secret指定secret的具体内容，会自动生成；

[root@tjwq01-sys-bs003007 ~]# cat build-robot-secret.yaml
apiVersion: v1
kind: Secret
metadata:
  name: build-robot-secret
  annotations:
    kubernetes.io/service-account.name: build-robot
data:   # 和该secret关联的数据，k8s会自动加上其它三个数据域：namespace、token、ca
  data: dGVzdCBkYXRhCg==     # test data的base64编码值
type: kubernetes.io/service-account-token

[root@tjwq01-sys-bs003007 ~]# kubectl get secrets
NAME                      TYPE                                  DATA      AGE
build-robot-secret        kubernetes.io/service-account-token   4         23s   # data数量为4个
build-robot-token-5p1jw   kubernetes.io/service-account-token   3         3h
default-token-8khbh       kubernetes.io/service-account-token   3         20h

[root@tjwq01-sys-bs003007 ~]# kubectl get secrets build-robot-secret -o yaml
apiVersion: v1
data:
  ca.crt: ...
  namespace: ZGVmYXVsdA==
  token: ...
  data: dGVzdCBkYXRhCg==   # 自定义的值；
kind: Secret
metadata:
  annotations:
    kubernetes.io/service-account.name: build-robot
    kubernetes.io/service-account.uid: cb6472b6-088f-11e7-8101-8cdcd4b3be48
  creationTimestamp: 2017-03-14T11:44:13Z
  name: build-robot-secret
  namespace: default
  resourceVersion: "939648"
  selfLink: /api/v1/namespaces/default/secrets/build-robot-secret
  uid: 8b737c8c-08ab-11e7-8101-8cdcd4b3be48
type: kubernetes.io/service-account-token

[root@tjwq01-sys-bs003007 ~]# kubectl get serviceaccount build-robot
NAME          SECRETS   AGE
build-robot   1         3h  # 默认只显示自动生成的token数量；
[root@tjwq01-sys-bs003007 ~]# kubectl get serviceaccount build-robot -o yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  creationTimestamp: 2017-03-14T08:25:34Z
  name: build-robot
  namespace: default
  resourceVersion: "926583"
  selfLink: /api/v1/namespaces/default/serviceaccounts/build-robot
  uid: cb6472b6-088f-11e7-8101-8cdcd4b3be48
secrets:  # 默认只显示可以被Pod容器Mount的Secret
- name: build-robot-token-5p1jw

[root@tjwq01-sys-bs003007 ~]# kubectl describe serviceaccount build-robot
Name:           build-robot
Namespace:      default
Labels:         <none>

Tokens:                 build-robot-secret   # build-robot自动关联了该API Token Secret；
                        build-robot-token-5p1jw

Image pull secrets:     <none>

Mountable secrets:      build-robot-token-5p1jw

如果创建secret时的service account不存在，则该secret将被token controller清理；

## serviceaccount不会自动关联Opaque类型的Secret，解决方法参考后文的replace serviceaccount定义的方式
[root@tjwq01-sys-bs003007 ~]# cat test-secret-3.yaml
apiVersion: v1
kind: Secret
metadata:
  name: test-secret-2
  annotations:
    kubernetes.io/service-account.name: build-robot-2
data:
  data-3: dmFsdWUtMQ0K
  data-4: dmFsdWUtMg0KDQo=
[root@tjwq01-sys-bs003007 ~]# kubectl get secret test-secret-2 -o yaml
apiVersion: v1
data:
  data-3: dmFsdWUtMQ0K
  data-4: dmFsdWUtMg0KDQo=
kind: Secret
metadata:
  annotations:
    kubernetes.io/service-account.name: build-robot-2
  creationTimestamp: 2017-03-14T17:41:20Z
  name: test-secret-2
  namespace: default
  resourceVersion: "963281"
  selfLink: /api/v1/namespaces/default/secrets/test-secret-2
  uid: 6f63c621-08dd-11e7-8101-8cdcd4b3be48
type: Opaque
[root@tjwq01-sys-bs003007 ~]# kubectl describe serviceaccount build-robot-2
Name:           build-robot-2
Namespace:      default
Labels:         <none>

Image pull secrets:     <none>

Mountable secrets:      build-robot-2-token-95hdz   # 没有关联Opque类型的Secret test-secret-2
                        test-secret

Tokens:                 build-robot-2-token-95hdz
                        build-robot-secret-3

##创建一个Opaque类型Secret，创建serviceaccount时指定使用该Secret

如果创建secret未指定type，则默认为Opaque类型；
[root@tjwq01-sys-bs003007 ~]# cat secret.yaml
apiVersion: v1
kind: Secret
metadata:
  name: test-secret
data:
  data-1: dmFsdWUtMQ0K
  data-2: dmFsdWUtMg0KDQo=

[root@tjwq01-sys-bs003007 ~]# kubectl get secret test-secret -o yaml
apiVersion: v1
data:
  data-1: dmFsdWUtMQ0K
  data-2: dmFsdWUtMg0KDQo=
kind: Secret
metadata:
  creationTimestamp: 2017-03-14T12:07:46Z
  name: test-secret
  namespace: default
  resourceVersion: "941222"
  selfLink: /api/v1/namespaces/default/secrets/test-secret
  uid: d5f76adf-08ae-11e7-8101-8cdcd4b3be48
type: Opaque

[root@tjwq01-sys-bs003007 ~]# cat test-serviceaccount.yaml # 创建test-secret serviceaccount时指定关联的secrets；
apiVersion: v1
kind: ServiceAccount
metadata:
  name: test-serviceaccount
secrets:
- name: test-secret

[root@tjwq01-sys-bs003007 ~]# kubectl create -f test-serviceaccount.yaml
serviceaccount "test-serviceaccount" created

k8s自动为serviceaccount生成和关联一个自动mount的secret
[root@tjwq01-sys-bs003007 ~]# kubectl get serviceaccount  test-serviceaccount  # 注意secrets的数目为2；
NAME                  SECRETS   AGE
test-serviceaccount   2         3h
[root@tjwq01-sys-bs003007 ~]# kubectl get serviceaccount  test-serviceaccount -o yaml 
apiVersion: v1
kind: ServiceAccount
metadata:
  creationTimestamp: 2017-03-14T12:10:31Z
  name: test-serviceaccount
  namespace: default
  resourceVersion: "941408"
  selfLink: /api/v1/namespaces/default/serviceaccounts/test-serviceaccount
  uid: 38886423-08af-11e7-8101-8cdcd4b3be48
secrets:
- name: test-secret                      # 创建serviceaccount时指定的secret；
- name: test-serviceaccount-token-tvpb9  # 自动生成的token，如果创建pod时没有指定secret则自动使用该sercret；

kubectl get serviceaccount 的secrets字段显示的是可以被Mount到Pod容器的Secret列表；

[root@tjwq01-sys-bs003007 ~]# kubectl describe serviceaccount  test-serviceaccount
Name:           test-serviceaccount
Namespace:      default
Labels:         <none>

Image pull secrets:     <none>

Mountable secrets:      test-secret
                        test-serviceaccount-token-tvpb9

Tokens:                 test-serviceaccount-token-tvpb9

[root@tjwq01-sys-bs003007 ~]# cat test-serviceaccount-pod.yaml
apiVersion: v1
kind: Pod
metadata:
  name: busybox-serviceaccount
  namespace: default
spec:
  containers:
  - image: busybox
    command:
      - sleep
      - "3600"
    imagePullPolicy: IfNotPresent
    name: busybox
  restartPolicy: Always
  serviceAccount: test-serviceaccount   # 指定使用的serviceaccount, 自动mount它关联的自动生成的secret token test-serviceaccount-token-tvpb9

虽然test-serviceaccount包含两个secrets，但是当前k8s不支持自动mount手动创建的secret如test-secret：
https://github.com/kubernetes/kubernetes/issues/9902


[root@tjwq01-sys-bs003007 ~]# kubectl get pods busybox-serviceaccount -o yaml
apiVersion: v1
kind: Pod
metadata:
  creationTimestamp: 2017-03-14T12:16:41Z
  name: busybox-serviceaccount
  namespace: default
  resourceVersion: "941822"
  selfLink: /api/v1/namespaces/default/pods/busybox-serviceaccount
  uid: 15071345-08b0-11e7-8101-8cdcd4b3be48
spec:
  containers:
  - command:
    - sleep
    - "3600"
    image: busybox
    imagePullPolicy: IfNotPresent
    name: busybox
    resources: {}
    terminationMessagePath: /dev/termination-log
    volumeMounts:
    - mountPath: /var/run/secrets/kubernetes.io/serviceaccount
      name: test-serviceaccount-token-tvpb9                     # 可见，mount了test-serviceaccount自动关联的secret
      readOnly: true
  dnsPolicy: ClusterFirst
  nodeName: 127.0.0.1
  restartPolicy: Always
  securityContext: {}
  serviceAccount: test-serviceaccount
  serviceAccountName: test-serviceaccount
  terminationGracePeriodSeconds: 30
  volumes:
  - name: test-serviceaccount-token-tvpb9
    secret:
      defaultMode: 420
      secretName: test-serviceaccount-token-tvpb9

[root@tjwq01-sys-bs003007 ~]# cat test-serviceaccount-pod2.yaml
apiVersion: v1
kind: Pod
metadata:
  name: nginx-serviceaccount-2
  namespace: default
spec:
  containers:
  - image: nginx:1.7.9
    imagePullPolicy: IfNotPresent
    name: nginx
    volumeMounts:
    - name: foo
      mountPath: /var/run/secrets/kubernetes.io/serviceaccount
      readOnly: true
  restartPolicy: Always
  serviceAccount: test-serviceaccount
  volumes:
  - name: foo
    secret:
      secretName: test-secret

[root@tjwq01-sys-bs003007 ~]# kubectl create -f test-serviceaccount-pod2.yaml
pod "nginx-serviceaccount-2" created

[root@tjwq01-sys-bs003007 ~]# kubectl get pods nginx-serviceaccount-2   -o yaml
apiVersion: v1
kind: Pod
metadata:
  creationTimestamp: 2017-03-14T12:38:26Z
  name: nginx-serviceaccount-2
  namespace: default
  resourceVersion: "943272"
  selfLink: /api/v1/namespaces/default/pods/nginx-serviceaccount-2
  uid: 1e6ee9c1-08b3-11e7-8101-8cdcd4b3be48
spec:
  containers:
  - image: nginx:1.7.9
    imagePullPolicy: IfNotPresent
    name: nginx
    resources: {}
    terminationMessagePath: /dev/termination-log
    volumeMounts:
    - mountPath: /var/run/secrets/kubernetes.io/serviceaccount
      name: foo
      readOnly: true
  dnsPolicy: ClusterFirst
  nodeName: 127.0.0.1
  restartPolicy: Always
  securityContext: {}
  serviceAccount: test-serviceaccount
  serviceAccountName: test-serviceaccount
  terminationGracePeriodSeconds: 30
  volumes:
  - name: foo
    secret:
      defaultMode: 420
      secretName: test-secret
...

[root@tjwq01-sys-bs003007 ~]# kubectl exec -i -t nginx-serviceaccount-2 -c nginx bash
root@nginx-serviceaccount-2:/# ls -l /var/run
lrwxrwxrwx 1 root root 4 Jan 27  2015 /var/run -> /run
root@nginx-serviceaccount-2:/# ls /run/secrets/kubernetes.io/serviceaccount/
data-1  data-2
root@nginx-serviceaccount-2:/# head /run/secrets/kubernetes.io/serviceaccount/data-*
==> /run/secrets/kubernetes.io/serviceaccount/data-1 <==
value-1

==> /run/secrets/kubernetes.io/serviceaccount/data-2 <==
value-2

# serviceaccount 默认不对sercret进行检查
[root@tjwq01-sys-bs003007 ~]# kubectl describe serviceaccount  build-robot
Name:           build-robot
Namespace:      default
Labels:         <none>

Image pull secrets:     <none>

Mountable secrets:      build-robot-token-wp9mc

Tokens:                 build-robot-secret
                        build-robot-token-wp9mc
创建一个使用build-robot serviceaccount的pod，但是成功引用不在build-robot中的secret：
[root@tjwq01-sys-bs003007 ~]# cat test-serviceaccount-pod3.yaml 
apiVersion: v1
kind: Pod
metadata:
  name: nginx-serviceaccount-3
  namespace: default
spec:
  containers:
  - image: nginx:1.7.9
    imagePullPolicy: IfNotPresent
    name: nginx
    volumeMounts:
    - name: foo
      mountPath: /var/run/secrets/kubernetes.io/serviceaccount
      readOnly: true
  restartPolicy: Always
  serviceAccount: build-robot
  volumes:
  - name: foo
    secret:
      secretName: test-secret


# 创建enforce-mountable-secrets的serviceaccount
[root@tjwq01-sys-bs003007 ~]# cat build-robot-serviceaccount-2.yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: build-robot-2
  annotations:
    kubernetes.io/enforce-mountable-secrets: "true"  # 注意value是字符串，而不是bool值；

[root@tjwq01-sys-bs003007 ~]# kubectl get serviceaccount build-robot-2 -o yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  annotations:
    kubernetes.io/enforce-mountable-secrets: "true"
  creationTimestamp: 2017-03-14T16:43:05Z
  name: build-robot-2
  namespace: default
  resourceVersion: "959426"
  selfLink: /api/v1/namespaces/default/serviceaccounts/build-robot-2
  uid: 4c529d78-08d5-11e7-8101-8cdcd4b3be48
secrets:
- name: build-robot-2-token-95hdz

[root@tjwq01-sys-bs003007 ~]# cat test-serviceaccount-pod4.yaml  # 创建一个pod，引用 build-robot-2 不存在的secret时会出错
apiVersion: v1
kind: Pod
metadata:
  name: nginx-serviceaccount-4
  namespace: default
spec:
  containers:
  - image: nginx:1.7.9
    imagePullPolicy: IfNotPresent
    name: nginx
    volumeMounts:
    - name: foo
      mountPath: /var/run/secrets/kubernetes.io/serviceaccount
      readOnly: true
  restartPolicy: Always
  serviceAccount: build-robot-2
  volumes:
  - name: foo
    secret:
      secretName: test-secret
[root@tjwq01-sys-bs003007 ~]# kubectl create -f test-serviceaccount-pod4.yaml
Error from server (Forbidden): error when creating "test-serviceaccount-pod4.yaml": pods "nginx-serviceaccount-4" is forbidden: volume with secret.secretName="test-secret" is not allowed because service account build-robot-2 does not reference that secret


[root@tjwq01-sys-bs003007 ~]# cat build-robot-secret-3.yaml
apiVersion: v1
kind: Secret
metadata:
  name: build-robot-secret-3
  annotations:
    kubernetes.io/service-account.name: build-robot-2
data:
  data: dGVzdCBkYXRhCg==
type: kubernetes.io/service-account-token
[root@tjwq01-sys-bs003007 ~]# kubectl create -f build-robot-secret-3.yaml
secret "build-robot-secret-3" created
[root@tjwq01-sys-bs003007 ~]# kubectl get serviceaccount  build-robot-2 -o yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  annotations:
    kubernetes.io/enforce-mountable-secrets: "true"
  creationTimestamp: 2017-03-14T16:43:05Z
  name: build-robot-2
  namespace: default
  resourceVersion: "959426"
  selfLink: /api/v1/namespaces/default/serviceaccounts/build-robot-2
  uid: 4c529d78-08d5-11e7-8101-8cdcd4b3be48
secrets:
- name: build-robot-2-token-95hdz
[root@tjwq01-sys-bs003007 ~]# kubectl describe serviceaccount  build-robot-2
Name:           build-robot-2
Namespace:      default
Labels:         <none>

Mountable secrets:      build-robot-2-token-95hdz

Tokens:                 build-robot-2-token-95hdz
                        build-robot-secret-3

Image pull secrets:     <none>
[root@tjwq01-sys-bs003007 ~]# cat  test-serviceaccount-pod5.yaml
apiVersion: v1
kind: Pod
metadata:
  name: nginx-serviceaccount-5
  namespace: default
spec:
  containers:
  - image: nginx:1.7.9
    imagePullPolicy: IfNotPresent
    name: nginx
    volumeMounts:
    - name: foo
      mountPath: /run/secret-mount
      readOnly: true
  restartPolicy: Always
  serviceAccount: build-robot-2
  volumes:
  - name: foo
    secret:
      secretName: build-robot-secret-3
[root@tjwq01-sys-bs003007 ~]# kubectl create -f test-serviceaccount-pod5.yaml
Error from server (Forbidden): error when creating "test-serviceaccount-pod5.yaml": pods "nginx-serviceaccount-5" is forbidden: volume with secret.secretName="build-robot-secret-3" is not allowed because service account build-robot-2 does not reference that secret

如果为serviceaccount指定了  annotations kubernetes.io/enforce-mountable-secrets: "true"，则创建pod时会对sercret是否和service account关联进行检查。

[root@tjwq01-sys-bs003007 ~]# kubectl get serviceaccount  build-robot-2 -o yaml >build-robot-serviceaccount-2.new.yaml
[root@tjwq01-sys-bs003007 ~]# vim build-robot-serviceaccount-2.new.yaml
[root@tjwq01-sys-bs003007 ~]# cat build-robot-serviceaccount-2.new.yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  annotations:
    kubernetes.io/enforce-mountable-secrets: "true"
  creationTimestamp: 2017-03-14T16:43:05Z
  name: build-robot-2
  namespace: default
  resourceVersion: "959426"
  selfLink: /api/v1/namespaces/default/serviceaccounts/build-robot-2
  uid: 4c529d78-08d5-11e7-8101-8cdcd4b3be48
secrets:
- name: build-robot-2-token-95hdz
- name: test-secret  # 相比build-robot-serviceaccount-2.yaml，多加了这一行；

[root@tjwq01-sys-bs003007 ~]# kubectl replace -f build-robot-serviceaccount-2.new.yaml
serviceaccount "build-robot-2" replaced
[root@tjwq01-sys-bs003007 ~]# kubectl get serviceaccount build-robot-2
NAME            SECRETS   AGE
build-robot-2   2         20m
[root@tjwq01-sys-bs003007 ~]# kubectl get serviceaccount build-robot-2 -o yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  annotations:
    kubernetes.io/enforce-mountable-secrets: "true"
  creationTimestamp: 2017-03-14T16:43:05Z
  name: build-robot-2
  namespace: default
  resourceVersion: "960511"
  selfLink: /api/v1/namespaces/default/serviceaccounts/build-robot-2
  uid: 4c529d78-08d5-11e7-8101-8cdcd4b3be48
secrets:
- name: build-robot-2-token-95hdz
- name: test-secret
[root@tjwq01-sys-bs003007 ~]# kubectl describe serviceaccount  build-robot-2
Name:           build-robot-2
Namespace:      default
Labels:         <none>

Image pull secrets:     <none>

Mountable secrets:      build-robot-2-token-95hdz
                        test-secret  # 注意增加了这一个secret；

Tokens:                 build-robot-2-token-95hdz
                        build-robot-secret-3   # 该secret类型为kubernetes.io/service-account-token，serviceaccount默认只能有一个该类型的sercret可以被Mountable，即自动创建的build-robot-2-token-95hdz；虽然该secret不能被pod的容器mount，但是如果知道了它的值，
                        也可以用来向apiserver发送请求；

[root@tjwq01-sys-bs003007 ~]# diff test-serviceaccount-pod5.yaml test-serviceaccount-pod6.yaml
4c4
<   name: nginx-serviceaccount-5
---
>   name: nginx-serviceaccount-6
20c20
<       secretName: build-robot-secret-3
---
>       secretName: test-secret
[root@tjwq01-sys-bs003007 ~]# kubectl create -f test-serviceaccount-pod6.yaml  # 创建成功
pod "nginx-serviceaccount-6" created
[root@tjwq01-sys-bs003007 ~]# kubectl exec -i -t nginx-serviceaccount-6 -c nginx bash
root@nginx-serviceaccount-6:/# ls /run/secret-mount/   # 成功挂载了secret和 API Token Secret
data-1  data-2
root@nginx-serviceaccount-6:/# ls /run/secrets/kubernetes.io/serviceaccount/
ca.crt  namespace  token
root@nginx-serviceaccount-6:/# cat /run/secrets/kubernetes.io/serviceaccount/token
eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJrdWJlcm5ldGVzL3NlcnZpY2VhY2NvdW50Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9uYW1lc3BhY2UiOiJkZWZhdWx0Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9zZWNyZXQubmFtZSI6ImJ1aWxkLXJvYm90LTItdG9rZW4tOTVoZHoiLCJrdWJlcm5ldGVzLmlvL3NlcnZpY2VhY2NvdW50L3NlcnZpY2UtYWNjb3VudC5uYW1lIjoiYnVpbGQtcm9ib3QtMiIsImt1YmVybmV0ZXMuaW8vc2VydmljZWFjY291bnQvc2VydmljZS1hY2NvdW50LnVpZCI6IjRjNTI5ZDc4LTA4ZDUtMTFlNy04MTAxLThjZGNkNGIzYmU0OCIsInN1YiI6InN5c3RlbTpzZXJ2aWNlYWNjb3VudDpkZWZhdWx0OmJ1aWxkLXJvYm90LTIifQ.GF-b9wTPLcMnHTfXnkkV1VPpdq0mlWrgrsisJTLsd95m_i7xz9YxeZZlgAlZGFnfd2SC6gWv9BX6uI_ZyhjLg8sGtkOjQV1vI9RXMxrymU-bDck5SuhABSMG8Nzuhj9JMvLRjCpBijFdDUJAHaGt8i2LhvpXZ4cF1e6R_8ARSRhFOoXak0qNzVushJqTlev5Y-oHhC6DpPKsb-4Wjc6vUha0bx85Nwq0PBhHB8ce5sQXzS5LQO8m6y-ZwTVrT6VYU-ToE4nznmGByVdqgE3mOlBdqq1vOzjZirkDLg-UM7AseXfM5sLQD9jSeVk1vHu9UZmGnnpkgrJ5ls-J7ftK9Qroot@nginx-serviceaccount-6:/#
[root@tjwq01-sys-bs003007 ~]# kubectl describe secrets build-robot-2-token-95hdz  # 上面pod容器自动挂载的ca和build-robot-2自动创建的 API Token Secret一致；
Name:           build-robot-2-token-95hdz
Namespace:      default
Labels:         <none>
Annotations:    kubernetes.io/service-account.name=build-robot-2
                kubernetes.io/service-account.uid=4c529d78-08d5-11e7-8101-8cdcd4b3be48

Type:   kubernetes.io/service-account-token

Data
====
ca.crt:         1208 bytes
namespace:      7 bytes
token:          eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJrdWJlcm5ldGVzL3NlcnZpY2VhY2NvdW50Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9uYW1lc3BhY2UiOiJkZWZhdWx0Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9zZWNyZXQubmFtZSI6ImJ1aWxkLXJvYm90LTItdG9rZW4tOTVoZHoiLCJrdWJlcm5ldGVzLmlvL3NlcnZpY2VhY2NvdW50L3NlcnZpY2UtYWNjb3VudC5uYW1lIjoiYnVpbGQtcm9ib3QtMiIsImt1YmVybmV0ZXMuaW8vc2VydmljZWFjY291bnQvc2VydmljZS1hY2NvdW50LnVpZCI6IjRjNTI5ZDc4LTA4ZDUtMTFlNy04MTAxLThjZGNkNGIzYmU0OCIsInN1YiI6InN5c3RlbTpzZXJ2aWNlYWNjb3VudDpkZWZhdWx0OmJ1aWxkLXJvYm90LTIifQ.GF-b9wTPLcMnHTfXnkkV1VPpdq0mlWrgrsisJTLsd95m_i7xz9YxeZZlgAlZGFnfd2SC6gWv9BX6uI_ZyhjLg8sGtkOjQV1vI9RXMxrymU-bDck5SuhABSMG8Nzuhj9JMvLRjCpBijFdDUJAHaGt8i2LhvpXZ4cF1e6R_8ARSRhFOoXak0qNzVushJqTlev5Y-oHhC6DpPKsb-4Wjc6vUha0bx85Nwq0PBhHB8ce5sQXzS5LQO8m6y-ZwTVrT6VYU-ToE4nznmGByVdqgE3mOlBdqq1vOzjZirkDLg-UM7AseXfM5sLQD9jSeVk1vHu9UZmGnnpkgrJ5ls-J7ftK9Q

# service account也用于创建ImagePullSecrets
Adding ImagePullSecrets to a service account

First, create an imagePullSecret, as described here Next, verify it has been created. For example:
$ kubectl get secrets myregistrykey
NAME             TYPE                              DATA
myregistrykey    kubernetes.io/.dockerconfigjson   1
Next, read/modify/write the service account for the namespace to use this secret as an imagePullSecret
$ kubectl get serviceaccounts default -o yaml > ./sa.yaml
$ cat sa.yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  creationTimestamp: 2015-08-07T22:02:39Z
  name: default
  namespace: default
  resourceVersion: "243024"
  selfLink: /api/v1/namespaces/default/serviceaccounts/default
  uid: 052fb0f4-3d50-11e5-b066-42010af0d7b6
secrets:
- name: default-token-uudge
$ vi sa.yaml
[editor session not shown]
[delete line with key "resourceVersion"]
[add lines with "imagePullSecret:"]
$ cat sa.yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  creationTimestamp: 2015-08-07T22:02:39Z
  name: default
  namespace: default
  selfLink: /api/v1/namespaces/default/serviceaccounts/default
  uid: 052fb0f4-3d50-11e5-b066-42010af0d7b6
secrets:
- name: default-token-uudge
imagePullSecrets:
- name: myregistrykey

$ kubectl replace serviceaccount default -f ./sa.yaml
serviceaccounts/default
Now, any new pods created in the current namespace will have this added to their spec:
spec:
  imagePullSecrets:
  - name: myregistrykey


# Secrets

Kubernetes提供了Secret来处理敏感信息，目前Secret的类型有3种：

1. Opaque(default): 任意字符串
2. kubernetes.io/service-account-token: 用做ServiceAccount的API Token Secret，被pod的容器自动挂载；绝大多数场景下，不应该人工创建该类型的secret；
3. kubernetes.io/dockerconfigjson: 作用于Docker registry