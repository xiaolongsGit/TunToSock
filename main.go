package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"tuntosock/conf"
	"tuntosock/engine"
)

var (
	File bool = true
	key       = new(conf.Key)
)

func init() {
	flag.BoolVar(&File, "file", true, "是否使用配置文件(配置文件与exe在同一目录下)")
	flag.StringVar(&key.LogLevel, "loglevel", "info", "日志等级：[debug|info|warning|error|silent]")
	flag.StringVar(&key.DevName, "devname", "tun", "网卡名称")
	flag.IntVar(&key.DevMask, "devmask", 24, "设备掩码")
	flag.StringVar(&key.DevIP, "devip", "192.168.55.24", "设备IP地址")
	flag.BoolVar(&key.Client, "client", true, "是否开启客户端")
	flag.StringVar(&key.ClientIP, "clientip", "0.0.0.0", "客户端使用的地址")
	flag.IntVar(&key.ClientPort, "clientport", 11455, "客户端使用的端口")
	flag.BoolVar(&key.Server, "server", false, "是否开启服务端")
	flag.StringVar(&key.ServerIP, "serverip", "0.0.0.0", "服务端监听地址")
	flag.IntVar(&key.ServerPort, "serverport", 11451, "服务端监听端口")
	flag.StringVar(&key.RemoteServerIP, "remoteip", "127.0.0.1", "客户端的服务端地址")
	flag.IntVar(&key.RemoteServerPort, "remoteport", 11451, "客户端的服务端端口")
	flag.Parse()
}
func main() {

	if File {
		data, err := os.ReadFile("config.json")
		if err != nil {
			log.Fatalf("读取配置文件失败：%v", err)
		}
		if err = json.Unmarshal(data, key); err != nil {
			log.Fatalf("反序列化配置文件失败：%v", err)
		}
	}
	engine.Start(key)
	defer engine.Stop(key)
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
}
