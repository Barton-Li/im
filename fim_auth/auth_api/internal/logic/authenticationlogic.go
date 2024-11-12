package logic

import (
	"context"
	"errors"
	"fim/fim_auth/auth_api/internal/svc"
	"fim/fim_auth/auth_api/internal/types"
	"fim/utils"
	"fim/utils/jwts"
	"fmt"

	"github.com/zeromicro/go-zero/core/logx"
)

type AuthenticationLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAuthenticationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AuthenticationLogic {
	return &AuthenticationLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// Authentication 进行用户身份验证。
// 它首先检查请求的路径是否在白名单中，如果是，则直接返回成功响应。
// 如果请求的Token为空，则返回认证失败的错误。
// 对提供的Token进行解析，检查是否有效。如果解析失败，则返回认证失败的错误。
// 检查用户是否已登出。如果用户已登出，则返回认证失败的错误。
// 如果所有验证步骤都通过，则返回认证成功的响应，包括用户ID和角色信息。
//
// 参数:
//
//	req *types.AuthenticationRequest - 包含验证路径和Token的请求。
//
// 返回值:
//
//	*types.AuthenticationResponse - 包含用户ID和角色的响应。
//	error - 如果认证失败，则返回错误。
func (l *AuthenticationLogic) Authentication(req *types.AuthenticationRequest) (resp *types.AuthenticationResponse, err error) {
	// 检查请求的路径是否在白名单中，如果是，则直接返回成功响应。
	if utils.InListByRegex(l.svcCtx.Config.WhiteList, req.ValiPath) {
		logx.Infof("白名单访问:%s", req.ValiPath)
		return
	}

	// 如果请求的Token为空，则返回认证失败的错误。
	if req.Token == "" {
		err = errors.New("认证失败")
		return
	}

	// 解析提供的Token，检查是否有效。如果解析失败，则返回认证失败的错误。
	payload, err := jwts.ParseToken(req.Token, l.svcCtx.Config.Auth.AccessSecret)
	if err != nil {
		err = errors.New("认证失败")
		return
	}

	// 检查用户是否已登出。如果已登出，则返回认证失败的错误。
	_, err = l.svcCtx.Redis.Get(fmt.Sprintf("logout:%d", payload.UserID)).Result()
	if err == nil {
		logx.Error("用户已退出")
		err = errors.New("认证失败")
		return
	}

	// 如果所有验证步骤都通过，则准备并返回认证成功的响应，包括用户ID和角色信息。
	resp = &types.AuthenticationResponse{
		UserID: payload.UserID,
		Role:   int(payload.Role),
	}
	return resp, nil
}
