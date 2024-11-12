package logic

import (
	"context"
	"encoding/json"
	"errors"
	"fim/common/models/ctype"
	"fim/fim_chat/chat_rpc/chat"
	"fim/fim_user/user_api/internal/svc"
	"fim/fim_user/user_api/internal/types"
	"fim/fim_user/user_models"

	"github.com/zeromicro/go-zero/core/logx"
)

type ValidStatusLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewValidStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ValidStatusLogic {
	return &ValidStatusLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ValidStatusLogic) ValidStatus(req *types.FriendValidStatusRequest) (resp *types.FriendValidStatusResponse, err error) {
	var friendVerify user_models.FriendVerifyModel
	err = l.svcCtx.DB.Take(&friendVerify, "id=?", req.VerifyID).Error
	if err != nil {
		return nil, errors.New("验证记录不存在")
	}
	if friendVerify.SendUserID == friendVerify.RevUserID && friendVerify.RevUserID == req.UserID {
		switch req.Status {
		case 1, 2, 3:
			if friendVerify.RevStatus != 0 {
				return nil, errors.New("不可变更状态")
			}
		case 4:
			if friendVerify.RevStatus == 0 {
				return nil, errors.New("不可删除未处理状态")
			}
		default:
			return nil, errors.New("接收方状态错误")
		}
	} else if friendVerify.SendUserID == req.UserID {
		switch req.Status {
		case 4:
			if friendVerify.RevStatus == 0 {
				return nil, errors.New("接收方未处理，不可删除")
			}
		default:
			return nil, errors.New("发送方状态错误")
		}
	} else {
		switch req.Status {
		case 1, 2, 3:
			if friendVerify.SendStatus != 0 {
				return nil, errors.New("不可变更状态")
			}
		case 4:
			if friendVerify.SendStatus == 0 {
				return nil, errors.New("不可删除未处理状态")
			}
		default:
			return nil, errors.New("接收方状态错误")
		}
	}
	switch req.Status {
	case 1: //同意
		friendVerify.RevStatus = 1
		l.svcCtx.DB.Create(&user_models.FriendModel{
			SendUserID: friendVerify.SendUserID,
			RevUserID:  friendVerify.RevUserID,
		})
		msg := ctype.Msg{
			Type: ctype.TextMsgType,
			TextMsg: &ctype.TextMsg{
				Content: "我们已经是好友了，开始聊天吧！",
			},
		}
		byteData, _ := json.Marshal(msg)
		_, err = l.svcCtx.ChatRpc.UserChat(l.ctx, &chat.UserChatRequest{
			SnedUserId: uint32(friendVerify.SendUserID),
			RevUserId:  uint32(friendVerify.RevUserID),
			Msg:        byteData,
			SystemMsg:  nil,
		})
		if err != nil {
			logx.Error(err)
		}

	case 2: //拒绝
		friendVerify.RevStatus = 2
	case 3: //忽略
		friendVerify.RevStatus = 3
	case 4: //删除
		if friendVerify.SendUserID == req.UserID {
			friendVerify.SendStatus = 4
		} else {
			friendVerify.RevStatus = 4
		}

	}
	l.svcCtx.DB.Save(&friendVerify)
	return

}
