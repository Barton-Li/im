package logic

import (
	"context"
	"errors"
	"fim/fim_group/group_models"

	"fim/fim_group/group_api/internal/svc"
	"fim/fim_group/group_api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GroupValidLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGroupValidLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupValidLogic {
	return &GroupValidLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GroupValidLogic) GroupValid(req *types.GroupValidRequest) (resp *types.GroupValidResponse, err error) {
	var member group_models.GroupMemberModel
	err = l.svcCtx.DB.Take(&member, "group_id=?and user_id=?", req.GroupID, req.UserID).Error
	if err == nil {
		return nil, errors.New("user already in group")
	}
	var group group_models.GroupModel
	err = l.svcCtx.DB.Take(&group, req.GroupID).Error
	if err != nil {
		return nil, errors.New("group not found")
	}
	resp = new(types.GroupValidResponse)
	resp.Verification = group.Verification
	switch group.Verification {
	case 0: //不允许任何人加入
	case 1: //允许任何人加入
	case 2: //需要验证
	case 3, 4: //需要正确回答问题
		if group.VerificationQuestion != nil {
			resp.VerificationQuestion = types.VerificationQuestion{
				Problem1: group.VerificationQuestion.Problem1,
				Problem2: group.VerificationQuestion.Problem2,
				Problem3: group.VerificationQuestion.Problem3,
			}
		}
	default:
	}

	return
}
