package engine

import (
	"tuntosock/conf"
	"tuntosock/log"
	"tuntosock/sock/client"
	"tuntosock/sock/server"
	"tuntosock/tun"
)

// siliconvn.dscloud.me
var (
	clientStartUp bool     = false
	tunPrt        *tun.TUN = nil
)

func Start(k *conf.Key) {
	level, err := log.ParseLevel(k.LogLevel)
	if err != nil {
		log.Errorf("设置日志等级失败:%v", err)
		return
	}
	log.SetLevel(level)
	if k.Server {
		log.Infof("正在开启服务端")
		if !server.StartUDPServer(k) {
			log.Infof("启动服务器失败，请退出程序")
		}
	}
	if k.Client {
		t, err := tun.NewTun(k.DevName, 1400)
		if err != nil {
			log.Errorf("创建虚拟网卡失败:%v", err)
			return
		}
		log.Infof("创建虚拟网卡成功")
		err2 := t.SetIPAndRoute(k.DevIP, 5, k.DevMask)
		if err2 != nil {
			log.Errorf("设置IP与路由失败:%v", err2)
			return
		}
		log.Infof("设置网卡路由成功")
		tunPrt = t
		log.Infof("正在开启客户端")
		clientStartUp = client.StartClient(t, k)
		if !clientStartUp {
			log.Infof("客户端启动失败，请退出程序")
		}
	}
}
func Stop(k *conf.Key) {
	if k.Client && clientStartUp {
		log.Infof("正在关闭客户端")
		if tunPrt != nil {
			err := tunPrt.CloseDevice()
			if err != nil {
				log.Errorf("关闭虚拟网卡失败:%v", err)
			}
			log.Infof("正在关闭虚拟网卡")
			err2 := tunPrt.DelRoute(k.DevIP, k.DevMask)
			if err2 != nil {
				log.Errorf("删除网卡路由失败:%v", err2)
			}
			log.Infof("正在删除网卡路由")
		}
	}
}
