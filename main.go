package main

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/sftp"
	"github.com/robfig/cron"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path"
	"runtime/debug"
	"strconv"
	"strings"
	. "syncdb/config"
	"syncdb/logger"
	"time"

	_ "github.com/tim1020/godaemon"
)

func main() {
	a := make(chan int)
	initt()

	pathList, err := readJson()
	if err != nil {
		logger.Error("app", fmt.Sprintln("读取json文件出错 err:", err))
		return
	}

	for k, v := range pathList {
		go sync(k, v)
	}

	cr := cron.New()
	cr.AddFunc("0 27 11,18 * * *", func() {
		pathList, err := readJson()
		if err != nil {
			logger.Error("app", fmt.Sprintln("读取json文件出错 err:", err))
			return
		}
		for k, v := range pathList {
			go sync(k, v)
		}
	})
	cr.Start()

	<-a
}

func sync(k string, dir Dir) {
	var (
		err        error
		sftpClient *sftp.Client
	)
	dir.LocalDir = fmt.Sprintf("%s%s", LocalDir, dir.LocalDir)
	sftpClient, err = connect(dir.UserName, dir.Password, dir.Host, dir.Port)
	if err != nil {
		logger.Error("app", fmt.Sprintf("创建连接出错 err:%v", err))
		return
	}
	defer sftpClient.Close()

	fileList, err := sftpClient.ReadDir(dir.RemoteDir)
	if err != nil {
		logger.Error("app", fmt.Sprintf("读取远程文件列表出错 k:%s err:%v", k, err))
		return
	}
	var lastFile os.FileInfo
	for _, v := range fileList {
		if v.IsDir() {
			continue
		}
		if !strings.Contains(v.Name(), dir.FileName) {
			continue
		}
		if lastFile == nil {
			lastFile = v
			continue
		}
		if !lastFile.ModTime().After(v.ModTime()) {
			lastFile = v
		}
	}
	logger.Info("app", fmt.Sprintf("服务器%s 文件%s", k, lastFile.Name()))
	// 用来测试的远程文件路径 和 本地文件夹
	if lastFile == nil {
		logger.Error("app", fmt.Sprintf("服务器 %s 路径:%s 文件不存在", k, dir.RemoteDir))
		return
	}
	var remoteFilePath = fmt.Sprintf("%s%s", dir.RemoteDir, lastFile.Name())
	var localDir = dir.LocalDir
	//logger.Info("app", fmt.Sprintf(localDir))
	_, err = os.Stat(localDir[:len(localDir)-1])
	if os.IsNotExist(err) {
		err = os.Mkdir(localDir[:len(localDir)-1], 0777)
		if err != nil {
			log.Fatalf("创建日志文件夹出错 %v dir:%s", err, localDir)
		}
	}

	srcFile, err := sftpClient.Open(remoteFilePath)
	if err != nil {
		logger.Error("app", fmt.Sprintf("开启远程文件夹出错 err:%v", err))
		return
	}
	defer srcFile.Close()

	var localFileName = path.Base(remoteFilePath)
	dstFile, err := os.Create(path.Join(localDir, localFileName))
	if err != nil {
		logger.Error("app", fmt.Sprintf("创建本地文件出错 err:%v", err))
		return
	}
	defer dstFile.Close()

	if _, err = srcFile.WriteTo(dstFile); err != nil {
		logger.Error("app", fmt.Sprintf("写入本地文件出错 err:%v", err))
		return
	}
	logger.Info("app", fmt.Sprintf("同步远程服务器文件成功 %s", k))
}

func initt() {
	defer PacnicStack()
	InitServer()
	//设置日志前台可见
	consoleFlag := false
	if strings.EqualFold(Log.IsConsole, "1") {
		consoleFlag = true
	}
	logger.SetConsole(consoleFlag)

	//根据日志文件设置日志等级
	loglevel := logger.ERROR
	switch Log.Level {
	case "Debug":
		loglevel = logger.DEBUG
	case "Info":
		loglevel = logger.INFO
	case "Warn":
		loglevel = logger.WARN
	case "Fatal":
		loglevel = logger.FATAL
	}
	logger.SetLevel(loglevel)
	//根据配置文件设置　日志路径，日志名，日志切割大小限制
	num, err := strconv.Atoi(Log.Num)
	if err != nil {
		num = 10
	}

	max, err := strconv.Atoi(Log.Max)
	if err != nil {
		max = 50
	}

	logger.SetRollingFile(Log.Path, int32(num), int64(max), logger.MB)
}

func connect(user, password, host string, port int) (*sftp.Client, error) {
	var (
		auth         []ssh.AuthMethod
		addr         string
		clientConfig *ssh.ClientConfig
		sshClient    *ssh.Client
		sftpClient   *sftp.Client
		err          error
	)
	// get auth method
	auth = make([]ssh.AuthMethod, 0)
	auth = append(auth, ssh.Password(password))

	clientConfig = &ssh.ClientConfig{
		User:    user,
		Auth:    auth,
		Timeout: 30 * time.Second,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}

	// connet to ssh
	addr = fmt.Sprintf("%s:%d", host, port)

	if sshClient, err = ssh.Dial("tcp", addr, clientConfig); err != nil {
		logger.Error("app", fmt.Sprintf("连接远程服务器 创建ssh连接出错 %v", err))
		return nil, err
	}

	// create sftp client
	if sftpClient, err = sftp.NewClient(sshClient); err != nil {
		logger.Error("app", fmt.Sprintf("连接远程服务器 创建sftp客户端 %v", err))
		return nil, err
	}

	return sftpClient, nil
}

//PacnicStack 崩溃日志
func PacnicStack() {
	if err := recover(); err != nil {
		str := string(debug.Stack()) + "\n--------------------------------------\n"
		str += fmt.Sprintln(err) + "\n"
		file, err := os.OpenFile(Log.Path+"/crash.log", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			fmt.Printf("%v %v", Log, err.Error())
			return
		}
		defer file.Close()
		_, err = file.WriteString(str)
		if err != nil {
			logger.Error(err.Error())
			return
		}
		fmt.Println("stack := ", str)
		logger.Error("app", fmt.Sprintf("panic %v", str))
		return
	}
}

func readJson() (map[string]Dir, error) {
	pathList := make(map[string]Dir)
	f, err := ioutil.ReadFile("./config/app.json")
	if err != nil {
		logger.Error("app", fmt.Sprintln("读取json文件出错 err:", err))
		return pathList, err
	}

	err = json.Unmarshal(f, &pathList)
	if err != nil {
		logger.Error("app", fmt.Sprintln("解析json文件出错 err:", err))
		return pathList, err
	}
	return pathList, nil
}
