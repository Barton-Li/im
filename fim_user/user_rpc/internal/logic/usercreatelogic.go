package logic

import (
	"context"
	"errors"
	"fim/fim_user/user_models"

	"fim/fim_user/user_rpc/internal/svc"
	"fim/fim_user/user_rpc/types/user_rpc"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserCreateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUserCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserCreateLogic {
	return &UserCreateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// UserCreate 通过给定的用户信息创建新用户。
// 参数 in 包含待创建用户的详细信息，如昵称、头像、角色等。
// 返回创建成功的用户ID和可能的错误。
func (l *UserCreateLogic) UserCreate(in *user_rpc.UserCreateRequest) (*user_rpc.UserCreateResponse, error) {
	// 初始化一个 UserModel 对象，用于后续的数据操作。
	var user user_models.UserModel

	// 尝试根据 OpenID 查找已存在的用户，如果存在则更新，否则创建新用户。
	err := l.svcCtx.DB.Take(&user, "open_id =?", in.OpenId).Error
	if err != nil {
		// 如果查找失败，返回“用户未找到”的错误。
		return nil, errors.New("user not found")
	}

	// 更新用户信息，准备创建新用户。
	user = user_models.UserModel{
		Nickname:       in.NickName,
		Avatar:         in.Avatar,
		Role:           int8(in.Role),
		OpenID:         in.OpenId,
		RegisterSource: in.RegisterSource,
	}

	// 创建新用户。
	err = l.svcCtx.DB.Create(&user).Error
	if err != nil {
		// 如果创建失败，记录错误并返回“创建用户失败”的错误。
		logx.Error(err)
		return nil, errors.New("create user failed")
	}
	//创建用户配置
	l.svcCtx.DB.Create(&user_models.UserConfModel{
		UserID:        user.ID,
		RecallMessage: nil, //撤回消息提示内容，
		FriendOnline:  false,
		Sound:         true,
		SecureLink:    false,
		SavePwd:       false,
		SearchUser:    2, //通过id和昵称搜索用户
		Verification:  2, //需要验证消息

		Online: true,
	})
	// 返回创建成功的用户ID。
	return &user_rpc.UserCreateResponse{UserId: int32(user.ID)}, nil
}
