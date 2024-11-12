package logic

import (
	"context"
	"errors"
	"fim/fim_user/user_models"
	"fim/fim_user/user_rpc/types/user_rpc"

	"fim/fim_user/user_api/internal/svc"
	"fim/fim_user/user_api/internal/types"

	"encoding/json"
	"github.com/zeromicro/go-zero/core/logx"
)

type UserInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserInfoLogic {
	return &UserInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// UserInfo 是一个处理用户信息请求的方法，它接收一个 *types.UserInfoRequest 结构体作为参数，
// 并返回一个 *types.UserInfoResponse 结构体和一个错误。
// 如果在处理过程中发生错误，错误将通过第二个返回值返回。
// 方法首先调用 UserRpc 的 UserInfo RPC 方法，然后将响应数据反序列化为 UserModel。
// 如果反序列化失败，方法将记录错误并返回一个带有 "数据错误" 消息的自定义错误。
// 最后，方法将 UserModel 的字段填充到响应结构体中，并根据需要创建 VerificationQuestion 字段。
func (l *UserInfoLogic) UserInfo(req *types.UserInfoRequest) (resp *types.UserInfoResponse, err error) {
	// 调用服务上下文中的 UserRpc，并传入背景上下文和用户ID
	res, err := l.svcCtx.UserRpc.UserInfo(context.Background(), &user_rpc.UserInfoRequest{
		UserId: uint32(req.UserID),
	})
	if err != nil {
		return nil, err // 如果RPC调用出错，返回错误
	}

	// 反序列化响应数据到 UserModel
	var user user_models.UserModel
	err = json.Unmarshal(res.Data, &user)
	if err != nil {
		logx.Error(err)                // 记录错误
		return nil, errors.New("数据错误") // 返回数据错误
	}

	// 填充响应结构体
	resp = &types.UserInfoResponse{
		UserID:        user.ID,
		Nickname:      user.Nickname,
		Abstract:      user.Abstract,
		Avatar:        user.Avatar,
		RecallMessage: user.UserConfModel.RecallMessage,
		FriendOnline:  user.UserConfModel.FriendOnline,
		Sound:         user.UserConfModel.Sound,
		SearchUser:    user.UserConfModel.SearchUser,
		SavePwd:       user.UserConfModel.SavePwd,
		Verification:  user.UserConfModel.Verification,
	}

	// 如果存在验证问题，创建并填充 VerificationQuestion 字段
	if user.UserConfModel.VerificationQuestion != nil {
		resp.VerificationQuestion = &types.VerificationQuestion{
			Problem1: user.UserConfModel.VerificationQuestion.Problem1,
			Problem2: user.UserConfModel.VerificationQuestion.Problem2,
			Problem3: user.UserConfModel.VerificationQuestion.Problem3,

			Answer1: user.UserConfModel.VerificationQuestion.Answer1,
			Answer2: user.UserConfModel.VerificationQuestion.Answer2,
			Answer3: user.UserConfModel.VerificationQuestion.Answer3,
		}
	}

	return resp, nil // 返回响应和nil错误
}
