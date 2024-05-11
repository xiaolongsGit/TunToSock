package sock

import (
	"net"
)

// 服务器：SRC->remote
type UDPTab struct {
	Remote net.Addr //远端真实地址
	SRC    string   //远端虚拟ip
	EffIP  string   //掩码有效位IP
}
