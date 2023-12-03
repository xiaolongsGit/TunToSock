package client

import (
	"math"
	"net"
	"time"
	"tuntosock/conf"
	"tuntosock/log"
	"tuntosock/packet"
	"tuntosock/tun"
)

var (
	emptyData                        = []byte("1")
	client_defaultMask               = uint8(24)
	client_defauleDevIP              = ""
	client_TCP          *net.TCPConn = nil
)

func StartClient(t *tun.TUN, k *conf.Key) bool {
	ns, err := net.LookupHost(k.RemoteServerIP)
	if err != nil {
		log.Errorf("客户端-获取远程IP地址失败")
		return false
	}
	k.RemoteServerIP = ns[0]
	client_defauleDevIP = k.DevIP
	client_defaultMask = uint8(k.DevMask)
	TCPLocal := net.TCPAddr{IP: net.ParseIP(k.ClientIP), Port: k.ClientPort}
	TCPRemote := net.TCPAddr{IP: net.ParseIP(k.RemoteServerIP), Port: k.RemoteServerPort}

	ok, tcp, err := clientTryTCP(&TCPLocal, &TCPRemote)
	if !ok {
		log.Errorf("客户端-尝试TCP连接服务器失败:%v", err)
		return false
	}
	client_TCP = tcp
	go clientWrite(t)
	go clientTCPReadHandle(t)
	return true
}
func clientTryTCP(local, remote *net.TCPAddr) (bool, *net.TCPConn, error) {
	log.Infof("客户端正在尝试TCP连接%v", remote)
	tcp, err := net.DialTCP("tcp", local, remote)
	if err != nil {
		log.Errorf("客户端尝试TCP连接失败:%v", err)
		return false, nil, err
	}
	login := packet.TransPacket{
		TransType: 2,
		Len:       1,
		SRC:       packet.PackIP(client_defauleDevIP),
		DST:       packet.PackIP("127.0.0.1"),
		Bro:       2,
		Pro:       20,
		Mask:      client_defaultMask,
		Data:      emptyData,
	}
	_, err = tcp.Write(packet.PackPacket(login))
	if err != nil {
		log.Errorf("客户端尝试TCP连接失败:%v", err)
		return false, nil, err
	}
	retData := make([]byte, 14)
	_, err = tcp.Read(retData)
	if err != nil {
		log.Errorf("客户端尝试TCP连接失败:%v", err)
		return false, nil, err
	}
	retData = append(retData, byte(1))
	packet := packet.UnpackPacket(retData)
	log.Debugf("客户端尝试TCP 收到数据：%v", packet)
	if packet.TransType == 3 {
		log.Infof("客户端尝试TCP连接成功")
		data := make([]byte, packet.Len)
		tcp.Read(data)
		return true, tcp, nil
	}
	if packet.TransType == 4 {
		log.Infof("该IP已被占用,请尝试其他IP")
	}
	return false, nil, err
}

func clientTCPWriteHandle(data []byte, len int, ipinfo packet.IPPacket) bool {
	log.Debugf("客户端-TCP数据长度:%v", len)
	if len > 1400 {
		len = 1400
	}
	pac := packet.TransPacket{
		TransType: 6,
		Len:       len,
		SRC:       packet.PackIP(client_defauleDevIP),
		DST:       ipinfo.DST,
		Bro:       2,
		Pro:       ipinfo.Protocol,
		Mask:      client_defaultMask,
		Data:      data[0:len],
	}
	if IsBroadcast(ipinfo.DST, client_defaultMask) {
		pac.Bro = 1
	}
	log.Debugf("客户端-TCP数据:%v", pac)
	_, err := client_TCP.Write(packet.PackPacket(pac))
	if err != nil {
		log.Errorf("客户端-TCP发数据给服务器错误(连接或许已经关闭):%v", err)
		return false
	}
	return true
}
func clientTCPReadHandle(t *tun.TUN) {
	for {
		receive := make([]byte, 14)
		_, err := client_TCP.Read(receive)
		log.Debugf("客户端-TCP读取到数据(头部):%v", receive)
		if err != nil {
			log.Errorf("客户端-TCP读取数据错误,请关闭程序:%v", err)
			return
		}
		receive = append(receive, byte(1))
		receivePac := packet.UnpackPacket(receive)
		buf := make([]byte, receivePac.Len)
		_, err = client_TCP.Read(buf)
		if err != nil {
			log.Errorf("客户端-TCP读取数据错误,请关闭程序:%v", err)
			return
		}
		_, err = t.Write(buf)
		if err != nil {
			log.Errorf("客户端-TCP写入设备失败,请关闭程序:%v", err)
			return
		}
	}
}
func clientWrite(t *tun.TUN) {
	go heartBeat()
	for {
		data := make([]byte, 1400)
		n, err := t.Read(data)
		if err != nil {
			log.Errorf("客户端-虚拟设备读取出错:%v", err)
			return
		}
		pac := packet.UnpackIPPacket(data)
		if pac.Version != 4 {
			continue
		}
		if !clientTCPWriteHandle(data, n, pac) {
			log.Infof("客户端写入失败")
			return
		}
	}
}
func heartBeat() {
	hb := packet.TransPacket{
		TransType: 5,
		Len:       1,
		SRC:       packet.PackIP(client_defauleDevIP),
		DST:       packet.PackIP("127.0.0.1"),
		Bro:       2,
		Pro:       20,
		Mask:      client_defaultMask,
		Data:      emptyData,
	}
	hbByte := packet.PackPacket(hb)
	for {
		time.Sleep(time.Second * 10)
		_, err := client_TCP.Write(hbByte)
		if err != nil {
			log.Errorf("客户端-发送心跳数据出错:%v", err)
			return
		}
	}
}
func IsBroadcast(dst net.IP, mask uint8) bool {
	//判断是否是广播信息
	result := mask / 8
	remain := mask % 8
	if result > 4 {
		return false
	}
	bronum := 0
	for i := result; i < 4; i++ {
		if i == result {
			bronum = int(dst[i] << remain >> remain)
		} else {
			bronum = bronum*256 + int(dst[i])
		}
	}
	//2的n次方减1
	return bronum == int(math.Pow(2, float64(32-mask)))-1
}
