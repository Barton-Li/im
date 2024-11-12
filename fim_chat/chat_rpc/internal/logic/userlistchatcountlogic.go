package logic

import (
	"context"

	"fim/fim_chat/chat_rpc/internal/svc"
	"fim/fim_chat/chat_rpc/types/chat_rpc"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserListChatCountLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUserListChatCountLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserListChatCountLogic {
	return &UserListChatCountLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UserListChatCountLogic) UserListChatCount(in *chat_rpc.UserListChatCountRequest) (*chat_rpc.UserListChatCountResponse, error) {
	// todo: add your logic here and delete this line

	return &chat_rpc.UserListChatCountResponse{}, nil
}
