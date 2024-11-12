package logic

import (
	"context"
	"encoding/json"
	"errors"
	"fim/common/models/ctype"
	"fim/fim_group/group_models"
	user_models "fim/fim_user/user_models"
	"fim/fim_user/user_rpc/types/user_rpc"

	"fim/fim_group/group_api/internal/svc"
	"fim/fim_group/group_api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GroupValidAddLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGroupValidAddLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupValidAddLogic {
	return &GroupValidAddLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// GroupValidAdd 处理群组验证添加逻辑。
// req 是群组验证添加的请求数据。
// 返回响应数据和可能的错误。
func (l *GroupValidAddLogic) GroupValidAdd(req *types.GroupValidAddRequest) (resp *types.GroupValidAddResponse, err error) {
	// 通过用户ID获取用户信息。
	userInfo, err := l.svcCtx.UserRpc.UserInfo(l.ctx, &user_rpc.UserInfoRequest{
		UserId: uint32(req.UserID),
	})
	if err != nil {
		// 如果获取用户信息失败，记录错误并返回。
		logx.Error(err)
		return nil, errors.New("用户服务错误")
	}
	// 解析用户信息，检查用户是否被限制加入群聊。
	var userInfoModel user_models.UserModel
	json.Unmarshal(userInfo.Data, &userInfoModel)
	if userInfoModel.UserConfModel.CurtailInGroupChat {
		return nil, errors.New("用户被限制加入群聊")
	}
	// 检查用户是否已经在群组中。
	var member group_models.GroupMemberModel
	err = l.svcCtx.DB.Take(&member, "group_id =? AND user_id =?", req.GroupID, req.UserID).Error
	if err == nil {
		return nil, errors.New("用户已经在群聊中")
	}
	// 获取群组信息以进行后续验证。
	var group group_models.GroupModel
	err = l.svcCtx.DB.Take(&group, req.GroupID).Error
	if err != nil {
		return nil, errors.New("群聊不存在")
	}
	// 初始化响应对象。
	resp = new(types.GroupValidAddResponse)
	// 初始化群组验证模型。
	var verifyModel = group_models.GroupVerifyModel{
		GroupID:            req.GroupID,
		UserID:             req.UserID,
		Status:             0,
		AdditionalMessages: req.Verify,
		Type:               1, // 加群
	}
	// 根据群组的验证类型进行不同的处理。
	switch group.Verification {
	case 0:
		return nil, errors.New("群聊不允许加群")
	case 1:
		// 自动同意加入群组。
		verifyModel.Status = 1
		var groupMember = group_models.GroupMemberModel{
			GroupID: req.GroupID,
			UserID:  req.UserID,
			Role:    3,
		}
		l.svcCtx.DB.Create(&groupMember)
	case 2, 3:
		// 处理其他验证类型，如问题验证。
		if req.VerificationQuestion != nil {
			verifyModel.VerificationQuestion = &ctype.VerificationQuestion{
				Problem1: group.VerificationQuestion.Problem1,
				Problem2: group.VerificationQuestion.Problem2,
				Problem3: group.VerificationQuestion.Problem3,
				Answer1:  req.VerificationQuestion.Answer1,
				Answer2:  req.VerificationQuestion.Answer2,
				Answer3:  req.VerificationQuestion.Answer3,
			}
		}
	case 4:
		// 处理答案验证逻辑。
		if req.VerificationQuestion != nil && group.VerificationQuestion != nil {
			var count int
			if group.VerificationQuestion.Answer1 != nil && req.VerificationQuestion.Answer1 != nil {
				if *group.VerificationQuestion.Answer1 == *req.VerificationQuestion.Answer1 {
					count += 1
				}
			}
			if group.VerificationQuestion.Answer2 != nil && req.VerificationQuestion.Answer2 != nil {
				if *group.VerificationQuestion.Answer2 == *req.VerificationQuestion.Answer2 {
					count += 1
				}
			}
			if group.VerificationQuestion.Answer3 != nil && req.VerificationQuestion.Answer3 != nil {
				if *group.VerificationQuestion.Answer3 == *req.VerificationQuestion.Answer3 {
					count += 1
				}
			}
			if count != group.ProblemCount() {
				return nil, errors.New("答案错误")
			}
			// 答案正确，自动加群。
			verifyModel.Status = 1
			verifyModel.VerificationQuestion = group.VerificationQuestion
			var groupMember = group_models.GroupMemberModel{
				GroupID: req.GroupID,
				UserID:  req.UserID,
				Role:    3,
			}
			l.svcCtx.DB.Create(&groupMember)
		} else {
			return nil, errors.New("答案错误")
		}
	}
	// 创建验证记录。
	err = l.svcCtx.DB.Create(&verifyModel).Error
	if err != nil {
		return nil, err
	}
	return resp, nil
}
