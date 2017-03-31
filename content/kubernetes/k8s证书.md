# SSL证书的Subject和Subject Alternative Name域

1. Subject的内容： 
  示例：CN=Sample Cert, OU=R&D, O=Company Ltd., L=Dublin 4, S=Dublin, C=IE
  CN: CommonName
  OU: OrganizationalUnit
  O: Organization
  L: Locality
  S: StateOrProvinceName
  C: CountryName

2. Subject Alternative Name(SANs)的内容：
  示例：IP Address:10.64.3.7, IP Address:10.64.3.7, IP Address:10.254.0.1, DNS:kubernetes, DNS:kubernetes.default, DNS:kubernetes.default.svc, DNS:kubernetes.default.svc.cluster.local
  + Email addresses
  + IP addresses
  + URIs
  + DNS names (This is usually also provided as the Common Name RDN within the Subject field of the main certificate.)
  + directory names (alternative Distinguished Names to that given in the Subject)
  + other names, given as a General Name: a registered[3] Object identifier followed by a value

apiserver从客户端请求证书的 CN 中提取username，O中提取group信息(可以指定多个O，表示用户属于多个group)；如果客户端证书SANs中包含如下内容：
IP Address:10.64.3.7, IP Address:10.64.3.7, IP Address:10.254.0.1, DNS:kubernetes, DNS:kubernetes.default, DNS:kubernetes.default.svc, ，
则apiserver对客户端的来源IP进行检查，如果不在上面的IP列表中，则拒绝接收请求；

同理，客户端收到apiserver的证书后，也会从Subject、SANs中提取出认证信息，然后进行检查&校验；

查看证书的内容：openssl x509 -noout -text -in ca.crt

在kublet的kubcofnig文件中，如果为user中了证书，则username应该和证书中的CN一致；

https://kubernetes.io/docs/admin/authentication/

# Creating Certificates

When using client certificate authentication, you can generate certificates using an existing deployment script or manually through easyrsa or openssl.

## Using an Existing Deployment Script

Using an existing deployment script is implemented at cluster/saltbase/salt/generate-cert/make-ca-cert.sh.
https://raw.githubusercontent.com/GoogleCloudPlatform/kubernetes/v0.21.1/cluster/saltbase/salt/generate-cert/make-ca-cert.sh

Execute this script with two parameters. The first is the IP address of API server. The second is a list of subject alternate names in the form IP:<ip-address> or DNS:<dns-name>.

The script will generate three files: ca.crt, server.crt, and server.key.

Finally, add the following parameters into API server start parameters:
--client-ca-file=/srv/kubernetes/ca.crt
--tls-cert-file=/srv/kubernetes/server.crt
--tls-private-key-file=/srv/kubernetes/server.key

## easyrsa

easyrsa can be used to manually generate certificates for your cluster.

Download, unpack, and initialize the patched version of easyrsa3.
  curl -L -O https://storage.googleapis.com/kubernetes-release/easy-rsa/easy-rsa.tar.gz
  tar xzf easy-rsa.tar.gz
  cd easy-rsa-master/easyrsa3
  ./easyrsa init-pki

Generate a CA. (--batch set automatic mode. --req-cn default CN to use.)
  ./easyrsa --batch "--req-cn=${MASTER_IP}@`date +%s`" build-ca nopass

Generate server certificate and key. (build-server-full [filename]: Generate a keypair and sign locally for a client or server)
  ./easyrsa --subject-alt-name="IP:${MASTER_IP}" build-server-full server nopass

--subject-alt-name=指定能使用该证书的IP列表，这里是apiserver使用，所以应该是apiserver的IP，注意应该同时包含node ip和cluster ip；
--subject-alt-name="IP:${MASTER_NODEIP},IP:${MASTER_CLUSTERIP}"

Copy pki/ca.crt, pki/issued/server.crt, and pki/private/server.key to your directory.

Fill in and add the following parameters into the API server start parameters:
  --client-ca-file=/yourdirectory/ca.crt
  --tls-cert-file=/yourdirectory/server.crt
  --tls-private-key-file=/yourdirectory/server.key

## openssl

openssl can also be use to manually generate certificates for your cluster.

1. 生成ca证书私钥；
  openssl genrsa -out ca.key 2048

2. 生成ca证书；
  openssl req -x509 -new -nodes -key ca.key -subj "/CN=${MASTER_IP}" -days 10000 -out ca.crt

3. 创建server私钥；
  openssl genrsa -out server.key 2048

4. 创建server证书请求的配置文件

[root@tjwq01-sys-bs003007 certs]# cat server_ssl.cnf
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

[ alt_names ]
DNS.1 = kubernetes
DNS.2 = kubernetes.default
DNS.3 = kubernetes.default.svc
DNS.4 = kubernetes.default.svc.cluster.local
DNS.5 = k8s-master
IP.1 = 10.64.3.7
IP.2 = 10.254.0.1

5. 创建server私钥的证书签名请求；

openssl req -new -key server.key -subj "/CN=${MASTER_IP}" -config server_ssl.cnf -out server.csr

apiserver从证书的CN中提取username，所以如果该证书是客户端使用，CN应该设置为客户端名称；

6. 使用ca对证书签名请求进行签名，生成server的证书；
openssl x509 -req -in server.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out server.crt -days 10000 -extensions v3_req -extfile server_ssl.cnf

7. 查看证书：
  openssl x509  -noout -text -in ./server.crt

Finally, do not forget to fill out and add the same parameters into the API server start parameters.


# 自签名证书

1. 创建私钥
openssl genrsa -out server.key 2048

2. 创建自签名证书(不需要CA签名)
https://docs.docker.com/registry/insecure/#using-self-signed-certificates

$ openssl req -newkey rsa:2048 -nodes -sha256 -keyout certs/domain.key -x509 -days 365 -out certs/domain.crt # 将提示的CN设置为为mydockerhub.com；

如果需要对上面的domain.crt进行验证，可以将ca.crt设置为domain.crt(因为domain.crt是自签名)；