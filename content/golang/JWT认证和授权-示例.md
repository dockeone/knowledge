# express-jwt实例
下面给一个具体的实例，这个例子的客户端是web app，使用AngularJS框架。服务端使用NodeJS做的RESTful API接口，客户端直接调用接口数据，其中使用了token认证机制。
当用户把他的授权信息发过来的时候， Node.js 服务检查是否正确，然后返回一个基于用户信息的唯一 token 。 AngularJS 应用把 token 保存在用户的 SessionStorage ，
之后的在发送请求的时候，在请求头里面加上包含这个 token 的 Authorization。如果 endpoint 需要确认用户授权，服务端检查验证这个 token，然后如果成功了就返回数据，
如果失败了返回 401 或者其它的异常。

用到的技术：
+ AngularJS
+ NodeJS （ express.js, express-jwt 和 moongoose）
+ MongoDB
+ Redis （备用，用于记录用户退出登录时候还没有超时的token）

## 客户端 : AngularJS 部分
首先，我们来创建我们的 AdminUserCtrl controller 和处理 login/logout 动作。
appControllers.controller('AdminUserCtrl', ['$scope', '$location', '$window', 'UserService', 'AuthenticationService',
    function AdminUserCtrl($scope, $location, $window, UserService, AuthenticationService) {
 
        //Admin User Controller (login, logout)
        $scope.logIn = function logIn(username, password) {
            if (username !== undefined && password !== undefined) {
                UserService.logIn(username, password).success(function(data) {
                    AuthenticationService.isLogged = true;
                    $window.sessionStorage.token = data.token;
                    $location.path("/admin");
                }).error(function(status, data) {
                    console.log(status);
                    console.log(data);
                });
            }
        }
 
        $scope.logout = function logout() {
            if (AuthenticationService.isLogged) {
                AuthenticationService.isLogged = false;
                delete $window.sessionStorage.token;
                $location.path("/");
            }
        }
    }
]);

这个 controller 用了两个 service: UserService 和 AuthenticationService。第一个处理调用 REST api 用证书。后面一个处理用户的认证。它只有一个布尔值，用来表示用户是否被授权。
appServices.factory('AuthenticationService', function() {
    var auth = {
        isLogged: false
    }
 
    return auth;
});

appServices.factory('UserService', function($http) {
    return {
        logIn: function(username, password) {
            return $http.post(options.api.base_url + '/login', {username: username, password: password});
        },
 
        logOut: function() {
 
        }
    }
});
好了，我们需要做张登陆页面:

<form class="form-horizontal" role="form">
    <div class="form-group">
        <label for="inputUsername" class="col-sm-4 control-label">Username</label>
        <div class="col-sm-4">
            <input type="text" class="form-control" id="inputUsername" placeholder="Username" ng-model="login.email">
        </div>
    </div>
    <div class="form-group">
        <label for="inputPassword" class="col-sm-4 control-label">Password</label>
        <div class="col-sm-4">
            <input type="password" class="form-control" id="inputPassword" placeholder="Password" ng-model="login.password">
        </div>
    </div>
    <div class="form-group">
        <div class="col-sm-offset-4 col-sm-10">
            <button type="submit" class="btn btn-default" ng-click="logIn(login.email, login.password)">Log In</button>
        </div>
    </div>
</form>

当用户发送他的信息过来，我们的 controller 把内容发送到 Node.js 服务器，如果信息可用，我们把 AuthenticationService里面的 isLogged 设为 true。我们把从服务端发过来的 token 
存起来，以便下次请求的时候使用。等讲到 Node.js 的时候我们会看看怎么处理。

好了，我们要往每个请求里面追加一个特殊的头信息了:[Authorization: Bearer ] 。为了实现这个需求，我们建立一个服务，叫 TokenInterceptor。

appServices.factory('TokenInterceptor', function ($q, $window, AuthenticationService) {
    return {
        request: function (config) {
            config.headers = config.headers || {};
            if ($window.sessionStorage.token) {
                config.headers.Authorization = 'Bearer ' + $window.sessionStorage.token;
            }
            return config;
        },
 
        response: function (response) {
            return response || $q.when(response);
        }
    };
});
然后我们把这个interceptor 追加到 $httpProvider :

app.config(function ($httpProvider) {
    $httpProvider.interceptors.push('TokenInterceptor');
});
然后，我们要开始配置路由了，让 AngularJS 知道哪些需要授权，在这里，我们需要检查用户是否已经被授权，也就是查看 AuthenticationService 的 isLogged 值。

app.config(['$locationProvider', '$routeProvider', 
  function($location, $routeProvider) {
    $routeProvider.
        when('/', {
            templateUrl: 'partials/post.list.html',
            controller: 'PostListCtrl'
        }).
        when('/post/:id', {
            templateUrl: 'partials/post.view.html',
            controller: 'PostViewCtrl'
        }).
        when('/tag/:tagName', {
            templateUrl: 'partials/post.list.html',
            controller: 'PostListTagCtrl'
        }).
        when('/admin', {
            templateUrl: 'partials/admin.post.list.html',
            controller: 'AdminPostListCtrl',
            access: { requiredLogin: true }
        }).
        when('/admin/post/create', {
            templateUrl: 'partials/admin.post.create.html',
            controller: 'AdminPostCreateCtrl',
            access: { requiredLogin: true }
        }).
        when('/admin/post/edit/:id', {
            templateUrl: 'partials/admin.post.edit.html',
            controller: 'AdminPostEditCtrl',
            access: { requiredLogin: true }
        }).
        when('/admin/login', {
            templateUrl: 'partials/admin.login.html',
            controller: 'AdminUserCtrl'
        }).
        when('/admin/logout', {
            templateUrl: 'partials/admin.logout.html',
            controller: 'AdminUserCtrl',
            access: { requiredLogin: true }
        }).
        otherwise({
            redirectTo: '/'
        });
}]);

app.run(function($rootScope, $location, $window, AuthenticationService) {
    $rootScope.$on("$routeChangeStart", function(event, nextRoute, currentRoute) {
        //redirect only if both isLogged is false and no token is set
        if (nextRoute != null && nextRoute.access != null && nextRoute.access.requiredLogin 
            && !AuthenticationService.isLogged && !$window.sessionStorage.token) {

            $location.path("/admin/login");
        }
    });
});

# 服务端: Node.js + MongoDB 部分
为了在我们的 RESTful api 处理授权信息，我们要用到 express-jwt (JSON Web Token) 来生成一个唯一 Token，基于用户的信息。以及验证 Token。
首先，我们在 MongoDB 里面创建一个用户的 Schema。我们还要创建调用一个中间件，在创建和保存用户信息到数据库之前，用于加密密码。还有我们需要一个方法来解密密码，当收到用户请求的时候，
检查是否在数据库里面有匹配的。

var Schema = mongoose.Schema;
 
// User schema
var User = new Schema({
    username: { type: String, required: true, unique: true },
    password: { type: String, required: true}
});
 
// 使用bcrypt对用户的密码进行加密，然后存到数据库；
// Bcrypt middleware on UserSchema
User.pre('save', function(next) {
  var user = this;
 
  if (!user.isModified('password')) return next();
 
  bcrypt.genSalt(SALT_WORK_FACTOR, function(err, salt) {
    if (err) return next(err);
 
    bcrypt.hash(user.password, salt, function(err, hash) {
        if (err) return next(err);
        user.password = hash;
        next();
    });
  });
});
 
//Password verification
User.methods.comparePassword = function(password, cb) {
    bcrypt.compare(password, this.password, function(err, isMatch) {
        if (err) return cb(err);
        cb(isMatch);
    });
};

然后我们开始写授权用户和创建 Token 的方法:

exports.login = function(req, res) {
    var username = req.body.username || '';
    var password = req.body.password || '';
 
    if (username == '' || password == '') {
        return res.send(401);
    }
 
    db.userModel.findOne({username: username}, function (err, user) {
        if (err) {
            console.log(err);
            return res.send(401);
        }
 
        user.comparePassword(password, function(isMatch) {
            if (!isMatch) {
                console.log("Attempt failed to login with " + user.username);
                return res.send(401);
            }
 
            var token = jwt.sign(user, secret.secretToken, { expiresInMinutes: 60 });
 
            return res.json({token:token});
        });
 
    });
};

最后，我们需要把 jwt 中间件加到所有的，访问时需要授权的路由上面:

/*
Get all published posts
*/
app.get('/post', routes.posts.list);
/*
    Get all posts
*/
app.get('/post/all', jwt({secret: secret.secretToken}), routes.posts.listAll);
 
/*
    Get an existing post. Require url
*/
app.get('/post/:id', routes.posts.read);
 
/*
    Get posts by tag
*/
app.get('/tag/:tagName', routes.posts.listByTag);
 
/*
    Login
*/
app.post('/login', routes.users.login);
 
/*
    Logout
*/
app.get('/logout', routes.users.logout);
 
/*
    Create a new post. Require data
*/
app.post('/post', jwt({secret: secret.secretToken}), routes.posts.create);
 
/*
    Update an existing post. Require id
*/
app.put('/post', jwt({secret: secret.secretToken}), routes.posts.update);
 
/*
    Delete an existing post. Require id
*/

app.delete('/post/:id', jwt({secret: secret.secretToken}), routes.posts.delete);
上面这个实例就采用了token的验证方式构建了api接口，但是有两个问题需要解决：

1. 用户退出登录，但是token并没有失效(因为token没有过期且值正确，服务器会校验通过，只有等到了过期时间或者token被修改，服务器才校验失败)，因为服务端没有删除这个token;
2. token失效了，怎么办，如果还是让用于登录重新获取token，会体验不好。应该有token刷新机制。

###使用Redis解决问题1
解决方法是：当用户点了 logout 按钮的时候，Token 只会保存一段时间，就是你用 jsonwebtoken 登陆之后，token 有效的这段时间，我们将这个token存放在Redis中，
生存时间也是jwt获取这个token的时间。这个时间到期后，token 会被 redis 自动删掉。最后，我们创建一个 nodejs 的中间件，检查所有受限 endopoint 用的 token 是否存在 Redis 数据库中。

redis充当黑名单的功能：用户点击logout后，token被保存到redis中，过期时间为当前时间+该token剩余的有效时间(下面的代码使用固定时间)；收到用户请求后，先校验token正确性，再检查token是否在该名单中，
如果在则报错，如果不在，在继续后续处理；

####NodeJS 配置 Reids
var redis = require('redis');
var redisClient = redis.createClient(6379);
 
redisClient.on('error', function (err) {
    console.log('Error ' + err);
});
 
redisClient.on('connect', function () {
    console.log('Redis is ready');
});
 
exports.redis = redis;
exports.redisClient = redisClient;

然后，我们来创建一个方法，用来检查提供的 token 是不是被删除了。

#### Token 管理和中间件
为了在 Redis 中保存 Token，我们要创建一个方法来拿到请求中的 Header 的 Token 参数，然后把它作为 Redis 的 key 保存起来。值是什么我们不管它。

var redisClient = require('./redis_database').redisClient;
var TOKEN_EXPIRATION = 60;
var TOKEN_EXPIRATION_SEC = TOKEN_EXPIRATION * 60;
 
exports.expireToken = function(headers) {
    var token = getToken(headers);
 
    if (token != null) {
        redisClient.set(token, { is_expired: true });
        redisClient.expire(token, TOKEN_EXPIRATION_SEC);
    }
};
 
var getToken = function(headers) {
    if (headers && headers.authorization) {
        var authorization = headers.authorization;
        var part = authorization.split(' ');
 
        if (part.length == 2) {
            var token = part[1];
 
            return part[1];
        }
        else {
            return null;
        }
    }
    else {
        return null;
    }
};

然后，再创建一个中间件来验证一下 token，当用户发起请求的时候:

// Middleware for token verification
exports.verifyToken = function (req, res, next) {
    var token = getToken(req.headers);
 
    redisClient.get(token, function (err, reply) {
        if (err) {
            console.log(err);
            return res.send(500);
        }
 
        if (reply) { //token在redis记录的logout黑名单中；
            res.send(401);
        }
        else {
            next();
        }
 
    });
};

verifyToken 这个方法，是一个中间件，用来拿到请求头中的 token，然后在 Redis 里面查找它。如果 token 被发现了，我们就发 HTTP 401.否则我们就继续工作流，让请求访问 API。

我们要在用户点 logout 的时候，执行 expireToken 方法:

exports.logout = function(req, res) {
    if (req.user) {
        tokenManager.expireToken(req.headers);
 
        delete req.user;
        return res.send(200);
    }
    else {
        return res.send(401);
    }
}

最后我们更新路由，用上新的中间件:

//Login
app.post('/user/signin', routes.users.signin);
 
//Logout
app.get('/user/logout', jwt({secret: secret.secretToken}), routes.users.logout);
 
//Get all posts
app.get('/post/all', jwt({secret: secret.secretToken}), tokenManager.verifyToken, routes.posts.listAll);
 
//Create a new post
app.post('/post', jwt({secret: secret.secretToken}), tokenManager.verifyToken , routes.posts.create);
 
//Edit the post id
app.put('/post', jwt({secret: secret.secretToken}), tokenManager.verifyToken, routes.posts.update);
 
//Delete the post id
app.delete('/post/:id', jwt({secret: secret.secretToken}), tokenManager.verifyToken, routes.posts.delete);

好了，现在我们每次发送请求的时候，我们都去解析 token， 然后看看是不是有效的。

##refresh token解决问题2
服务端提供一个refreToken API，使用token验证保护；客户端检查当前的token是否快过期了，如果是，则向服务器的refreToken API发送更新请求，获得新的Token，服务器将旧的token加到redis黑名单中；

整个项目的源码：https://github.com/kdelemme/blogjs