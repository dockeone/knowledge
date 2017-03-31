# Cassandra

NoSQL数据库的选择之痛，目前市面上有近150多种NoSQL数据库，如何在这么庞杂的队伍中选中适合业务场景的佼佼者，实非易事。
好的是经过大量的筛选，大家比较肯定的几款NoSQL数据库分别是HBase、MongoDB和Cassandra。

Cassandra在哪些方面吸引住了大量的开发人员呢？下面仅做一个粗略的分析。

1.1 高可靠性

Cassandra采用gossip作为集群中结点的通信协议，该协议整个集群中的节点都处于**同等地位**，没有主从之分，这就使得任一节点的退出都不会导致整个集群失效。
Cassandra和HBase都是借鉴了Google BigTable的思想来构建自己的系统，但Cassandra另一重要的创新就是将原本存在于将**文件共享架构的p2p(peer to peer)引入了NoSQL**。
P2P的一大特点就是**去中心化，集群中的所有节点享有同等地位**，这极大避免了单个节点退出而使整个集群不能工作的可能。
与之形成对比的是HBase采用了Master/Slave的方式，这就存在单点失效的可能。

1.2 高可扩性

随着时间的推移，集群中原有的规模不足以存储新增加的数据，此时进行系统扩容。Cassandra级联可扩，非常容易实现添加新的节点到已有集群，操作简单。

1.3 最终一致性

分布式存储系统都要面临CAP定律问题，任何一个分布式存储系统不可能同时满足一致性(consistency)，可用性(availability)和分区容错性(partition tolerance)。
Cassandra是**优先保证AP**，即可用性和分区容错性。

Cassandra为写操作和读操作提供了不同级别的一致性选择，用户可以根据具体的应用场景来选择不同的一致性级别。

1.4 高效写操作

写入操作非常高效，这对于实时数据非常大的应用场景，Cassandra的这一特性无疑极具优势。
数据读取方面则要视情况而定：

1. 如果是单个读取即指定了键值，会很快的返回查询结果。
2. 如果是范围查询，由于查询的目标可能存储在多个节点上，这就需要对多个节点进行查询，所以返回速度会很慢
3. 读取全表数据，非常低效。

1.5 结构化存储

Cassandra是一个**面向列**的数据库，对那些从RDBMS方面转过来的开发人员来说，其学习曲线相对平缓。
Cassandra同时提供了较为友好CQL语言，与SQL语句相似度很高。

1.6 维护简单

从系统维护的角度来说，由于Cassandra的对等系统架构，使其维护操作简单易行。如添加节点，删除节点，甚至于添加新的数据中心，操作步骤都非常的简单明了。

# Cassandra数据模型

2.1 单表查询

2.1.1 单表主键查询

在建立个人信息数据库的时候，以个人身份证id为主键，查询的时候也只以身份证为关键字进行查询，则表可以设计成为：

create table person (
	userid text primary key,
	fname text,
	lname text,
	age	int,
	gender int);

Primary key中的**第一个列**名是作为Partition key。也就是说根据针对partition key的hash结果决定将记录存储在哪一个partition中，如果不湊巧的情况下单一主键导致所有的hash结果
全部落在同一分区，则会导致该分区数据被撑满。

解决这一问题的办法是通过**组合分区键(compsoite partition key)**来使得数据尽可能的均匀分布到各个节点上。

举例来说，可能将(userid,fname)设置为**复合主键**。那么相应的表创建语句可以写成

create table person (
userid text,
fname text,
lname text,
gender int,
age int,
primary key((userid,fname),lname);
) with clustering order by (lname desc);

稍微解释一下primary key((userid, fname),lname)的含义：

+ 其中(userid,fname)称为组合分区键(composite partition key)
+ lname是聚集列(clustering column)
+ ((userid,fname),lname)一起称为复合主键(composite primary key)

2.1.2 单表非主键查询

如果要查询表person中具有相同的first name的人员，那么就必须针对fname创建相应的**索引**，否则查询速度会非常缓慢。

Create index on person(fname);

Cassandra目前只能对表中的**某一列**建立索引，不允许对多列建立联合索引。

2.2 多表关联查询

Cassandra并**不支持关联查询，也不支持分组和聚合操作**。
那是不是就说明Cassandra只是看上去很美其实根本无法解决实际问题呢？答案显然是No,只要你不坚持用RDBMS的思路来解决问题就是了。
比如我们有两张表，一张表(Departmentt)记录了公司部门信息，另一张表(employee)记录了公司员工信息。显然每一个员工必定有归属的部门，如果想知道每一个部门拥有的所有员工。
如果是用RDBMS的话，SQL语句可以写成：

select * from employee e , department d where e.depId = d.depId;

要用Cassandra来达到同样的效果，就必须在employee表和department表之外，再创建一张额外的表(dept_empl)来记录每一个部门拥有的员工信息。

Create table dept_empl (
deptId text,

看到这里想必你已经明白了，在Cassandra中**通过数据冗余来实现高效的查询效果。将关联查询转换为单一的表操作**。

2.3 分组和聚合

在RDBMS中常见的group by和max、min在Cassandra中是不存在的。
如果想将所有人员信息按照姓进行分组操作的话，那该如何创建数据模型呢？

Create table fname_person (
fname text,
userId text,
primary key(fname)
);

 2.4 子查询 
Cassandra不支持子查询，下图展示了一个在MySQL中的子查询例子： 
要用Cassandra来实现，必须通过**添加额外的表**来存储冗余信息。 

Create table office_empl (
officeCode text,
country text,
lastname text,
firstname,
primary key(officeCode,country));
create index on office_empl(country);

 2.5 小结

**总的来说，在建立Cassandra数据模型的时候，要求对数据的读取需求进可能的清晰，然后利用反范式的设计方式来实现快速的读取，原则就是以空间来换取时间。**

# 利用Spark强化Cassandra的实时分析功能

在Cassandra数据模型一节中，讲述了通过数据冗余和反范式设计来达到快速高效的查询效果。

但如果对存储于cassandra数据要做更为复杂的实时性分析处理的话，使用原有的技巧无法实现目标，那么可以通过与Spark相结合，利用Spark这样一个快速高效的分析平台来实现复杂的数据分析功能。

3.1 整体架构
  
利用spark-cassandra-connector连接Cassandra，读取存储在Cassandra中的数据，然后就可以使用Spark RDD中的支持API来对数据进行各种操作。

3.2 Spark-cassandra-connector

在Spark中利用datastax提供的spark-cassandra-connector来连接Cassandra数据库是最为简单的一种方式。

如果是用sbt来管理scala程序的话，只需要在build.sbt中加入如下内容即可由sbt自动下载所需要的spark-cassandra-connector驱动

datastax.spark" %% "spark-cassandra-connector" % "1.1.0-alpha3" withSources() withJavadoc()

3.2.1 driver的配置

使用spark-cassandra-connector的时候需要编辑一些参数，比如指定Cassandra数据库的地址，每次最多获取多少行，一个线程总共获取多少行等。

这些参数即可以硬性的写死在程序中，如

[cpp] view plaincopy在CODE上查看代码片派生到我的代码片
val conf = new SparkConf()  
conf.set(“spark.cassandra.connection.host”, cassandra_server_addr)  
conf.set(“spark.cassandra.auth.username”, “cassandra”)  
conf.set(“spark.cassandra.auth.password”,”cassandra”)  

硬编码的方式是发动不灵活，其实这些配置参数完全可以写在spark-defaults.conf中，那么上述的配置可以写成

spark.cassandra.connection.host 192.168.6.201
spark.cassandra.auth.username cassandra
spark.cassandra.auth.password cassandra

3.2.2 依赖包的版本问题

sbt会自动下载spark-cassandra-connector所依赖的库文件，这在程序编译阶段不会呈现出任何问题。

但在执行阶段问题就会体现出来，即程序除了spark-cassandra-connector之外还要依赖哪些文件呢，这个就需要重新回到maven版本库中去看spark-cassandra-connector的依赖了。

总体上来说spark-cassandra-connector严重依赖于这几个库

cassandra-clientutil
cassandra-driver-core
cassandra-all

另外一种解决的办法就是查看$HOME/.ivy2目录下这些库的最新版本是多少 

find ~/.ivy2 -name “cassandra*.jar”

取最大的版本号即可，就alpha3而言，其所依赖的库及其版本如下 
com.datastax.spark/spark-cassandra-connector_2.10/jars/spark-cassandra-connector_2.10-1.1.0-alpha3.jar
org.apache.cassandra/cassandra-thrift/jars/cassandra-thrift-2.1.0.jar
org.apache.thrift/libthrift/jars/libthrift-0.9.1.jar
org.apache.cassandra/cassandra-clientutil/jars/cassandra-clientutil-2.1.0.jar
com.datastax.cassandra/cassandra-driver-core/jars/cassandra-driver-core-2.1.0.jar
io.netty/netty/bundles/netty-3.9.0.Final.jar
com.codahale.metrics/metrics-core/bundles/metrics-core-3.0.2.jar
org.slf4j/slf4j-api/jars/slf4j-api-1.7.7.jar
org.apache.commons/commons-lang3/jars/commons-lang3-3.3.2.jar
org.joda/joda-convert/jars/joda-convert-1.2.jar
joda-time/joda-time/jars/joda-time-2.3.jar
org.apache.cassandra/cassandra-all/jars/cassandra-all-2.1.0.jar
org.slf4j/slf4j-log4j12/jars/slf4j-log4j12-1.7.2.jar 

3.3 Spark的配置

程序顺利通过编译之后，准备在Spark上进行测试，那么需要做如下配置

3.3.1 spark-default.env

Spark-defaults.conf的作用范围要搞清楚，编辑driver所在机器上的spark-defaults.conf，该文件会影响到driver所提交运行的application，及专门为该application提供计算资源的executor的启动参数

只需要在driver所在的机器上编辑该文件，不需要在worker或master所运行的机器上编辑该文件

举个实际的例子

spark.executor.extraJavaOptions	   -XX:MaxPermSize=896m
spark.executor.memory		   5g
spark.serializer        org.apache.spark.serializer.KryoSerializer
spark.cores.max		32
spark.shuffle.manager	SORT
spark.driver.memory	2g
上述配置表示为该application提供计算资源的executor启动时, heap memory需要有5g。

这里需要引起注意的是，如果worker在加入cluster的时候，申明自己所在的机器只有4g内存，那么为上述的application分配executor是，该worker不能提供任何资源，因为4g<5g，无法满足最低的资源需求。

3.3.2 spark-env.sh

Spark-env.sh中最主要的是指定ip地址，如果运行的是master，就需要指定SPARK_MASTER_IP，如果准备运行driver或worker就需要指定SPARK_LOCAL_IP，要和本机的IP地址一致，否则启动不了。

配置举例如下

export SPARK_MASTER_IP=127.0.0.1
export SPARK_LOCAL_IP=127.0.0.1
3.3.3 启动Spark集群

第一步启动master

$SPARK_HOME/sbin/start-master.sh
第二步启动worker

$SPARK_HOME/bin/spark-class org.apache.spark.deploy.worker.Worker spark://master:7077
将master替换成MASTER实际运行的ip地址

如果想在一台机器上运行多个worker(主要是用于测试目的),那么在启动第二个及后面的worker时需要指定—webui-port的内容，否则会报端口已经被占用的错误,启动第二个用的是8083，第三个就用8084，依此类推。

$SPARK_HOME/bin/spark-class org.apache.spark.deploy.worker.Worker spark://master:7077
    –webui-port 8083
这种启动worker的方式只是为了测试是启动方便，正规的方式是用$SPARK_HOME/sbin/start-slaves.sh来启动多个worker，由于涉及到ssh的配置，比较麻烦，我这是图简单的办法。

用$SPARK_HOME/sbin/start-slave.sh来启动worker时有一个默认的前提，即在每台机器上$SPARK_HOME必须在同一个目录。

注意：

使用相同的用户名和用户组来启动Master和Worker，否则Executor在启动后会报连接无法建立的错误。

我在实际的使用当中，遇到”no route to host”的错误信息，起初还是认为网络没有配置好，后来网络原因排查之后，忽然意识到有可能使用了不同的用户名和用户组，使用相同的用户名/用户组之后，问题消失。

3.3.4 Spark-submit

spark集群运行正常之后，接下来的问题就是提交application到集群运行了。

Spark-submit用于Spark application的提交和运行，在使用这个指令的时候最大的困惑就是如何指定应用所需要的依赖包。

首先查看一下spark-submit的帮助文件

$SPARK_HOME/bin/submit --help

有几个选项可以用来指定所依赖的库，分别为

--driver-class-path driver所依赖的包，多个包之间用冒号(:)分割
--jars   driver和executor都需要的包，多个包之间用逗号(,)分割
为了简单起见，就通过—jars来指定依赖，运行指令如下

$SPARK_HOME/bin/spark-submit –class 应用程序的类名 \
--master spark://master:7077 \
--jars 依赖的库文件 \
spark应用程序的jar包
3.3.5 RDD函数使用的一些问题

collect

如果数据集特别大，不要贸然使用collect，因为collect会将计算结果统统的收集返回到driver节点，这样非常容易导致driver结点内存不足，程序退出

repartition

在所能提供的core数目不变的前提下，数据集的分区数目越大，意味着计算一轮所花的时间越多，因为中间的通讯成本较大，而数据集的分区越小，通信开销小而导致计算所花的时间越短，但数据分区越小意味着内存压力越大。

假设为每个spark application提供的最大core数目是32,那么将partition number设置为core number的两到三倍会比较合适，即parition number为64～96。

/tmp目录问题

由于Spark在计算的时候会将中间结果存储到/tmp目录，而目前linux又都支持tmpfs，其实说白了就是将/tmp目录挂载到内存当中。

那么这里就存在一个问题，中间结果过多导致/tmp目录写满而出现如下错误

No Space Left on the device
解决办法就是针对tmp目录不启用tmpfs,修改/etc/fstab，如果是archlinux，仅修改/etc/fstab是不够的，还需要执行如下指令：

systemctl mask tmp.mount
3.4 Cassandra的配置优化

3.4.1 表结构设计

Cassandra表结构设计的一个重要原则是先搞清楚要对存储的数据做哪些操作，然后才开始设计表结构。如：

只对表进行添加，查询操作
对表需要进行添加，修改，查询
对表进行添加和修改操作
一般来说，针对Cassandra中某张具体的表进行“添加，修改，查询”并不是一个好的选择，这当中会涉及到效率及一致性等诸多问题。

Cassandra比较适合于添加，查询这种操作模式。在这种模式下，需要先搞清楚要做哪些查询然后再来定义表结构。

加深对Cassandra中primary key及其变种的理解有利于设计出高效查询的表结构。

create test ( k int, v int , primary key(k,v))
上述例子中primary key由(k,v)组成，其中k是partition key,而v是clustering columns，如果k相同，那么这些记录在物理存储上其实是存储在同一行中，即Cassandra中常会提及的wide rows.

有了这个基础之后，就可以进行范围查询了

select * from test where k = ? and v > ? and v < ?
当然也可以对k进行范围查询，不过要加token才行，但一般这样的范围查询结果并不是我们想到的 
select * from test where token(k) > ? and token(k) < ?
Cassandra中针对二级索引是不支持范围查询的，一切的一切都在主键里打主意。

3.4.2 参数设置

Cassandra的配置参数项很多，对于新手来说主要集中于对这两个文件中配置项的理解。

cassandra.yaml   Cassandra系统的运行参数
cassandra-env.sh  JVM运行参数
在cassandra-env.sh中针对JVM的设置

JVM_OPTS="$JVM_OPTS -XX:+UseParNewGC" 
JVM_OPTS="$JVM_OPTS -XX:+UseConcMarkSweepGC" 
JVM_OPTS="$JVM_OPTS -XX:+CMSParallelRemarkEnabled" 
JVM_OPTS="$JVM_OPTS -XX:SurvivorRatio=8" 
JVM_OPTS="$JVM_OPTS -XX:MaxTenuringThreshold=1"
JVM_OPTS="$JVM_OPTS -XX:CMSInitiatingOccupancyFraction=80"
JVM_OPTS="$JVM_OPTS -XX:+UseCMSInitiatingOccupancyOnly"
JVM_OPTS="$JVM_OPTS -XX:+UseTLAB"
JVM_OPTS="$JVM_OPTS -XX:ParallelCMSThreads=1"
JVM_OPTS="$JVM_OPTS -XX:+CMSIncrementalMode"
JVM_OPTS="$JVM_OPTS -XX:+CMSIncrementalPacing"
JVM_OPTS="$JVM_OPTS -XX:CMSIncrementalDutyCycleMin=0"
JVM_OPTS="$JVM_OPTS -XX:CMSIncrementalDutyCycle=10"
如果nodetool无法连接到Cassandra的话，在cassandra-env.sh中添加如下内容 
JVM_OPTS="$JVM_OPTS -Djava.rmi.server.hostname=ipaddress_of_cassandra"
在cassandra.yaml中，注意memtable_total_space_in_mb的设置，不要将该值设的特别大。将其配置成为JVM HEAP的1/4会是一个比较好的选择。如果该值设置太大，会导致不停的FULL GC，那么在这种情况下Cassandra基本就不可用了。

3.4.3 nodetool使用

Cassandra在运行期间可以通过nodetool来看内部的一些运行情况。

如看一下读取的完成情况

nodetool -hcassandra_server_address tpstats
检查整个cluster的状态 
nodetool -hcassandra_server_address status
检查数据库中每个表的数据有多少 
nodetool -hcassandra_server_address cfstats
To Be Contunued……

关于作者：许鹏，一个喜欢读点文学的老程序员，长期混迹于通信领域，研究过点Linux内核，目前迷上了大数据计算框架Spark 。


# Cassandra高并发数据读取实现剖析

本文就spark-cassandra-connector的一些实现细节进行探讨，主要集中于如何快速将大量的数据从Cassandra中读取到本地内存或磁盘。 

## 数据分区

存储在Cassandra中的数据一般都会比较多，记录数在千万级别或上亿级别是常见的事。如何将这些表中的内容快速加载到本地内存就是一个非常现实的问题。
解决这一挑战的思路从大的方面来说是比较简单的，那就是**将整张表中的内容分成不同的区域，然后分区加载，不同的分区可以在不同的线程或进程中加载，利用并行化来减少整体加载时间**。
顺着这一思路出发，要问的问题就是Cassandra中的数据如何才能分成不同的区域。

不同于MySQL，在Cassandra中是不存在Sequence Id这样的类型的，也就是说无法简单的使用seqId来指定查询或加载的数据范围。

既然没有SequenceID，在Cassandra中是否就没有办法了呢？答案显然是否定的，如果只是仅仅支持串行读取，Cassandra早就会被扔进垃圾桶了。
数据分区在Cassandra中至少可以通过两种途径实现 ，一是通过token range，另一个是slice range。这里主要讲解利用token range来实现目的。

1. Token Range
Cassandra将要存储的记录存储在不同的区域中，判断某一记录具体存储在哪个区域的依据是partition key的Hash值。 
在Cassandra 1.2之前，组成Cassandra集群的所有节点（Node），都需要手动指定该节点的Hash值范围也就是Token Range。
手工计算Token Range显然是很繁琐，同时也不怎么容易维护，在Cassandra 1.2之后，引进了虚拟节点（vnode）的概念，主要目的是减少不必要的人工指定，同时也将token range的划分变得更为细粒度。
比如原先手工指定token range，只能达到10000这样一个精度，而有了vnode之后，默认安装是每一个物理节点上有256个虚拟节点，这样子的话每一个range的范围就是10000/256，这样变的更为精细。

**有关token range的信息存储在cassandra的system命名空间(keyspace)下的local和peers两张表中**。其中local表示本节点的token range情况，而peers表示集群中其它节点的token range情况。
这两张表中的tokens字段就存储有详细的信息。如果集群中只由一台机器组成，那么peers中的就会什么内容都没有。

简单实验，列出本节点的token range：

use system;
desc table local;
select tokens from local;

2. Thrift接口

Token Range告诉我们Cassandra的记录是分片存储的，也就意味着可以分片读取。现在的问题转换成为如何知道每一个Token Range的起止范围。
Cassandra支持的Thrift接口中describe_ring就是用来获取token range的具体起止范围的。我们常用的nodetool工具使用的就是thrift接口，nodetool 中有一个describering指令使用的就是
describe_ring原语。

可以做一个简单的实验，利用nodetool来查看某个keyspace的token range具体情况。

nodetool -hcassandra_server_addr describering keyspacename

注意将cassandra_server和keyspacename换成实际的内容。

## Spark-Cassandra-Connector

在第一节中讲解了Cassandra中Token Range信息的存储位置，以及可以使用哪些API来获取token range信息。
接下来就分析spark-cassandra-connector是如何以cassandra为数据源将数据加载进内存的。
以简单的查询语句为例，假设用户要从demo这个keyspace的tableX表中加载所有数据，用CQL来表述就是：

select * from demo.tableX

上述的查询使用spark-cassandra-connector来表述就是：
sc.cassandraTable(“demo”,”tableX”)

尽管上述语句没有触发Spark Job的提交，也就是说并不会将数据直正的从Cassandra的tableX表中加载进来，但spark-cassandra-connector还是需要进行一些数据库的操作。要解决的主要问题就是
schema相关。
cassandraTable(“demo”,”tableX”)只是说要从tableX中加载数据，并没有告诉connector有哪些字段，每个字段的类型是什么。这些信息对后面使用诸如get[String](“fieldX”)来说却是非常关键的。
为了获取字段类型信息的元数据，需要读取system.schema_columns表，利用如下语句可以得到schema_columns表结构的详细信息：

desc table system.schema_columns

如果在conf/log4j.properties中将日志级别设置为DEBUG，然后再执行sc.cassandraTable语句就可以看到具体的CQL查询语句是什么。

### CassandraRDDPartitioner

Spark-cassandra-connector添加了一种新的RDD实现，即**CassandraRDD**。我们知道对于一个Spark RDD来说，非常关键的就是确定getPartitions和compute函数。
getPartitions函数会调用CassandraRDDPartitioner来获取分区数目：

override def getPartitions: Array[Partition] = {
    verify // let's fail fast
    val tf = TokenFactory.forCassandraPartitioner(cassandraPartitionerClassName)
    val partitions = new CassandraRDDPartitioner(connector, tableDef, splitSize)(tf).partitions(where)
    logDebug(s"Created total ${partitions.size} partitions for $keyspaceName.$tableName.")
    logTrace("Partitions: \n" + partitions.mkString("\n"))
    partitions
  }

CassandraRDDPartitioner中的partitions的处理逻辑大致如下： 
1. 首先确定token range，使用describe_ring
2. 然后根据Cassandra中使用的Partitioner来确定某一个token range中可能的记录条数，这么做的原因就是为进一步控制加载的数据，**提高并发度**。否则并发度就永远是256了，比如有一个物理节点，
其中有256个vnodes，也就是256个token分区。如果每个分区中大致的记录数是20000，而每次加载最大只允许1000的话，整个数据就可以分成256x2=512个分区。
3. 对describeRing返回的token range进一步拆分的话，需要使用splitter，splitter的构建需要根据keyspace中使用了何种Partitioner来决定，Cassandra中默认的Partitioner是
Murmur3Partitioner，Murmur3Hash算法可以让Hash值更为均匀的分布到不同节点。
4. splitter中会利用到配置项spark.cassandra.input.split.size和spark.cassandra.page.row.size，分别表示一个线程最多读取多少记录，另一个表示每次读取多少行。

partitions的源码详见CasssandraRDDParitioner.scala

compute函数就利用确定的token的起止范围来加载内容，这里在理解的时候需要引起注意的就是flatMap是惰性执行的，也就是说只有在真正需要值的时候才会被执行，延迟触发。

数据真正的加载是发生在fetchTokenRange函数，这时使用到的就是Cassandra Java Driver了，平淡无奇。

### fetchTokenRange 
fetcchTokenRange函数使用Cassandra Java Driver提供的API接口来读取数据，利用Java API读取数据一般遵循以下步骤：

val cluster = ClusterBuilder.addContactPoint(“xx.xx.xx.xx”).build
val session = cluster.connect
val stmt = new SimpleStatement(queryCQL)
session.execute(session)
session.close
cluster.close

addContactPoint的参数是cassandra server的ip地址，在后面真正执行cql语句的时候，如果集群有多个节点构成，那么不同的cql就会在不同的节点上执行，自动实现了负载均衡。
可以在addContactPoint的参数中设定多个节点的地址，这样可以防止某一节点挂掉，无法获取集群信息的情况发生。

session是线程安全的，在不同的线程使用同一个session是没有问题的，建议针对一个keySpace只使用一个session。

### RDD中使用Session

在Spark RDD中是无法使用SparkContext的，否则会形成RDD嵌套的现象，因为利用SparkContext很容易构造出RDD，如果在RDD的函数中如map中调用SparkContext创建一个新的RDD，
则形成深度嵌套进而导致Spark Job有嵌套。

但在实际的情况下，我们需要根据RDD中的值再去对数据库进行操作，那么有什么办法来打开数据库连接呢？

解决的办法就是直接使用Cassandra Java Driver而不再使用spark-cassandra-connector的高级封装，因为不能像这样子来使用cassandraRDD。

sc.cassandraRDD(“ks”,”tableX”)
.map(x=>sc.cassandraRDD(“ks”,”tableX”).where(filter))
如果是直接使用Cassandra Java Driver，为了避免每个RDD中的iterator都需要打开一个session，那么可以使用foreachPartition函数来进行操作，减少打开的session数。
val  rdd1 = sc.cassandraTable(“keyspace”,”tableX”)
	rdd1.foreachPartition( lst => {
		val cluster = ClusterBuilder.addContactPoint(“xx.xx.xx.xx”).build
		val session = cluster.connect
		while ( iter.hasNext ) {
		 	val  elem = iter.next
			//do something by using session and elem
		}
		session.close
		cluster.close
	})

其实最好的办法是在外面建立一个session，然后在不同的partition中使用同一个session，但这种方法不行的原因是在执行的时候会需要”Task not Serializable”的错误，于是只有在foreachPartition函数内部新建session。

## 数据备份

尽管Cassandra号称可以做到宕机时间为零，但为了谨慎起见，还是需要对数据进行备份。

Cassandra提供了几种备份的方法

1. 将数据导出成为json格式
2. 利用copy将数据导出为csv格式
3. 直接复制sstable文件

导出成为json或csv格式，当表中的记录非常多的时候，这显然不是一个好的选择。于是就只剩下备份sstable文件了。

问题是将sstable存储到哪里呢？放到HDFS当然没有问题，那有没有可能对放到HDFS上的sstable直接进行读取呢，在没有经过任务修改的情况下，这是不行的。

试想一下，sstable的文件会被拆分为多个块而存储到HDFS中，这样会破坏记录的完整性，HDFS在存储的时候并不知道某一block中包含有完成的记录信息。

为了做到记录信息不会被拆分到多个block中，需要根据sstable的格式自行提取信息，并将其存储到HDFS上。这样存储之后的文件就可以被并行访问。

Cassandra中提供了工具**sstablesplit**来将大的sstable分割成为小的文件。

DataStax的DSE企业版中提供了和Hadoop及Spark的紧密结合，其一个很大的基础就是先将sstable的内容存储到CFS中，大体的思路与刚才提及的应该差不多。

对sstable存储结构的分析是一个研究的热门，可以参考如下的链接。

https://www.fullcontact.com/blog/cassandra-sstables-offline/

之所以要研究备份策略是想将对数据的分析部分与业务部分相分离开，避免由于后台的数据分析导致Cassandra集群响应变得缓慢而致前台业务不可用，即将OLTP和OLAP的数据源分离开。

通过近乎实时的数据备份，后台OLAP就可以使用Spark来对数据进行分析和处理。

## 高级查询 Cassandra+Solr

与传统的RDBMS相比，Cassandra所能提供的查询功能实在是弱的可以，如果想到实现非常复杂的查询功能的，需要将Cassandra和Solr进行结合。

DSE企业版提供了该功能，如果想手工搭建的话，可以参考下面的链接：

http://www.slideshare.net/planetcassandra/an-introduction-to-distributed-search-with-cassandra-and-solr 
https://github.com/Stratio/stratio-cassandra开源方面的尝试 Cassandra和Lucene的结合

## 共享SparkContext

SparkContext可以被多个线程使用，这意味着同个Spark Application中的Job可以同时提交到Spark Cluster中，减少了整体的等待时间。

在同一个线程中， Spark只能逐个提交Job，当Job在执行的时候，Driver Application中的提交线程是处于等待状态的。如果Job A没有执行完，Job B就无法提交到集群，就更不要提分配资源真正执行了。

那么如何来减少等待时间呢，比如在读取Cassandra数据的过程中，需要从两个不同的表中读取数据，一种办法就是先读取完成表A与读取表B，总的耗时是两者之和。

如果利用共享SparkContext的技术，在不同的线程中去读取，则耗时只是两者之间的最大值。

在Scala中有多种不同的方式来实现多线程，现仅以Future为例来说明问题：

val ll  = (1 to 3 toList).map(x=>sc.makeRDD(1 to 100000 toList, 3))
val futures = ll.map ( x => Future {
		x.count()
	})
val fl = Future.sequencce(futures)
Await.result(fl,3600 seconds)

简要说明一下代码逻辑
创建三个不同的RDD
在不同的线程(Future)中通过count函数来提交Job
使用Await来等待Future执行结束