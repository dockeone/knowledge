https://skyao.gitbooks.io/leaning-etcd3/content/documentation/dev-guide/interacting_v3.html

etcdctl兼容V2 API和V3 API，默认是V2 API，通过指定环境变量ETCDCTL_API=3的形式使用V3 API：

$ export ETCDCTL_API=3

其它全局参数，可以使用环境变量指定：
ETCDCTL_DIAL_TIMEOUT=3s
ETCDCTL_CACERT=/tmp/ca.pem
ETCDCTL_CERT=/tmp/cert.pem
ETCDCTL_KEY=/tmp/key.pem

# version

./etcdctl version
etcdctl version: 3.1.0-alpha.0+git
API version: 3.1


# 写入key: PUT [options] <key> <value>

Options
lease -- lease ID (in hexadecimal) to attach to the key.
prev-kv -- return the previous key-value pair before modification.

应用通过写入 key 来储存 key 到 etcd 中。每个存储的 key 被通过 Raft 协议复制到所有 etcd 集群成员来达到一致性和可靠性。
这是设置 key foo 的值为 bar 的命令:

$ etcdctl put foo bar
OK

写入key的时候可以指定一个lease ID, 这时该key的生命周期由该lease控制；
$ ./etcdctl put foo bar --lease=1234abcd
OK

$ ./etcdctl get foo
foo
bar

$ ./etcdctl put foo bar1 --prev-kv
OK
foo
bar

$./etcdctl get foo
foo
bar1

# 读取 key: GET [options] <key> [range_end]

Options

hex -- print out key and value as hex encode string // 适用于key和value包含非可打印字符的情况(如二进制内容)
limit -- maximum number of results // 限制返回的key数目
prefix -- get keys by matching prefix  // 返回前缀为key的所有key
order -- order of results; ASCEND or DESCEND // 排序方式
sort-by -- sort target; CREATE, KEY, MODIFY, VALUE, or VERSION // 可以对key、value、创建时间、修改时间和版本等排序；
rev -- specify the kv revision // 返回指定revision版本的key
print-value-only -- print only value when used with write-out=simple // 只返回value
consistency -- Linearizable(l) or Serializable(s)
// 返回比指定key名称大的key，如指定的key为a，则不但返回a1, a2，还可以返回b, c,d 等比a大的key；
// 而prefix只返回以a开头的key；
from-key -- Get keys that are greater than or equal to the given key using byte compare 
keys-only -- Get only the keys // 只返回key

应用可以从 etcd 集群中读取 key 的值。查询可以读取单个 key，或者[key, range_end)范围的key。

./etcdctl put foo bar
OK

./etcdctl put foo1 bar1
OK

./etcdctl put foo2 bar2
OK

./etcdctl put foo3 bar3
OK

./etcdctl get foo
foo
bar

./etcdctl get --from-key foo1
foo1
bar1
foo2
bar2
foo3
bar3

./etcdctl get foo1 foo3
foo1
bar1
foo2
bar2

注： 测试中发现，这个命令的有效区间是这样：[foo foo9)， 即包含第一个参数，但是不包含第二个参数。因此如果第二个参数是 foo3，上面的命令是不会返回 key foo3 的值的。具体可以参考 API 说明。

# 读取 key 过往版本的值

应用可能想读取 key 的被替代的值。例如，应用可能想通过访问 key 的过往版本来回滚到旧的配置。或者，应用可能想通过访问 key 历史记录的多个请求来得到一个覆盖多个 key 上的统一视图。
因为 etcd 集群上键值存储的每个修改都会增加 etcd 集群的**全局修订版本**，应用可以通过提供旧有的 etcd 版本来读取被替代的 key。

假设 etcd 集群已经有下列 key：
$ etcdctl put foo bar         # revision = 2
$ etcdctl put foo1 bar1       # revision = 3
$ etcdctl put foo bar_new     # revision = 4
$ etcdctl put foo1 bar1_new   # revision = 5

这里是访问 key 的过往版本的例子：
$ etcdctl get foo foo9 # 访问 key 的最新版本
foo
bar_new
foo1
bar1_new

$ etcdctl get --rev=4 foo foo9 # 访问 key 的修订版本4
foo
bar_new
foo1
bar1

$ etcdctl get --rev=3 foo foo9 # 访问 key 的修订版本3
foo
bar
foo1
bar1

$ etcdctl get --rev=2 foo foo9 # 访问 key 的修订版本2
foo
bar

$ etcdctl get --rev=1 foo foo9 # 访问 key 的修订版本1

# 获取当前的全局修订版本

$ etcdctl --endpoints=:2379 endpoint status
[{"Endpoint":"127.0.0.1:2379","Status":{"header":{"cluster_id":8925027824743593106,"member_id":13803658152347727308,"revision":1516,"raft_term":2},"version":"2.3.0+git","dbSize":17973248,"leader":13803658152347727308,"raftIndex":6359,"raftTerm":2}}]

# 删除 key DEL [options] <key> [range_end]

Removes the specified key or range of keys [key, range_end) if range-end is given.

Options

prefix -- delete keys by matching prefix // 删除以prefix为前缀的key
prev-kv -- return deleted key-value pairs
from-key -- delete keys that are greater than or equal to the given key using byte compare // 删除key名称被指定key大的key

应用可以从 etcd 集群中删除一个 key 或者特定范围的 key。

下面是删除 key foo 的命令：
$ etcdctl del foo
1 # 删除了一个key

这是删除从 foo to foo9 范围的 key 的命令：
$ etcdctl del foo foo9
2 # 删除了两个key

./etcdctl put a 123
OK 

./etcdctl put b 456
OK 

./etcdctl put z 789
OK 

./etcdctl del --from-key a
3

./etcdctl get --from-key a

./etcdctl put zoo val
OK 

./etcdctl put zoo1 val1
OK 

./etcdctl put zoo2 val2
OK 

./etcdctl del --prefix zoo
3
./etcdctl get zoo2

# 观察 key 的变化

应用可以观察一个 key 或者特定范围内的 key 来监控任何更新。

这是在 key foo 上进行观察的命令：
$ etcdctl watch foo

在另外一个终端: 
$ etcdctl put foo bar
foo
bar

这是观察从 foo to foo9 范围key的命令：
$ etcdctl watch foo foo9

在另外一个终端: 
$ etcdctl put foo bar
foo
bar

在另外一个终端: 
$ etcdctl put foo1 bar1
foo1
bar1

# 观察 key 的历史改动

应用可能想观察 etcd 中 key 的历史改动。例如，应用想接收到某个 key 的所有修改。如果应用一直连接到etcd，那么 watch 就足够好了。但是，如果应用或者
etcd 出错，改动可能发生在出错期间，这样应用就没能实时接收到这个更新。为了保证更新被接收，应用必须能够观察到 key 的历史变动。为了做到这点，应用可以在观察时指定一个历史修订版本，
就像读取 key 的过往版本一样。

假设我们完成了下列操作序列：
etcdctl put foo bar         # revision = 2
etcdctl put foo1 bar1       # revision = 3
etcdctl put foo bar_new     # revision = 4
etcdctl put foo1 bar1_new   # revision = 5

这是观察历史改动的例子：

从修订版本 2 开始观察key `foo` 的改动
$ etcdctl watch --rev=2 foo
PUT
foo
bar
PUT
foo
bar_new

从修订版本 3 开始观察key `foo` 的改动
$ etcdctl watch --rev=3 foo
PUT
foo
bar_new

# 压缩修订版本

如我们提到的，etcd 保存修订版本以便应用可以读取 key 的过往版本。但是，为了避免积累无限数量的历史数据，压缩过往的修订版本就变得很重要。压缩之后，etcd 删除历史修订版本，释放资源来提供未来使用。
所有修订版本在压缩修订版本之前的被替代的数据将不可访问。

这是压缩修订版本的命令：
$ etcdctl compact 5
compacted revision 5

在压缩修订版本之前的任何修订版本都不可访问
$ etcdctl get --rev=4 foo
Error:  rpc error: code = 11 desc = etcdserver: mvcc: required revision has been compacted

# 授予租约

应用可以为 etcd 集群里面的 key 授予租约。当 key 被附加到租约时，它的生存时间被绑定到租约的生存时间，而租约的生存时间相应的被 time-to-live (TTL)管理。租约的实际 TTL 值是不低于最小
TTL，由 etcd 集群选择。一旦租约的 TTL 到期，租约就过期并且所有附带的 key 都将被删除。

这是授予租约的命令：

授予租约，TTL为10秒
$ etcdctl lease grant 10
lease 32695410dcc0ca06 granted with TTL(10s)

附加key foo到租约32695410dcc0ca06
$ etcdctl put --lease=32695410dcc0ca06 foo bar
OK

# 撤销租约

应用通过租约 id 可以撤销租约。撤销租约将删除所有它附带的 key。

假设我们完成了下列的操作：
$ etcdctl lease grant 10
lease 32695410dcc0ca06 granted with TTL(10s)

$ etcdctl put --lease=32695410dcc0ca06 foo bar
OK

这是撤销同一个租约的命令：
$ etcdctl lease revoke 32695410dcc0ca06
lease 32695410dcc0ca06 revoked

$ etcdctl get foo
空应答，因为租约撤销导致foo被删除

# 维持租约

应用可以通过刷新 key 的 TTL 来维持租约，以便租约不过期。

假设我们完成了下列操作：
$ etcdctl lease grant 10
lease 32695410dcc0ca06 granted with TTL(10s)

这是维持同一个租约的命令：
$ etcdctl lease keep-alive 32695410dcc0ca0
lease 32695410dcc0ca0 keepalived with TTL(100)
lease 32695410dcc0ca0 keepalived with TTL(100)
lease 32695410dcc0ca0 keepalived with TTL(100)
...
注： 上面的这个命令中，etcdctl 不是单次续约，而是 etcdctl 会一直不断的发送请求来维持这个租约。


# 维护

etcd 集群需要定期维护来保持可靠。基于 etcd 应用的需要，这个维护通常可以自动执行，不需要停机或者显著的降低性能。

所有 etcd 的维护是指管理被 etcd 键空间消耗的存储资源。通过存储空间的配额来控制键空间大小; 如果 etcd 成员运行空间不足，将触发集群级警告，这将使得系统进入有限操作的维护模式。为了避免没有
空间来写入键空间，etcd 键空间历史必须被压缩。存储空间自身可能通过碎片整理 etcd 成员来回收。最后，etcd 成员状态的定期快照备份使得恢复任何非故意的逻辑数据丢失或者操作错误导致的损坏变成可能。

## 历史压缩

因为 etcd 保持它的键空间的确切历史，这个历史应该定期压缩来避免性能下降和最终的存储空间枯竭。压缩键空间历史删除所有关于被废弃的在给定键空间修订版本之前的键的信息。这些key使用的空间随机变得
可用来继续写入键空间。键空间可以使用 etcd 的时间窗口历史保持策略自动压缩，或者使用 etcdctl 手工压缩。 etcdctl 方法在压缩过程上提供细粒度的控制，反之自动压缩适合仅仅需要一定时间长度的键
历史的应用。etcd 可以使用带有小时时间单位的 --auto-compaction 选项来设置为自动压缩键空间:

保持一个小时的历史
$ etcd --auto-compaction-retention=1

压缩到修订版本3
$ etcdctl compact 3

在压缩修订版本之前的修订版本变得无法访问：
$ etcdctl get --rev=2 somekey
Error:  rpc error: code = 11 desc = etcdserver: mvcc: required revision has been compacted

## 反碎片化

在压缩键空间之后，后端数据库可能出现碎片。内部碎片是指可以被后端使用但是依然消耗存储空间的空间。反碎片化过程释放这个存储空间到文件系统。反碎片化在每个成员上发起，因此集群范围的延迟尖峰
(latency spike)可能可以避免。通过留下间隔在后端数据库，压缩旧有修订版本会内部碎片化 etcd 。碎片化的空间可以被 etcd 使用，但是对于主机文件系统不可用。

$ time ./etcdctl defrag # 命令执行超时而失败(默认5s)
Failed to defragment etcd member[127.0.0.1:2379] (context deadline exceeded)

real    0m5.034s
user    0m0.030s
sys     0m0.020s

$ time ./etcdctl defrag  --command-timeout=100s  # 指定超时时间
Finished defragmenting etcd member[127.0.0.1:2379]

real    0m6.153s

## 空间配额

在 etcd 中空间配额确保集群以可靠方式运作。没有空间配额， etcd 可能会收到低性能的困扰，如果键空间增长的过度的巨大，或者可能简单的超过存储空间，导致不可预测的集群行为。如果键空间的任何成员
的后端数据库超过了空间配额， etcd 发起集群范围的警告，让集群进入维护模式，仅接收键的读取和删除。在键空间释放足够的空间之后，警告可以被解除，而集群将恢复正常运作。
默认，etcd 设置适合大多数应用的保守的空间配额，但是它可以在命令行中设置，单位为字节：

设置非常小的 16MB 配额
$ etcd --quota-backend-bytes=16777216

空间配额可以用循环触发：

消耗空间
$ while [ 1 ]; do dd if=/dev/urandom bs=1024 count=1024  | etcdctl put key  || break; done
...
Error:  rpc error: code = 8 desc = etcdserver: mvcc: database space exceeded

确认配额空间被超过
$ etcdctl --write-out=table endpoint status
+----------------+------------------+-----------+---------+-----------+-----------+------------+
|    ENDPOINT    |        ID        |  VERSION  | DB SIZE | IS LEADER | RAFT TERM | RAFT INDEX |
+----------------+------------------+-----------+---------+-----------+-----------+------------+
| 127.0.0.1:2379 | bf9071f4639c75cc | 2.3.0+git | 18 MB   | true      |         2 |       3332 |
+----------------+------------------+-----------+---------+-----------+-----------+------------+

确认警告已发起
$ etcdctl alarm list
memberID:13803658152347727308 alarm:NOSPACE

删除多读的键空间将把集群带回配额限制，因此警告能被解除:

获取当前修订版本
$ etcdctl --endpoints=:2379 endpoint status
[{"Endpoint":"127.0.0.1:2379","Status":{"header":{"cluster_id":8925027824743593106,"member_id":13803658152347727308,"revision":1516,"raft_term":2},"version":"2.3.0+git","dbSize":17973248,"leader":13803658152347727308,"raftIndex":6359,"raftTerm":2}}]

压缩所有旧有修订版本
$ etdctl compact 1516
compacted revision 1516

反碎片化过度空间
$ etcdctl defrag
Finished defragmenting etcd member[127.0.0.1:2379]

解除警告
$ etcdctl alarm disarm
memberID:13803658152347727308 alarm:NOSPACE

测试put被再度容许
$ etdctl put newkey 123
OK

# 快照备份

在正规基础上执行 etcd 集群快照可以作为 etc 键空间的持久备份。通过获取 etcd 成员的候选数据库的定期快照，etcd 集群可以被恢复到某个有已知良好状态的时间点。

通过 etcdctl 获取快照：
$ etcdctl snapshot save backup.db
$ etcdctl --write-out=table snapshot status backup.db
+----------+----------+------------+------------+
|   HASH   | REVISION | TOTAL KEYS | TOTAL SIZE |
+----------+----------+------------+------------+
| fe01cf57 |       10 |          7 | 2.1 MB     |
+----------+----------+------------+------------+