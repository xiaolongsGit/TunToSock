package server

import (
	"net"
	"tuntosock/conf"
	"tuntosock/log"
	"tuntosock/packet"
	"tuntosock/sock"
)

var (
	emptyData          = []byte("1")
	server_TCPListener *net.TCPListener
	server_TCPTable    = make(map[string]*sock.ServerTCP, 0)
	loginSucByte       = make([]byte, 0)
	occupiedByte       = make([]byte, 0)
)

func init() {
	loginSuc := packet.TransPacket{
		TransType: 3,
		Len:       1,
		SRC:       packet.PackIP("127.0.0.1"),
		DST:       packet.PackIP("127.0.0.1"),
		Bro:       2,
		Pro:       20,
		Mask:      0,
		Data:      emptyData,
	}
	loginSucByte = packet.PackPacket(loginSuc)
	occpied := packet.TransPacket{
		TransType: 4,
		Len:       1,
		SRC:       packet.PackIP("127.0.0.1"),
		DST:       packet.PackIP("127.0.0.1"),
		Bro:       2,
		Pro:       20,
		Mask:      0,
		Data:      emptyData,
	}
	occupiedByte = packet.PackPacket(occpied)
}
func StartServer(k *conf.Key) bool {
	TCPAddr := net.TCPAddr{IP: net.ParseIP(k.ServerIP), Port: k.ServerPort}
	TCPListener, err := net.ListenTCP("tcp", &TCPAddr)
	if err != nil {
		log.Errorf("服务器监听TCP失败:%v", err)
		return false
	}
	log.Infof("服务器TCP启动成功")
	server_TCPListener = TCPListener
	go serverTCP()
	return true
}
func serverTCP() {
	defer server_TCPListener.Close()
	for {
		acc, err := server_TCPListener.Accept()
		if err != nil {
			log.Errorf("服务器-TCP Accept失败:%v", err)
		}
		go serverTCPHandle(acc)
	}
}
func serverTCPHandle(con net.Conn) {
	defer func() {
		err := recover()
		if err != nil {
			log.Errorf("发生意外错误:%v", err)
		}
	}()
	defer con.Close()
	//远端地址
	remote := con.RemoteAddr()
	src := ""
	for {
		//读取数据
		head := make([]byte, 14)
		_, err := con.Read(head)
		if err != nil {
			log.Errorf("服务器TCP读取数据错误(可能对方主动退出):%v", err)
			if src != "" {
				delete(server_TCPTable, src)
				log.Infof("删除了虚拟地址:%v 路由", src)
			}
			return
		}
		head = append(head, byte(1))
		//解析自定义包
		trans := packet.UnpackPacket(head)
		//传输过来的设备IP
		srcStr := packet.UnpackIP(trans.SRC)
		src = srcStr
		//包目的地址
		dstStr := packet.UnpackIP(trans.DST)
		//接收数据
		data := make([]byte, trans.Len)
		_, err = con.Read(data)
		if err != nil {
			log.Errorf("服务器TCP读取数据错误(可能对方主动退出):%v", err)
			if src != "" {
				delete(server_TCPTable, src)
				log.Infof("删除了虚拟地址:%v 路由", src)
			}
			return
		}
		//无论是不是广播，都需要修改传输数据
		transcopy := packet.TransPacket{
			TransType: 1,
			Len:       trans.Len,
			SRC:       trans.SRC,
			DST:       trans.DST,
			Bro:       trans.Bro,
			Pro:       trans.Pro,
			Mask:      trans.Mask,
			Data:      data,
		}
		transByte := packet.PackPacket(transcopy)
		log.Debugf("服务器TCP收到数据(type已修改):%v", transcopy)
		_, ok := server_TCPTable[srcStr]
		log.Debugf("服务器TCP 源地址:%v 是否找到路由表:%v", srcStr, ok)
		switch trans.TransType {
		case 2:
			if ok {
				_, err := con.Write(occupiedByte)
				if err != nil {
					log.Errorf("服务器TCP返回[IP被占用]发生错误:%v", err)
				}
				return
			}
			server_TCPTable[srcStr] = &sock.ServerTCP{
				Addr: remote,
				SRC:  srcStr,
				Mask: trans.Mask,
				Dial: con,
			}
			_, err := con.Write(loginSucByte)
			if err != nil {
				log.Errorf("服务器-TCP返回[登录成功]发生错误:%v", err)
			}
			log.Infof("虚拟IP:%v,掩码:%v,远程地址:%v  登录成功", srcStr, trans.Mask, remote)
		case 6:
			if trans.Bro == 1 {
				for _, v := range server_TCPTable {
					if v.Addr != remote && v.Mask == trans.Mask {
						_, err := v.Dial.Write(transByte)
						if err != nil {
							log.Errorf("服务器TCP写数据出错:%v", err)
						}
					}
				}
			} else {
				tab, ok := server_TCPTable[dstStr]
				if ok {
					_, err := tab.Dial.Write(transByte)
					if err != nil {
						log.Errorf("服务器TCP写数据出错:%v", err)
					}
				}
			}
		}
	}
}
