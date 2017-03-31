<!-- toc -->

# 配置 dockerd

1. dockerd启动时会读取`flanneld`启动时创建的环境变量文件`/run/flannel/docker`，然后设置docker0的bip、mtu等参数；
1. 为了加快pull image的速度，可以使用`registry-mirror`，增加`max-concurrent-downloads`的值(默认为3)；

    ``` bash
    $ cat /etc/docker/daemon.json
    {
    "registry-mirrors": ["https://docker.mirrors.ustc.edu.cn", "hub-mirror.c.163.com"],
    "max-concurrent-downloads": 10
    }
    $ systemctl start docker
    ```

1. docker 1.13版本后，`ip-forward`参数默认为`ture`(自动配置`net.ipv4.ip_forward = 1`)，dockerd**将iptables FORWARD的默认策略设置为DROP**，从而导致Node ping其它Node的Pod IP失败；

  原因分析：
        1. kube-proxy只会创建ClusterIP和nodePort的DNAT规则(也可能创建SNAT规则，参考[kue-proxy和iptables.md](kue-proxy和iptables.md)), 一般不创建FORWARD规则；参考[docker创建的iptables.md](../docker/docker创建的iptables.md)
        1. dockerd一般只会为nodePort映射的的容器创建FORWARD规则；
        1. pod IP不属于第一种情况，一般也不属于第二种情况，所以会被docker创建的FORWARD默认policy drop；

  解决办法：
        1. `$ sudo iptables -P FORWARD ACCEPT`;
        1. 或者在 /etc/docker/daemon.json中设置`"ip-forward": false`;
        https://github.com/docker/docker/pull/28257
        https://github.com/kubernetes/kubernetes/issues/40182

1. dockerd默认启用`--iptabes`、`--ip-masq`参数，不能关闭:
        1. 用于**对Pod发出的请求做SNAT**；
        1. 用于创建**hostPort端口映射的iptables NAT规则**(将访问host:port的请求DNAT到container:port)，如果需要运行有**hostPort端口映射的容器**(如手动docker run -p80:80或pod的spec里面指定hostPort);则**不能关闭**这两个选项，否则访问会被Refuse。

        原因分析：

        1. docker不再生成访问hostPort的DNAT规则，目的地址hostIP:hostPort的包转发到容器后，容器会拒绝接收；

1. 如果指定的私有registry需要登录验证(HTTPS证书、basic账号密码)，则需要放置CA证书和生成认证信息：

        1. 将registry的CA证书放置到 `/etc/docker/certs.d/{registryIP:Port}}/ca.crt`；
        1. 执行 `docker login` 命令，docker自动将认证信息保存到`~/.docker/config.json`：
        ``` bash
        $ cat ~/.docker/config.json
        {
                "auths": {
                        "10.64.3.7:8000": {
                                "auth": "Zm9vMjpmb28y"
                        }
                }
        }
        ```