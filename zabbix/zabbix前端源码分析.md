# zabbix前端源码分析
## 目录和代码结构
zabbix 3.0 前端引入了MVC架构，将存储模型、展示视图和代码逻辑分离，代码更清晰、模块化；但是目前还没有完成，老架构和MVC架构共存；

### 通用文件
/index.php: zabbix前端的入口文件，login、logout时调用，如果有sessionid且检查通过，则直接条状到/zabbix.php?action=dashboadr.view;
/conf: 配置文件目录，/conf/zabbix.conf.php是前端的配置文件；
/include/classes/core/CConfigFile.php: 加载/conf/zabbix.conf.php配置文件；
/setup.php: 前端配置程序：检查php版本和参数，配置数据库连接，然后生成
/conf/zabbix.conf.php文件，注意是在ZBase::EXEC_MODE_SETUP模式下运行；
/conf/api\_jsonrpc.php: zabbix API入口文件，在ZBase::EXEC_MODE_API模式下运行；
/include/*.php: 按功能划分的函数文件, 初始化时被自动加载;
/include/classes: 按功能划分的类定义文件，重要的目录如api、core、mvc、html, 所有类均以C开头，如CAutoloader.php，注意C和A均大写；
/include/\*.php和/include/classes下的所有文件，在初始化时(/include/classess/core/Z/ZBase.php中)导入, 供前端代码使用；
/include/schema.inc.php: 定义数据库表结构, key字段为表的primay key，一般为xxxid；


### 老架构
/php/\*.php: 老架构的前端入口文件，分别和前端Menu对应，如点击Monitoring->Overview时，调用/php/overview.php文件；这些文件校验表单参数、生成/include/views下对应view文件所需的数据结构，生成浏览器页面；
/include/views: 老架构的view文件；使用/php/*.php生成的data生成浏览器页面；

### MVC架构
/php/zabbix.php: 新的MVC架构的入口文件，通过action参数来指定view文件，如点击Monitoring->Dashboard时action=dashboard.view，点击Monitor->Web时action=web.view；

/app: app目录下存放MVC架构的controllers和views文件；
    + view: 根据操作类型的不通，同一个页面分为edit、list和对应的js文件，如: administration.proxy.edit.php/administration.proxy.list.php/administration.proxy.edit.js.php
    + controller: 继承自抽象类CController(/include/class/mvc/CController.php);

## 请求处理路径
老架构前端页面/php/\*.php和新MVC架构/php/zabbix.php都会引入/include/config.inc.php文件，该文件是前端界面的关键文件：
    1. 调用/include/classes/core/Z.php文件，Z.php调用/include/classes/core/ZBase.php完成初始化；
    2. 调用Z::getInstance()->run(ZBase::EXEC\_MODE_DEFAULT)生成MVC架构的浏览器界面;
    3. 如果请求中有action参数，则表明是MVC架构，ZBase.php的processRequest()会生成页面内容；否则是老架构，/php/*.php代码负责生成浏览器界面;

### /inlcude/classess/core/ZBase.php文件分析
1. 导入/include/classes/core/CAutoloader.php文件，CAutoloader的作用是告诉php如何根据class name找到找到对应class file, 相应的关键代码是spl_autoload_register([$this, 'loadClass']);.
2. init()方法：
   1. 调用CAutoloader的register方法将/include/classes目录及其子目录、/local/app/controllers、/app/controllers目录导入到php的spl逻辑，这样php在使用相应类时，自动在注册的目录中查找相关类的定义文件；
   2. 初始化API实例, API实例会在/php/\*.php 和 controller文件中使用, 先关的代码如下:
   3. 导入/include/*.inc.php文件，这些文件定义了系统和页面相关的通用函数；
3. run()方法是ZBase.php文件的入口，在/include/config.inc.php中被调用, 传入ZBase::EXEC\_MODE_DEFAULT参数, 这种类型执行的流程如下:
   1. 加载配置文件(/include/classes/core/CConfigFile.php), 位置为/conf/zabbix.conf.php；
   2. 初始化DB, 调用/include/db.inc.php的DBconnect()函数， 生成全局的database connection，保存在全局的$DB["DB"]变量中，同时检查DbVersion和Config表是否满足:
   3. 认证用户:
      1. 调用CWebUser::checkAuthentication(CWebUser::getSessionCookie())获取$sessionId;
      2. 使用$sessionId更新API:getWrapper()->auth字段;
      3. 设置API的debug模式；
   4. 初始化locale;
   5. 判断请求URL是否有action参数，如果有则按照MVC类型处理，否则不做任何处理(按老架构方式处理, 即导入/include/config.inc.php的/*.php文件负责处理);
      1. 获取$router: $router = new CRouter(getRequest('action'));
      2. 获取$router对应的Controller,如果不为null，调用$this.processRequest()方法；
      3. processRequest(): 根据请求的action参数，生成$router、$controller对象，然后调用$controller->run()方法，获得$response, 其类型可能是：
         1. CControllerResponseData：
            1. 如果action没有关联的view，则直接将$respnose->getData()结果传给$layout, 然后输出CView的getOutput()的结果；
            2. 否则：
                1. 初始化$view = new CView($router->getView(), $response->getData())；
                2. 从$response中提取action、page title、main_block、files/pre/post类型的JavaScript，填充$data变量；
                3. 用action关联的layout，第二步生成的$data，生成$layout对象：$layout = new CView($router->getLayout(), $data);
                4. echo $layout->getOutput()的结果；
            3. 不论如何，最终都是由/app/view/下的layout文件生成生成页面显示结果；
        2. 其它类型的错误；

### /app/views/layout.*.php 文件分析
1. layout.htmlpage.footer.php：显示相关messages、如果登陆用户处于debug模式，则停止CProfiler，并显示CProfiler的结果；
2. layout.htmlpage.header.php
3. layout.htmlpage.menu.php
4. layout.htmlpage.php
5. layout.javascript.php、layout.json.php、layout.warning.php、layout.widget.php: 分别为相应页面类型的布局类；

### /include/classes/mvc/*.php 文件分析
1. CRouter是MVC的核心，用于获取Controller、View；
   1. 内部封装了$layout、$controller、$view、$action对象；
   2. 将action参数与相应的controller、layout、view关联起来，如：
      		// action					controller							layout					view
		'acknowledge.create'	=> ['CControllerAcknowledgeCreate',		null,					null],
		'acknowledge.edit'		=> ['CControllerAcknowledgeEdit',		'layout.htmlpage',		'monitoring.acknowledge.edit'],
		'dashboard.favourite'	=> ['CControllerDashboardFavourite',	'layout.javascript',	null],
   3. controller为/app/controllers目录下的各类文件；
   4. view和layout为/app/views目录下的各类文件；layout为特殊类型的view；
   5. CRouter根据action参数和第二步的映射表，来获取action对应的controller、layout和view；
2. CController: 抽象类, 被/app/controllers目录下的各文件实现, 用户获取和校验用户的输入，然后生成view或layout需要的$data:
   1. validateInput(): 对输入进行校验；
   2. checkSID(): SID为sessionId的最开始16字节，该方法check请求头部的sid参数是否与sessionId提取的一致；
   3. doAction(): 执行action，生成响应对象；
   4. getResponse(): 返回响应对象；
   5. run(): controller主方法，返回响应对象: data、redirect或者fatal redirect;
   6. CController对应与CView是无关的；
3. CView: 根据传入的$data或者set(), get()传入的参数，生成views:
   1. $filePath: view文件的路径；
   1. $template: include($filePath)后返回的结果；
   2. $scripts: 页面上的scripts;
   3. $jsIncludePost、$jsIncludeFiles、$jsFiles: 页面包含的各类型js code；
   4. CView依次在['local/app/views', 'app/views', 'include/views']目录下查找view文件；
   5. render():
      1. 设置view file用到的$data变量；
      2. include($this->filePath), 结果赋值给$this->template，期间的输出js代码，赋值给$this->scripts并显示；
   6. show(): 必须在reder()之后调用，执行$this->template->show()方法，即调用view类对象的show()方法；
   7. getOutput(): render和show方法逐渐被淘汰，query代之的是本方法；调用include($this->filePath), 返回 view template的HTML/JSON/CVS/etc等文本输出；

### /include/classes/user/CWebUser.php 文件分析
1. setDefault(): 将User对象的$data属性设置为guest用户；
2. setSessionCookie($sessionId): 调用zbx\_setcookie()将$sessionId设置到cookie中，如果$autoLogin为true，则有效期1个月，否则为会话期间有效；
2. login($login, $password):
   1. 调用API:User()->login()登陆，结果赋值给$data属性; 如果认证失败或不允许GUI access，则抛出异常；userData=True, 结果$data中将含有users表中的各字段信息；
   2. 检查$data['attempt\_failed']是否为空，否则输出历史登陆失败的信息；
   3. 调用setSessionCookie(self::$data['sessionid']) 设置session cookie；
3. logout(): 调用getSessionCookie()函数获得sessionid，然后调用API:User()->logout()退出，CSession::destroy()删除session记录，zbx_unsetcookie()删除zbx_sessionid cookie；
4. checkAuthentication($sessionId):
   1. 调用API::User()->checkAuthentication([$sessionId])(/include/classes/api/services/CUser.php)检查$sessionId是否正确；
   2. 如果失败，则设置$sessionId为guest用户的sessionid，否则为传入的sesionId;
   3. 调用setSessionCookie($sessionId)函数，设置session和cookie；

### /include/classes/api分析：
1. /include/classes/api/API.php：
   1. API的入口，定义了各个API对应的方法，各方法返回CApiWrapper或CApiService对象，然后调用相应对象的方法；
   2. CApiWrapper对象实际为CFrontedApiWrapper;
   3. CApiService对象实际为/include/classes/api/services/\*.php中定义的对象；
2. CApiService:
   1. 所有services/\*.php API Class的基类，
   2. 定义了最重要的select()方法, 以及services使用的其它通用方法；
   3. services/*.php的各类，实现了对应API的方法， 如CUser实现了get、create、update等方法；
3. CApiWrapper: 将CApiClient封装到内部$client属性，同时保存api name和auth token；调用它的方法时，实际调用的是client->callMethod($this->api, $method, $params, $this->auth)方法；
   1. CApiServiceFactory: 将API的endpoint name与/include/classes/api/serivices目录下的所有CApiService子类关联起来；
   2. CApiClient：重新定义了__call()方法，内部调用抽象方法callMethod()；这样当调用CApiClient对象的method时，实际调用的是callMethod方法；
   3. CLocalApiClient继承自CApiClient, 实现了callMethod方法;
      1. 对API的name、method、auth进行校验；
      2. 使用$serviceFactory->getObject($api)，获取实际的CApiService对象，然后调用它的$method，获得结果；
  4. 综上，CApiWrapper对API的参数进行校验，然后通过CApiServiceFactory的getObject方法获取CApiService对象，进而调用它的API方法；
4. 初始化(/include/classes/core/ZBaes.php->init())：
		// initialize API classes
		$apiServiceFactory = new CApiServiceFactory();

		$client = new CLocalApiClient();
		$client->setServiceFactory($apiServiceFactory);
		$wrapper = new CFrontendApiWrapper($client);
		$wrapper->setProfiler(CProfiler::getInstance());
		API::setWrapper($wrapper);
		API::setApiServiceFactory($apiServiceFactory);

### /include/classes/api/API.php 文件分析
1. 内部封装了$wrapper、$apiServiceFactory对象，前者是/include/classes/api/wrappers/CFrontedApiWrapper对象的实例，后者是CApiServiceFactory对象的实例；
2. getApi($name): 返回一个可以用于API调用的对象，如果$wrapper非空，则返回该对象(CApiWrapper)，否则调用CApiService方法；
3. getApiService($name): 调用$apiServiceFactory->getObject($name ? $name : "api")方法，返回CApiServiceFactory对象；

### /include/classes/api/CApiService.php 文件分析:
1. pk($tableName): 返回$tableName的Primary Key.
2. createRelationMap(array $objects, $baseField, $foreignField, $table = null): 为$objcts各object创建$baseField和$foreignField，如MapMany、MapOne;
3. select($tableName, array $options):
   1. 调用this->createSelectQuery()创建SQL SELECT query，返回从数据库查询的结果；
   2. 该方法是CApiService类的入口方法，它调用或间接调用了该类的其它方法；
4. createSelectQuery($tableName, array $options): 创建数据库查询语句；
5. createSelectQueryParts($tableName, $tableAlias, array $options):
   1. 将$options与$globalGetOptions 合并；
   2. 创建Select SQL Parts语句数组；
   3. 添加filter、output、sort options；
   4. 返回生成的$sqlParts;
6. createSelectQueryFromParts(array $sqlParts): 利用$sqlParts创建具体的SQL SELECT query语句；

### /include/classes/api/CApiServiceFactory.php 文件分析
1. CApiServiceFactory extends CRegistryFactory， 故具有继承的 getObject、hasObject方法；
2. 向CRegistryFactory注册API Name和 API Class的对应关系，如： 'api'->'CApiService', 'action'->'CAction'; 这样可以使用api endpoint自动调用响应的类定义；
### /include/classes/core/CRegistryFactory.php 文件分析：
1. CRegistryFactory 类用于存储全局对象的实例；
2. getObject($object): $object为对象名称，如果不存在其实例，则创建一个对象实例；
3. hasObject($object): 判断$object对象的定义(不是其实例)是否存在；

### /include/classes/api/wrappers/*.php 文件分析：
1. CApiWrapper.php: 封装了api name、auth token，然后调用CApiClient对象的API services；
2. CFrontedApiWrapper.php: 继承自CApiWrapper类，用于前端调用API时使用， 主要是封装了CProfiler对象；

### /include/classes/api/clients/*.php 文件分析
1. CApiClient.php: 定义API Service使用的Client对象，定义了抽象方法callMethod($api, $method, array $params, $auth);
2. CLocalApiClient.php: CLocalApiClient extends CapiClient类，是实际调用 API Service的对象：
   1. 内部封装了$serviceFactory属性；
   2. callMethod($requestApi, $requestMethod, array $params, $auth):
      1. 实例化CApiClientResponse对象，作为$response;
      2. 对API的name、method进行验证；
      3. 调用authenticate($auth)进行$auth验证；
      4. 调用API method， 注意使用了$serviceFactory的getObject方法：
         call_user_func_array([$this->serviceFactory->getObject($api), $method], [$params]);
  3. authenticate($auth): 调用user API的checkAuthentication方法，对$auth进行验证；

### /include/classes/user/CProfile.php 文件分析
user关联的value cache，主要是提升性能考虑；

### /include/classes/api/services/CUser.php
1. ldapLoin(array $user): 使用LDAP认证$user，看是否能够登陆，使用php-ldap module;
2. dbLogin($user): 检查$user的用户名和密码是否正确， 对passowrd做md5校验，然后检查数据库是否正确；
3. get($options): 获取用户数据，实际上是利用$options构造$sqlParts；
4. checkInput(&$users, $method): 对用户$user和$method进行检查, $method为update或create；
5. create(): 创建用户；
6. login($user): $user包含两个key: user、明文password：
   1. zabbix对password进行md5计算，然后与数据库记录比较，看是否正确；
   2. 如果user的attempt_failed数目大于阈值，且time()-attempt_clock小于ZBX_LOGIN_BLOCK则，阻止登陆，同时用当前时间更新attempt_clock；
   3. 检查$authType值，进行ZBX_AUTH_HTTP、ZBX_AUTH_LDAP、ZBX_AUTH_INTERNAL、ZBX_AUTH_HTTP认证；如果出错，则获取REMOTE_ADDR值，更新user的attempt_*值；
   4. 使用$sessionid = md5(time().$password.$name.rand(0, 10000000));创建sessionid，然后向sessions表中为userid创建一条status为ZBX_SESSION_ACTIVE的记录；
   5. 更新全局变量: CWebUser::$data、CUser:$userData;
7. logout(): 不需要传入参数；从session中获取sessionId，然后更新sessions表中的userid记录，将status设置为ZBX_SESSION_PASSIVE(登陆状态为ZBX_SESSION_ACTIVE)；
8. checkAuthentication(array $sessionid):
   1. 如果用户已经login且CUser::$userData非空，返回$userData；
   2. 用$sesisonid、ZBX_SESSION_ACTIVE、s.lastaccess+u.autologout>time()或者autologout=0条件查询sessions和users表，如果结果为空，则报：'Session terminated, re-login, please.'；
      1. autologout: 0表示不过期；其它时间表示session过期时间；
   3. 如果当前时间不等于 lastaccess(login的时候设置)，则用time()更新lastaccess;
   4. 更新全局变量：CWebUser::$data、CUser:$userData;


### API分析
### /include/db.inc.php文件分析
1. 该文件定义了操作数据库的高层函数，屏蔽各数据库的差异：
   1. DBconnect(): 创建全局的database connection；
       1. DBconnect()使用第一步加载的配置，调用php mysql/pg等扩展提供的mysqli\_connect等函数建立连接；
       2. 检查第一步判断的DbBackend(/include/classes/db/*DbBackend.php)的Version和Config()方法, 如MysqlDbBackend:
          1. MysqlDbBackend extends DbBackend;
          2. DbBackend:
            1. checkDbVersionTable(): 检查dbversion是否存在, 如果不存在则提示: "The frontend does not match Zabbix database."
            2. 检查前端定义的ZABBIX_DB_VERSION变量是否满足dbversion表中的mandatory值，如果不满足，就提示：
               "The frontend does not match Zabbix database. Current database version (mandatory/optional): %d/%d. Required mandatory version:%d. Contact your system administrator"
            3. checkConfig()检查configuration表是否存在；
  2. DBclose(): 关闭全局数据库连接；
  3. DBStart(): 开始数据库事务，全局变量DB["TRANSACTIONS"]记录当前事务的数目，如果不能与0，则拒绝执行，防止事务嵌套；
  4. DBend()、 DBcommit()、DBrollback()：结束、提交和回滚最近一次的数据库事务；
  5. DBselect($query, $limit = null, $offset = 0): 执行select语句， 返回cursor对象；
  6. DBaddLimit($query, $limit = null, $offset = 0): 将$query后面添加LIMIT子句；
  7. DBexecute($query, $skip_error_message = 0)： 执行$query语句；
  8. DBfetch($cursor, $convertNulls = true): 返回$cursor包含的下一批数据集，如果没有则返回false；$cursor一般是DBselect()的返回结果；
  9. get_dbid($table, $field): 查找ids表，返回$table表$filed下一个可用id，同时将对应的nextid加1，共后续使用；
  10. zbx_db_search($table, $options, &$sql\_parts): 根据$options的$excludeSearch、$search、$searchByAny、$searchWildcardsEnabled、$startSearch等参数，更新$sql_parts['search']内容；$options为用户调用API接口时指定的参数；$search是一个object，key对应的value为array类型，可以为key指定多个patterns；
  11. check_db_fields($dbFields, &$args): 检查$dbFields的key是否都在$args中，如果不在，就在$args中创建一个；
  12. dbConditionInt($fieldName, array $values, $notIn = false, $sort = true): 创建一个Where condition， 其中filedName是整型值，IN或者notIn $values中；
  13. dbConditionString($fieldName, array $values, $notIn = false): 创建一个Where condition，其中fileName是字符串，IN或者notIn $values中；
  14. DBfetchArray($crusor): 持续调用DBfetch($cursor)， 结果作为数组返回；
  15. DBfetchArrayAssoc($cursor, $field): 持续调用DBfetch($cursor)，返回一个关联数组，key为每次提取$field的值；
  16. DBfetchColumn($cursor, $column, $asHash = false): 持续调用DBfetch($cursor), 每项取$column列对应的值， 返回一个数组或关联数组(取决于$asHash);

### /include/class/db/DB.php 文件分析
1. getDbBackend(): 根据$DB['TYPE']的配置，选择相应的backend(/include/classes/db/);
2. getSchema($table): 从/include/schema.inc.php文件获取$table表的定义(不是从数据库中获取)，其中主键保存在$tableSchema['key']中；
3. refreshIds($table, $count): 更新ids表中$table表的id记录($table表的primay key)， 先删除ids表中$table标的id记录，然后用$table表中最大的id+$count来更新ids表。
4. reserveIds($table, $count): 为$table保留$count个ids, 内部调用refreshIds()函数;
5. addMissingFields($tableSchema, $values): 遍历$tableSchema的field name，如果不在$values中，则赋值为空字符串;
6. getDefaults($table): 返回$table有'default'字段的key和它的default值dict；
7. checkValueType($table, &$values): 检查$values中key对应的field name的value是否符合$table的定义:
   1. 如果field name是CHAR类型，检查values是否大于定义的length，如果是，则报错:
   'Value "%1$s" is too long for field "%2$s" - %3$d characters. Allowed length is %4$d characters.'
   2. 对int、uint、id、float数据类型进行检查；
8. find($tableName, array $criteria = []): 利用$criteria列表中的key-value，构造'select * from '.$tableName 的where查询条件(AND), 然后执行DBfetchArray(DBSelect($sql));
9. inert($table, $values, $getids = true): 将$values每次一条地插入到$table中，插入前使用reserveIds($table, count($values))来更新ids表；如果$getids为true，这返回插入的ids；
10. insertBatch($table, $values, $getids = true): 与insert函数类似，但是使用 INSERT INTO xxx VALUES(),(),...的方式，一次插入所有值；
11. update($table, $data): $data[...]['values']、$data[...]['where'], where字段是必须的，一次一行地update；
12. updateByPk($tableName, $pk, array $values): 使用$pk值来构造where条件，然后更新;
13. save($tableName, array $data): 如果$data中的记录有Primary key，则更新表记录，否则插入新的表记录，返回插入后的field及其pk值；
14. replace($tableName, array $oldRecords, array $newRecords): 使用$newRecords更新$oldRecords:
    1. 对于有相同PK的记录，使用new更新old；
    2. $newRecords中新的记录，则save到数据库，返回生成的PK列表；
    3. $oldRecords中剩余的, 不在$newRecords中的记录将被删除;
15. replaceByPosition($tableName, array $groupedOldRecords, array $groupedNewRecords): 使用newRecords更新oldRecords:
    1. 对于newRecords中的每一项，如果old中存在则update，否则新建；
    2. 对于newRecords没有的项，将其从old中删除；
16. recordModified($tableName, array $oldRecord, array $newRecord): 比较old中的记录是否都在new中，返回true或false；
17. mergeRecords(array $oldRecords, array $newRecords, $pk): 使用newRecords更新oldRecords;
18. delete($table, $wheres, $use_or = false): 删除$table中，符合$wheres的记录；