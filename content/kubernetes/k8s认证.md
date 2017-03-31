kube-apiserver启动后，会同时监听insecure port(–insecure-bind-address=0.0.0.0 –insecure-port=8080)和 secure port(–bind-address=0.0.0.0 –secure-port=6443)。

对于insecret port的流量没有任何安全机制限制，一般是为了集群bootstrap或集群开发调试使用的，但是走sercret port的流量将会遇到验证、授权等安全机制的限制。官方文档建议：集群外部流量都应该走secure port。

通过将apiserver的--insecure-port设置为0，可以关闭非安全端口；

进入secret port的流量，会经过如下关卡：
安全通道(tls) -> Authentication(认证) -> Authorization（授权）-> Admission Control(入口条件控制)

1. 安全通道：端口6443(通过 /opt/bin/kube-apiserver --help查看options说明可以得到)，公钥证书server.cert ，私钥文件：server.key，即基于tls的https的安全通道建立，对流量进行加密，防止嗅探、身份冒充和篡改；

2. Authentication：即身份验证，这个环节它面对的输入是整个http request。它负责对来自client的请求进行身份校验，支持的方法包括：client证书验证（https双向验证）、basic auth、普通token以及jwt token(用于serviceaccount)。APIServer启动时，可以指定一种Authentication方法，也可以指定多种方法。如果指定了多种方法，那么APIServer将会逐个使用这些方法对客户端请求进行验证，只要请求数据通过其中一种方法的验证，APIServer就会认为Authentication成功；

3. Authorization：授权。这个阶段面对的输入是http request context中的各种属性，包括：user、group、request path（比如：/api/v1、/healthz、/version等）、request verb(比如：get、list、create等)。APIServer会将这些属性值与事先配置好的访问策略(access policy）相比较。APIServer支持多种authorization mode，包括AlwaysAllow、AlwaysDeny、ABAC、RBAC和Webhook。APIServer启动时，可以指定一种authorization mode，也可以指定多种authorization mode，如果是后者，只要Request通过了其中一种mode的授权，那么该环节的最终结果就是授权成功；

4. Admission Control：从技术的角度看，Admission control就像a chain of interceptors（拦截器链模式），它拦截那些已经顺利通过authentication和authorization的http请求。
http请求沿着APIServer启动时配置的admission control chain顺序逐一被拦截和处理，如果某个interceptor拒绝了该http请求，那么request将会被直接reject掉，而不是像authentication

# Users

https://kubernetes.io/docs/admin/authentication/

k8s有两种用户账户：useraccount和普通users。

普通users来源：
1. SSL 证书的Subject、SANs中提取；
2. 外部系统账户，如：Keystone or Google Accounts；
3. 文件中配置，如账号、密码，Token；

k8s不提供创建普通user的API，也没有代表普通user的对象；但是k8s提供了创建和管理useraccount的API，它们属于特定的namespace，被apiserver自动创建和管理；

用户请求可以关联useraccount、普通user或者匿名请求(anonymous requests)，apiserver对它们进行认证或当做匿名用户处理；

# JWT

https://jwt.io/introduction/
http://blog.leapoahead.com/2015/09/06/understanding-jwt/
JWT 是 JSON Web Token的简称，遵守RFC 7519，用来代表两个parties安全共享claims信息，JWT的组成：
1. Header 
{
  "alg": "HS256", // 共享密码的签名算法，RSA256是基于X.509的公私钥证书签名算法；
  "typ": "JWT"
}
2. Payload
{
  "sub": "1234567890",
  "name": "John Doe",
  "admin": true
}
3. 验证签名：
HMACSHA256(
  base64UrlEncode(header) + "." +
  base64UrlEncode(payload),
  hmac-secret-key
)
将上面各三部分信息分别base64编码，然后用 . 连接，形成一个Token：
eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWV9.TJVA95OrM7E2cBab30RMHrHDcEfxjoYZgeFONFh7HgQ

生成的JWT一般使用Bearer Scheam保存在HTTP(s)请求的Authorization字段中：
Authorization: Bearer <token>
这是一种无状态的认证机制，JWT是自包含所有信息，不需要来回在数据库中查询；server收到JWT后，base64解码，获取Payload、签名算法和签名，然后用
签名算法对Payload进行计算，看是否和签名一致。

签名算法可以是基于共享密码的HMAC算法，也可以是基于公私钥的X.509证书签名算法；


# 认证策略(Authentication strategies)

kube-apiserver提供如下认证策略，对来自client的请求进行认证，只要有一种方式认证成功，就响应请求：
1. HTTP Basic；
2. X.509证书；
3. Token认证；
4. ServiceAccount认证；
5. OpenID Connect Tokens;
6. Webhook Token Authentication
7. Authenticating Proxy
8. Keystone Password
9. Anonymous requests

客户端使用HTTP Base或Token与apiserver认证时需要走apiserver的安全端口(https)，但是是单向加密的：服务端->客户端加密，但是客户端->服务端不加密；

kubectl命令行工具比较特殊，它同时支持CA双向认证和简单认证两种模式和apiserver通信，其它客户端组件只能使用CA双向认证或通过非安全端口的非
安全模式和apiserver通信；

apiserver的认证系统会试着将下列属性和请求关联，这些属性稍后被**授权系统**使用：
1. Username：用户名字符串；
2. UID： 代表用户的字符串；
3. Groups：用户所属组列表；
4. 其它fields：a map of strings to list of strings which holds additional information authorizers may find useful.

认证通过的user将被加入 system:authenticated组中；

# HTPPT Basic认证

kube-apiserver启动时如果--basic-auth-file参数不为空则开启HTTP Basci认证，该参数指定csv格式的basicauthfile文件位置该参数在运行过程中不能修改，文件的格式如下，组是可选的：
  password,user,uid,"group1,group2,group3"

当使用HTTP Client时，apiserver期望在Authorization头部中包含  Basic BASE64ENCODED(USER:PASSWORD) 内容（不带group信息），apiserver提取出Basic后面的值后，查找--basic-auth-file，从而确定user、uid和groups信息；

## 示例：使用Basic authentication + ABAC model设置API server

1. 首先配置Basic authentication文件
  $ cat basic_auth.csv：
  admin_passwd,admin,admin
  test_passwd,test,test

2. 然后配置ABAC访问策略, 设置admin具有任何权限，test用户只能访问pods：
  $ cat policy_file.jsonl:
  {"user":"admin"}
  {"user":"test", "resource": "pods", "readonly": true}

3. 然后启动API Server：
  $ kube-apiserver --basic-auth-file=basic_auth.csv --authorization-mode=ABAC --authorization-policy-file=policy_file.jsonl ...

4. 访问API，可以看到test用户无法访问Pod之外的资源：
  $ curl --basic -u admin:admin_passwd https://192.168.3.146:6443/api/v1/pods -k


# Client HTTPS证书认证

kube-apiserver启动时如果--client-ca-file参数不为空则开启HTTPS客户端证书认证，只要是ca签名过的证书都可以通过验证。
kube-apiserver通过--tls-cert-file以及--tls-private-key-file参数指定自己的证书和私钥，如果没有指定这两个参数，kube-apiserver自动生成自签名的证书和私钥(不生成ca)，并放在CertDirectory: "/var/run/kubernetes"目录下；

可以使用make-ca-cert.sh脚本快速创建ca和签名的公私钥：
https://raw.githubusercontent.com/GoogleCloudPlatform/kubernetes/v0.21.1/cluster/saltbase/salt/generate-cert/make-ca-cert.sh

注意：上面脚本创建了ca.crt但是没有ca.key，所以没法对其它crt进行签名。也可以手动重新创建ca，并将apiserver使用的.key、.crt以及各个components的client.key和client.crt都生成一份，并用你生成的Ca签发。

apiserver验证client的证书后，提取Common Name(CN)作为请求的用户名，提取Organization作为用户所属的组(证书中可以有多个Organization)。

使用openssl创建csr时，可以指定CN和Organizaion的Subject：
  openssl req -new -key jbeda.pem -out jbeda-csr.pem -subj "/CN=jbeda/O=app1/O=app2"
创建的证书，包含用户名jbeda，属于两个组：app1、app2；

https://kubernetes.io/docs/admin/authentication/

[root@tjwq01-sys-bs003007 ssl]# openssl x509 -noout -text -in server.cert |grep Subject
        Subject: CN=kubernetes-master
[root@tjwq01-sys-bs003007 ssl]# openssl x509 -noout -text -in kubecfg.crt |grep Subject
        Subject: O=system:masters, CN=kubecfg

# Token认证

kube-apiserver启动时如果--token-auth-file参数不为空则在secret port上启动token认证；
token-auth-file是一个csv文件，运行过程中不能动态修改；

token文件的格式如下，组信息是可选的：
  token,user,uid,"group1,group2,group3"

客户请求secret port时，会从Authorization中提取出携带token值(标记token的关键字Bearer)，然后校验是否正确, 如：
  Authorization: Bearer 31ada4fd-adec-460c-809a-9e56ceb75269

apiserver提取出Bearer后面的值后，查找--tokenauth-file，从而确定user、uid和groups信息；


# ServiceAccount认证

ServiceAccount实际是token认证的变形，它使用X.509证书对JWT Payload进行签名和验证；

apiserver默认启动ServiceAccount，它使用两个可选参数：
1. --service_account_key_file： 签名bearer token的证书，如果未指定则使用apiserver的私钥，controller-manager需要指定同一个私钥；
2. --service-account-lookup： 如果为true，则apiserver未查到token时，则从etcd获取；

serviceaccount主要用于pod中的容器访问apiserver服务：
1. kube-controller-manager中的Admission Controller 会自动给每个namespace创建default serviceaccount, 并生成和关联一个API Token Secret，该Secret中包含namespace、token、ca(指定了--root-ca-file参数后才会生成ca文件)三个文件; 其中token是jwt字符串，未编码的类似是 struct Token。
2. kube-controller-manager 自动将serviceaccount文件关联的API Token Secret挂载到pod容器中；
3. pod使用挂载的ca验证apiserver的证书，然后将token内容添加到HTTP的Authorization头部中，格式如：Authorization: bearer token，访问apiserver；
4. apiserver从--service_account_key_file指定的私钥中提取出公钥；
5. apiserver从客户端请求的HTTP头部Authorization字段提取出bearer token，然后使用上面的公钥对token(jwt字符串)进行解密，获取paredToken，格式如下：
  type Token struct { 
    Raw string // The raw token. Populated when you Parse a token 
    Method    SigningMethod // The signing method used or to be used 
    Header     map[string]interface{} // The first segment of the token       
    Claims       map[string]interface{} // The second segment of the token     
    Signature string // The third segment of the token. Populated when you Parse a token 
    Valid bool // Is the token valid? Populated when you Parse/Verify a token 
  }
  Claims字段包含namespace、secretName、serviceAccountName、token等内容，具体如下：
  https://github.com/kubernetes/kubernetes/blob/e73e25422f4fee69782a5b1fba531de27a8ee86e/pkg/serviceaccount/jwt.go#L166

    // Identify the issuer
    claims[IssuerClaim] = Issuer
    // Username
    claims[SubjectClaim] = apiserverserviceaccount.MakeUsername(serviceAccount.Namespace, serviceAccount.Name)
    // Persist enough structured info for the authenticator to be able to look up the service account and secret
    claims[NamespaceClaim] = serviceAccount.Namespace
    claims[ServiceAccountNameClaim] = serviceAccount.Name
    claims[ServiceAccountUIDClaim] = serviceAccount.UID
    claims[SecretNameClaim] = secret.Name

6. apiserver对token中的Claims字段信息进行校验(Claims中的各数据是否存在，namespace和SubjectClaim是否正确)，看是否符合要求；如果指定了--service-account-lookup参数，则会根据Claims中的namespace ，secretName ，serviceAccountName从etcd中取出已设置好的serviceaccount以及secret来进行身份验证，验证通过之后会返回user信息。

ServiceAccount校验通过后，将请求关联到用户名system:serviceaccount:(NAMESPACE):(SERVICEACCOUNT)，组system:serviceaccounts和system:serviceaccounts:(NAMESPACE)

# OpenID Connect

OpenID是OAuth2的变形，Identity Provider在返回access_token的同时，返回id_token; id_token是JWT，包含well known的fields，如username、email、groups等；

1. Login to your identity provider
2. Your identity provider will provide you with an access_token, id_token and a refresh_token
3. When using kubectl, use your id_token with the --token flag or add it directly to your kubeconfig
4. kubectl sends your id_token in a header called Authorization to the API server
5. The API server will make sure the JWT signature is valid by checking against the certificate named in the configuration;
  apiserver根据配置的--oidc-issuer-url参数从identity provider获取签名证书(一般是静态的，可以一直缓冲到本地)，然后对id_token的签名进行验证；
6. Check to make sure the id_token hasn’t expired （iat+ exp）
7. Make sure the user is authorized
8. Once authorized the API server returns a response to kubectl
9. kubectl provides feedback to the user

所有的用户信息都保存在 id_token中，apiserver不会像identity provider获取user的其它信息(如电话号码)，jwt是无状态的，增强了可扩展性；

apiserver对identity provider的要求是：
1. 支持 OpenID connect discovery; 
2. 必须支持TLS通信；
3. 必须使用CA签名的证书(自签名证书也可以），且CA证书的CA flag值必须为TRUE；

  [root@tjwq01-sys-bs003007 ssl]# openssl x509 -noout -text -in ca.crt |grep CA
                  CA:TRUE
  [root@tjwq01-sys-bs003007 ssl]# openssl x509 -noout -text -in kubecfg.crt |grep CA
                  CA:FALSE

google的OepenID Connect服务：https://developers.google.com/identity/protocols/OpenIDConnect#discovery

## 配置apiserver使用OpenID Connect
1. --oidc-issuer-url： 必选，必须是https://，一般是identity provider的discovery URL(不包含.well-known/openid-configuration), apiserver从这个URL获取公共签名key(jws_uri字段指定的URL)；
2. --oidc-client-id: 必选，JWT关联的client id;
3. --oidc-username-claim：可选，默认值为sub；
4. --oidc-groups-claim：可选，JWT Payload中指定user所属groups的名称(默认为groups)，类型是字符串；
5. --oidc-ca-file：可选(自签名证书的情况下是可选的，否则必须指定)；签名identity provider使用的证书的CA证书，证书的CA flag值必须为TRUE；

## 配置kubectl使用OpenID Connect
方法1： 使用 oidc authenticator。一旦id_token过期，kubectl自动自动向identity provider刷新token，然后将新的id_token保存到 ~/kube/.confg文件；

kubectl config set-credentials USER_NAME \
   --auth-provider=oidc
   --auth-provider-arg=idp-issuer-url=( issuer url ) \
   --auth-provider-arg=client-id=( your client id ) \
   --auth-provider-arg=client-secret=( your client secret ) \
   --auth-provider-arg=refresh-token=( your refresh token ) \
   --auth-provider-arg=idp-certificate-authority=( path to your ca certificate ) \
   --auth-provider-arg=id-token=( your id_token )
对应的配置文件：
users:
- name: mmosley
  user:
    auth-provider:
      config:
        client-id: kubernetes
        client-secret: 1db158f6-177d-4d9c-8a8b-d36869918ec5
        id-token: ...
        idp-certificate-authority: /root/ca.pem
        idp-issuer-url: https://oidcidp.tremolo.lan:8443/auth/idp/OidcIdP
        refresh-token: q1bKLFOyUiosTfawzA93TzZIDzH2TNa2SMm0zEiPKTUwME6BkEo6Sql5yUWVBSWpKUGphaWpxSVAfekBOZbBhaEW+VlFUeVRGcluyVF5JT4+haZmPsluFoFu5XkpXk5BXq
      name: oidc
方法2：直接在kubectl命令行参数中将id_token设置为--token选择的值；

# Webhook Token Authentication

Webhook authentication is a hook for verifying bearer tokens.
1. --authentication-token-webhook-config-file a kubeconfig file describing how to access the remote webhook service.
2. --authentication-token-webhook-cache-ttl how long to cache authentication decisions. Defaults to two minutes.

webhook-config-file的格式类似于Kubeconfgi，users指向apiserver的webhook，clusters指向远程认证服务； apisever的wehook向远程认证服务发送认证请求；
  clusters:
    - name: name-of-remote-authn-service
      cluster:
        certificate-authority: /path/to/ca.pem         # CA for verifying the remote service.
        server: https://authn.example.com/authenticate # URL of remote service to query. Must use 'https'.
  users:
    - name: name-of-api-server
      user:
        client-certificate: /path/to/cert.pem # cert for the webhook plugin to use
        client-key: /path/to/key.pem          # key matching the cert
  current-context: webhook
  contexts:
  - context:
      cluster: name-of-remote-authn-service
      user: name-of-api-sever
    name: webhook

apiserver收到包含bearer token的请求后，向远程认证服务器发送请求，请求body如下：
{
  "apiVersion": "authentication.k8s.io/v1beta1",
  "kind": "TokenReview",
  "spec": {
    "token": "(BEARERTOKEN)"
  }
}
远程认证服务器收到请求后，如果认证成功给apiserver发送如下响应：
{
  "apiVersion": "authentication.k8s.io/v1beta1",
  "kind": "TokenReview",
  "status": {
    "authenticated": true,
    "user": {
      "username": "janedoe@example.com",
      "uid": "42",
      "groups": [
        "developers",
        "qa"
      ],
      "extra": {
        "extrafield1": [
          "extravalue1",
          "extravalue2"
        ]
      }
    }
  }
}
认证失败则发送：
{
  "apiVersion": "authentication.k8s.io/v1beta1",
  "kind": "TokenReview",
  "status": {
    "authenticated": false
  }
}

# Authenticating Proxy

这种情况下，客户端请求必须提供SSL证书，apiserver对证书进行验证，通过后从请求头中提取username，然后看是否在已知的列表中，如果在则允许，否则拒绝；如果列表为空，则默认允许；
--requestheader-username-headers：必选，指定请求头中包含username的字段名称；
--requestheader-client-ca-file：必选，签名请求客户端的CA证书；
--requestheader-allowed-names： 允许的username列表，如果为空则默认都允许；

# Keystone Password

keystone是openstack的认证系统，使用账号、密码进行认证，对应的参数：
--experimental-keystone-url=<AuthURL>
--experimental-keystone-ca-file=SOMEFILE

# Anonymous requests
apiserver默认允许匿名访问，可以使用--anonymous-auth=false来关闭该功能；

当开启时，如果请求没有被所有其它认证方法拒绝，也没有被他们匹配，则认为是匿名请求，关联的用户名 system:anonymous和组 system:unauthenticated；

如果要求必须认证后才能访问系统，可以设置授权方式为非 AlwaysAllow 或 设置--anonymous-auth=false；

# Admission Controllers

在认证和授权之外，kube-apiserver的Admission Controller机制也会对请求进行一些类验证，任何一环拒绝了请求，则会返回错误。

Admission Controller是作为Kubernetes API Serve的一部分，并以插件代码的形式存在，在API Server启动的时候，可以配置需要哪些Admission Controller，以及它们的顺序，如：
--admission_control=NamespaceLifecycle,NamespaceExists,LimitRanger,SecurityContextDeny,ServiceAccount,ResourceQuota