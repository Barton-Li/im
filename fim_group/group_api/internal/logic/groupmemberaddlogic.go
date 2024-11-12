package logic

import (
	"context"
	"errors"
	"fim/fim_group/group_models"

	"fim/fim_group/group_api/internal/svc"
	"fim/fim_group/group_api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GroupMemberAddLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGroupMemberAddLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupMemberAddLogic {
	return &GroupMemberAddLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// GroupMemberAdd 实现了添加群成员的功能。
// req 是一个包含群ID、用户ID和成员ID列表的请求对象。
// 返回值 resp 是一个包含操作结果的响应对象，err 是操作过程中可能产生的错误。
func (l *GroupMemberAddLogic) GroupMemberAdd(req *types.GroupMemberAddRequest) (resp *types.GroupMemberAddResponse, err error) {
	// 初始化一个群成员模型实例。
	var member group_models.GroupMemberModel

	// 查询数据库，检查当前用户是否有权限邀请成员。
	err = l.svcCtx.DB.Preload("GroupModel").Take(&member, "group_id =? and user_id =?", req.ID, req.UserID).Error
	if err != nil {
		// 如果查询出错，返回非法调用错误。
		return nil, errors.New("非法调用")
	}

	// 如果当前用户是普通成员且群组未开启邀请入群功能，则返回错误。
	if member.Role == 3 {
		if member.GroupModel.IsInvite {
			return nil, errors.New("管理员未开放好友邀请入群功能")
		}
	}

	// 查询数据库，检查待添加的成员是否已存在于群组中。
	var memberList []group_models.GroupMemberModel
	l.svcCtx.DB.Find(&memberList, "group_id=?and user_id in?", req.ID, req.MemberIDList)
	if len(memberList) > 0 {
		// 如果有成员已经存在于群组中，返回错误。
		return nil, errors.New("群成员已存在")
	}

	// 遍历成员ID列表，创建新的群成员模型实例，并添加到memberList中。
	for _, memberID := range req.MemberIDList {
		memberList = append(memberList, group_models.GroupMemberModel{
			GroupID: req.ID,
			UserID:  memberID,
			Role:    3,
		})
	}

	// 将新的群成员信息批量添加到数据库。
	err = l.svcCtx.DB.Create(&memberList).Error
	if err != nil {
		// 如果添加过程中出现错误，记录错误日志。
		logx.Error(err)
	}

	// 返回成功添加成员的响应对象或错误。
	return
}
