package logic

import (
	"context"
	"errors"
	auth_models "fim/fim_auth/auth_models"
	"fim/fim_user/user_rpc/types/user_rpc"
	"fim/utils/jwts"
	"fim/utils/open_login"

	"fim/fim_auth/auth_api/internal/svc"
	"fim/fim_auth/auth_api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type Open_loginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewOpen_loginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *Open_loginLogic {
	return &Open_loginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// Open_login 实现了开放登录逻辑。
// 根据请求中的标志（例如QQ），它会尝试使用相应的开放平台进行登录验证。
// 如果用户已存在，它将返回登录令牌；如果用户不存在，它将创建新用户并返回登录令牌。
func (l *Open_loginLogic) Open_login(req *types.OpenLoginRequest) (resp *types.LoginResponse, err error) {
	// 定义开放平台用户信息的结构体
	type OpenInfo struct {
		Nickname string
		OpenID   string
		Avatar   string
	}
	// 初始化开放平台用户信息
	var info OpenInfo

	// 根据请求中的标志来处理不同的开放登录方式
	switch req.Flag {
	case "qq":
		// 使用QQ登录配置和请求中的代码尝试进行QQ登录
		qqInfo, openErr := open_login.NewQQLogin(open_login.QQConfig{
			AppID:    l.svcCtx.Config.QQ.AppID,
			AppKey:   l.svcCtx.Config.QQ.AppKey,
			Redirect: l.svcCtx.Config.QQ.Redirect,
		}, req.Code)
		// 如果QQ登录失败，则返回错误
		info = OpenInfo{
			OpenID:   qqInfo.OpenID,
			Nickname: qqInfo.Nickname,
			Avatar:   qqInfo.Avator,
		}
		err = openErr
	default:
		// 如果标志不是预期的QQ，则返回标志错误
		err = errors.New("flag error")
	}

	// 如果在此处出现任何错误，则记录错误并返回开放登录错误
	if err != nil {
		logx.Error(err)
		return nil, errors.New("open login error")
	}

	// 尝试根据开放ID查找已存在的用户
	var user auth_models.UserModel
	err = l.svcCtx.DB.Take(&user, "open_id=?", info.OpenID).Error
	// 如果用户不存在，则创建新用户
	if err != nil {
		// 调用用户创建RPC接口来创建新用户
		res, err := l.svcCtx.UserRpc.UserCreate(context.Background(), &user_rpc.UserCreateRequest{
			NickName:       info.Nickname,
			Password:       "",
			Role:           2,
			Avatar:         info.Avatar,
			OpenId:         info.OpenID,
			RegisterSource: "qq",
		})
		// 如果创建用户失败，则记录错误并返回注册错误
		if err != nil {
			logx.Error(err)
			return nil, errors.New("register error")
		}
		// 更新本地用户变量以反映新创建的用户
		user.Model.ID = uint(res.UserId)
		user.Role = 2
		user.Nickname = info.Nickname
	}

	// 生成登录令牌
	// 登录逻辑
	token, err1 := jwts.GenerateToken(jwts.JwtPayLoad{
		UserID:   user.ID,
		NickName: user.Nickname,
		Role:     user.Role,
	}, l.svcCtx.Config.Auth.AccessSecret, l.svcCtx.Config.Auth.AccessExpire)
	// 如果生成令牌失败，则记录错误并返回生成令牌错误
	if err1 != nil {
		logx.Error(err1)
		err = errors.New("generate token error")
		return nil, err1
	}

	// 返回登录响应，包含生成的令牌
	return &types.LoginResponse{Token: token}, nil

	return
}
