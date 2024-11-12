package logic

import (
	"context"
	"errors"
	"fim/fim_auth/auth_api/internal/svc"
	"fim/fim_auth/auth_api/internal/types"
	auth_models "fim/fim_auth/auth_models"
	"fim/utils/jwts"
	"github.com/zeromicro/go-zero/core/logx"
)

type LoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// Login 执行用户登录操作。
//
// 参数:
// req - 包含登录请求信息的类型为`types.LoginRequest`的指针。
//
// 返回值:
// resp - 成功登录时返回的类型为`types.LoginResponse`的指针，包含生成的访问令牌。
// err  - 登录过程中遇到的任何错误。
func (l *LoginLogic) Login(req *types.LoginRequest) (resp *types.LoginResponse, err error) {
	var user auth_models.UserModel
	// 从数据库中查找与用户名匹配的用户记录
	err = l.svcCtx.DB.Take(&user, "id=?", req.UserName).Error
	if err != nil {
		// 如果查找失败，设置错误信息并返回
		err = errors.New("用户名或密码错误")
		return
	}

	// 生成JWT令牌
	token, err := jwts.GenerateToken(jwts.JwtPayLoad{
		UserID:   user.ID,
		NickName: user.Nickname,
		Role:     user.Role,
	}, l.svcCtx.Config.Auth.AccessSecret, l.svcCtx.Config.Auth.AccessExpire)
	if err != nil {
		// 如果生成令牌失败，记录错误信息并返回通用服务内部错误
		logx.Error(err)
		err = errors.New("服务内部错误")
		return
	}
	// 登录成功，返回生成的令牌
	return &types.LoginResponse{
		Token: token,
	}, nil
}
