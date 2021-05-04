package main

import (
	"fake-atom/bot"
	"fake-atom/config"
	_ "fake-atom/modules/autoReply"
	_ "fake-atom/modules/logging"
	"fake-atom/utils"
	"os"
	"os/signal"
)

func init() {
	utils.WriteLogToFS()
	config.Init()
	//bot.GenRandomDevice()
}

func main() {
	// 生成device.json
	bot.GenRandomDevice()
	
	// 快速初始化
	bot.Init()

	// 初始化 Modules
	bot.StartService()

	// 使用协议
	// 不同协议可能会有部分功能无法使用
	// 在登陆前切换协议
	bot.UseProtocol(bot.AndroidPhone)

	// 登录
	bot.Login()

	// 刷新好友列表，群列表
	bot.RefreshList()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, os.Kill)
	<-ch
	bot.Stop()
}
