1. 区分好程序员和差程序员的一项标准是：好程序员写error处理代码，好程序员一直谨记代码要等抵御bug和failures；
	1. 检查输入的各种情况；
	2. 返回有意义的error信息；
	3. 添加必要的logger信息，logger要分级；

2. 为代码写测试代码，应该是一项标准的开发实践；

3. Go’s type system focuses on composition rather than inheritance.

4. PROBLEM
	You’re writing code that depends on types defined in external libraries, and you want to write test code that can verify that those libraries are correctly used.
SOLUTION
	Create interfaces to describe the types you need to test. Use those interfaces in your code, and then write stub or mock implementations for your tests.

如果要测试第三方库的方法，可以将这些方法封装到一个interface中，然后自己的代码使用这个接口对象，测试的时候，mock一个实现该接口的对象；
例如Send是一个第三方库的方法, 自己的程序需要调用它来发送告警，则可以这样写代码：

type Messager interface {
	Send(email, subject string, body []byte) error
}
func Alert(m Messager, problem []byte) error {
	return m.Send("noc@example.com", "Critical Error", problem)
}

// 测试代码
package msg
import (
	"testing"
)
// mock 对象
type MockMessage struct {
	email, subject string
	body []byte
}
func (m *MockMessage) Send(email, subject string, body []byte) error
	m.email = email
	m.subject = subject
	m.body = body
	return nil
	}
// 测试自己的发送逻辑
func TestAlert(t *testing.T) {
	msgr := new(MockMessage)
	body := []byte("Critical Error")
	Alert(msgr, body)
	if msgr.subject != "Critical Error" {
		t.Errorf("Expected 'Critical Error', Got '%s'", msgr.subject)	
	}
	// ...
}

5. 金丝雀canary测试
当依赖第三方库里的interface类型断言时，使用canary测试，可以在编译时发现不匹配的情况并出错退出；

func main() {
	m := map[string]interface{}{
	"w": &MyWriter{},
}
}

func doSomething(m map[string]interface{}) {
	w := m["w"].(io.Writer)
}

// 测试方法
func TestWriter(t *testing.T) {
	var _ io.Writer = &MyWriter{}
}

$ go test
# _/Users/mbutcher/Code/go-in-practice/chapter5/tests/canary
./canary_test.go:15: cannot use MyWriter literal (type *MyWriter)
as type io.Writer in assignment:
*MyWriter does not implement io.Writer (wrong type for Write method)
have Write([]byte) error
want Write([]byte) (int, error)
FAIL _/Users/mbutcher/Code/go-in-practice/chapter5/tests/canary
[build failed]

6. 如果template.Exec()，出现了错误则可以将错误写入到response中，这样可以了解错误的原因；

7. template.ParseFiles()会使用文件的basename作为template name，所以对应文件中不需要再使用{{define basefilename}}的形式定义tmplate；

错误：
tmpl, err := template.New("test").ParseFiles("file.txt")

正确：
tmpl, err := template.New("file.txt").ParseFiles("file.txt")

1. interface、map、slice、chan等变量是引用类型，在函数传参的过程中虽然pass by value，但是引用的是相同的数据；故没必要声明引用类型指针的函数参数；

2. TCP/HTTP Listen时，如果指定":"，则随机选择一个端口，在所有接口上监听；

3. 如果字符串为空，则Split后的列表有一个元素，为空字符串，而非空列表；
	$ cat test.go
	package main

	import "fmt"
	import "strings"

	func main() {
		s := ""
		s1 := strings.Split(s, ",")
		fmt.Printf("%#v\n", s1)
	}
	[zhangjun3@bjlg-b22-op101 pssh]$ go run test.go
	[]string{""}

4. 追加写文件时，OpenFile的flag不能仅仅是os.O_APPEND|os.O_CREATE，还需要包含os.O_WRONLY，文件的mode为0开头的八进制权限:

    if f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0755); err != nil {
        log.Fatalln(err)
    } else {
        log.SetOutput(f)
    }

5. 不要在for循环里使用defer语句：
m = new(sync.Mutex)
for {
	m.Lock()
	// do something
	defer m.Unlock()
}

5. defer只会在外层函数return或退出后才会执行：

func main() {
	prof, err := os.Create("/tmp/profile")
	if err != nil {
		fmt.Println(err)
	}
	if err := pprof.StartCPUProfile(prof); err != nil {
		echo(err.Error())
	}
	defer func() {
		pprof.StopCPUProfile()
		prof.Close()
	}()

	sig := make(chan os.Signal, 2)
	signal.Notify(sig, os.Interrupt, os.Kill, syscall.SIGTERM)
	go func() {
		<-sig
		fmt.Println("exited")
		os.Exit(1)
	}()
	for{
		os.Sleep(1)
	}
}
类似于上面的代码，如果C-c Interrupt了程序，程序直接退出，而不会执行main里面定义的defer代码；

5. 读取文件中每一行的方式

f, err := os.Open("/path/to/filename", os.O_RONLY)
if err != nil {
	panic(err)
}
scanner := bufio.NewScanner(f)
for scanner.Scan() {
	text := scanner.Text()
}
return scanner.Err()

5. 执行shell命令，捕获输出并设置超时时间：
	func execProcess(taskID int, ops string) (err error) {
		processCMD, err := filepath.Abs("./resource/process.py")
		if err != nil {
			return err
		}

		outputDIR := fmt.Sprintf("./output/hostInfo/%d", taskID)
		cmd := exec.Command(processCMD, ops, outputDIR)

		// 收集命令的输出和出错信息
		output := &bytes.Buffer{}
		cmd.Stdout = output
		cmd.Stderr = output

		if err := cmd.Start(); err != nil {
			return err
		}
		// 超时控制，最多执行10分钟
		timer := time.AfterFunc(10*time.Minute, func() {
			cmd.Process.Kill()
		})
		if err := cmd.Wait(); err != nil {
			return fmt.Errorf("exec process.py failed: \n%s", output.String())
		}  
		// else正常退出时，也可以使用output.String()打印出标准输出；
		timer.Stop()
		return nil
	}

6. 可寻址变量的情况：变量、指针重定向、slice索引、可寻址的struct的field、可寻址array的index操作符；组合字面量也可以寻址；
map的元素不是addressable的(因为map是动态扩容的，扩容后的地址可能与以前不一致)，所以不支持表达式m["foo"].f = x，但是支持m["a"]++运算符；
https://tip.golang.org/ref/spec#Address_operators
https://github.com/golang/go/issues/3117g
解决方法是：生成struct值，然后赋值给m["foo"]，如：
不能使用：
	ipPools[ipInfo.ISPName][ipInfo.ProvinceName][ipInfo.IP].Count += 1
需要修改为：
	ipPools[ipInfo.ISPName][ipInfo.ProvinceName][ipInfo.IP] = model.Stat{
		Low:   oldStat.Low,
		Mean:  oldStat.Mean,
		High:  oldStat.High,
		Count: oldStat.Count + 1,
	}

7. 在迭代过程中删除map元素，只能使用单元素迭代形式，不能使用key、value的两元素迭代形式：
正确：
	for hostName := range NodeCache["slave"].Nodes {
		if time.Now().Sub(NodeCache["slave"].Nodes[hostName].LastSeen) > 30*time.Second {
			delete(NodeCache["slave"].Nodes, hostName)
			log.Printf("slave node %s timeout, removed!\n", hostName)
		}
	}
无效：
	for hostName, node := range NodeCache["slave"].Nodes {
		if time.Now().Sub(node.LastSeen) > 30*time.Second {
			delete(NodeCache["salve"].Nodes, hostName)
			log.Printf("slave nodes: %v\n", NodeCache["slave"].Nodes)
		}
	}


8. HTTP服务器允许CORE请求
	ret, _ := json.Marshal(response)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write(ret)

9. 自定义http client的场景：
	1. 指定transport，调用它的CancelRequest可以随时cacel request；
	2. 指定timeout，超时后client调用底层transport的CancelRequest;

注意：建议使用Request.WithContext创建可以cancel的Request。CancelRequest不能cancel http/2请求；


9. 并发读写map时会panic，可以在build的时候加-race参数，这样panic的时候，会提示读写的位置：
	==================
	WARNING: DATA RACE
	Read at 0x00000128f6c0 by goroutine 1070:
	git.op.ksyun.com/zhangjun3/netmap/slave.doFping()
		/home/ksyun/golang/src/git.op.ksyun.com/zhangjun3/netmap/slave/fping.go:127 +0xb05

	Previous write at 0x00000128f6c0 by goroutine 1066:
	git.op.ksyun.com/zhangjun3/netmap/slave.(*SlaveNodeType).Post()
		/home/ksyun/golang/src/git.op.ksyun.com/zhangjun3/netmap/slave/rpc.go:70 +0x14e
	git.op.ksyun.com/zhangjun3/netmap/slave.doFping()
		/home/ksyun/golang/src/git.op.ksyun.com/zhangjun3/netmap/slave/fping.go:147 +0x150f

10. 读map元素时，如果不存在，则返回元素对应的zero value，不一定是nil，如:
	type node struct {
		age int
	}
	m := map[string]node{}
	m["123"]的结果是空node struct，不是nil；

11. range可以迭代nil的slive、map，但不能迭代nil的channel:
$ cat demo.go
package main

import (
        "fmt"
)

func main() {
  var m map[string]int
  var l []int
  var c chan int

  for k,v := range m {
      fmt.Println(k, v)
  }
  fmt.Println("nil map ok!")

  for i,v := range l {
      fmt.Println(i, v)
  }
  fmt.Println("nil slice ok!")

  //range nil channel时将一直被block；
  //for v := range c {
  //    fmt.Println(v)
  //}
  //fatal error: all goroutines are asleep - deadlock!
  
  //写nil channel时将一直被block
  c <- 2
  //fatal error: all goroutines are asleep - deadlock!
}

12. 调用函数时，函数参数或返回值中的slice变量只是被声明而未初始化：地址为0x0，值为nil，不可以index;但可以对nil slice使用append函数；
slice变量被空字面量方式初始化后，地址不再是0x0，但是长度和容量为0，不能index，可以append；
% cat demo.go
package main
import "fmt"

func add(a, b int) (output []int) {
   fmt.Printf("begin: output:\n\ttype: %T\n\taddress: %p\n\tvalue: %v\n\toutput == nil? %t\n", output, output, output, output==nil)

   //output[0] = 1
   //Output: panic: runtime error: index out of range

   fmt.Println("append a, b")
   output = append(output, a, b)
   fmt.Printf("result: %#v\n\n", output)

   output = []int{} //等效为 output = make([]int, 0, 0)
   fmt.Printf("after: output:\n\ttype: %T\n\taddress: %p\n\tvalue: %v\n\tlength: %v\n\tcap: %v\n\toutput == nil? %t\n", output, output, output, len(output), cap(output), output==nil)
   //output[0] = 1
   //Output: panic: runtime error: index out of range

   output = append(output, a, b)

   output = make([]int, 2, 2)
   output[0] = 1
   return
}

func main(){
  output := add(1, 2)
  fmt.Printf("%v\n", output)
}

% go run demo.go
begin: output:
	type: []int
	address: 0x0
	value: []
	output == nil? true
append a, b
result: []int{1, 2}

after: output:
	type: []int
	address: 0x116300
	value: []
	length: 0
	cap: 0
	output == nil? false
[1 0]

13. 函数参数或返回值中的map、channel也只是被声明但未初始化：地址为0x0, 值为nil；对于nil map不可以赋值，但是可以获取结果；
% cat demo1.go
package main
import "fmt"

func add(a, b int) (mmap map[int]int) {
   fmt.Printf("begin: mmap:\n\ttype: %T\n\taddress: %p\n\tvalue: %v\n\tmmap == nil? %t\n", mmap, mmap, mmap, mmap==nil)
   fmt.Printf("result: %#v\n\n", mmap)

   //mmap[1] = 1
   //panic: assignment to entry in nil map

   fmt.Println(mmap[1])
   //输出： 0

   return
}

func main(){
  add(1, 2)
}
% go run demo1.go
begin: mmap:
	type: map[int]int
	address: 0x0
	value: map[]
	mmap == nil? true
result: map[int]int(nil)

14. HTTP Post的时候需要对body的参数进行URL编码：
value := url.Values{}
value.Add("id", "wj5027")
value.Add("MD5_td_code", "5fe022d21cd6a964b3fe7538af93b356")
//必须设置Content-Type头部，否则Server会出现解析错误；
resp, err := http.Post(postURL, "application/x-www-form-urlencoded", bytes.NewBufferString(value.Encode()))

或者：
form := url.Values{}
form.Add("ln", c.ln)
req, err := http.NewRequest("POST", url, strings.NewReader(form.Encode()))
//必须设置Content-Type头部，否则Server会出现解析错误；
req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
hc := http.Client{}
resp, err := hc.Do(req)

注意： 
1. http.Post()和http.NewRequest()最后一个参数都是io.Reader，读取后的数据将作为Request的Body；
2. http.PostForm和http.Post类似，自动设置Content-Type为application/x-www-form-urlencoded；


15. for循环内部生成goroutine时，需要将遍历的结果作为参数传给goroutine函数，不能依赖闭包特性在goroutine内部引用，因为goroutine可能是延迟执行，多个goroutine可能同时获得相同的闭包值；
for i, file = range os.Args[1:] {
       wg.Add(1)
       go func(filename string) {
              compress(filename)
              wg.Done()
       }(file)  //遍历的值通过函数参数传给goroutine；
} 
或者，重新定义循环变量为本地变量：
for i, file = range os.Args[1:] {
       wg.Add(1)
	   file2 := file
       go func() {
              compress(file2)
              wg.Done()
       }
} 

类似于下面的遍历channel的方式是有问题的：
for devicePower := range devicePowerChan {
		fmt.Printf("device power: %v\n", devicePower)
		// 将采集的的device power发送到MQ
		producer.Input() <-&sarama.ProducerMessage{
				Topic: "power",
				Key:   sarama.StringEncoder(devicePower.Rack), // 使用设备Rack作为key，保证同一个机柜的设备在一个分区
				Value: &devicePower,                         // model.Power实现了sarama.Encoder接口
		}
}
这会导致给producer.Input() channel发送重复的消息；

正确的形式为：
for devicePower := range devicePowerChan {
		fmt.Printf("device power: %v\n", devicePower)
		// 将采集的的device power发送到MQ
		value := devicePower
		producer.Input() <-&sarama.ProducerMessage{
				Topic: "power",
				Key:   sarama.StringEncoder(devicePower.Rack), // 使用设备Rack作为key，保证同一个机柜的设备在一个分区
				Value: &value,                         // model.Power实现了sarama.Encoder接口
		}
}

16. select一般嵌套在for循环中使用，用来重复选择可以接收或发送的chann分支：
	for {
		select {
		case buf := <-echo:
			os.Stdout.Write(buf)
		case <-done:
			fmt.Println("Timed out")
			os.Exit(0)
		} 
	}

17. channel应该在它的sender代码中关闭，而不能在其它地方关闭；否则sender向关闭的channel发送数据会引起panic；
常用的解决方法是定义一个布尔类型的done channel和for循环包含的select语句，将发送数据的操作放在default子句中，这样select每次选择时先看done是否可读，如果不可读再执行default channel
中的发送语句：

package main

import (
    "fmt"
	"time"
	)

func main() {
     msg := make(chan string)
     done := make(chan bool) //用于指示是否结束的channel
     until := time.After(5 * time.Second)
     go send(msg, done)
	 for {
		select {
			case m := <-msg:
				fmt.Println(m)
			case <-until:
				done <- true //指示结束
				time.Sleep(500 * time.Millisecond)
				return 
		}
	 }
}

func send(ch chan<- string, done <-chan bool) {
	for {
		select {
			case <-done: //如果结束，则close channel；
				println("Done")
				close(ch)
				return
			default: //实际发送语句位于default子句中；只有当done不可读时，select才会选择执行该子句中的语句；
				ch <- "hello" 
				time.Sleep(500 * time.Millisecond)
  } }
}

18. 通过buffer为1的channel，可以实现全局锁：
package main
import (
"fmt"
"time"
)
func main() {
	lock := make(chan bool, 1)
	for i := 1; i < 7; i++ {
		go worker(i, lock)
	}
	time.Sleep(10 * time.Second)
}

func worker(id int, lock chan bool) {
	fmt.Printf("%d wants the lock\n", id)
	lock <- true
	fmt.Printf("%d has the lock\n", id)
	time.Sleep(500 * time.Millisecond)
	fmt.Printf("%d is releasing the lock\n", id)
	<-lock
}

另一个例子，多个goroutine竞争处理一个query，返回第一个结果：
func Query(conns []Conn, query string) Result {
    ch := make(chan Result, 1)
    for _, conn := range conns {
        go func(c Conn) {
            select {
            case ch <- c.DoQuery(query):
            default:
            }
        }(conn)
    }
    return <-ch
}

18. 对任意函数或方法调用执行超时控制的方法：
		c := make(chan error, 1)
		// 超时控制
		t := time.After(5 * time.Second)
		go func() { c <- client.Call("KeepAliveType.Join", SlaveNode, nil) }()
		select {
		case err := <-c:
			// 函数执行结束，检查返回值；
		case <-t:
			// 函数执行超时
		}
对于命令，可以使用下面的方法，对执行时间进行控制:
	cmd := exec.Command("fping", args...)
	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start fping failed: %v", err)
	}

	// 超时控制：最长执行15分钟
	timer := time.AfterFunc(900*time.Second, func() {
		cmd.Process.Kill()
	})
	defer timer.Stop()

18. 尽管Go提供了无符号的整形和数字，但是倾向于使用int类型(即使对于非负的情形，如数组的长度)，无符号整形一般只在位运算、加密等场景使用；
float32是6位数字的精度，float64是15位数字的精度；一般情况下，优先选择float64类型；
http://www.cnblogs.com/BradMiller/archive/2010/11/25/1887945.html

19. 当一个package的返回的错误类型比较多时，可以定义一些error 变量，这样调用者就可以对结果的错误类型进行判断，进而决定如何处理：
package main
import (
     "errors"
	 "fmt"
     "math/rand"
	）

var ErrTimeout = errors.New("The request timed out")
var ErrRejected = errors.New("The request was rejected")
var random = rand.New(rand.NewSource(35))

func main() {
     response, err := SendRequest("Hello")
     for err == ErrTimeout {
            fmt.Println("Timeout. Retrying.")
            response, err = SendRequest("Hello")
              }
     if err != nil {
            fmt.Println(err)
     } else {
            fmt.Println(response)
	} 
}

func SendRequest(req string) (string, error) {
     switch random.Int() % 3 {
     case 0:
            return "Success", nil
     case 1:
            return "", ErrRejected
     default:
            return "", ErrTimeout
	}
}

20. go惯例是函数最后一个返回值是error类型，另外自定义类型也可以实现error接口；

20. 如果在function calling stack的顶部还没有被处理panic，则程序会异常退出；
对于goroutine而言，它不能jump到初始化该goroutine的函数(即调用 go handler() 语句的函数), groutine
的函数调用栈顶层是handler()函数；如果goroutine发生panic且一直到handler()没有处理，则会终止整个程序的执行；

解决办法是在handler里面defer recover并处理；也可以将会引起panic的代码包在匿名函数代码块中，然后在函数的顶部定义recover代码；

21. 可以在程序运行的过程中打印函数调用stack：方法是调用 runtime.Stack()函数，用于调试，runtime package还提供了其它类似函数可以用于获取runtime的信息；

22. Go的测试代码文件应放置在被测试的文件目录中，而不是专门把测试文件放到一个目录下；另外测试文件应该与被测试文件位于一个packae中，这样可以测试public和非public的代码；

23. 测试方法：
	1. 依赖于第三方的类型及其方法：可以将依赖的方法抽象为接口，程序使用接口对象；这样在测试时，可以mock一个实现该接口的对象，进而测试；
	2. 类型断言的测试方法：在测试代码里使用var _ io.Writer = &MyWriter{} 来测试MyWritter{}对象实现了io.Writer接口；

24. HTTP Post发送和接收JSON数据

编码JSON：
type User struct{
    Id      string
    Balance uint64
}

func main() {
    u := User{Id: "US123", Balance: 8}
    b := new(bytes.Buffer)
    json.NewEncoder(b).Encode(u)
    res, _ := http.Post("https://httpbin.org/post", "application/json; charset=utf-8", b)
    io.Copy(os.Stdout, res.Body)
}

解码JSON：
func main() {
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        var u User
        if r.Body == nil {
            http.Error(w, "Please send a request body", 400)
            return
        }
        err := json.NewDecoder(r.Body).Decode(&u) // JSON数据位于HTTP Body中；
        if err != nil {
            http.Error(w, err.Error(), 400)
            return
        }
        fmt.Println(u.Id)
    })
    log.Fatal(http.ListenAndServe(":8080", nil))
}

24. 接收和处理信号的方式
	ch := make(chan os.Signal)
	signal.Notify(ch, os.Interrupt, os.Kill) // 注册关注的信号
	go listenForShutdown(ch)

	func listenForShutdown(ch <-chan os.Signal) {
		<-ch // 收到信号
		manners.Close()
	}

25. 如果向http.ServeMux注册一个subtree，即path以"/"结尾，如http.HandleFunc("/zim/", xxx), 则收到不带"/"的请求如"/zim"时，
ServeMux会将请求重定向到"/zim/"，除非重新注册了一个不带"/"的path如http.HandleFunc("/zim", xxx)，这时请求"/zim"精确匹配到该Handler；
http.HandleFunc("/test/", func(w http.ResponseWriter, r *http.Request) { // 访问/test/正常，访问/test时会被重定向到"/test/"
		fmt.Fprintf(w, "hello!")
	})
http.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) { // 访问/test正常，访问/test/提示404；
	fmt.Fprintf(w, "hello!")
})
dir := http.Dir("./github/Notes/Zim")
handler := http.StripPrefix("/zim", http.FileServer(dir)) // 可以是/zim或/zim/
http.Handle("/zim/", handler) // 请求的是以"/zim"开头的文件路径，所以"/zim/"最后的"/"不能省，否则会提示404；

26. 大文件上传，server调用request.FormFile()，request.MultipartFrom时，底层会调用ParseMultipartForm()，这个方法会等待所有的文件上传完毕，
并将文件内容保存到内存或硬盘。为了避免本地保存大量的临时文件，应该使用Request.MultipartReader()来流式获取上传的各mutil part文件，保存到本地或远
端(如云存储KS3)；

示例：
1. mutipart表单，同时包含text和file field；
<!doctype html>
<html>
	<head>
		<title>File Upload</title>
	</head>
	<body>
		<form action="/" method="POST" enctype="multipart/form-data">
			<label for="name">Name:</label>
			<input type="text" name="name" id="name">  // text field
			<br>
			<label for="file">File:</label>
			<input type="file" name="file" id="file">  // file field
			<br>
			<button type="submit" name="submit">Submit</button>
		</form>
	</body>
</html>

2. 
func fileForm(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		t, _ := template.ParseFiles("file_plus.html")
		t.Execute(w, nil)
	} else {
		mr, err := r.MultipartReader()
		if err != nil {
			panic("Failed to read multipart message")
		}
		values := make(map[string][]string)
		maxValueBytes := int64(10 << 20)
		for {
			part, err := mr.NextPart()
			if err == io.EOF {
				break
			}
			name := part.FormName() // 检查field name
			if name == "" {
				continue
			}
			filename := part.FileName() // 检查file name
			var b bytes.Buffer
			if filename == "" { // 文件名为空，说明是非file field
				n, err := io.CopyN(&b, part, maxValueBytes)
				if err != nil && err != io.EOF {
					fmt.Fprint(w, "Error processing form")
					return
				}
				maxValueBytes -= n
				if maxValueBytes == 0 {
					msg := "multipart message too large"
					fmt.Fprint(w, msg)
					return
				}
				values[name] = append(values[name],b.String())
				continue
			}
			dst, err := os.Create("/tmp/" + filename)
			defer dst.Close()
			if err != nil {
				return
			}
			for {
				buffer := make([]byte, 100000)
				cBytes, err := part.Read(buffer)
				if err == io.EOF {
					break
				}
				dst.Write(buffer[0:cBytes])
			}
		}
		fmt.Fprint(w, "Upload complete")
	}
}

27. 判断http package返回的err是否是timeout的方法
func hasTimedOut(err error) bool {
	switch err := err.(type) {
	case *url.Error: // 如果匹配 *url.Error类型，则err变为 url.Error的实例， err.Err是url.Error的成员；
		if err, ok := err.Err.(net.Error); ok && err.Timeout() {
			return true
		}
	case net.Error:
		if err.Timeout() {
			return true
		}
	case *net.OpError:  // 这一个判断实际上是多余的，因为net.Error是接口，*net.OpError实现了net.Error接口；
		if err.Timeout() {
			return true
		}
	}
	errTxt := "use of closed network connection"
	if err != nil && strings.Contains(err.Error(), errTxt) {
		return true
	}
	return false
}

28. 大文件下载的断点续传+重试
func download(location string, file *os.File, retries int64) error {
	req, err := http.NewRequest("GET", location, nil)
	if err != nil {
		return err
	}
	fi, err := file.Stat() // 获取当前文件的大小
	if err != nil {
		return err
	}
	current := fi.Size()
	if current > 0 {
		start := strconv.FormatInt(current, 10)
		req.Header.Set("Range", "bytes="+start+"-") // 如果文件当前大小大于0，则设置Range请求的开始位置；
	}
	cc := &http.Client{Timeout: 5 * time.Minute}	// 设置下载文件所需的时间
	res, err := cc.Do(req)
	if err != nil && hasTimedOut(err) {				// 如果超时，则递归重试
		if retries > 0 {
			return download(location, file, retries-1)
		}
		return err
	} else if err != nil {
		return err
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		errFmt := "Unsuccess HTTP request. Status: %s"
		return fmt.Errorf(errFmt, res.Status)
	}
	if res.Header.Get("Accept-Ranges") != "bytes" {
		retries = 0
	}
	_, err = io.Copy(file, res.Body)
	if err != nil && hasTimedOut(err) {
		if retries > 0 {
			return download(location, file, retries-1)
		}
		return err
	} else if err != nil {
		return err
	}
	return nil
}

func main() {
	file, err := os.Create("file.zip")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()
	location := https://example.com/file.zip
	err = download(location, file, 100)
	if err != nil {
		fmt.Println(err)
		return
	}
	fi, err := file.Stat()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Got it with %v bytes downloaded", fi.Size())
}

在断点重试的时候，可以考虑引入指数回退计数算法：https://github.com/jpillora/backoff；

29. 为了给消费者提供稳定的API服务，服务端应该实现基于版本的API，使用大版本如v1,v2,v3来标记不兼容的大版本更新，使用小版本如v1.1,v1.2来标记兼容的
版本更新；常用的两种方法是：
1. 在URL中包含版本信息如：http://netbench.ksyun.com/api/v1/todos，对于开发者来说更简单，但不符合REST的在Request中包含所有信息的语义；
2. 在HTTP Request的Header如Content-Type中包含版本信息，也可以是Customize Header；

30. 程序尽量做到无配置启动，从运行环境中获取有用的运行信息；

31. net package中的错误实现了net.Error接口，通过转换后，可以判断是否是Timeout还是其它原因引起的，如：
	var conn net.Conn
	var err error
	// 如果是dial timeout，则最多重试三次
	for i := 0; i < 3; i++ {
		conn, err = net.DialTimeout("tcp", hostIP+":22", dialTimeout)
		if err == nil {
			break
		} else if err.(net.Error).Timeout() {
			continue
		} else {
			return nil, err
		}
	}
	if err != nil && err.(net.Error).Timeout() {
		return nil, fmt.Errorf("tried 3 times, but all timeout")
	}