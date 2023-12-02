package conf

type Key struct {
	LogLevel string
	//设备信息
	DevName string //设备名
	DevMask int    //掩码长度
	DevIP   string //设备IP地址

	//作为客户端时
	Client     bool
	ClientIP   string //客户端IP
	ClientPort int    //客户端端口
	//作为服务器时
	Server     bool   //是否开启服务器
	ServerIP   string //服务端IP
	ServerPort int    //服务端端口
	//远程服务器信息
	RemoteServerIP   string //远程服务器IP
	RemoteServerPort int    //远程服务器端口
}
