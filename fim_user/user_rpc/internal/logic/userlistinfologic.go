package logic

import (
	"context"

	"fim/fim_user/user_rpc/internal/svc"
	"fim/fim_user/user_rpc/types/user_rpc"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserListInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUserListInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserListInfoLogic {
	return &UserListInfoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UserListInfoLogic) UserListInfo(in *user_rpc.UserListInfoRequest) (*user_rpc.UserListInfoResponse, error) {
	// todo: add your logic here and delete this line

	return &user_rpc.UserListInfoResponse{}, nil
}
