package logic

import (
	"context"

	"fim/fim_user/user_rpc/internal/svc"
	"fim/fim_user/user_rpc/types/user_rpc"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserOnlineListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUserOnlineListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserOnlineListLogic {
	return &UserOnlineListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UserOnlineListLogic) UserOnlineList(in *user_rpc.UserOnlineListRequest) (*user_rpc.UserOnlineListResponse, error) {
	// todo: add your logic here and delete this line

	return &user_rpc.UserOnlineListResponse{}, nil
}
