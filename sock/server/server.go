package server

import (
	"net"
	"tuntosock/conf"
	"tuntosock/log"
	"tuntosock/packet"
	"tuntosock/sock"
)

var (
	emptyData      = []byte("1")
	defaultIP      = packet.PackIP("127.0.0.1")
	server_UDPConn *net.UDPConn
	server_UDPTab  = make(map[string]sock.UDPTab)
	loginSucByte   = make([]byte, 0)
	occupiedByte   = make([]byte, 0)
)

// 初始化 把登录成功和被占用 回复初始化为byte
func init() {
	loginSuc := packet.TransPacket{
		TransType: 203,
		Len:       1,
		SRC:       defaultIP,
		DST:       defaultIP,
		EffIP:     defaultIP,
		Bro:       2,
		Pro:       20,
		Mask:      24,
		Data:      emptyData,
	}
	loginSucByte = packet.PackPacket(loginSuc)
	occpied := packet.TransPacket{
		TransType: 204,
		Len:       1,
		SRC:       defaultIP,
		DST:       defaultIP,
		EffIP:     defaultIP,
		Bro:       2,
		Pro:       20,
		Mask:      24,
		Data:      emptyData,
	}
	occupiedByte = packet.PackPacket(occpied)
}
func StartUDPServer(k *conf.Key) bool {
	UDPAddr := net.UDPAddr{IP: net.ParseIP(k.ServerIP), Port: k.ServerPort}
	UDPConn, err := net.ListenUDP("udp", &UDPAddr)
	if err != nil {
		log.Errorf("服务器监听UDP失败:%v", err)
		return false
	}
	log.Infof("服务器监听中...")
	server_UDPConn = UDPConn
	go serverUDP()
	return true
}
func serverUDP() {
	defer server_UDPConn.Close()
	for {
		data := make([]byte, 1500)
		len, remote, err := server_UDPConn.ReadFrom(data)
		//读取出错
		if err != nil {
			log.Errorf("服务器读取数据错误:%v", err)
			//忽略这个包
			continue
		}
		data = data[0:len]
		//解析自定义包
		trans := packet.UnpackPacket(data)
		//传输过来的设备IP
		srcStr := packet.UnpackIP(trans.SRC)
		//有效位DST
		effIPStr := packet.UnpackIP(trans.EffIP)
		//包目的地址
		dstStr := packet.UnpackIP(trans.DST)
		//无论是不是广播，都需要修改传输数据
		data[0] = 201
		log.Debugf("服务器 收到数据(type已修改):%v", data)
		switch trans.TransType {
		case 202:
			//登录
			newTab := sock.UDPTab{
				Remote: remote,
				SRC:    srcStr,
				EffIP:  effIPStr,
			}
			//获取源地址路由表
			srcTab, ok := server_UDPTab[srcStr]
			//存在且远端地址不相同（该路由已经存在，有新客户申请这个IP
			//因为UDP没有断开连接状态，只能抢占
			if ok && remote.String() != srcTab.Remote.String() {
				_, err := server_UDPConn.WriteTo(occupiedByte, srcTab.Remote)
				if err != nil {
					log.Errorf("服务器返回[IP被占用]发生错误:%v", err)
					continue
				}
			}
			//遍历表，查找是否存在同一个remote但是不同IP
			for k, v := range server_UDPTab {
				if v.Remote == remote {
					delete(server_UDPTab, k)
					break
				}
			}
			//插入新表
			server_UDPTab[srcStr] = newTab
			_, err := server_UDPConn.WriteTo(loginSucByte, remote)
			if err != nil {
				log.Errorf("服务器返回[登录成功]发生错误:%v", err)
				continue
			}
			log.Infof("虚拟IP:%v,掩码:%v,远程地址:%v  登录成功", srcStr, trans.Mask, remote)
		case 200:
			//传输数据
			if trans.Bro == 1 {
				for _, v := range server_UDPTab {
					if v.Remote.String() != remote.String() && v.EffIP == effIPStr {
						_, err := server_UDPConn.WriteTo(data, v.Remote)
						if err != nil {
							log.Errorf("服务器写数据出错:%v", err)
							continue
						}
					}
				}
			} else {
				value, ok := server_UDPTab[dstStr]
				if ok {
					_, err := server_UDPConn.WriteTo(data, value.Remote)
					if err != nil {
						log.Errorf("服务器写数据出错:%v", err)
						continue
					}
				}
			}
		}
	}
}
