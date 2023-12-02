package sock

import (
	"net"
)

type TableUDP struct {
	WLock chan int
	Dial  *net.UDPConn
}

// 客户端只有一个表 没有对应关系
type TableTCP struct {
	WLock chan int
	Dial  *net.TCPConn
}
type ServerTCPCon struct {
	Remote net.Addr
	Dial   net.Conn
	RLock  chan int
}

// 服务器：SRC->remote
type ServerTCP struct {
	Addr net.Addr //远端真实地址
	SRC  string   //远端虚拟ip
	Mask uint8    //远端虚拟ip掩码
	Dial net.Conn
	Lock chan int
}
type ServerUDP struct {
	SRC  string
	Addr net.Addr
}
