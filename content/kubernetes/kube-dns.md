# kube-dns

kube-dns schedules DNS Pods and Service on the cluster, other pods in cluster can use the DNS Service’s IP to resolve DNS names.

$ kubectl get services kube-dns --namespace=kube-system
NAME       CLUSTER-IP   EXTERNAL-IP   PORT(S)         AGE
kube-dns   10.0.0.10    <none>        53/UDP,53/TCP   8m

# What things get DNS names?

Every Service defined in the cluster (including the DNS server itself) is assigned a DNS name. By default, a client Pod’s DNS search list will include the Pod’s own namespace 
and the cluster’s default domain. This is best illustrated by example:
Assume a Service named foo in the Kubernetes namespace bar. A Pod running in namespace bar can look up this service by simply doing a DNS query for foo. A Pod running in namespace quux can look up this service by doing a DNS query for foo.bar

# Supported DNS schema

The following sections detail the supported record types and layout that is supported. Any other layout or names or queries that happen to work are considered implementation details and are subject to change without warning.

##Services

1. A records

“Normal” (not headless) Services are assigned a DNS A record for a name of the form my-svc.my-namespace.svc.cluster.local. This resolves to the cluster IP of the Service.

“Headless” (without a cluster IP) Services are also assigned a DNS A record for a name of the form my-svc.my-namespace.svc.cluster.local. Unlike normal Services, this resolves to the set of IPs of the pods selected by the Service. 
Clients are expected to consume the set or else use standard round-robin selection from the set.

2. SRV records

SRV Records are created for named ports that are part of normal or Headless Services. For each named port, the SRV record would have the form _my-port-name._my-port-protocol.my-svc.my-namespace.svc.cluster.local. For a regular service, this resolves to the port number and the CNAME: my-svc.my-namespace.svc.cluster.local. For a headless service, this resolves to multiple answers, one for each pod that is backing the service, and contains the port number and a CNAME of the pod of the form auto-generated-name.my-svc.my-namespace.svc.cluster.local.

3. Backwards compatibility

Previous versions of kube-dns made names of the form my-svc.my-namespace.cluster.local (the ‘svc’ level was added later). This is no longer supported.

## Pods

1. A Records

When enabled, pods are assigned a DNS A record in the form of pod-ip-address.my-namespace.pod.cluster.local.

For example, a pod with ip 1.2.3.4 in the namespace default with a dns name of cluster.local would have an entry: 1-2-3-4.default.pod.cluster.local.

2. A Records and hostname based on Pod’s hostname and subdomain fields

Currently when a pod is created, its hostname is the Pod’s metadata.name value.

With v1.2, users can specify a Pod annotation, pod.beta.kubernetes.io/hostname, to specify what the Pod’s hostname should be. The Pod annotation, if specified, takes precendence over the Pod’s name, to be the hostname of the pod. 
For example, given a Pod with annotation pod.beta.kubernetes.io/hostname: my-pod-name, the Pod will have its hostname set to “my-pod-name”.

With v1.3, the PodSpec has a hostname field, which can be used to specify the Pod’s hostname. This field value takes precedence over the pod.beta.kubernetes.io/hostname annotation value.

v1.2 introduces a beta feature where the user can specify a Pod annotation, pod.beta.kubernetes.io/subdomain, to specify the Pod’s subdomain. The final domain will be “ ...svc.". 
For example, a Pod with the hostname annotation set to "foo", and the subdomain annotation set to "bar", in namespace "my-namespace", will have the FQDN "foo.bar.my-namespace.svc.cluster.local"

With v1.3, the PodSpec has a subdomain field, which can be used to specify the Pod’s subdomain. This field value takes precedence over the pod.beta.kubernetes.io/subdomain annotation value.

Example:
apiVersion: v1
kind: Pod
metadata:
  name: busybox
  namespace: default
spec:
  hostname: busybox-1
  subdomain: default
  containers:
  - image: busybox
    command:
      - sleep
      - "3600"
    name: busybox

If there exists a headless service in the same namespace as the pod and with the same name as the subdomain, the cluster’s KubeDNS Server also returns an A record for the Pod’s fully qualified hostname. Given a Pod with the hostname set to “foo” and the subdomain set to “bar”, and a headless Service named “bar” in the same namespace, the pod will see it’s own FQDN as “foo.bar.my-namespace.svc.cluster.local”. DNS serves an A record at that name, pointing to the Pod’s IP.

With v1.2, the Endpoints object also has a new annotation endpoints.beta.kubernetes.io/hostnames-map. Its value is the json representation of map[string(IP)][endpoints.HostRecord], for example: ‘{“10.245.1.6”:{HostName: “my-webserver”}}’. If the Endpoints are for a headless service, an A record is created with the format ...svc. For the example json, if endpoints are for a headless service named "bar", and one of the endpoints has IP "10.245.1.6", an A record is created with the name "my-webserver.bar.my-namespace.svc.cluster.local" and the A record lookup would return "10.245.1.6". This endpoints annotation generally does not need to be specified by end-users, but can used by the internal service controller to deliver the aforementioned feature.

With v1.3, The Endpoints object can specify the hostname for any endpoint, along with its IP. The hostname field takes precedence over the hostname value that might have been specified via the endpoints.beta.kubernetes.io/hostnames-map annotation.
With v1.3, the following annotations are deprecated: pod.beta.kubernetes.io/hostname, pod.beta.kubernetes.io/subdomain, endpoints.beta.kubernetes.io/hostnames-map

# How it Works

The running Kubernetes DNS pod holds 3 containers - kubedns, dnsmasq and a health check called healthz. 

1. The kubedns process watches the Kubernetes master for changes in Services and Endpoints, and maintains in-memory lookup structures to service DNS requests. 
2. The dnsmasq container adds DNS caching to improve performance. 
3. The healthz container provides a single health check endpoint while performing dual healthchecks (for dnsmasq and kubedns).

The DNS pod is exposed as a Kubernetes Service with a static IP. Once assigned the kubelet passes DNS configured using the --cluster-dns=10.0.0.10 flag to each container.
DNS names also need domains. The local domain is configurable, in the kubelet using the flag --cluster-domain=<default local domain>

The Kubernetes cluster DNS server (based off the SkyDNS library) supports forward lookups (A records), service lookups (SRV records) and reverse IP address lookups (PTR records).

Dnsmasq 提供 DNS 缓存和 DHCP 服务功能。作为域名解析服务器(DNS)，dnsmasq可以通过缓存 DNS 请求来提高对访问过的网址的连接速度。作为DHCP 服务器，dnsmasq 可以用于为局域网电脑分配内网ip地址和提供路由。
DNS和DHCP两个功能可以同时或分别单独实现。dnsmasq轻量且易配置，适用于个人用户或少于50台主机的网络。此外它还自带了一个 PXE 服务器。


# Inheriting DNS from the node

When running a pod, kubelet will prepend the cluster DNS server and search paths to the node’s own DNS settings. If the node is able to resolve DNS names specific to the larger environment, pods should be able to, also. See “Known issues” below for a caveat.
If you don’t want this, or if you want a different DNS config for pods, you can use the kubelet’s --resolv-conf flag. Setting it to “” means that pods will not inherit DNS. Setting it to a valid file path means that kubelet will use this file instead of /etc/resolv.conf for DNS inheritance.
Known issues

Kubernetes installs do not configure the nodes’ resolv.conf files to use the cluster DNS by default, because that process is inherently distro-specific. This should probably be implemented eventually.
Linux’s libc is impossibly stuck (see this bug from 2005) with limits of just 3 DNS nameserver records and 6 DNS search records. Kubernetes needs to consume 1 nameserver record and 3 search records. This means that if a local installation already uses 3 nameservers or uses more than 3 searches, some of those settings will be lost. As a partial workaround, the node can run dnsmasq which will provide more nameserver entries, but not more search entries. You can also use kubelet’s --resolv-conf flag.
If you are using Alpine version 3.3 or earlier as your base image, dns may not work properly owing to a known issue with Alpine. Check here for more information.

# 使用和部署

1. kubelet 配置中增加启动项：
$ vi /etc/kubernetes/kubelet
KUBELET_ARGS="--cluster_dns=10.254.0.10 --cluster_domain=kube.local
重启kubelet；

2. 创建DNS Deployment和Service，可以参考：https://github.com/kubernetes/kubernetes/tree/master/cluster/addons/dns

3. 创建pod和服务，看是否能解析名称；