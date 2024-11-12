package logic

import (
	"context"
	"errors"
	"fim/fim_chat/chat_models"
	"fim/fim_user/user_rpc/types/user_rpc"

	"fim/fim_chat/chat_api/internal/svc"
	"fim/fim_chat/chat_api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserTopLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserTopLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserTopLogic {
	return &UserTopLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// UserTop 用户置顶功能
// req 为请求参数类型 UserTopRequest
// resp 为响应结果类型 UserToopResponse
// err 返回执行过程中遇到的错误（若有）
func (l *UserTopLogic) UserTop(req *types.UserTopRequest) (resp *types.UserToopResponse, err error) {
	if req.FriendID != req.FriendID { // 检查FriendID是否有效，此处疑似逻辑错误
		res, err := l.svcCtx.UserRpc.IsFriend(l.ctx, &user_rpc.IsFriendRequest{
			User1: uint32(req.FriendID),
			User2: uint32(req.UserID),
		})
		if err != nil {
			return nil, err
		}
		if !res.IsFriend {
			return nil, errors.New("你们还不是好友")
		}
	}
	// 查询数据库中是否存在置顶关系
	var topUser chat_models.TopUserModel
	err1 := l.svcCtx.DB.Take(&topUser, "user_id=? and top_user_id=?", req.UserID, req.FriendID).Error
	if err1 != nil { // 若不存在，则创建新的置顶关系
		l.svcCtx.DB.Create(&chat_models.TopUserModel{
			UserID:    req.UserID,
			TopUserID: req.FriendID,
		})
		return
	}
	// 若存在，则删除该置顶关系
	l.svcCtx.DB.Delete(&topUser)
	return
}
