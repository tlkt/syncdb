package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type Dirr struct {
	Host      string
	Port      int
	RemoteDir string
	LocalDir  string
	UserName  string
	Password  string
	FileName  string
}

func main1() {
	f, err := ioutil.ReadFile("./config/app.json")
	if err != nil {
		fmt.Println("read err:", err)
	}

	data := make(map[string]Dirr)
	err = json.Unmarshal(f, &data)
	if err != nil {
		fmt.Println("JSON ERR:", err)
	}
	for k, v := range data {
		fmt.Println(k, v)
	}
}
