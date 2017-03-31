当使用一个API时，其中一个挑战就是认证（authentication）。在传统的web应用中，服务端成功的返回一个响应（response）依赖于两件事:
1. 他通过一种存储机制保存了会话信息（Session）。每一个会话都有它独特的信息（id），常常是一个长的，随机化的字符串，它被用来让未来的请求（Request）检索信息。
2. 包含在响应头（Header）里面的信息使客户端保存了一个Cookie。服务器自动的在每个子请求里面加上了会话ID，这使得服务器可以通过检索Session中的信息来辨别用户。
这就是传统的web应用逃避HTTP面向无连接的方法（This is how traditional web applications get around the fact that HTTP is stateless）。

在使用中，并不会每次都让用户提交用户名和密码，通常的情况是客户端通过一些可靠信息和服务器交换取token，这个token作为客服端再次请求的权限钥匙。Token通常比密码更加长而且复杂。
比如说，JWTs通常会长达150个字符。一旦获得了token，在每次调用API的时候都要附加上它。然后，这仍然比直接发送账户和密码更加安全，哪怕是HTTPS。
把token想象成一个安全的护照。你在一个安全的前台验证你的身份（通过你的用户名和密码），如果你成功验证了自己，你就可以取得这个。当你走进大楼的时候（试图从调用API获取资源），你会被要求
验证你的护照，而不是在前台重新验证。

API应该被设计成无状态的（Stateless）。这意味着没有登陆，注销的方法，也没有sessions，API的设计者同样也不能依赖Cookie，因为不能保证这些request是由浏览器所发出的。
自然，我们需要一个新的机制。

不依赖cookies和sessions，使用jwt-go生成的token来认证(下面假设都启用了HTTPS)：
1. client注册时提供usernmae和password，服务器使用bcrypt将密码加密，保存到数据库；
2. client登陆时，提供username和password，服务端使用bcrypt对提供的密码进行校验，如果符合，则生成一个time-limited的token；
3. client将token保存并在后续的请求中附带该token；
4. 在服务端为需要登陆认证的router添加token检查中间件；JWT token有过期(exp)和截止(nbf)时间戳，JWT解析Header时对它们进行检查；检查通过后才继续后续的处理(一般是调用传入的handler)。
5. 用于logout时，它的token可能还有效，服务端需要将该token加入失效名单中(token在失效名单中的有效期为token剩余的有效期)，后续校验通过的token时，再检查是否在该名单中，如果再则拒绝；
一般使用redis来实现；
5. 客户端周期刷新token(通过重新请求生成token来实现，旧的token会自动因过期失效)，例如token过期时间是5min，则每4分钟刷新一次，用户需要使用旧token来刷新新token；

对于client，token可以保存到cookies、浏览器local storage或者app的内存：
1. 如果保存到cookies中，则后续的每次请求会自动发送；
2. 如果是localstorage或app内存，则在需要验证身份的场合(如RESTFulAPI)，使用JavaScript将token添加到请求Header、添加URL的query参数或body的表单中(依赖于服务端的校验方式)，如：
    var token = window.localStorage.getItem('token');
    if (token) {
    $.ajaxSetup({
        headers: {
        'x-access-token': token
        }
    });
    }

使用jwt的服务端代码示例：
http://blog.brainattica.com/restful-json-api-jwt-go/
https://github.com/brainattica/golang-jwt-authentication-api-sample

注意：下面示例代码是v2版本的jwt-go，现在最新的是v3版本，部分API发生了变化如：
1. ParseFromRequest and its new companion ParseFromRequestWithClaims被删除，被subpackage request里面的API所代替；

#返回给client的，包含token的JSON格式
golang-jwt-authentication-api-sample/api/parameters/auth.go
type TokenAuthentication struct {
	Token string `json:"token" form:"token"`
}

https://github.com/brainattica/golang-jwt-authentication-api-sample/blob/master/core/authentication/jwt_backend.go
type JWTAuthenticationBackend struct {
	privateKey *rsa.PrivateKey
	PublicKey  *rsa.PublicKey
}

const (
	tokenDuration = 72
	expireOffset  = 3600
)

var authBackendInstance *JWTAuthenticationBackend = nil

func InitJWTAuthenticationBackend() *JWTAuthenticationBackend {
	if authBackendInstance == nil {
		authBackendInstance = &JWTAuthenticationBackend{
			privateKey: getPrivateKey(),
			PublicKey:  getPublicKey(),
		}
	}

	return authBackendInstance
}

func (backend *JWTAuthenticationBackend) GenerateToken(userUUID string) (string, error) {
	token := jwt.New(jwt.SigningMethodRS512)
	token.Claims["exp"] = time.Now().Add(time.Hour * time.Duration(settings.Get().JWTExpirationDelta)).Unix()
	token.Claims["iat"] = time.Now().Unix()
	token.Claims["sub"] = userUUID
	tokenString, err := token.SignedString(backend.privateKey)
	if err != nil {
		panic(err)
		return "", err
	}
	return tokenString, nil
}

//验证用户名、密码是否正确；
func (backend *JWTAuthenticationBackend) Authenticate(user *models.User) bool {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("testing"), 10)

	testUser := models.User{
		UUID:     uuid.New(),
		Username: "haku",
		Password: string(hashedPassword),
	}

	return user.Username == testUser.Username && bcrypt.CompareHashAndPassword([]byte(testUser.Password), []byte(user.Password)) == nil
}

//计算token剩余的有效时间
func (backend *JWTAuthenticationBackend) getTokenRemainingValidity(timestamp interface{}) int {
	if validity, ok := timestamp.(float64); ok {
		tm := time.Unix(int64(validity), 0)
		remainer := tm.Sub(time.Now())
		if remainer > 0 {
			return int(remainer.Seconds() + expireOffset)
		}
	}
	return expireOffset
}

//将logout的token加到redis黑名单中，过期时间为计算的有效时间
func (backend *JWTAuthenticationBackend) Logout(tokenString string, token *jwt.Token) error {
	redisConn := redis.Connect()
	return redisConn.SetValue(tokenString, tokenString, backend.getTokenRemainingValidity(token.Claims["exp"]))
}

//检查token是否在黑名单中
func (backend *JWTAuthenticationBackend) IsInBlacklist(token string) bool {
	redisConn := redis.Connect()
	redisToken, _ := redisConn.GetValue(token)

	if redisToken == nil {
		return false
	}

	return true
}

func getPrivateKey() *rsa.PrivateKey {
	privateKeyFile, err := os.Open(settings.Get().PrivateKeyPath)
	if err != nil {
		panic(err)
	}

	pemfileinfo, _ := privateKeyFile.Stat()
	var size int64 = pemfileinfo.Size()
	pembytes := make([]byte, size)

	buffer := bufio.NewReader(privateKeyFile)
	_, err = buffer.Read(pembytes)

	data, _ := pem.Decode([]byte(pembytes))

	privateKeyFile.Close()

	privateKeyImported, err := x509.ParsePKCS1PrivateKey(data.Bytes)

	if err != nil {
		panic(err)
	}

	return privateKeyImported
}

func getPublicKey() *rsa.PublicKey {
	publicKeyFile, err := os.Open(settings.Get().PublicKeyPath)
	if err != nil {
		panic(err)
	}

	pemfileinfo, _ := publicKeyFile.Stat()
	var size int64 = pemfileinfo.Size()
	pembytes := make([]byte, size)

	buffer := bufio.NewReader(publicKeyFile)
	_, err = buffer.Read(pembytes)

	data, _ := pem.Decode([]byte(pembytes))

	publicKeyFile.Close()

	publicKeyImported, err := x509.ParsePKIXPublicKey(data.Bytes)

	if err != nil {
		panic(err)
	}

	rsaPub, ok := publicKeyImported.(*rsa.PublicKey)

	if !ok {
		panic(err)
	}

	return rsaPub
}

# 验证token的中间件，该中间件将放在需要验证的路由前面
https://github.com/brainattica/golang-jwt-authentication-api-sample/blob/master/core/authentication/middlewares.go
func RequireTokenAuthentication(rw http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	authBackend := InitJWTAuthenticationBackend()

    //从请求中获取解析后的token；
	token, err := jwt.ParseFromRequest(req, func(token *jwt.Token) (interface{}, error) {
        //检查签名方法是否正确
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		} else {
			return authBackend.PublicKey, nil
		}
	})
    //检查token是否valid，是否在redis logout黑名单中，如果不在，则继续下一个处理；这一步实际上还可以将token对应的信息加到请求的context中，供后续使用；
	if err == nil && token.Valid && !authBackend.IsInBlacklist(req.Header.Get("Authorization")) {
		next(rw, req)
	} else {
		rw.WriteHeader(http.StatusUnauthorized)
	}
}

#controller，调用相应的service
https://github.com/brainattica/golang-jwt-authentication-api-sample/blob/master/controllers/auth_controller.go
func Login(w http.ResponseWriter, r *http.Request) {
	requestUser := new(models.User)
	decoder := json.NewDecoder(r.Body)
	decoder.Decode(&requestUser)

	responseStatus, token := services.Login(requestUser)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(responseStatus)
	w.Write(token)
}

func RefreshToken(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	requestUser := new(models.User)
	decoder := json.NewDecoder(r.Body)
	decoder.Decode(&requestUser)

	w.Header().Set("Content-Type", "application/json")
	w.Write(services.RefreshToken(requestUser))
}

func Logout(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	err := services.Logout(r)
	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
	}
}

func HelloController(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	w.Write([]byte("Hello, World!"))
}

#服务端router
https://github.com/brainattica/golang-jwt-authentication-api-sample/blob/master/routers/hello.go
func SetHelloRoutes(router *mux.Router) *mux.Router {
	router.Handle("/test/hello",
		negroni.New(
			negroni.HandlerFunc(authentication.RequireTokenAuthentication), //引入认证中间件；
			negroni.HandlerFunc(controllers.HelloController),
		)).Methods("GET")

	return router
}

func SetAuthenticationRoutes(router *mux.Router) *mux.Router {
	router.HandleFunc("/token-auth", controllers.Login).Methods("POST")  //用户登陆
	router.Handle("/refresh-token-auth",
		negroni.New(
			negroni.HandlerFunc(authentication.RequireTokenAuthentication), //刷新token，需要验证token
			negroni.HandlerFunc(controllers.RefreshToken),
		)).Methods("GET")
	router.Handle("/logout",
		negroni.New(
			negroni.HandlerFunc(authentication.RequireTokenAuthentication), //用户登出，需要验证token
			negroni.HandlerFunc(controllers.Logout),
		)).Methods("GET")
	return router
}    

func InitRoutes() *mux.Router {
	router := mux.NewRouter()
	router = SetHelloRoutes(router)
	router = SetAuthenticationRoutes(router)
	return router
}

#各种services
https://github.com/brainattica/golang-jwt-authentication-api-sample/blob/master/services/auth_service.go
func Login(requestUser *models.User) (int, []byte) {
	authBackend := authentication.InitJWTAuthenticationBackend()

	if authBackend.Authenticate(requestUser) {
		token, err := authBackend.GenerateToken(requestUser.UUID)
		if err != nil {
			return http.StatusInternalServerError, []byte("")
		} else {
			response, _ := json.Marshal(parameters.TokenAuthentication{token})
			return http.StatusOK, response
		}
	}

	return http.StatusUnauthorized, []byte("")
}

func RefreshToken(requestUser *models.User) []byte {
	authBackend := authentication.InitJWTAuthenticationBackend()
	token, err := authBackend.GenerateToken(requestUser.UUID)
	if err != nil {
		panic(err)
	}
	response, err := json.Marshal(parameters.TokenAuthentication{token})  //这里还应该将老的token加到redis黑名单中；
	if err != nil {
		panic(err)
	}
	return response
}

func Logout(req *http.Request) error {
	authBackend := authentication.InitJWTAuthenticationBackend()
	tokenRequest, err := jwt.ParseFromRequest(req, func(token *jwt.Token) (interface{}, error) {
		return authBackend.PublicKey, nil
	})
	if err != nil {
		return err
	}
	tokenString := req.Header.Get("Authorization")
	return authBackend.Logout(tokenString, tokenRequest)
}

https://github.com/brainattica/golang-jwt-authentication-api-sample/blob/master/services/models/users.go
type User struct {
	UUID     string `json:"uuid" form:"-"`
	Username string `json:"username" form:"username"`
	Password string `json:"password" form:"password"`
}