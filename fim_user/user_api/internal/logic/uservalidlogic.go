package logic

import (
	"context"
	"errors"
	"fim/fim_user/user_models"

	"fim/fim_user/user_api/internal/svc"
	"fim/fim_user/user_api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserValidLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserValidLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserValidLogic {
	return &UserValidLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserValidLogic) UserValid(req *types.UserVaildRequest) (resp *types.UserVaildResponse, err error) {
	//判断是否为好友
	var friend user_models.FriendModel
	if friend.IsFriend(l.svcCtx.DB, req.UserID, req.FriendID) {
		return nil, errors.New("你们已经是好友了")
	}
	var userConf user_models.UserConfModel
	err = l.svcCtx.DB.Take(&userConf, "user_id=?", req.FriendID).Error
	if err != nil {
		return nil, errors.New("用户不存在")
	}
	resp = new(types.UserVaildResponse)
	resp.Verification = userConf.Verification
	switch userConf.Verification {
	case 0: //不允许任何验证
	case 1: //允许任何人添加
	case 2: //需要验证
	case 3, 4:
		if userConf.VerificationQuestion != nil {
			resp.VerificationQuestion = types.VerificationQuestion{
				Problem1: userConf.VerificationQuestion.Problem1,
				Problem2: userConf.VerificationQuestion.Problem2,
				Problem3: userConf.VerificationQuestion.Problem3,
				Answer1:  userConf.VerificationQuestion.Answer1,
				Answer2:  userConf.VerificationQuestion.Answer2,
			}
		}

	default:
	}
	return
}
