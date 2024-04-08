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
	client_defaultMask               = 24
	client_defauleDevIP              = ""
	client_UDP          *net.UDPConn = nil
)

func StartClient(t *tun.TUN, k *conf.Key) bool {
	ns, err := net.LookupHost(k.RemoteServerIP)
	if err != nil {
		log.Errorf("客户端-获取远程IP地址失败")
		return false
	}
	k.RemoteServerIP = ns[0]
	client_defauleDevIP = k.DevIP
	client_defaultMask = k.DevMask
	UDPLocal := net.UDPAddr{IP: net.ParseIP(k.ClientIP), Port: k.ClientPort}
	UDPRemote := net.UDPAddr{IP: net.ParseIP(k.RemoteServerIP), Port: k.RemoteServerPort}

	ok, udp, err := clientTryUDP(&UDPLocal, &UDPRemote)
	if !ok {
		log.Errorf("客户端-尝试连接服务器失败:%v", err)
		return false
	}
	client_UDP = udp
	go clientWrite(t)
	go clientRead(t)
	return true
}
func clientTryUDP(local, remote *net.UDPAddr) (bool, *net.UDPConn, error) {
	log.Infof("客户端正在尝试连接%v", remote)
	udp, err := net.DialUDP("udp", local, remote)
	if err != nil {
		log.Errorf("客户端尝试连接失败(1):%v", err)
		return false, nil, err
	}
	_, effIP := IsBroadcast(packet.PackIP(client_defauleDevIP), client_defaultMask)
	login := packet.TransPacket{
		TransType: 202,
		Len:       1,
		SRC:       packet.PackIP(client_defauleDevIP),
		DST:       packet.PackIP("127.0.0.1"),
		EffIP:     effIP,
		Bro:       2,
		Pro:       20,
		Mask:      uint8(client_defaultMask),
		Data:      emptyData,
	}
	_, err = udp.Write(packet.PackPacket(login))
	if err != nil {
		log.Errorf("客户端尝试连接失败(2):%v", err)
		return false, nil, err
	}
	data := make([]byte, 1500)
	len, err := udp.Read(data)
	if err != nil {
		log.Errorf("客户端尝试连接失败(3):%v", err)
		return false, nil, err
	}
	effData := data[0:len]
	pak := packet.UnpackPacket(effData)
	log.Debugf("服务器返回数据：%v", pak)
	if pak.TransType == 203 {
		log.Infof("客户连接服务器成功")
		return true, udp, nil
	}
	return false, nil, nil
}

func clientWriteHandle(data []byte, len int, ipinfo packet.IPPacket) bool {
	log.Debugf("客户端-数据长度:%v", len)
	if len > 1400 {
		len = 1400
	}
	pac := packet.TransPacket{
		TransType: 200,
		Len:       len,
		SRC:       ipinfo.SRC,
		DST:       ipinfo.DST,
		Pro:       ipinfo.Protocol,
		Mask:      uint8(client_defaultMask),
		Data:      data[0:len],
	}
	bro, effIP := IsBroadcast(ipinfo.DST, client_defaultMask)
	pac.Bro = bro
	pac.EffIP = effIP
	log.Debugf("客户端-数据:%v", pac)
	_, err := client_UDP.Write(packet.PackPacket(pac))
	if err != nil {
		log.Errorf("客户端-发数据给服务器错误(连接或许已经关闭):%v", err)
		return false
	}
	return true
}
func clientRead(t *tun.TUN) {
	for {
		receive := make([]byte, 1500)
		len, err := client_UDP.Read(receive)
		if err != nil {
			log.Errorf("客户端-读取数据错误,请关闭程序:%v", err)
			return
		}
		effReceive := receive[0:len]
		receivePac := packet.UnpackPacket(effReceive)
		buf := receivePac.Data
		if receivePac.TransType == 201 {
			_, err = t.Write(buf)
			if err != nil {
				log.Errorf("客户端-写入设备失败,请关闭程序:%v", err)
				return
			}
		}
		if receivePac.TransType == 204 {
			client_UDP.Close()
			log.Infof("该IP被另一人占用,请使用其他IP")
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
		if !clientWriteHandle(data, n, pac) {
			log.Infof("客户端-写入失败")
			return
		}
	}
}
func heartBeat() {
	hb := packet.TransPacket{
		TransType: 205,
		Len:       1,
		SRC:       packet.PackIP(client_defauleDevIP),
		DST:       packet.PackIP("127.0.0.1"),
		EffIP:     packet.PackIP("127.0.0.1"),
		Bro:       2,
		Pro:       20,
		Mask:      uint8(client_defaultMask),
		Data:      emptyData,
	}
	hbByte := packet.PackPacket(hb)
	for {
		time.Sleep(time.Second * 10)
		_, err := client_UDP.Write(hbByte)
		if err != nil {
			log.Errorf("客户端-发送心跳数据出错:%v", err)
			return
		}
	}
}
func IsBroadcast(dst net.IP, mask int) (uint8, net.IP) {
	//判断是否是广播信息
	result := mask / 8
	remain := mask % 8
	if result > 4 {
		return 2, dst
	}
	effIP := make([]byte, 4)
	bronum := 0
	for i := 0; i < 4; i++ {
		if i < result {
			//全广播阻拦
			//向目的地址网段广播
			effIP[i] = dst[i]
		}
		if i == result {
			effIP[i] = dst[i] >> (8 - remain) << (8 - remain)
			bronum = int(dst[i] << remain >> remain)
		}
		if i > result {
			effIP[i] = 0
			bronum = bronum*256 + int(dst[i])
		}
	}
	//2的n次方减1
	if bronum == int(math.Pow(2, float64(32-mask)))-1 {
		return 1, effIP
	}
	return 2, effIP
}
