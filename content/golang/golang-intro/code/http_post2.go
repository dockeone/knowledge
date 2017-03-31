package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

var NOCURL = "http://noc.ksyun.com/interface/rest?"

type IDC struct {
	Id    string `json:"id"`
	Name  string `json:"name"`
	CName string `json:"cname"`
}

func main() {
	idcData := url.Values{
		"user":         {"zhangjun3"},
		"access_token": {"25f005c25859ea67f0ff54ded752e3e7"},
		"tablename":    {"idc_info"}, "headers": {"id,name,cname"}, "size": {"-1"},
	}
	req, _ := http.NewRequest("POST", NOCURL, strings.NewReader(idcData.Encode()))
	req.Header.Set("X-Auth-Token", "SEC-1512")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()
	buf := new(bytes.Buffer)
	if _, err = io.Copy(buf, res.Body); err != nil {
		return
	}
	var idcInfo = struct {
		Data []IDC `json:"data"`
	}{}
	if err = json.Unmarshal(buf.Bytes(), &idcInfo); err == nil {
		fmt.Printf("%v\n", idcInfo)
	}
}
