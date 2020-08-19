package config

import (
	"fmt"
	"github.com/widuu/goini"
	"os"
	"strings"
)

//LogInfo 日志配置
type LogInfo struct {
	Level     string
	IsConsole string
	Path      string
	Num       string
	Max       string
	FileName  string
	ID        string
	SockURL   string
}

//RedisCache redis缓存配置
type RedisCache struct {
	AccessToken string
	LastTime    int64
	IP          string
}

//Log 日志配置
var Log LogInfo
var LogKinds []string
var LocalDir string

//InitServer 服务器初始化配置
func InitServer() {
	conf := goini.SetConfig("config/app.ini")

	LocalDir = conf.GetValue("backup", "local_dir")
	//配置log
	logkinds_str := conf.GetValue("LogInfo", "logKinds")
	LogKinds = strings.Split(logkinds_str, ",")

	if len(LogKinds) == 1 && len(LogKinds[0]) == 0 {
		fmt.Println("日志种类未配置")
		os.Exit(1)
	}

	Log.FileName = conf.GetValue("LogInfo", "filename")
	Log.Level = conf.GetValue("LogInfo", "level")
	Log.ID = conf.GetValue("LogInfo", "id")
	Log.IsConsole = conf.GetValue("LogInfo", "isConsole")
	Log.Max = conf.GetValue("LogInfo", "max")
	Log.Num = conf.GetValue("LogInfo", "num")
	Log.Path = conf.GetValue("LogInfo", "path")
	Log.SockURL = conf.GetValue("LogInfo", "SockUrl")

}
