package main

import (
	"os"
	"os/signal"

	"fake-atom/bot"
	"fake-atom/config"
	"fake-atom/utils"

	_ "fake-atom/modules/goodNight"
	_ "fake-atom/modules/joinGroupVerify"
	_ "fake-atom/modules/logging"
)

func init() {
	utils.WriteLogToFS()
	config.Init()
}

func main() {
	// 快速初始化
	bot.Init()

	// 初始化 Modules
	bot.StartService()

	// 使用协议
	// 不同协议可能会有部分功能无法使用
	// 在登陆前切换协议
	bot.UseProtocol(bot.AndroidWatch)

	// 登录
	bot.Login()
	bot.SaveToken()

	// 刷新好友列表，群列表
	bot.RefreshList()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, os.Kill)
	<-ch
	bot.Stop()
}
