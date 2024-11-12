package logic

import (
	"context"
	"errors"
	"fim/fim_user/user_models"

	"fim/fim_user/user_api/internal/svc"
	"fim/fim_user/user_api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type FriendNoticeUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewFriendNoticeUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FriendNoticeUpdateLogic {
	return &FriendNoticeUpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// FriendNoticeUpdate 用于更新好友通知设置。
// 当用户A更新其对用户B的通知设置时，此逻辑用于验证A和B是否为好友，并根据需要更新通知设置。
// 参数:
//
//	req.FriendNoticeUpdateRequest: 包含用户ID、好友ID和新的通知设置。
//
// 返回值:
//
//	FriendNoticeUpdateResponse: 更新后的响应数据，如果未更新则可能为空。
//	error: 如果操作失败，返回错误信息。
func (l *FriendNoticeUpdateLogic) FriendNoticeUpdate(req *types.FriendNoticeUpdateRequest) (resp *types.FriendNoticeUpdateResponse, err error) {
	// 加载好友信息以验证用户关系并进行后续更新。
	var friend user_models.FriendModel
	// 验证请求中的用户是否为好友关系。
	if !friend.IsFriend(l.svcCtx.DB, req.UserID, req.FriendID) {
		// 如果不是好友，则返回错误。
		return nil, errors.New("他还不是你的好友")
	}
	// 如果是发送方更新通知设置。
	if friend.SendUserID == req.UserID {
		// 如果新的通知设置与现有设置相同，则无需更新。
		if friend.SenUserNotice == req.Notice {
			return
		}
		// 更新发送方的通知设置。
		l.svcCtx.DB.Model(&friend).Update("sen_user_notice", req.Notice)
	}
	// 如果是接收方更新通知设置。
	if friend.RevUserID == req.UserID {
		// 如果新的通知设置与现有设置相同，则无需更新。
		if friend.RevUserNotice == req.Notice {
			return
		}
		// 更新接收方的通知设置。
		l.svcCtx.DB.Model(&friend).Update("rev_user_notice", req.Notice)
	}
	// 返回更新后的响应，如果未更新则响应可能为空。
	return
}
