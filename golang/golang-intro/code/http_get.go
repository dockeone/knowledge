package main

import (
  "fmt"
  "net/http"
  "io/ioutil"
)

func main(){
    resp, _ := http.Get("http://10.160.109.143/api/ebsinfo?idc=BJLG")
    defer resp.Body.Close()
    body, _ := ioutil.ReadAll(resp.Body)
    fmt.Printf("%v\n", string(body))
}
