# 安装 ceph rgw

docker registry 目前只能通过swift接口访问ceph存储driver，s3协议暂时不支持；

http://docs.ceph.com/docs/hammer/radosgw/admin/

[zhangjun3@tjwq01-sys-power003008 ceph-cluster]$ ceph-deploy rgw create tjwq01-sys-power003009.tjwq01 # rgw默认监听7480端口
[ceph_deploy.conf][DEBUG ] found configuration file at: /home/zhangjun3/.cephdeploy.conf
[ceph_deploy.cli][INFO  ] Invoked (1.5.37): /usr/bin/ceph-deploy rgw create tjwq01-sys-power003009.tjwq01
[ceph_deploy.cli][INFO  ] ceph-deploy options:
[ceph_deploy.cli][INFO  ]  username                      : None
[ceph_deploy.cli][INFO  ]  verbose                       : False
[ceph_deploy.cli][INFO  ]  rgw                           : [('tjwq01-sys-power003009.tjwq01', 'rgw.tjwq01-sys-power003009.tjwq01')]
[ceph_deploy.cli][INFO  ]  overwrite_conf                : False
[ceph_deploy.cli][INFO  ]  subcommand                    : create
[ceph_deploy.cli][INFO  ]  quiet                         : False
[ceph_deploy.cli][INFO  ]  cd_conf                       : <ceph_deploy.conf.cephdeploy.Conf instance at 0x1b14f80>
[ceph_deploy.cli][INFO  ]  cluster                       : ceph
[ceph_deploy.cli][INFO  ]  func                          : <function rgw at 0x1a598c0>
[ceph_deploy.cli][INFO  ]  ceph_conf                     : None
[ceph_deploy.cli][INFO  ]  default_release               : False
[ceph_deploy.rgw][DEBUG ] Deploying rgw, cluster ceph hosts tjwq01-sys-power003009.tjwq01:rgw.tjwq01-sys-power003009.tjwq01
The authenticity of host 'tjwq01-sys-power003009.tjwq01 (10.64.3.9)' can't be established.
ECDSA key fingerprint is aa:fd:72:c6:24:5e:73:f1:1b:a8:e8:ec:c7:61:b0:bc.
Are you sure you want to continue connecting (yes/no)? yes
Warning: Permanently added 'tjwq01-sys-power003009.tjwq01' (ECDSA) to the list of known hosts.
[tjwq01-sys-power003009.tjwq01][DEBUG ] connection detected need for sudo
[tjwq01-sys-power003009.tjwq01][DEBUG ] connected to host: tjwq01-sys-power003009.tjwq01
[tjwq01-sys-power003009.tjwq01][DEBUG ] detect platform information from remote host
[tjwq01-sys-power003009.tjwq01][DEBUG ] detect machine type
[ceph_deploy.rgw][INFO  ] Distro info: CentOS Linux 7.2.1511 Core
[ceph_deploy.rgw][DEBUG ] remote host will use systemd
[ceph_deploy.rgw][DEBUG ] deploying rgw bootstrap to tjwq01-sys-power003009.tjwq01
[tjwq01-sys-power003009.tjwq01][DEBUG ] write cluster configuration to /etc/ceph/{cluster}.conf
[tjwq01-sys-power003009.tjwq01][DEBUG ] create path recursively if it doesn't exist
[tjwq01-sys-power003009.tjwq01][INFO  ] Running command: sudo ceph --cluster ceph --name client.bootstrap-rgw --keyring /var/lib/ceph/bootstrap-rgw/ceph.keyring auth get-or-create client.rgw.tjwq01-sys-power003009.tjwq01 osd allow rwx mon allow rw -o /var/lib/ceph/radosgw/ceph-rgw.tjwq01-sys-power003009.tjwq01/keyring
[tjwq01-sys-power003009.tjwq01][INFO  ] Running command: sudo systemctl enable ceph-radosgw@rgw.tjwq01-sys-power003009.tjwq01
[tjwq01-sys-power003009.tjwq01][WARNIN] Created symlink from /etc/systemd/system/ceph-radosgw.target.wants/ceph-radosgw@rgw.tjwq01-sys-power003009.tjwq01.service to /usr/lib/systemd/system/ceph-radosgw@.service.
[tjwq01-sys-power003009.tjwq01][INFO  ] Running command: sudo systemctl start ceph-radosgw@rgw.tjwq01-sys-power003009.tjwq01
[tjwq01-sys-power003009.tjwq01][INFO  ] Running command: sudo systemctl enable ceph.target
[ceph_deploy.rgw][INFO  ] The Ceph Object Gateway (RGW) is now running on host tjwq01-sys-power003009.tjwq01 and default port 7480


# 创建账号

http://ivanjobs.github.io/2016/05/07/docker-registry-study/
http://bingdian.blog.51cto.com/94171/1893658

[zhangjun3@tjwq01-sys-power003009 ~]$ radosgw-admin user create --uid=demo --display-name="ceph sgw demo user"
{
    "user_id": "demo",
    "display_name": "ceph sgw demo user",
    "email": "",
    "suspended": 0,
    "max_buckets": 1000,
    "auid": 0,
    "subusers": [],
    "keys": [
        {
            "user": "demo",
            "access_key": "5Y1B1SIJ2YHKEHO5U36B",
            "secret_key": "nrIvtPqUj7pUlccLYPuR3ntVzIa50DToIpe7xFjT"
        }
    ],
    "swift_keys": [],
    "caps": [],
    "op_mask": "read, write, delete",
    "default_placement": "",
    "placement_tags": [],
    "bucket_quota": {
        "enabled": false,
        "max_size_kb": -1,
        "max_objects": -1
    },
    "user_quota": {
        "enabled": false,
        "max_size_kb": -1,
        "max_objects": -1
    },
    "temp_url_keys": []
}

# 创建swift子账号

[zhangjun3@tjwq01-sys-power003009 ~]$ radosgw-admin subuser create --uid demo --subuser=demo:swift --access=full --secret=secretkey --key-type=swift
2017-03-20 14:11:42.703810 7fcdd1db79c0  0 WARNING: detected a version of libcurl which contains a bug in curl_multi_wait(). enabling a workaround that may de
grade performance slightly.
{
    "user_id": "demo",
    "display_name": "ceph sgw demo user",
    "email": "",
    "suspended": 0,
    "max_buckets": 1000,
    "auid": 0,
    "subusers": [
        {
            "id": "demo:swift",
            "permissions": "full-control"
        }
    ],
    "keys": [
        {
            "user": "demo",
            "access_key": "5Y1B1SIJ2YHKEHO5U36B",
            "secret_key": "nrIvtPqUj7pUlccLYPuR3ntVzIa50DToIpe7xFjT"
        }
    ],
    "swift_keys": [
        {
            "user": "demo:swift",
            "secret_key": "secretkey"
        }
    ],
    "caps": [],
    "op_mask": "read, write, delete",
    "default_placement": "",
    "placement_tags": [],
    "bucket_quota": {
        "enabled": false,
        "max_size_kb": -1,
        "max_objects": -1
    },
    "user_quota": {
        "enabled": false,
        "max_size_kb": -1,
        "max_objects": -1
    },
    "temp_url_keys": []
}

# 生成swift子账号的key

[zhangjun3@tjwq01-sys-power003009 ~]$ radosgw-admin key create --subuser=demo:swift --key-type=swift --gen-secret
{
    "user_id": "demo",
    "display_name": "ceph sgw demo user",
    "email": "",
    "suspended": 0,
    "max_buckets": 1000,
    "auid": 0,
    "subusers": [
        {
            "id": "demo:swift",
            "permissions": "full-control"
        }
    ],
    "keys": [
        {
            "user": "demo",
            "access_key": "5Y1B1SIJ2YHKEHO5U36B",
            "secret_key": "nrIvtPqUj7pUlccLYPuR3ntVzIa50DToIpe7xFjT"
        }
    ],
    "swift_keys": [
        {
            "user": "demo:swift",
            "secret_key": "aCgVTx3Gfz1dBiFS4NfjIRmvT0sgpHDP6aa0Yfrh"
        }
    ],
    "caps": [],
    "op_mask": "read, write, delete",
    "default_placement": "",
    "placement_tags": [],
    "bucket_quota": {
        "enabled": false,
        "max_size_kb": -1,
        "max_objects": -1
    },
    "user_quota": {
        "enabled": false,
        "max_size_kb": -1,
        "max_objects": -1
    },
        "temp_url_keys": []
}

# 查看对象 list

[zhangjun3@tjwq01-sys-power003009 ~]$ swift -V 1.0 -A http://localhost:7480/auth -U demo:swift -K aCgVTx3Gfz1dBiFS4NfjIRmvT0sgpHDP6aa0Yfrh list
[zhangjun3@tjwq01-sys-power003009 ~]$

# 创建 docker registry

参考 docker 目录下的 [registry.md](../../docker/registry.md)

# 给本地的一个image添加私有docker registry的tag

[root@tjwq01-sys-bs003007 ~]# docker tag docker.io/kubernetes/pause localhost:8000/zhangjun3/pause
[root@tjwq01-sys-bs003007 ~]# docker images
REPOSITORY                                            TAG                 IMAGE ID            CREATED             SIZE
docker.io/registry                                    latest              047218491f8c        2 weeks ago         33.17 MB
docker.io/kubernetes/pause                            latest              f9d5de079539        2 years ago         239.8 kB
localhost:8000/zhangjun3/pause                        latest              f9d5de079539        2 years ago         239.8 kB

# 将image push到是有registry

[root@tjwq01-sys-bs003007 ~]# docker push localhost:8000/zhangjun3/pause
The push refers to a repository [localhost:8000/zhangjun3/pause]
5f70bf18a086: Pushed
e16a89738269: Pushed
latest: digest: sha256:9a6b437e896acad3f5a2a8084625fdd4177b2e7124ee943af642259f2f283359 size: 916

# 查看 ceph rgw中的数据

[zhangjun3@tjwq01-sys-power003009 ~]$ swift -V 1.0 -A http://localhost:7480/auth -U demo:swift -K aCgVTx3Gfz1dBiFS4NfjIRmvT0sgpHDP6aa0Yfrh list
registry

[zhangjun3@tjwq01-sys-power003009 ~]$ radosgw-admin bucket stats
[
    {
        "bucket": "registry",
        "pool": "default.rgw.buckets.data",
        "index_pool": "default.rgw.buckets.index",
        "id": "9c2d5a9d-19e6-4003-90b5-b1cbf15e890d.4310.1",
        "marker": "9c2d5a9d-19e6-4003-90b5-b1cbf15e890d.4310.1",
        "owner": "demo",
        "ver": "0#99",
        "master_ver": "0#0",
        "mtime": "2017-03-20 14:45:44.221337",
        "max_marker": "0#",
        "usage": {
            "rgw.main": {
                "size_kb": 73,
                "size_kb_actual": 108,
                "num_objects": 13
            }
        },
        "bucket_quota": {
            "enabled": false,
            "max_size_kb": -1,
            "max_objects": -1
        }
    }
]

# 查看 ceph 上是否已经有 pause容器
[zhangjun3@tjwq01-sys-power003009 ~]$ rados lspools
rbd
.rgw.root
default.rgw.control
default.rgw.data.root
default.rgw.gc
default.rgw.log
default.rgw.users.uid
default.rgw.users.keys
default.rgw.users.swift
default.rgw.buckets.index
default.rgw.buckets.data

[zhangjun3@tjwq01-sys-power003009 ~]$ rados --pool default.rgw.buckets.data ls|grep pause
9c2d5a9d-19e6-4003-90b5-b1cbf15e890d.4310.1_files/docker/registry/v2/repositories/zhangjun3/pause/_layers/sha256/f9d5de0795395db6c50cb1ac82ebed1bd8eb3eefcebb1aa724e01239594e937b/link
9c2d5a9d-19e6-4003-90b5-b1cbf15e890d.4310.1_files/docker/registry/v2/repositories/zhangjun3/pause/_layers/sha256/f72a00a23f01987b42cb26f259582bb33502bdb0fcf5011e03c60577c4284845/link
9c2d5a9d-19e6-4003-90b5-b1cbf15e890d.4310.1_files/docker/registry/v2/repositories/zhangjun3/pause/_layers/sha256/a3ed95caeb02ffe68cdd9fd84406680ae93d633cb16422d00e8a7c22955b46d4/link
9c2d5a9d-19e6-4003-90b5-b1cbf15e890d.4310.1_files/docker/registry/v2/repositories/zhangjun3/pause/_manifests/tags/latest/current/link
9c2d5a9d-19e6-4003-90b5-b1cbf15e890d.4310.1_files/docker/registry/v2/repositories/zhangjun3/pause/_manifests/tags/latest/index/sha256/9a6b437e896acad3f5a2a8084625fdd4177b2e7124ee943af642259f2f283359/link
9c2d5a9d-19e6-4003-90b5-b1cbf15e890d.4310.1_files/docker/registry/v2/repositories/zhangjun3/pause/_manifests/revisions/sha256/9a6b437e896acad3f5a2a8084625fdd4177b2e7124ee943af642259f2f283359/link
