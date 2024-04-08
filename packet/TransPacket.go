package packet

import (
	"strconv"
	"strings"
)

var MaxLen = 1418

type TransPacket struct {
	TransType uint8 //1byte 包类型
	//200：客户端发送给服务器的数据
	//201：服务器发送给客户端的数据
	//202：客户端发送给服务器的路由登录
	//203：服务器回复登录成功
	//204: 服务器回复 你申请的IP已经被占用
	//205：客户端发送给服务器的心跳
	Len   int    //2byte 数据长度
	SRC   []byte //4byte 设备IP
	DST   []byte //4byte 目的地址 从IP数据的DST解析上来的
	EffIP []byte //4byte 有效位IP
	Bro   uint8  //1byte 是否是广播
	//1：广播数据
	//2：非广播数据
	Pro  uint8 //1byte 数据包协议
	Mask uint8 //1byte 掩码

	//以上 8项 18byte
	Data []byte //最大1400byte长度 数据
}

// siliconvn.dscloud.me
func PackPacket(data TransPacket) []byte {
	packet := make([]byte, 0)
	packet = append(packet, data.TransType)
	packet = append(packet, IntToTwo(data.Len)...)
	packet = append(packet, data.SRC...)
	packet = append(packet, data.DST...)
	packet = append(packet, data.EffIP...)
	packet = append(packet, data.Bro)
	packet = append(packet, data.Pro)
	packet = append(packet, data.Mask)
	packet = append(packet, data.Data...)
	return packet
}

func UnpackPacket(data []byte) TransPacket {
	pac := TransPacket{}
	pac.TransType = data[0]
	pac.Len = TwoToInt(data[1:3])
	pac.SRC = data[3:7]
	pac.DST = data[7:11]
	pac.EffIP = data[11:15]
	pac.Bro = data[15]
	pac.Pro = data[16]
	pac.Mask = data[17]
	pac.Data = data[18:]
	return pac
}
func PackIP(ip string) []byte {
	ipByte := make([]byte, 4)
	strs := strings.Split(ip, ".")
	if len(strs) != 4 {
		return ipByte
	}
	for i := 0; i < 4; i++ {
		num, _ := strconv.Atoi(strs[i])
		ipByte[i] = byte(num)
	}
	return ipByte
}
func UnpackIP(ipbyte []byte) string {
	if len(ipbyte) != 4 {
		return "0.0.0.0"
	}
	ipstr := strconv.Itoa(int(ipbyte[0]))
	ipstr += "."
	ipstr += strconv.Itoa(int(ipbyte[1]))
	ipstr += "."
	ipstr += strconv.Itoa(int(ipbyte[2]))
	ipstr += "."
	ipstr += strconv.Itoa(int(ipbyte[3]))
	return ipstr
}
func IntToTwo(i int) []byte {
	two := make([]byte, 0, 2)
	two = append(two, byte(i>>8))
	two = append(two, byte(i))
	return two
}
func TwoToInt(b []byte) int {
	return int(b[0])<<8 + int(b[1])
}
