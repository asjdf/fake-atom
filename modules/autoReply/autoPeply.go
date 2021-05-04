package autoReply

import (
	"fake-atom/bot"
	"fmt"
	"github.com/Mrs4s/MiraiGo/client"
	"github.com/Mrs4s/MiraiGo/message"
	"sync"
)

func init() {
	instance = &replyer{}
	bot.RegisterModule(instance)
}

func (m *replyer) MiraiGoModule() bot.ModuleInfo {
	return bot.ModuleInfo{
		ID: "atom.autoReply",
		Instance: instance,
	}
}

var instance *replyer

type replyer struct {
}

func (m *replyer) Init() {
	//panic("implement me")
}

func (m *replyer) PostInit() {
	//panic("implement me")
}

func (m *replyer) Serve(bot *bot.Bot) {
	registerReplayer(bot)
}

func (m *replyer) Start(bot *bot.Bot) {
	//panic("implement me")
}

func (m *replyer) Stop(bot *bot.Bot, wg *sync.WaitGroup) {
	defer wg.Done()
}

func registerReplayer(b *bot.Bot) {
	b.OnPrivateMessage(func(qqClient *client.QQClient, message *message.PrivateMessage) {
		fmt.Println(message.ToString())
	})
}

