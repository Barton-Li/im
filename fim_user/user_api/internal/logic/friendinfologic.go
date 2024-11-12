package logic

import (
	"context"
	"errors"
	"fim/fim_user/user_models"
	"fim/fim_user/user_rpc/types/user_rpc"

	"encoding/json"
	"fim/fim_user/user_api/internal/svc"
	"fim/fim_user/user_api/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type FriendInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewFriendInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FriendInfoLogic {
	return &FriendInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// FriendInfo 用于获取好友的信息。
// 通过传入的UserID和FriendID，首先检查两者是否为好友关系，
// 然后获取好友的详细信息，并返回给请求者。
// 参数:
//
//	req *types.FriendInfoRequest: 包含请求者ID和好友ID的请求结构体。
//
// 返回值:
//
//	*types.FriendInfoResponse: 包含好友信息的响应结构体。
//	error: 如果操作失败，返回相应的错误。
func (l *FriendInfoLogic) FriendInfo(req *types.FriendInfoRequest) (resp *types.FriendInfoResponse, err error) {
	// 检查请求者和好友是否为好友关系
	var friend user_models.FriendModel
	if friend.IsFriend(l.svcCtx.DB, req.UserID, req.FriendID) {
		// 如果不是好友，则返回错误
		return nil, errors.New("他人不是你的好友")
	}

	// 调用UserRpc服务，获取好友的用户信息
	res, err := l.svcCtx.UserRpc.UserInfo(l.ctx, &user_rpc.UserInfoRequest{
		UserId: uint32(req.FriendID),
	})
	if err != nil {
		// 如果调用RPC服务失败，则返回错误
		return nil, errors.New(err.Error())
	}

	// 解析RPC响应中的好友用户信息
	var friendUser user_models.UserModel
	json.Unmarshal(res.Data, &friendUser)

	// 构建并返回好友信息的响应
	response := types.FriendInfoResponse{
		UserID:   friendUser.ID,
		Nickname: friendUser.Nickname,
		Avatar:   friendUser.Avatar,
		Abstract: friendUser.Abstract,
		Notice:   friend.GetUserNotice(req.FriendID),
	}
	return &response, nil
}
