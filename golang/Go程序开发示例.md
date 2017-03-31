# Go程序开发示例

+ go程序员倾向于把所有go代码放到一个工作空间(workspace)中;
+ 工作空间可能包含多个版本控制仓库(如git)目录；
+ 每个仓库包含一个或多个packages；
+ 每个package包含一个或多个位于同个目录的Go源文件；
+ package目录的路径决定了该package的导入路径(import path)；

go程序代码是按照package来组织的， 同一package可以包含多个源文件。

packge main比较特殊， 编译后生成可执行程序； 其它package，编译后生成可链接的静态库(*.a)。

go工具提供了一种标准的获取、构建(build)、安装package(包括可执行程序)的方法，它要求代码按照一种特殊的目录结构和约定进行组织.

# 工作空间(workspace)

工作空间是包含下面三个子目录的目录：

| 子目录 | 功能                                                          |
| ------ | ------------------------------------------------------------- |
| src    | 包含所有组织成packages的Go源文件，每个子目录只能有一个package。   |
| pkg    | 包含编译后生存的package二进制对象。                           |
| bin    | 包含可执行二进制文件                                          |


go工具构建源程序，将生成的二进制文件安装到pkg或bin目录；

示例如下：

``` bash
  bin/
    hello                          # command executable
    outyet                         # command executable
  pkg/
    linux_amd64/
        github.com/golang/example/
            stringutil.a           # package object
  src/
    github.com/golang/example/ #项目目录，包含多个packages
        .git/
	    hello/                     #package hello， 命令
	        hello.go
	    outyet/                    #package outyet， 命令
	        main.go
	        main_test.go
	    stringutil/                #package stringutil， 函数库
	        reverse.go
	        reverse_test.go
```

golang开发者一般只配置一个workspace， 通过项目路径(如示例中的github.com/golang/example/)来划分各packages。

# GOPATH和PATH

  GOPATH环境变量用于指定workspace的位置，多个workspace用冒号分割。通常需要将各workspace下的bin目录加到PATH中。

``` bash
$ mkdir $HOME/work
$ export GOPATH=$HOME/work
$ export PATH=$PATH:$GOPATH/bin
```

# 导入路径(import path)

在import package或使用go工具编译package的时候，需要指定package path。

+ 标准库提供的packages：导入路径可以使用简写如"fmt"和"net/http", 它们是相对于$GOROOT/pkg/$GOOS_$GOARCH目录的。
+ 自己写的packages：导入路径不能和标准库或第三库的导入路径冲突，所以需要选择一个相对于$GOPATH/src的路径作为代码的package path。

一般情况下将代码仓库的URL作为项目的base path，如 github.com/user。

base path + package_name才是import使用的package path。

# 开发程序

1. 建立workspace，设置环境变量$GOPATH和$PATH:

  ``` bash
  $ mkdir -p $HOME/go/{src,pkg,bin}
  $ export GOPATH=$HOME/go/
  $ export PATH=$GOPATH/bin:$PATH
  ```

2. 设置一个base path，一般为代码仓库URL:

  ```  bash
  $ mkdir -p $HOME/src/github.com/golang
  ```

3. 在base path下面开发各个packages，每个package一个子目录

  ``` bash
  $ mkdir -p $HOME/src/github.com/golang/{song,decode,encode,mixer}
  ```

## 可执行程序

例如：

``` bashsh
$ mkdir $GOPATH/src/github.com/user/hello # hello main package
$ cat $GOPATH/src/github.com/user/hello/hello.go
```

``` go
package main

import "fmt"

func main() {
  fmt.Printf("Hello, world.\n")
}
```

使用go工具来构建(build)和安装(install)该package：
     
``` bash
$ go install github.com/user/hello #第三个参数为package path
```

注意，可以在任何目录执行上面命令，go工具会在$GOPATH中的各workspace中查找上面的package path。还可以切换到package path目录，直接执行go install命令：

``` bash
$ cd $GOPATH/src/github.com/user/hello
$ go install
```

go install命令 先构建hello程序，然后安装到$GOPATH/bin目录中， 如果该目录在PATH中，则可以直接执行生成的命令：

``` bash
$ hello
Hello, world.
```

# 函数库

开发函数库的方法和可执行程序类似，区别在于：

1. package name不是main。
2. go build不会生成任何文件，只做语法检查。
3. go install会将构建的函数库安装到$GOPATH/pkg/$GOOS_GOARCH目录下相应的package path中。

如果package main的代码导入了该函数库package，那么：

1. 执行 go build会在当前目录生成可执行程序，但不会安装函数库。
2. 执行 go install会安装该可执行程序的同时安装它依赖的函数库。

# 包名称

每个package path下的所有go源文件只能属于一个package(用于测试的*_test.go源文件例外)。

package path用来查找各packages， 源文件中的package name语句决定了该package被import后使用的限定前缀。

``` go
package name
```

例如demo.go中声明package demo，则其它package import该demo的packge path后，使用的限定前缀是demo，而不管package path最后
一级目录是什么。

golang的惯例是package name与该package所在的目录名称一致，即与package path的最后一级目录名称一致，例如 package path为"test/demo"
时，test/demo下所有go文件中的name应该为dmeo。两者不一致也是允许的，以package name为准，但这样在看代码时就不好确定限定前缀来源
于哪个package path。

各packages可以含有相同的限定前缀，只要它们的import path(package path)不同即可。

# 测试

go test工具和testing package组成一个轻量级的测试框架。

test文件以_test.go结尾，和被测试文件放在同一个目录下。test文件可以和所测试的文件属于同一个package，也可以按惯例命名为package_test。

``` go
package stringutil //或者package stringutil_test

import "testing"

func TestReverse(t *testing.T) {
cases := []struct {
  in, want string
}{
  {"Hello, world", "dlrow ,olleH"},
  {"Hello, 世界", "界世 ,olleH"},
  {"", ""},
}
for _, c := range cases {
  got := Reverse(c.in)
  if got != c.want {
    t.Errorf("Reverse(%q) == %q, want %q", c.in, got, c.want)
  }
}
  }
```

运行go test命令:

``` go
$ go test github.com/user/stringutil
ok  	github.com/user/stringutil 0.165s
```

# 远程包

如果代码中import path在当前workspace中不存在，则go工具会尝试利用git或hg从远程仓库下载。
go get命令会自动下载、构建和安装相应仓库的packages， 默认会安装到$GOPATH中第一个workspace。

# 参考

1. [How to Write Go Code](http://golang.org/doc/code.html)
