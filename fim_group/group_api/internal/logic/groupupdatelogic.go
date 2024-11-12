package logic

import (
	"context"
	"errors"
	"fim/common/models/ctype"
	"fim/fim_group/group_api/internal/svc"
	"fim/fim_group/group_api/internal/types"
	"fim/fim_group/group_models"
	"fim/utils/maps"

	"github.com/zeromicro/go-zero/core/logx"
)

type GroupUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGroupUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupUpdateLogic {
	return &GroupUpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GroupUpdateLogic) GroupUpdate(req *types.GroupUpdateRequest) (resp *types.GroupUpdateResponse, err error) {
	var groupMember group_models.GroupMemberModel
	err = l.svcCtx.DB.Preload("GroupModel").Take(&groupMember, "group_id =? AND user_id =?", req.ID, req.UserID).Error
	if err != nil {
		return nil, errors.New("群不存在或用户不在群中")
	}
	if !(groupMember.Role == 1 || groupMember.Role == 2) {
		return nil, errors.New("权限不足")
	}
	groupMaps := maps.RefToMap(*req, "conf")
	if len(groupMaps) != 0 {
		verificationQuestion, ok := groupMaps["verification_question"]
		if ok {
			delete(groupMaps, "verification_question")
			data := ctype.VerificationQuestion{}
			maps.MapToStruct(verificationQuestion.(map[string]any), &data)
			l.svcCtx.DB.Model(&groupMember.GroupModel).Updates(&group_models.GroupModel{
				VerificationQuestion: &data,
			})

		}
		err = l.svcCtx.DB.Model(&groupMember.GroupModel).Updates(groupMaps).Error
		if err != nil {
			logx.Error(err)
			return nil, errors.New("更新失败")
		}

	}

	return
}
