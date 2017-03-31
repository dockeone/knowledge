
kubelet的功能是发现并运行pod，有三种方式发现pod：
1. 通过--config参数指定一个manifests目录，该目录下保存了pod的YAML或JSON文件，kubelet周期检查这个文件的状态；
2. 通过URL指定pod manifests；
3. 从--apiserver参数指定的服务器获取pods；

kubelet可以脱离apiserver运行！

pod里面的contianer共享一些资源：IP、端口、卷, 这是通过和pod container一起运行的基础设施容器(infrastructure container)来实现的：
1. kubelet将所有pod内各容器共享的资源放到基础设施容器；
2. pod内只有基础设施容器分配有IP，其它容器没有分配IP，可以通过命令验证： docker inspect --format '{{ .NetworkSettings.IPAddress  }}' f1a27680e401；
3. 其它容器的NetworkMode指向基础设施容器，这样达到复用基础设施容器网络的目的，可以通过命令验证：docker inspect --format '{{ .HostConfig.NetworkMode  }}' c5e357fc981a；
4. 共享基础设施容器其实是一个129byte的ELF二进制文件，它调用pause 系统调用，直到收到信号后退出；这可以保证基础设施容器一直运行，直到kubelet关闭它；

kubelet启动时，自动向apiserver注册它所在的node；
kubelt对外暴露了rest HTTP 接口，apiserver可以通过这个接口过去pod、headz等状态；

$ curl http://localhost:10255/healthz
ok

There are also a few status endpoints. For example, you can get a list of running pods at /pods:

$ curl --stderr /dev/null http://localhost:10255/pods | jq . | head
{
  "kind": "PodList",
  "apiVersion": "v1",
  "metadata": {},
  "items": [
    {
      "metadata": {
        "name": "nginx-kx",
        "namespace": "default",
        "selfLink": "/api/v1/pods/namespaces/nginx-kx/default",

You can also get specs of the machine the kubelet is running on at /spec/:

$ curl --stderr /dev/null  http://localhost:10255/spec/ | jq . | head
{
  "num_cores": 4,
  "cpu_frequency_khz": 2700000,
  "memory_capacity": 4051689472,
  "machine_id": "9eacc5220f4b41e0a22972d8a47ccbe1",
  "system_uuid": "818B908B-D053-CB11-BC8B-EEA826EBA090",
  "boot_id": "a95a337d-6b54-4359-9a02-d50fb7377dd1",
  "filesystems": [
    {
      "device": "/dev/mapper/kx--vg-root",

spec.containers[].imagePullPolicy: 
    获取镜像的策略，可选值Always、Never(只是用本地)、IfNotPresent(本地有时使用本地，否则下载);
    如果image没有指定tag则默认策略为Always，否则为IfNotPresent
    
spec.restartPolicy: Pod的重启策略，适用于Pod内的所有容器，可选值为Always(默认)、OnFailure、Never
    Always: 容器一旦终止运行，kubelet就重启它；
    OnFailure: 容器以非零状态码终止时，kubelet才会重启该容器，如果容器正常结束则kubelet不会重启它；
    Never：kubelet不重启容器；
    Master： 容器终止后，kubelet将退出码发给Master，而不重启它；

Pod的状态：
1. Pending：apiserver已创建该Pod，但是Pod内还有一个或多个容器的镜像没有创建，如正在下载镜像；
2. Running：Pod内的**所有容器**均已创建，且至少有一个容器处于运行、正在启动、正在重启的状态；
3. Succeeded：Pod内**所有容器**均成功退出，且不会重启；
4. Failed： Pod内**所有容器**均退出，但至少有一个容器失败退出；
5. Unknown：无法获取pod状态；

kublet重启失效容器的时间间隔以sync-frequency乘与2n来计算，最长延时5分钟，在成功重启后的10分钟重置该时间；
Pod的重启策略与控制方式相关，每种控制器对Pod的重启策略要求如下：
1. RC、Deployment、Replica Set、DaemonSet：必须设置为Always；
2. Job：OnFailure或Never, 确保容器执行完后不再重启；
3. kubelet：在Pod失效时无条件自动重启，不会对Pod进行健康检查；

kubectl get pods 返回当前正在运行的pod，--show-all参数指定返回所有的pods；
Pod挂载ConfigMap时，容器内部职能挂载为目录，无法挂载为文件。如果该目录下还有其它文件，则容器内的该目录会被挂载的ConfigMap说覆盖；

对Pod的健康检查是通过两类探针来实现：
1. LivenessProbe: 用于判断容器是否存活(Running)状态，如果不健康，则kubelet杀掉该容器，然后根据重启策略做相应处理；
2. ReadinessProbe: 用于判断容器是否启动完成，可以接收请求，如果探测到不健康，Endpoint controller将从Service的Endpoints中删除包含
该容器所在Pod的Endpoint；

LivenessProbe的方式：
1. ExecAction：在容器内部执行一条命令，根据返回值判断是否健康；
2. TCPSocketAction： 对容器的IP地址和端口号执行TCP连接检查；
3. HTTPGetAction: 对容器的IP地址、端口号和路径调用HTTP Get方法，如果返回码位于200和400间，则认为健康；

Ingress： HTTP 7 层路由机制
Service的表现形式是ClusterIP:Port，工作在TCP/IP层，而对于HTTP服务来说，不同的URL经常对应到不同的后端服务，Ingress就是用来实现HTTP层
的业务路由机制。Ingress的实现需要通过Ingress的定义、Ingress Controller结合起来，才能形成完整的HTTP负载分发机制；

如果pod的状态一直是terminating, 说明kubelet删除pod的逻辑失败了，可以通过查看kubelet的日志或docker的日志来定位原因；

# Graceful shutdown
https://pracucci.com/graceful-shutdown-of-kubernetes-pods.html

When a pod should be terminated:

1. A **SIGTERM** signal is sent to the main process (PID 1) in each container, and a **“grace period”** countdown starts (defaults to 30 seconds - see below to change it).
1. Upon the receival of the SIGTERM, each container should start a graceful shutdown of the running application and exit.
1. If a container doesn’t terminate within the grace period, a **SIGKILL** signal will be sent and the container violently terminated.

为了使app能收到这两个信号：
1. run the CMD in the exec form： CMD [ "myapp" ]
2. run the command with Bash: CMD [ "/bin/bash", "-c", "myapp --arg=$ENV_VAR" ]

bash会将收到的信号传递给child process，而其它shell如Alpine则不会；

修改 grace period:
1. 默认为30s；
2. 可以在 deployment .yaml 文件，或 命令行参数上指定(如kubectl apply、patch、delete、replace等会删除旧的pod的情景)
3. 其它情况，如 SIGTERM 会kill app的情况如Nginx，应该使用 /usr/sbin/nginx -s quit  来gracefully terminate it，可以使用preStop hook:


apiVersion: extensions/v1beta1
kind: Deployment
metadata:
    name: test
spec:
    replicas: 1
    template:
        spec:
            containers:
              - name: test
                image: ...
            terminationGracePeriodSeconds: 60

apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: nginx
spec:
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx
        ports:
        - containerPort: 80
        lifecycle:
          preStop:
            exec:
              # SIGTERM triggers a quick exit; gracefully terminate instead
              command: ["/usr/sbin/nginx","-s","quit"]