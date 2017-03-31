<!-- toc -->

## 修改配置/etc/kubernetes/scheduler

``` bash
$ grep -v '^#' scheduler |grep -v '^$'
KUBE_SCHEDULER_ARGS="--address=127.0.0.1"
```

## 重启进程

``` bash
$ systemctl start kube-scheduler
$ ps -e -o ppid,pid,args -H |grep kube-sch
1 42555   /root/local/bin/kube-scheduler --logtostderr=true --v=0 --master=http://10.64.3.7:8080 --address=127.0.0.1
$ netstat  -lnpt|grep kube-schedule
tcp        0      0 127.0.0.1:10251         0.0.0.0:*               LISTEN      42555/kube-schedule
```

## 查看日志

``` bash
$ journalctl -u kube-scheduler -o export|grep MESSAGE
$ journalctl -u kube-scheduler -o export|grep MESSAGE
MESSAGE_ID=39f53479d3a045ac8e11786248231fbf
MESSAGE=Started Kubernetes Scheduler Plugin.
MESSAGE_ID=7d4958e842da4a758f6c1cdc7b36dcc5
MESSAGE=Starting Kubernetes Scheduler Plugin...
MESSAGE=I0312 12:21:27.192438   46584 leaderelection.go:188] sucessfully acquired lease kube-system/kube-scheduler
MESSAGE=I0312 12:21:27.192534   46584 event.go:217] Event(api.ObjectReference{Kind:"Endpoints", Namespace:"kube-system", Name:"kube-scheduler", UID:"5c296e1f-06db-11e7-8472-8cdcd4b3be48", APIVersion:"v1", ResourceVersion:"682325", FieldPath:""}): type: 'Normal' reason: 'LeaderElection' tjwq01-sys-bs003007.tjwq01.ksyun.com became leader
```