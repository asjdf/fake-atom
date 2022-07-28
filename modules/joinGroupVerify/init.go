package joinGroupVerify

import (
	"fake-atom/bot"
	"fake-atom/config"
	"fake-atom/utils"
	"k8s.io/apimachinery/pkg/util/rand"
	"sync"
	"time"

	"github.com/hduhelp/api_open_sdk/transfer"
	"github.com/patrickmn/go-cache"
	"github.com/sirupsen/logrus"
)

func init() {
	instance = &module{verifyFLows{}}
	bot.RegisterModule(instance)
	logger = utils.GetModuleLogger(instance.MiraiGoModule().String())
}

type module struct {
	verifyFLows
}

var instance *module

var logger *logrus.Entry

var Cache *cache.Cache

func (m *module) MiraiGoModule() bot.ModuleInfo {
	return bot.ModuleInfo{
		ID:       "extend.joinGroupVerify",
		Instance: instance,
	}
}

func (m *module) Init() {
	transfer.Init(
		config.GlobalConfig.GetString("joinGroupVerify.appId"),
		config.GlobalConfig.GetString("joinGroupVerify.appKey"))
	Cache = cache.New(time.Minute*10, time.Hour*24)
	m.verifyFLows[551292799] = newVerifyChain(VerifyReject("本群已满 请加群714627058"))
	m.verifyFLows[726855751] = newVerifyChain(VerifyReject("本群已满 请加群714627058"))
	m.verifyFLows[714627058] = newVerifyChain(VerifyIsBan, VerifyByStuInfo())
	m.verifyFLows[477946200] = newVerifyChain(VerifyIsBan, VerifyByKshAndStuInfo())
}

func (m *module) PostInit() {

}

func (m *module) Serve(bot *bot.Bot) {
	bot.UserWantJoinGroupEvent.Subscribe(m.verifyHandler)
	bot.PrivateMessageEvent.Subscribe(banUserByUin)
}

func (m *module) Start(bot *bot.Bot) {
	for !bot.Online.Load() { // 等待登录成功
		time.Sleep(time.Second)
	}
	time.Sleep(8 * time.Second) // 等待列表更新完成
	for {
		if _, found := Cache.Get("UserWantJoinGroup-LastCheck"); !found { //只有在没有收到加群请求的情况下主动拉取，防止因为风控导致验证功能失效
			logger.Debug("fetching user want join group list")
			m.autoCheckJoinReq(bot)
		}
	}
}

func (m *module) Stop(bot *bot.Bot, wg *sync.WaitGroup) {
	defer wg.Done()
}

func (m *module) autoCheckJoinReq(bot *bot.Bot) {
	groupSystemMsg, err := bot.GetGroupSystemMessages()
	if err != nil {
		logger.Error("fetch group system message error: ", err)
		return
	}
	if groupSystemMsg.JoinRequests == nil {
		Cache.Set("UserWantJoinGroup-LastCheck", time.Now().Unix(), 20*time.Second)
		return
	}
	for _, v := range groupSystemMsg.JoinRequests {
		if !v.Checked {
			logger.Debug("checking join request from:", v.RequesterNick)
			m.verifyHandler(bot.QQClient, v)
			time.Sleep(time.Second * time.Duration(rand.IntnRange(5, 10)))
		}
	}
}
