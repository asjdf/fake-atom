package joinGroupVerify

import (
	"fake-atom/config"
	"fmt"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/rand"
	"math"
	"regexp"
	"strconv"
	"time"

	"fake-atom/pkg/student"

	"github.com/Mrs4s/MiraiGo/client"
	"github.com/Mrs4s/MiraiGo/message"
)

type verifyFLows map[int64]VerifyChain

type VerifyFunc func(v *verifyFlow)
type VerifyChain []VerifyFunc

type verifyFlow struct {
	Client  *client.QQClient
	Index   int
	JoinReq *client.UserJoinGroupRequest
	Handles VerifyChain
}

const abortIndex int = math.MaxInt / 2

func (f *verifyFlow) Next() {
	f.Index++
	for f.Index < len(f.Handles) {
		f.Handles[f.Index](f)
		f.Index++
	}
}

func (f *verifyFlow) Abort() {
	f.Index = abortIndex
}

func (m *module) verifyHandler(qqClient *client.QQClient, joinReq *client.UserJoinGroupRequest) {
	Cache.Set("UserWantJoinGroup-LastCheck", time.Now().Unix(), 20*time.Second) // 如果收到加群消息就意味着没有被风控，不用主动拉群加群请求
	if joinReq.Checked {
		return
	}
	time.Sleep(time.Second * time.Duration(rand.IntnRange(5, 10)))
	if handles, ok := m.verifyFLows[joinReq.GroupCode]; ok {
		logger.Info("收到 ", joinReq.RequesterNick, " ", joinReq.RequesterUin, " 的加群请求")
		verifyFlow := &verifyFlow{Client: qqClient, JoinReq: joinReq, Handles: handles, Index: -1}
		verifyFlow.Next()
	}
}

func newVerifyChain(f ...VerifyFunc) (c VerifyChain) {
	c = append(c, f...)
	return
}

func VerifyReject(reason string) func(f *verifyFlow) {
	return func(f *verifyFlow) {
		f.Client.SolveGroupJoinRequest(f.JoinReq, false, false, reason)
		logger.Info("拒绝了 ", f.JoinReq.RequesterNick, f.JoinReq.Message, " 加群 ", f.JoinReq.GroupName)
		f.Abort()
	}
}

// VerifyIsBan 检查是否在总黑名单内
func VerifyIsBan(f *verifyFlow) {
	banList := config.GlobalConfig.GetIntSlice("joinGroupVerify.banList")
	for _, v := range banList {
		if f.JoinReq.RequesterUin == int64(v) {
			f.Client.SendGroupMessage(config.GlobalConfig.GetInt64("joinGroupVerify.adminGroup"),
				message.NewSendingMessage().Append(message.NewText("黑名单用户加群 "+f.JoinReq.GroupName+"：")).
					Append(message.NewText(fmt.Sprintf("(%s:%d) %s", f.JoinReq.RequesterNick, f.JoinReq.RequesterUin, f.JoinReq.Message))))
			// f.Client.SolveGroupJoinRequest(f.JoinReq, false, false, "你在本群黑名单中")
			return
		}
	}
}

// VerifyByStuInfo 使用姓名-学号验证
func VerifyByStuInfo() func(f *verifyFlow) {
	answerReg := regexp.MustCompile("答案：(.*)$")
	staffIdReg := regexp.MustCompile("[0-9].*$")
	nameReg := regexp.MustCompile("^[\u4e00-\u9fa5·•.]*")
	return func(f *verifyFlow) {
		answer := ""
		if r := answerReg.FindStringSubmatch(f.JoinReq.Message); len(r) == 2 {
			answer = r[1]
		} else {
			f.Abort()
			return
		}
		stuffID := staffIdReg.FindString(answer)
		if stuffID == "" {
			f.Client.SolveGroupJoinRequest(f.JoinReq, false, false, "没有检测到您的学号，请确认格式：姓名+学号")
			f.Abort()
			return
		}
		realName := nameReg.FindString(answer)
		if realName == "" {
			f.Client.SolveGroupJoinRequest(f.JoinReq, false, false, "没有检测到您的姓名，请确认格式：姓名+学号")
			f.Abort()
			return
		}

		pass, err := verifyByStuffID(realName, stuffID)

		// 高风险账号处理 此类账号无法通过
		if f.JoinReq.Suspicious == true {
			logger.Info("出现高风险账号加群请求：", f.JoinReq.RequesterNick)
			if _, found := Cache.Get("UserWantJoinGroup-HighRisk-" + strconv.FormatInt(f.JoinReq.RequestId, 10)); !found {
				msg := ""
				if err != nil {
					msg = fmt.Sprintf("出现高风险账号：%v 加群(%v)请求，请手动处理，%v", f.JoinReq.RequesterNick, f.JoinReq.GroupCode, err.Error())
				} else {
					if pass {
						msg = fmt.Sprintf("出现高风险账号：%v 加群(%v)请求，请手动处理，审核通过", f.JoinReq.RequesterNick, f.JoinReq.GroupCode)
					} else {
						msg = fmt.Sprintf("出现高风险账号：%v 加群(%v)请求，请手动处理，审核未通过", f.JoinReq.RequesterNick, f.JoinReq.GroupCode)
					}
				}
				f.Client.SendGroupMessage(config.GlobalConfig.GetInt64("joinGroupVerify.adminGroup"),
					message.NewSendingMessage().Append(message.NewText(msg)))
				Cache.Set("UserWantJoinGroup-HighRisk-"+strconv.FormatInt(f.JoinReq.RequestId, 10), true, 15*time.Minute)
			}
			f.Abort()
			return
		}
		if err != nil {
			f.Client.SolveGroupJoinRequest(f.JoinReq, pass, false, err.Error())
			f.Abort()
			return
		} else {
			if !pass {
				f.Client.SolveGroupJoinRequest(f.JoinReq, false, false, "信息错误 审核未通过")
				return
			}
			f.Client.SolveGroupJoinRequest(f.JoinReq, pass, false, "")
			go func() {
				defer func() {
					if err := recover(); err != nil {
						logger.Error("加群审核处理出错：", err)
					}
				}()
				time.Sleep(2 * time.Second)
				studentInfo, err := student.GetInfo(stuffID)
				if err == nil {
					memberInfo, _ := f.Client.GetMemberInfo(f.JoinReq.GroupCode, f.JoinReq.RequesterUin)
					unitName := regexp.MustCompile("^.*?(?:学院)").FindString(studentInfo.UnitName)
					memberInfo.EditCard(fmt.Sprintf("%v-%v-%v", stuffID[:2], unitName, realName))
				}
			}()
		}
	}
}

func VerifyByKshAndStuInfo() func(f *verifyFlow) {
	answerReg := regexp.MustCompile("答案：(.*)$")
	staffIdReg := regexp.MustCompile("[0-9]+")
	nameReg := regexp.MustCompile("[\u4e00-\u9fa5·•.]+")
	return func(f *verifyFlow) {
		answer := ""
		if r := answerReg.FindStringSubmatch(f.JoinReq.Message); len(r) == 2 {
			answer = r[1]
		} else {
			f.Abort()
			return
		}
		stuffID := staffIdReg.FindString(answer)
		if stuffID == "" {
			f.Client.SolveGroupJoinRequest(f.JoinReq, false, false, "没有检测到您的学号/考生号")
			f.Abort()
			return
		}
		realName := nameReg.FindString(answer)
		if realName == "" {
			f.Client.SolveGroupJoinRequest(f.JoinReq, false, false, "没有检测到您的姓名")
			f.Abort()
			return
		}

		// 检查考生号是否存在
		passKSH, info, kshErr := verifyByKSH(realName, stuffID)
		if passKSH {
			logger.Info("通过考生号加群请求：", f.JoinReq.RequesterNick)
			f.Client.SolveGroupJoinRequest(f.JoinReq, true, false, "")
			go func() {
				defer func() {
					if err := recover(); err != nil {
						logger.Error("加群审核处理出错：", err)
					}
				}()
				time.Sleep(5 * time.Second)
				spec := info.Spec
				if s, ok := specShort[info.Spec]; ok {
					spec = s
				}
				memberInfo, _ := f.Client.GetMemberInfo(f.JoinReq.GroupCode, f.JoinReq.RequesterUin)
				memberInfo.EditCard(fmt.Sprintf("%s-%s", spec, info.Name))
			}()
			f.Abort()
			return
		}
		if kshErr != nil {
			logger.Error("检查考生号出错：", kshErr, "，请求信息：", f.JoinReq, "，答案：", answer)
			if errors.Is(kshErr, errors.New("network error")) {
				f.Client.SolveGroupJoinRequest(f.JoinReq, false, false, "校验时发生网络错误，请稍后重试")
				f.Abort()
				return
			}
		}

		// 检查学/工号是否存在
		passStuff, staffErr := verifyByStuffID(realName, stuffID)
		if passStuff {
			logger.Info("通过学/工号加群请求：", f.JoinReq.RequesterNick)
			f.Client.SolveGroupJoinRequest(f.JoinReq, true, false, "")
			go func() {
				defer func() {
					if err := recover(); err != nil {
						logger.Error("加群审核处理出错：", err)
					}
				}()
				time.Sleep(2 * time.Second)
				studentInfo, err := student.GetInfo(stuffID)
				if err == nil {
					memberInfo, _ := f.Client.GetMemberInfo(f.JoinReq.GroupCode, f.JoinReq.RequesterUin)
					unitName := regexp.MustCompile("^.*?(?:学院)").FindString(studentInfo.UnitName)
					memberInfo.EditCard(fmt.Sprintf("%v-%v-%v", stuffID[:2], unitName, realName))
				}
			}()
			f.Abort()
			return
		}
		if staffErr != nil {
			logger.Error("检查学/工号出错：", staffErr, "，请求信息：", f.JoinReq, "，答案：", answer)
			f.Client.SolveGroupJoinRequest(f.JoinReq, false, false, staffErr.Error())
			f.Abort()
			return
		}

		pass := passKSH || passStuff

		// 高风险账号处理 此类账号无法通过
		if f.JoinReq.Suspicious == true {
			logger.Info("出现高风险账号加群请求：", f.JoinReq.RequesterNick)
			if _, found := Cache.Get("UserWantJoinGroup-HighRisk-" + strconv.FormatInt(f.JoinReq.RequestId, 10)); !found {
				msg := ""
				if pass {
					msg = fmt.Sprintf("出现高风险账号：%v 加群(%v)请求，请手动处理，审核通过", f.JoinReq.RequesterNick, f.JoinReq.GroupCode)
				} else {
					msg = fmt.Sprintf("出现高风险账号：%v 加群(%v)请求，请手动处理，审核未通过\n 考生号校验结果：%v\n 学工号校验结果：%v", f.JoinReq.RequesterNick, f.JoinReq.GroupCode, kshErr.Error(), staffErr.Error())
				}
				f.Client.SendGroupMessage(config.GlobalConfig.GetInt64("joinGroupVerify.adminGroup"),
					message.NewSendingMessage().Append(message.NewText(msg)))
				Cache.Set("UserWantJoinGroup-HighRisk-"+strconv.FormatInt(f.JoinReq.RequestId, 10), true, 15*time.Minute)
			}
			f.Abort()
			return
		}

		if !pass && kshErr == nil && staffErr == nil {
			logger.Info("不通过加群请求：", f.JoinReq.RequesterNick)
			f.Client.SolveGroupJoinRequest(f.JoinReq, false, false, "验证失败，请检查信息是否正确")
			f.Abort()
			return
		}
	}
}
