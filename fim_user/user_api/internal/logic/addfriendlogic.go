package logic

import (
	"context"
	"errors"
	"fim/common/models/ctype"
	"fim/fim_user/user_models"

	"fim/fim_user/user_api/internal/svc"
	"fim/fim_user/user_api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AddFriendLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAddFriendLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddFriendLogic {
	return &AddFriendLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// AddFriend 尝试添加一个新朋友。这个方法检查用户是否被限制添加好友，双方是否已经是朋友，以及通过不同的验证机制来确认添加请求是否应该被批准。
// req 包含添加好友请求的详细信息，如请求者的ID，被请求者的ID和可能的验证信息。
// 返回一个添加好友的响应，包含操作的结果信息。
// 如果用户不存在、用户限制添加好友、双方已经是好友、验证失败或出现数据库错误，将返回错误。
func (l *AddFriendLogic) AddFriend(req *types.AddFriendRequest) (resp *types.AddFriendResponse, err error) {
	// 检查请求的用户是否存在，并且是否被限制添加好友
	//限制用户添加好友
	var conf user_models.UserConfModel
	err = l.svcCtx.DB.Take(&conf, "user_id = ?", req.UserID).Error
	if err != nil {
		return nil, errors.New("用户不存在")
	}
	if conf.CurtailAddUser {
		return nil, errors.New("用户限制添加好友")
	}

	// 检查请求者和被请求者是否已经是朋友
	var friend user_models.FriendModel
	if friend.IsFriend(l.svcCtx.DB, req.UserID, req.FriendID) {
		return nil, errors.New("你们已经是好友了")
	}

	// 检查被请求的用户是否存在
	var userConf user_models.UserConfModel
	err = l.svcCtx.DB.Take(&userConf, "user_id = ?", req.FriendID).Error
	if err != nil {
		return nil, errors.New("用户不存在")
	}

	// 初始化添加好友的响应
	resp = new(types.AddFriendResponse)

	// 根据用户的验证设置，处理添加好友的请求
	var verifyModel = user_models.FriendVerifyModel{
		SendUserID:         req.UserID,
		RevUserID:          req.FriendID,
		AdditionalMessages: req.Verify,
	}
	switch userConf.Verification {
	case 0:
		return nil, errors.New("该用户不允许任何人添加")
	case 1:
		//		允许任何人加好友
		//直接成为好友
		//先往验证表中加一条记录，然后通过
		verifyModel.RevStatus = 1
		var userFriend = user_models.FriendModel{
			SendUserID: req.UserID,
			RevUserID:  req.FriendID,
		}
		l.svcCtx.DB.Create(&userFriend)
	case 2:
		// 处理不需要验证的其他情况
	case 3:
		if req.VerificationQuestion != nil {
			verifyModel.VerificationQuestion = &ctype.VerificationQuestion{
				Problem1: req.VerificationQuestion.Problem1,
				Problem2: req.VerificationQuestion.Problem2,
				Problem3: req.VerificationQuestion.Problem3,
				Answer1:  req.VerificationQuestion.Answer1,
				Answer2:  req.VerificationQuestion.Answer2,
				Answer3:  req.VerificationQuestion.Answer3,
			}
		}
	case 4:
		if req.VerificationQuestion != nil && userConf.VerificationQuestion != nil {
			var count int
			// 验证答案是否正确
			if userConf.VerificationQuestion.Answer1 != nil && req.VerificationQuestion.Answer1 != nil {
				if *userConf.VerificationQuestion.Answer1 == *req.VerificationQuestion.Answer1 {
					count += 1
				}
			}
			if userConf.VerificationQuestion.Answer2 != nil && req.VerificationQuestion.Answer2 != nil {
				if *userConf.VerificationQuestion.Answer2 == *req.VerificationQuestion.Answer2 {
					count += 1
				}
			}
			if userConf.VerificationQuestion.Answer3 != nil && req.VerificationQuestion.Answer3 != nil {
				if *userConf.VerificationQuestion.Answer3 == *req.VerificationQuestion.Answer3 {
					count += 1
				}
			}
			// 检查答案的正确数量
			if count != userConf.ProblemCount() {
				return nil, errors.New("答案错误")
			}
			verifyModel.RevStatus = 1
			verifyModel.VerificationQuestion = userConf.VerificationQuestion
			var userFriend = user_models.FriendModel{
				SendUserID: req.UserID,
				RevUserID:  req.FriendID,
			}
			l.svcCtx.DB.Create(&userFriend)
		}
	default:
		return nil, errors.New("不支持的验证参数")
	}

	// 将验证信息记录到数据库
	err = l.svcCtx.DB.Create(&verifyModel).Error
	if err != nil {
		logx.Error(err)
		return nil, errors.New("添加好友失败")
	}

	return
}
