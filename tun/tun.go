package tun

import (
	"fmt"
	"net/netip"
	"tuntosock/winipcfg"

	"golang.zx2c4.com/wireguard/tun"
)

type TUN struct {
	nt     *tun.NativeTun
	mtu    int
	name   string
	offset int

	rSizes []int
	rBuffs [][]byte
	wBuffs [][]byte
}

func NewTun(n string, m int) (*TUN, error) {
	dev, err := tun.CreateTUN(n, m)
	if err != nil {
		return nil, err
	}
	tundev := dev.(*tun.NativeTun)
	t := &TUN{
		name:   n,
		mtu:    m, //虚拟设备mtu
		offset: 0,
		rSizes: make([]int, 1),
		rBuffs: make([][]byte, 1),
		wBuffs: make([][]byte, 1),
		nt:     tundev,
	}
	return t, nil
}
func (t *TUN) CloseDevice() error {
	return t.nt.Close()
}
func (t *TUN) SetIPAndRoute(addr string, metric uint32, mask int) error {
	link := winipcfg.LUID(t.nt.LUID())
	prefix, err := netip.ParsePrefix(fmt.Sprintf("%s/%v", addr, mask))
	if err != nil {
		return err
	}
	err2 := link.SetIPAddresses([]netip.Prefix{prefix})
	if err2 != nil {
		return err2
	}
	return nil
}
func (t *TUN) DelRoute(addr string, mask int) error {
	link := winipcfg.LUID(t.nt.LUID())
	prefix, err := netip.ParsePrefix(fmt.Sprintf("%s/%v", addr, mask))
	if err != nil {
		return err
	}
	ip, err := netip.ParseAddr(addr)

	if err != nil {
		return err
	}
	err3 := link.DeleteRoute(prefix, ip)
	if err3 != nil {
		return nil
	}
	return nil
}
func (t *TUN) Read(packet []byte) (int, error) {
	t.rBuffs[0] = packet
	_, err := t.nt.Read(t.rBuffs, t.rSizes, t.offset)
	return t.rSizes[0], err
}

func (t *TUN) Write(packet []byte) (int, error) {
	t.wBuffs[0] = packet
	return t.nt.Write(t.wBuffs, t.offset)
}
