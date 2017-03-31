http://grisha.org/blog/2014/06/03/graceful-restart-in-golang/
https://github.com/fvbock/endless
https://github.com/rcrowley/goagain
https://github.com/sevlyar/go-daemon
https://github.com/jdhenke/shared-conn/blob/master/main.go

Graceful Restart in Golang: 重启go程序(可能是因为程序二进制文件升级或更新了配置)，但是不中断已有的连接；需要解决两个问题：
重启但是不关闭socket：
1. Fork a new process which inherits the listening socket.
2. The child performs initialization and starts accepting connections on the socket.
3. Immediately after, child sends a signal to the parent causing the parent to stop accepting connecitons and terminate.

exec.Command返回的cmd的ExtraFiles可以指定除了(stdin/err/out)外被子进程继承的open files：
	// create fd to share
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	lnFile, err := ln.(*net.TCPListener).File()
	if err != nil {
		log.Fatal(err)
	}

	// assemble child process
	os.Setenv("IS_CHILD", "true")
	cmd := exec.Command(os.Args[0], os.Args[1:]...)
	cmd.Stdin = os.Stdin   // fd 0
	cmd.Stdout = os.Stdout // fd 1
	cmd.Stderr = os.Stderr // fd 2
	cmd.ExtraFiles = []*os.File{lnFile} // 将父进程打开的file传给子进程

	// start child process
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

注意：
lnFile, err := ln.(*net.TCPListener).File() 关闭ln对lnFile没影响，反之亦然；实际上是调用dup()并清除返回文件描述符的FD_CLOEXEC 
flag，这样在运行子进程前不会关闭该fd；
cmd.ExtraFiles = []*os.File{lnFile}： 如果非nil，entry i 在子进程内部的fd是 3+i;(因为默认继承了stdin/err/out);

在子进程内部：
    server := &http.Server{Addr: "0.0.0.0:8888"}

    var gracefulChild bool
    var l net.Listever
    var err error

    flag.BoolVar(&gracefulChild, "graceful", false, "listen on fd open 3 (internal use only)")

    if gracefulChild {
        log.Print("main: Listening to existing file descriptor 3.")
        f := os.NewFile(3, "")  // 获取fd 3对应的文件对象
        l, err = net.FileListener(f) // 获取对应的listener
    } else {
        log.Print("main: Listening on a new file descriptor.")
        l, err = net.Listen("tcp", server.Addr)
    }
	if gracefulChild {
		parent := syscall.Getppid()
		log.Printf("main: Killing parent pid: %v", parent)
		syscall.Kill(parent, syscall.SIGTERM) // kill父进程
	}
	server.Serve(l)