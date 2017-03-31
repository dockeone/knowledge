# About Registry

https://docs.docker.com/registry/introduction/#understanding-image-naming

A registry is a storage and content delivery system, holding named Docker images, available in different tagged versions.

Example: the image distribution/registry, with tags 2.0 and 2.1.

Users interact with a registry by using docker push and pull commands.
Example: docker pull registry-1.docker.io/distribution/registry:2.1.

Storage itself is delegated to **drivers**. The default storage driver is the local posix filesystem, which is suitable for development or small deployments. Additional **cloud-based storage drivers** like S3, Microsoft Azure, OpenStack Swift and Aliyun OSS are also supported. People looking into using other storage backends may do so by writing their own driver implementing the Storage API.

Since securing access to your hosted images is paramount, the Registry natively **supports TLS and basic authentication**.

注意：basic authentication必须在TLS情况下使用，即registry必须是安全模式的；

Finally, the Registry ships with **a robust notification system**, calling webhooks in response to activity, and both extensive logging and reporting, mostly useful for large installations that want to collect metrics.


# Understanding image naming

Image names as used in typical docker commands reflect their origin:
  1.   ```docker pull ubuntu```instructs docker to pull an image named ubuntu from **the official Docker Hub**. This is simply a shortcut for the longer ```docker pull docker.io/library/ubuntu```command
  2. ```docker pull myregistrydomain:port/foo/bar``` instructs docker to contact the registry located at myregistrydomain:port to find the image foo/bar

You can find out more about the various Docker commands dealing with images in the official Docker engine documentation.

# Use cases

Running your own Registry is a great solution to integrate with and complement your CI/CD system. In a typical workflow, **a commit to your source revision control system would trigger a build on your CI system, which would then push a new image to your Registry if the build is successful. A notification from the Registry would then trigger a deployment on a staging environment, or notify other systems that a new image is available**.

It’s also an essential component if you want to quickly deploy a new image over a large cluster of machines.
Finally, it’s the best way to distribute images inside an isolated network.


# Insecure registries

https://docs.docker.com/engine/reference/commandline/dockerd/#insecure-registries

Docker considers a private registry either secure or insecure. In the rest of this section, registry is used for private registry, and myregistry:5000 is a placeholder example for a private registry.

A secure registry **uses TLS and a copy of its CA certificate** is placed on the Docker host at **/etc/docker/certs.d/myregistry:5000/ca.crt**. An insecure registry is **either not using TLS (i.e., listening on plain text HTTP), or is using TLS with a CA certificate not known by the Docker daemon**. The latter can happen when the certificate was not found under /etc/docker/certs.d/myregistry:5000/, or if the certificate verification failed (i.e., wrong CA).

1. 安全的registry: 需要启用TLS，用CA检查registry的证书(registry可以不对发起请求的docker cli证书进行检查)；
2. 非安全的registry: 不启用TLS，或者启用了TLS但是不对**registry的证书**进行检查；

registry的CA证书位置和它的名称&端口有关，如myregistry:5000的默认证书位置是/etc/docker/certs.d/myregistry:5000/ca.crt，这样做的好处是一个docker daemon可以连接多个使用不同
CA证书签名的registry；

缺省请情况下，docker假设除了本地registry外的其它registries都是使用secure模式，所以和非安全的registies进行通信会失败。

如果在docker daemon的配置文件中指定了 --insecure-registry 参数，则docker可以和对应的registry使用非HTTPs模式通信；

By default, Docker assumes all, but local (see local registries below), registries are secure. Communicating with an insecure registry is not possible if Docker assumes that registry is secure. In order to communicate with an insecure registry, the Docker daemon requires ```--insecure-registry``` in one of the following two forms:
1. --insecure-registry myregistry:5000 tells the Docker daemon that myregistry:5000 should be considered insecure.
2. --insecure-registry 10.1.0.0/16 tells the Docker daemon that all registries whose domain resolve to an IP address is part of the subnet described by the CIDR syntax, should be considered insecure.

The flag can be **used multiple times** to allow multiple registries to be marked as insecure.

If an insecure registry is not marked as insecure, docker pull, docker push, and docker search will result in an error message prompting the user to either secure or pass the --insecure-registry flag to the Docker daemon as described above.

Local registries, whose IP address falls in the 127.0.0.0/8 range, are automatically marked as insecure as of Docker 1.3.2. It is not recommended to rely on this, as it may change in the future.

本地仓库自动被标记为insecure，故不需要再命令行或配置文件中指定--insecure-registry;

Enabling --insecure-registry, i.e., allowing un-encrypted and/or untrusted communication, can be useful when running a local registry. However, because its use creates security vulnerabilities it should ONLY be enabled for testing purposes. For increased security, users should add their CA to their system’s list of trusted CAs instead of enabling --insecure-registry.


# 创建 docker registry的配置文件

https://docs.docker.com/registry/configuration/#list-of-configuration-options

[root@tjwq01-sys-bs003007 ~]# cat docker_registry_config.yml
version: 0.1
log:
  fields:
    service: registry

storage:  # 这里使用 ceph rgw 存储
  swift:
    authurl: http://tjwq01-sys-power003009.tjwq01:7480/auth/v1
    username: demo:swift
    password: aCgVTx3Gfz1dBiFS4NfjIRmvT0sgpHDP6aa0Yfrh
    container: registry

http:
  addr: 0.0.0.0:8000
  headers:
    X-Content-Type-Options: [nosniff]

health:
  storagedriver:
    enabled: true
    interval: 10s
    threshold: 3


# 创建 docker registry

[root@tjwq01-sys-bs003007 ~]# sudo docker run -d -p 8000:8000 \
    -v /root/docker_registry_config.yml:/etc/docker/registry/config.yml \
    -v `pwd`/data:/var/lib/registry \
    --name registry registry
82e121ea2698981d91d3813a53fc900ac1b213d3ed192055c7e18f1ea4340477

[root@tjwq01-sys-bs003007 ~]# sudo docker logs registry
time="2017-03-20T06:48:29Z" level=warning msg="No HTTP secret provided - generated random secret. This may cause problems with uploads if multiple registries are behind a load-balancer. To provide a shared secret, fill in http.secret in the configuration file or set the REGISTRY_HTTP_SECRET environment variable." go.version=go1.7.3 instance.id=0dda4977-cb70-47ce-a35e-b6e62bcb84e1 version=v2.6.0
time="2017-03-20T06:48:29Z" level=info msg="Starting upload purge in 27m0s" go.version=go1.7.3 instance.id=0dda4977-cb70-47ce-a35e-b6e62bcb84e1 version=v2.6.0
time="2017-03-20T06:48:29Z" level=info msg="redis not configured" go.version=go1.7.3 instance.id=0dda4977-cb70-47ce-a35e-b6e62bcb84e1 version=v2.6.0
time="2017-03-20T06:48:29Z" level=info msg="listening on [::]:8000" go.version=go1.7.3 instance.id=0dda4977-cb70-47ce-a35e-b6e62bcb84e1 version=v2.6.0

# 存储

By default, your registry data is persisted as **a docker volume** on the host filesystem. Properly understanding volumes is essential if you want to stick with a local filesystem storage.

Specifically, you might want to point your volume location to a specific place in order to more easily access your registry data. To do so you can:

    docker run -d -p 5000:5000 --restart=always --name registry -v `pwd`/data:/var/lib/registry registry:2

除了默认使用的Host文件系统外，也可以使用其它的存储后端如S3、GCE等；

# 删除 docker registry

docker stop registry && docker rm -v registry  # 删除时需要指定 -v 参数才能删除创建的volume

# 给本地的一个image添加私有docker registry的tag

[root@tjwq01-sys-bs003007 ~]# docker tag docker.io/kubernetes/pause localhost:8000/zhangjun3/pause
[root@tjwq01-sys-bs003007 ~]# docker images
REPOSITORY                                            TAG                 IMAGE ID            CREATED             SIZE
docker.io/registry                                    latest              047218491f8c        2 weeks ago         33.17 MB
docker.io/kubernetes/pause                            latest              f9d5de079539        2 years ago         239.8 kB
localhost:8000/zhangjun3/pause                        latest              f9d5de079539        2 years ago         239.8 kB

# 在本地将image push到是有registry

[root@tjwq01-sys-bs003007 ~]# docker push localhost:8000/zhangjun3/pause
The push refers to a repository [localhost:8000/zhangjun3/pause]
5f70bf18a086: Pushed
e16a89738269: Pushed
latest: digest: sha256:9a6b437e896acad3f5a2a8084625fdd4177b2e7124ee943af642259f2f283359 size: 916

# 在本地push和pull进行均OK

[root@tjwq01-sys-bs003007 ~]# docker pull localhost:8000/zhangjun3/pause
Using default tag: latest
Trying to pull repository localhost:8000/zhangjun3/pause ...
latest: Pulling from localhost:8000/zhangjun3/pause
Digest: sha256:9a6b437e896acad3f5a2a8084625fdd4177b2e7124ee943af642259f2f283359
Status: Image is up to date for localhost:8000/zhangjun3/pause:latest

# 从其它机器push或pull镜像则需要验证

[root@tjwq01-sys-power003008 docker]# docker pull tjwq01-sys-bs003007.tjwq01:8000/zhangjun3/pause
Using default tag: latest
Error response from daemon: Get https://tjwq01-sys-bs003007.tjwq01:8000/v1/_ping: http: server gave HTTP response to HTTPS client

这是由于docker缺省假设registry是安全模式的，如果registry不提供证书，或提供的证书没有使用docker daemon本地保存的CA签名，则认为是非安全的，拒绝通信；

解决办法：
http://tonybai.com/2016/02/26/deploy-a-private-docker-registry/

1. 在需要pull或push image的docker机器上(不需要修改registry机器的docker参数)，修改docker daemon的参数，然后重启docker daemon：

    a. 对于新版的docker，配置文件是 /etc/docker/daemon.json
    [root@tjwq01-sys-power003008 ~]# cat /etc/docker/daemon.json
    {
        "insecure-registries": ["tjwq01-sys-bs003007.tjwq01:8000"]
    }
    [root@tjwq01-sys-power003008 ~]# systemctl restart docker
    [root@tjwq01-sys-power003008 ~]# docker pull tjwq01-sys-bs003007.tjwq01:8000/zhangjun3/pause
    Using default tag: latest
    latest: Pulling from zhangjun3/pause
    a3ed95caeb02: Pull complete
    f72a00a23f01: Pull complete
    Digest: sha256:9a6b437e896acad3f5a2a8084625fdd4177b2e7124ee943af642259f2f283359

    b. 老版的docker，配置文件是：/etc/sysconfig/docker

2. 配置registry使用TLS模式，证书经过CA签名，同时将CA保存到docker client机器上(如果CA是知名的，就不需要在每台机器上保存)：

# 配置安全的registry(tjwq01-sys-power003008.tjwq01)

创建目录
$ mkdir -p /root/registry/certs
$ cd /root/registry/certs

生成ca证书私钥
$ openssl genrsa -out ca.key 2048

生成ca证书
$ openssl req -x509 -new -nodes -key ca.key -subj "/CN=${MASTER_IP}" -days 10000 -out ca.crt

生成server私钥
$ openssl genrsa -out server.key 2048

创建server证书请求的配置文件
$ cat server_ssl.cnf
[ req ]
req_extensions                = v3_req
distinguished_name            = req_distinguished_name

[ req_distinguished_name ]

[ v3_ca ]
basicConstraints              = CA:TRUE
subjectKeyIdentifier          = hash
authorityKeyIdentifier        = keyid:always,issuer:always

[ v3_req ]
basicConstraints              = CA:FALSE
keyUsage                      = nonRepudiation, digitalSignature, keyEncipherment
subjectAltName                = @alt_names

[ alt_names ]  # 根据client请求server的方式（域名、IP）修改下面的DNS和IP记录
DNS.1 = kubernetes
DNS.2 = kubernetes.default
DNS.3 = kubernetes.default.svc
DNS.4 = kubernetes.default.svc.cluster.local
DNS.5 = k8s-master
IP.1 = 10.64.3.7  # apiserver 的 NodeIP
IP.2 = 10.254.0.1 # apiserver 的 ClusterIP

创建server私钥的证书签名请求；
$ openssl req -new -key server.key -subj "/CN=${MASTER_IP}" -config server_ssl.cnf -out server.csr

apiserver从证书的CN中提取username，所以如果该证书是客户端使用，CN应该设置为客户端名称；

使用ca对证书签名请求进行签名，生成server的证书；
$ openssl x509 -req -in server.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out server.crt -days 10000 -extensions v3_req -extfile server_ssl.cnf

查看证书：
$ openssl x509  -noout -text -in ./server.crt

$ ls certs/
ca.crt  ca.key  ca.srl  server.crt  server.csr  server.key  server_ssl.cnf

创建registry的配置文件
配置文件启用的功能：
1. 使用 ceph rgw 的swift接口保存image；
2. 允许删除image（默认不允许），需要指定digest;

$ cd /root/registry
$ cat config.yml
# https://docs.docker.com/registry/configuration/#list-of-configuration-options
version: 0.1
log:
  level: debug
  fromatter: text
  fields:
    service: registry

storage:  # 使用ceph rgw对象存储，使用swfit接口
  cache:  # proxy缓存，目前**只支持缓存layer metadata**，不能缓存layer data;
    blobdescriptor: inmemory # 使用内存缓存，也可以使用redis，需要再指定redis的配置参数
  delete:
    enabled: true
  swift:
    authurl: http://tjwq01-sys-power003009.tjwq01:7480/auth/v1
    username: demo:swift
    password: aCgVTx3Gfz1dBiFS4NfjIRmvT0sgpHDP6aa0Yfrh
    container: registry

http:
  addr: 0.0.0.0:8000
  headers:
    X-Content-Type-Options: [nosniff]
  tls:
    certificate: /certs/server.crt  # registry证书
    key: /certs/server.key          # registry证书对应的私钥
    #clientcas:
    #  - /certs/ca.crt              # 签名客户端证书的CA，这里不对客户端证书进行校验

proxy:                              # 开启pull-through cache，一旦开启，就不支持向该registry push镜像了；
   remoteurl: http://hub-mirror.c.163.com  # 目前支持和docker hub同步的registry，故不能proxy自己的私有registry

health:
  storagedriver:
    enabled: true
    interval: 10s
    threshold: 3

运行启用代理缓存和TLS安全的registry
$ sudo docker run -d -p 8000:8000 -v `pwd`/config.yml:/etc/docker/registry/config.yml -v `pwd`/certs:/certs  --name registry registry:2

# registry本地push和pull image

将image打本地registry的tag
$ docker tag docker.io/kubernetes/pause localhost:8000/zhangjun3/pause
$ docker images
REPOSITORY                                            TAG                 IMAGE ID            CREATED             SIZE
docker.io/registry                                    latest              047218491f8c        2 weeks ago         33.17 MB
docker.io/kubernetes/pause                            latest              f9d5de079539        2 years ago         239.8 kB
localhost:8000/zhangjun3/pause                        latest              f9d5de079539        2 years ago         239.8 kB

将image push到是有registry
$ docker push localhost:8000/zhangjun3/pause
The push refers to a repository [localhost:8000/zhangjun3/pause]
5f70bf18a086: Pushed
e16a89738269: Pushed
latest: digest: sha256:9a6b437e896acad3f5a2a8084625fdd4177b2e7124ee943af642259f2f283359 size: 916


# 在另外一台机器（tjwq01-sys-power003008.tjwq01）上测试私有registry

确保无inscure-registry相关的配置
$ cat /etc/docker/daemon.json 

使用主机名请求registry时，docker cli对registry提供的证书校验出错，因为registry的server.crt中的alt subject names指定了证书的有效
范围
$ docker pull tjwq01-sys-bs003007.tjwq01:8000/zhangjun3/pause
Using default tag: latest
Error response from daemon: Get https://tjwq01-sys-bs003007.tjwq01:8000/v1/_ping: x509: certificate is valid for kubernetes, kubernetes.default, kubernetes.default.svc, kubernetes.default.svc.cluster.local, k8s-master, not tjwq01-sys-bs003007.tjwq01

使用registry的NodeIP请求，这次是证书校验出错
$ docker pull 10.64.3.7:8000/zhangjun3/pause
Using default tag: latest
Error response from daemon: Get https://10.64.3.7:8000/v1/_ping: x509: certificate signed by unknown authority

将签署registry证书的ca证书拷贝到/etc/docker/certs.d/10.64.3.7:8000目录下
$ mkdir -p /etc/docker/certs.d/10.64.3.7:8000
$ mv /path/to/registry/ca.crt /etc/docker/certs.d/10.64.3.7:8000

测试请求是否OK
$ curl -I --cacert /etc/docker/certs.d/10.64.3.7\:8000/ca.crt https://10.64.3.7:8000/v2/ 
HTTP/1.1 200 OK
Content-Length: 2
Content-Type: application/json; charset=utf-8
Docker-Distribution-Api-Version: registry/2.0
X-Content-Type-Options: nosniff
Date: Tue, 21 Mar 2017 04:14:17 GMT

拉取私有rgistry本地保存的image成功
$ docker pull 10.64.3.7:8000/zhangjun3/pause
Using default tag: latest
latest: Pulling from zhangjun3/pause
a3ed95caeb02: Pull complete
f72a00a23f01: Pull complete
Digest: sha256:9a6b437e896acad3f5a2a8084625fdd4177b2e7124ee943af642259f2f283359
Status: Downloaded newer image for 10.64.3.7:8000/zhangjun3/pause:latest

拉取官方repo中的image失败
$ docker pull 10.64.3.7:8000/redis
Using default tag: latest
Pulling repository 10.64.3.7:8000/redis
Error: image redis:latest not found

解决方法是：
在daocker daemon的配置文件中添加registry-mirrors参数，指定私有registry的地址，这样后续执行docker cli时无需指定registry的地址；
docker默认尝试了所有registry-mirrors都没有找到对应的image，**docker最终会尝试从https://registry-1.docker.io获取镜像**：
$ cat /etc/docker/daemon.json
{
  "registry-mirrors": ["10.64.3.7:8000"]
}
$ systemctl restart docker

pull私有registry中的image
$ docker pull zhangjun3/pause
Using default tag: latest
latest: Pulling from zhangjun3/pause
a3ed95caeb02: Pull complete
f72a00a23f01: Pull complete
Digest: sha256:9a6b437e896acad3f5a2a8084625fdd4177b2e7124ee943af642259f2f283359
Status: Downloaded newer image for zhangjun3/pause:latest

pull位于docker hub中的image，私有registry正确地做了proxy：
$ docker pull redis
Using default tag: latest
latest: Pulling from library/redis
693502eb7dfb: Pulling fs layer
338a71333959: Pulling fs layer
83f12ff60ff1: Downloading
4b7726832aec: Waiting
19a7e34366a6: Waiting
622732cddc34: Waiting
3b281f2bcae3: Waiting
latest: Pulling from library/redis
693502eb7dfb: Pull complete
338a71333959: Pull complete
83f12ff60ff1: Pull complete
4b7726832aec: Pull complete
19a7e34366a6: Pull complete
622732cddc34: Pull complete
3b281f2bcae3: Pull complete
Digest: sha256:4c8fb09e8d634ab823b1c125e64f0e1ceaf216025aa38283ea1b42997f1e8059
Status: Downloaded newer image for redis:latest

如果registry启用了proxy，则会变成**只读的pull-through cache，不能再向其中push镜像**
$ docker tag zhangjun3/pause 10.64.3.7:8000/zhangjun3/pause2
$ docker push 10.64.3.7:8000/zhangjun3/pause2
The push refers to a repository [10.64.3.7:8000/zhangjun3/pause2]
5f70bf18a086: Retrying in 12 seconds
e16a89738269: Retrying in 12 seconds

查看registry的容器日志，可以看到 405 The operation is unsupported.的日志。https://docs.docker.com/registry/spec/api/

解决办法，关闭registry中的proxy相关配置，重启registry容器；

# 最佳实践

1. 运行一个私有registry，只用来保存自己的镜像，可以pull和push；
1. 运行另一个私有registry，配置proxy，指定一个registry-mirror的地址如：http://hub-mirror.c.163.com；
1. 将第二个registry的地址加到需要执行 git pull 的docker daemon配置文件中，如：
$ cat /etc/docker/daemon.json
{
  "registry-mirrors": ["10.64.3.7:8000"]
}
$ systemctl restart docker

可以使用 ustc 的 registry-mirror，速度比较快
$ cat /etc/docker/daemon.json
{
  "registry-mirrors": ["https://docker.mirrors.ustc.edu.cn", "http://hub-mirror.c.163.com/"]
}

1. 拉取私有镜像的image(需要指定registry的IP和端口)
$ docker pull 10.64.3.7：8000/zhangjun3/pause
Using default tag: latest
latest: Pulling from zhangjun3/pause
a3ed95caeb02: Pull complete
f72a00a23f01: Pull complete
Digest: sha256:9a6b437e896acad3f5a2a8084625fdd4177b2e7124ee943af642259f2f283359
Status: Downloaded newer image for zhangjun3/pause:latest

1. 通过 pull-through 的私有registry拉取docker hub的镜像：
$ docker pull redis
Using default tag: latest
latest: Pulling from library/redis
693502eb7dfb: Pulling fs layer
338a71333959: Pulling fs layer
83f12ff60ff1: Downloading
4b7726832aec: Waiting
19a7e34366a6: Waiting
622732cddc34: Waiting
3b281f2bcae3: Waiting


# 查询私有镜像中的images
$ curl  --cacert /etc/docker/certs.d/10.64.3.7\:8000/ca.crt https://10.64.3.7:8000/v2/_catalog
{"repositories":["library/redis","zhangjun3/busybox","zhangjun3/pause","zhangjun3/pause2"]}

# 查询某个镜像的tags列表
$ curl  --cacert /etc/docker/certs.d/10.64.3.7\:8000/ca.crt https://10.64.3.7:8000/v2/zhangjun3/busybox/tags/list
{"name":"zhangjun3/busybox","tags":["latest"]}

# 获取image或layer的digest

先向 v2/<repoName>/manifests/<tagName> 发GET请求，从响应的头部 Docker-Content-Digest 中获取digest，从响应的body的 fsLayers.blobSum 中获取 layDigests:
注意，必须包含请求头：Accept: application/vnd.docker.distribution.manifest.v2+json，否则获取的 Docker-Content-Digest 内容不对；

不带Accept头部，响应头中的Docker-Content-Digest**不是**image的digest，fsLayers.blobSum是各layout的digest
$ curl -v --cacert /etc/docker/certs.d/10.64.3.7\:8000/ca.crt https://10.64.3.7:8000/v2/zhangjun3/busybox/manifests/latest
< HTTP/1.1 200 OK
< Content-Length: 2745                                                                                                                               [34/1164]
< Content-Type: application/vnd.docker.distribution.manifest.v1+prettyjws
< Docker-Content-Digest: sha256:ea84cffd5ac1809f01a5a9a1d23c9dfa3dd3fd0a2c1a743e3dcb83e8b3b52381
< Docker-Distribution-Api-Version: registry/2.0
< Etag: "sha256:ea84cffd5ac1809f01a5a9a1d23c9dfa3dd3fd0a2c1a743e3dcb83e8b3b52381"
< X-Content-Type-Options: nosniff
< Date: Tue, 21 Mar 2017 15:21:10 GMT
<
{
   "schemaVersion": 1,
   "name": "zhangjun3/busybox",
   "tag": "latest",
   "architecture": "amd64",
   "fsLayers": [
      {
         "blobSum": "sha256:a3ed95caeb02ffe68cdd9fd84406680ae93d633cb16422d00e8a7c22955b46d4"
      },
      {
         "blobSum": "sha256:04176c8b224aa0eb9942af765f66dae866f436e75acef028fe44b8a98e045515"
      }
   ],
...

带Accept头部，响应头中的Docker-Content-Digest**是**image的digest
$ curl -v -H "Accept: application/vnd.docker.distribution.manifest.v2+json" --cacert /etc/docker/certs.d/10.64.3.7\:8000/ca.crt https://10.64.3.7:8000/v2/zhangjun3/busybox/manifests/latest

> GET /v2/zhangjun3/busybox/manifests/latest HTTP/1.1
> User-Agent: curl/7.29.0
> Host: 10.64.3.7:8000
> Accept: application/vnd.docker.distribution.manifest.v2+json
>
< HTTP/1.1 200 OK
< Content-Length: 527
< Content-Type: application/vnd.docker.distribution.manifest.v2+json
< Docker-Content-Digest: sha256:68effe31a4ae8312e47f54bec52d1fc925908009ce7e6f734e1b54a4169081c5
< Docker-Distribution-Api-Version: registry/2.0
< Etag: "sha256:68effe31a4ae8312e47f54bec52d1fc925908009ce7e6f734e1b54a4169081c5"
< X-Content-Type-Options: nosniff
< Date: Tue, 21 Mar 2017 15:19:42 GMT
<
{
   "schemaVersion": 2,
   "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
   "config": {
      "mediaType": "application/vnd.docker.container.image.v1+json",
      "size": 1465,
      "digest": "sha256:00f017a8c2a6e1fe2ffd05c281f27d069d2a99323a8cd514dd35f228ba26d2ff"
   },
   "layers": [
      {
         "mediaType": "application/vnd.docker.image.rootfs.diff.tar.gzip",
         "size": 701102,
         "digest": "sha256:04176c8b224aa0eb9942af765f66dae866f436e75acef028fe44b8a98e045515"
      }
   ]
}

# 删除image

向 /v2/<name>/manifests/<reference> 发送 DELETE请求，reference为上一步返回的HTTP头部中的Docker-Content-Digest字段内容，删除image；

$ curl -X DELETE  --cacert /etc/docker/certs.d/10.64.3.7\:8000/ca.crt https://10.64.3.7:8000/v2/zhangjun3/busybox/manifests/sha256:68effe31a4ae8312e47f54bec52d1fc925908009ce7e6f734e1b54a4169081c5

git pull删除的image时提示找不到
$ docker pull 10.64.3.7:8000/zhangjun3/busybox
Using default tag: latest
Pulling repository 10.64.3.7:8000/zhangjun3/busybox
Error: image zhangjun3/busybox:latest not found

但是通过API删除image只是soft delete，相关文件仍然存在于repo中，Garbge Collect阶段才会将这些不用的文件删除：
$ curl  --cacert /etc/docker/certs.d/10.64.3.7\:8000/ca.crt https://10.64.3.7:8000/v2/_catalog
{"repositories":["library/redis","zhangjun3/busybox","zhangjun3/pause","zhangjun3/pause2"]}

$ curl  --cacert /etc/docker/certs.d/10.64.3.7\:8000/ca.crt https://10.64.3.7:8000/v2/zhangjun3/busybox/tags/list
{"errors":[{"code":"NAME_UNKNOWN","message":"repository name not known to registry","detail":{"name":"zhangjun3/busybox"}}]}

# 删除layer

向/v2/<name>/blobs/<digest> 发送 DELETE 请求，其中 digest 是上面步骤响应body的 fsLayers.blobSum 中获取 layDigests

$ curl -X DELETE  --cacert /etc/docker/certs.d/10.64.3.7\:8000/ca.crt https://10.64.3.7:8000/v2/zhangjun3/busybox/blobs/sha256:a3ed95caeb02ffe68cdd9fd84406680ae93d633cb16422d00e8a7c22955b46d4

$ curl -X DELETE  --cacert /etc/docker/certs.d/10.64.3.7\:8000/ca.crt https://10.64.3.7:8000/v2/zhangjun3/busybox/blobs/sha256:04176c8b224aa0eb9942af765f66dae866f436e75acef028fe44b8a98e045515


# 垃圾回收和garbage-collect命令

https://docs.docker.com/registry/garbage-collection/#why-garbage-collection

## Why Garbage Collection?
Registry data can occupy considerable amounts of disk space and freeing up this disk space is an oft-requested feature. Additionally for reasons of security it can be desirable to ensure that certain layers no longer exist on the filesystem.

## Garbage Collection in the Registry
Filesystem layers are stored by their content address in the Registry. This has many advantages, one of which is that data is **stored once and referred to by many manifests**. See here for more details.

**Layers are therefore shared amongst manifests**; each manifest maintains a reference to the layer. As long as a layer is referenced by one manifest, it cannot be garbage collected.

Manifests and layers can be ‘deleted` with the registry API. **This API removes references to the target and makes them eligible for garbage collection**. It also makes them unable to be read via the API.

If a layer is deleted it will be removed from the filesystem **when garbage collection is run**. If a manifest is deleted the layers to which it refers will be removed from the filesystem if no other manifests refers to them.

## How Garbage Collection works
Garbage collection runs in two phases. First, in the ‘mark’ phase, the process scans all the manifests in the registry. From these manifests, it constructs a set of content address digests. This set is the ‘mark set’ and denotes the set of blobs to not delete. Secondly, in the ‘sweep’ phase, the process scans all the blobs and if a blob’s content address digest is not in the mark set, the process will delete it.

NOTE You should **ensure that the registry is in read-only mode or not running at all**. If you were to upload an image while garbage collection is running, there is the risk that the image’s layers will be mistakenly deleted, leading to a corrupted image.

This type of garbage collection is known as stop-the-world garbage collection. In future registry versions the intention is that garbage collection will be an automated background action and this manual process will no longer apply.

registry配置文件/etc/docker/registry/config.yml中的storage.maintenance中的相关内容就是用来定义GC相关的参数，默认启用，参数如下：
  
  maintenance:
    uploadpurging:  # 周期删除upload directories中的orphaned文件
      enabled: true
      age: 168h     # 一周
      interval: 24h # 每24小时删除一周前的无用数据
      dryrun: false
    readonly:
      enabled: false  # 删除数据时，是否将registry设置为readonly模式，可以在手动GC的时候将该参数设置为true，重启registry，然后执行GC命令，结束后设置为false再重启registry；


# 真正删除image数据

https://github.com/docker/docker-registry/issues/988
https://docs.docker.com/registry/garbage-collection/#how-garbage-collection-works
可以把脚本https://github.com/TranceMaster86/docker-reg-gc放到crontab中周期执行

docker exec -it <registry_container_id> bin/registry garbage-collect <path_to_registry_config>
<path_to_registry_config>=/etc/docker/registry/config.yml


# 启用 HTTP Basic认证

创建 HTTP Baisc 认证文件
$ mkdir auth
$ docker run --entrypoint htpasswd registry:2 -Bbn foo foo123  > auth/htpasswd
$ cat auth/htpasswd
foo:$2y$05$I60z69MdluAQ8i1Ka3x3Neb332yz1ioow2C4oroZSOE0fqPogAmZm

在config.yml文件中加入auth.htpasswd配置
auth:
  htpasswd:
    realm: basic-realm
    path: /auth/htpasswd

启动registry
$ docker run -d -p 8000:8000  -v /root/registry/auth/:/auth -v `pwd`/config.yml:/etc/docker/registry/config.yml -v `pwd`/certs:/certs  --name registry registry:2

在两台一台机器(tjwq01-sys-power003008) push 镜像失败，提示缺少 basic auth
$ docker tag zhangjun3/pause 10.64.3.7:8000/zhangjun3/pause2
$ docker push  10.64.3.7:8000/zhangjun3/pause2
The push refers to a repository [10.64.3.7:8000/zhangjun3/pause2]
5f70bf18a086: Preparing
e16a89738269: Preparing
no basic auth credentials

在两台一台机器(tjwq01-sys-power003008) pull 镜像时，提示**未找到镜像**
$ docker pull 10.64.3.7:8000/zhangjun3/nginx:1.7.9
Pulling repository 10.64.3.7:8000/zhangjun3/nginx
Error: image zhangjun3/nginx:1.7.9 not found

用上面创建的账号、密码登陆
$ docker login 10.64.3.7:8000
Username: foo
Password:
Login Succeeded

登陆信息写入config.json文件
$ cat ~/.docker/config.json
{
        "auths": {
                "10.64.3.7:8000": {
                        "auth": "Zm9vOmZvbzEyMw=="
                }
        }
}

再次push成功
$ docker push  10.64.3.7:8000/zhangjun3/pause2
The push refers to a repository [10.64.3.7:8000/zhangjun3/pause2]
5f70bf18a086: Layer already exists
e16a89738269: Layer already exists
latest: digest: sha256:2cf6d44ba259ddf16ebcc8a9dff94449c6082c17f900fa8eda7533525932512a size: 938

再次pull成功
$ docker pull 10.64.3.7:8000/zhangjun3/nginx:1.7.9
1.7.9: Pulling from zhangjun3/nginx
Digest: sha256:ba6612116006c334fcad0a1beb33e688129f52de9d4bc730bb3a0011fd3ab625
Status: Downloaded newer image for 10.64.3.7:8000/zhangjun3/nginx:1.7.9

向registry挂载的htpasswd中写入额外的认证信息，也可以用来登陆
$ docker run --rm --entrypoint htpasswd registry:2 -Bbn foo2 foo2 >>auth/htpasswd
$ docker login 10.64.3.7:8000
Username (foo): foo2
Password:
Login Succeeded

$ curl  --cacert /etc/docker/certs.d/10.64.3.7\:8000/ca.crt https://10.64.3.7:8000/v2/_catalog
{"errors":[{"code":"UNAUTHORIZED","message":"authentication required","detail":[{"Type":"registry","Class":"","Name":"catalog","Action":"*"}]}]}

$ curl -u foo:foo123  --cacert /etc/docker/certs.d/10.64.3.7\:8000/ca.crt https://10.64.3.7:8000/v2/_catalog
{"repositories":["library/redis","zhangjun3/busybox","zhangjun3/pause","zhangjun3/pause2"]}