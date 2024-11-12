package logic

import (
	"context"
	"encoding/json"
	"errors"
	"fim/fim_user/user_models"

	"fim/fim_user/user_rpc/internal/svc"
	"fim/fim_user/user_rpc/types/user_rpc"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUserInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserInfoLogic {
	return &UserInfoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// UserInfo 根据用户ID获取用户信息。
//
// 参数:
//   in - 包含用户ID的请求结构体。
// 返回值:
//   *user_rpc.UserInfoResponse - 包含用户信息的响应结构体。
//   error - 如果查询过程中出现错误，则返回错误信息。
func (l *UserInfoLogic) UserInfo(in *user_rpc.UserInfoRequest) (*user_rpc.UserInfoResponse, error) {
	// 初始化用户模型变量，用于存储查询到的用户数据。
	var user user_models.UserModel

	// 使用预加载查询用户信息，预加载UserConfModel数据，然后根据用户ID获取用户信息。
	// 如果查询过程中出现错误，则返回错误信息。
	err := l.svcCtx.DB.Preload("UserConfModel").Take(&user, in.UserId).Error
	if err != nil {
		// 如果查询到的用户不存在，则返回自定义的错误信息。
		return nil, errors.New("用户不存在")
	}

	// 将查询到的用户信息转换为JSON格式。
	byteData, _ := json.Marshal(user)

	// 返回包含用户信息的响应结构体。
	return &user_rpc.UserInfoResponse{Data: byteData}, nil
}

