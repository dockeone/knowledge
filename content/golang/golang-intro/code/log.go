package main
import (
    "log"; "os"
)
func main(){
    arr := []int {2,3}
    log.Printf("Printf array with item [%d,%d]\n",arr[0],arr[1])

    logFile,err  := os.Create("/tmp/log_demo.log")
    if err != nil {log.Fatalln("open file error !")}
    defer logFile.Close()

    debugLog := log.New(logFile,"[Debug]",log.Llongfile)
    debugLog.Println("A debug message here")
    debugLog.SetPrefix("[Info]")
    debugLog.Println("A Info Message here ")
    debugLog.SetFlags(debugLog.Flags() | log.LstdFlags)
    debugLog.Println("A different prefix")
}
