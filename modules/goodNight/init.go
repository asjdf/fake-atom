package goodNight

import (
	"fake-atom/bot"
	"github.com/Mrs4s/MiraiGo/client"
	"github.com/Mrs4s/MiraiGo/message"
	"sync"
	"time"
)

func init() {
	instance = &module{}
	bot.RegisterModule(instance)
}

var instance *module

type module struct {
}

func (m *module) MiraiGoModule() bot.ModuleInfo {
	return bot.ModuleInfo{
		ID:       "atom.goodNight",
		Instance: instance,
	}
}

func (m *module) Init() {

}

func (m *module) PostInit() {

}

func (m *module) Serve(bot *bot.Bot) {
	bot.PrivateMessageEvent.Subscribe(func(client *client.QQClient, msg *message.PrivateMessage) {
		if msg.Sender.Uin == 243852814 && msg.ToString() == "晚安" {
			client.SendPrivateMessage(msg.Sender.Uin, message.NewSendingMessage().Append(message.NewText("晚安")))
		}
	})
}

func (m *module) Start(bot *bot.Bot) {
	time.Sleep(1 * time.Minute)
	bot.QQClient.SendPrivateMessage(243852814, message.NewSendingMessage().Append(message.NewText("晚安")))
}

func (m *module) Stop(bot *bot.Bot, wg *sync.WaitGroup) {
	defer wg.Done()
}
