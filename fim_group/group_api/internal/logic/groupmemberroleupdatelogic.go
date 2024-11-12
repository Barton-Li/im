package logic

import (
	"context"
	"errors"
	"fim/fim_group/group_models"

	"fim/fim_group/group_api/internal/svc"
	"fim/fim_group/group_api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GroupMemberRoleUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGroupMemberRoleUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupMemberRoleUpdateLogic {
	return &GroupMemberRoleUpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// GroupMemberRoleUpdate 修改群组成员角色。
// req 是一个指向 GroupMemberRoleUpdateRequest 结构体的指针，包含修改角色所需的信息。
// 返回值 resp 是一个指向 GroupMemberRoleUpdateResponse 结构体的指针，包含修改结果。
// 如果发生错误，返回错误信息。
func (l *GroupMemberRoleUpdateLogic) GroupMemberRoleUpdate(req *types.GroupMemberRoleUpdateRequest) (resp *types.GroupMemberRoleUpdateResponse, err error) {
	// 初始化群成员模型。
	var member group_models.GroupMemberModel
	// 从数据库中查询指定群组和用户ID的群成员。
	err = l.svcCtx.DB.Take(&member, "group_id =? AND user_id =?", req.ID, req.UserID).Error
	if err != nil {
		// 如果查询错误，返回自定义错误信息。
		return nil, errors.New("用户不在群组中")
	}
	// 检查当前用户是否为群主。
	if member.Role != 1 {
		// 如果不是群主，返回自定义错误信息。
		return nil, errors.New("只有群主才能修改群成员角色")
	}
	// 初始化另一个群成员模型。
	var member1 group_models.GroupMemberModel
	// 从数据库中查询指定群组和待修改成员ID的群成员。
	err = l.svcCtx.DB.Take(&member1, "group_id =? AND user_id =?", req.ID, req.MemberID).Error
	if err != nil {
		// 如果查询错误，返回自定义错误信息。
		return nil, errors.New("成员不在群组中")
	}
	// 检查待修改的角色是否为合法值。
	if !(req.Role == 3 || req.Role == 2) {
		// 如果角色不合法，返回自定义错误信息。
		return nil, errors.New("角色只能是3或者2")
	}
	// 如果成员当前角色与待修改的角色相同，不需要进行修改。
	if member1.Role == req.Role {
		return
	}
	// 更新成员角色。
	l.svcCtx.DB.Model(&member1).Update("role", req.Role)
	return
}
