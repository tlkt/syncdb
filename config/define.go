package config

type Dir struct {
	Host      string //主机ip
	Port      int    //sftp端口
	RemoteDir string //远程地址
	LocalDir  string //保存本地地址
	UserName  string //用户名
	Password  string //密码
	FileName  string //备份文件区别名称
}
