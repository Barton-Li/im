package logic

import (
	"context"
	"errors"
	"fim/fim_group/group_models"

	"fim/fim_group/group_api/internal/svc"
	"fim/fim_group/group_api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GroupMemberNicknameUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGroupMemberNicknameUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupMemberNicknameUpdateLogic {
	return &GroupMemberNicknameUpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// GroupMemberNicknameUpdate 用于更新群组成员的昵称。
// req 是一个指向 GroupMemberNicknameUpdateRequest 类型的指针，包含更新昵称所需的信息。
// 返回值 resp 是一个指向 GroupMemberNicknameUpdateResponse 类型的指针，包含操作结果的信息。
// 如果操作过程中出现错误，会返回错误信息。
func (l *GroupMemberNicknameUpdateLogic) GroupMemberNicknameUpdate(req *types.GroupMemberNicknameUpdateRequest) (resp *types.GroupMemberNicknameUpdateResponse, err error) {
	// 尝试从数据库中加载群组成员信息，判断该用户是否为群组成员。
	var member group_models.GroupMemberModel
	err = l.svcCtx.DB.Take(&member, "group_id=? and user_id=?", req.ID, req.UserID).Error
	if err != nil {
		return nil, errors.New("group member not found")
	}

	// 尝试从数据库中加载被修改昵称的群组成员信息，判断该用户是否为群组成员。
	var member1 group_models.GroupMemberModel
	err = l.svcCtx.DB.Take(&member1, "group_id=? and user_id=?", req.ID, req.MemberID).Error
	if err != nil {
		return nil, errors.New("该用户不是群成员")
	}

	// 如果操作用户和被修改昵称的用户是同一个用户，则直接更新昵称。
	if req.UserID == req.MemberID {
		l.svcCtx.DB.Model(&member).Updates(map[string]any{
			"member_nickname": req.Nickname,
		})
		return
	}

	// 检查操作用户和被修改昵称用户的权限，确保操作用户有权限修改被修改昵称用户的昵称。
	if !((member.Role == 1 && (member1.Role == 2 || member1.Role == 3)) || (member.Role == 2 && member1.Role == 3)) {
		return nil, errors.New("权限不足")
	}

	// 更新被修改昵称用户的昵称。
	l.svcCtx.DB.Model(&member1).Updates(map[string]any{
		"member_nickname": req.Nickname,
	})

	return
}
