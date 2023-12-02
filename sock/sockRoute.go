package sock

import (
	"net"
)

// 服务器：SRC->remote
type ServerTCP struct {
	Addr net.Addr //远端真实地址
	SRC  string   //远端虚拟ip
	Mask uint8    //远端虚拟ip掩码
	Dial net.Conn
}
