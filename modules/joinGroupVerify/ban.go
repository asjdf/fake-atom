package joinGroupVerify

import (
	"fake-atom/config"
	"fmt"
	"github.com/Mrs4s/MiraiGo/client"
	"github.com/Mrs4s/MiraiGo/message"
	"math/rand"
	"regexp"
	"strconv"
	"time"
)

func banUserByUin(client *client.QQClient, privateMessage *message.PrivateMessage) {
	commands := regexp.MustCompile("@ban ([0-9]+)").FindStringSubmatch(privateMessage.ToString())
	if len(commands) != 0 {
		isAdmin := false
		for _, v := range config.GlobalConfig.GetIntSlice("joinGroupVerify.adminList") {
			if int64(v) == privateMessage.Sender.Uin {
				isAdmin = true
				break
			}
		}
		if !isAdmin {
			return
		}

		go func() {
			time.Sleep(time.Duration(0.5+rand.Float64()) * time.Second)
			client.SendPrivateMessage(privateMessage.Sender.Uin,
				message.NewSendingMessage().Append(message.NewText("收到")))
		}()

		banList := config.GlobalConfig.GetIntSlice("joinGroupVerify.banList")
		banUin, err := strconv.Atoi(commands[1])
		if err != nil {
			return
		}

		inBanList := false
		for _, v := range banList {
			if v == banUin {
				inBanList = true
				break
			}
		}
		if inBanList {
			client.SendPrivateMessage(privateMessage.Sender.Uin,
				message.NewSendingMessage().Append(message.NewText(fmt.Sprintf("%v 已在黑名单中", banUin))))
			return
		}

		banList = append(banList, banUin)
		config.GlobalConfig.Set("joinGroupVerify.banList", banList)
		config.GlobalConfig.WriteConfig()

		time.Sleep(time.Duration(2.0+rand.Float64()*2.0) * time.Second)
		client.SendPrivateMessage(privateMessage.Sender.Uin,
			message.NewSendingMessage().Append(message.NewText(fmt.Sprintf("%v 已加入黑名单", banUin))))

		//从现有的群翻出来然后踢出
		for _, groupInfo := range client.GroupList {
			if groupInfo.AdministratorOrOwner() {
				if banUserInfo := groupInfo.FindMemberWithoutLock(int64(banUin)); banUserInfo != nil {
					banUserInfo.Kick("您已进入本群黑名单", false)
				}
			}
		}
	}
}
