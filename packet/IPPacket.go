package packet

type IPPacket struct {
	Version  uint8
	HeadLen  int
	TotalLen int
	Protocol uint8 //TCP 6 UDP 17 ICMP 1 IGMP 2
	SRC      []byte
	DST      []byte
}

func UnpackIPPacket(data []byte) IPPacket {
	pro := IPPacket{}
	bs := data[0]
	pro.Version = bs >> 4
	pro.HeadLen = int(bs << 4 >> 4)
	pro.TotalLen = int(data[2])<<8 + int(data[3])
	pro.SRC = data[12:16]
	pro.DST = data[16:20]
	pro.Protocol = data[9]
	return pro
}
