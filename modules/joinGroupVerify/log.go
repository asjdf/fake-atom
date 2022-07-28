package joinGroupVerify

import (
	"github.com/Mrs4s/MiraiGo/client"
)

func logUserWantJoinGroup(joinReq *client.UserJoinGroupRequest) {
	logger.
		WithField("from", "UserWantJoinGroup").
		WithField("RequesterUin", joinReq.RequesterUin).
		WithField("RequestId", joinReq.RequestId).
		WithField("GroupCode", joinReq.GroupCode).
		Info(joinReq.Message)
}