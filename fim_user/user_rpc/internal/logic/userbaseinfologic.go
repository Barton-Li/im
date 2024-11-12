package logic

import (
	"context"

	"fim/fim_user/user_rpc/internal/svc"
	"fim/fim_user/user_rpc/types/user_rpc"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserBaseInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUserBaseInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserBaseInfoLogic {
	return &UserBaseInfoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UserBaseInfoLogic) UserBaseInfo(in *user_rpc.UserBaseInfoRequest) (*user_rpc.UserBaseInfoResponse, error) {
	// todo: add your logic here and delete this line

	return &user_rpc.UserBaseInfoResponse{}, nil
}
