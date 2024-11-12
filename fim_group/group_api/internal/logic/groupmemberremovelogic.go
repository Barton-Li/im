package logic

import (
	"context"
	"errors"
	"fim/fim_group/group_models"

	"fim/fim_group/group_api/internal/svc"
	"fim/fim_group/group_api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GroupMemberRemoveLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGroupMemberRemoveLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupMemberRemoveLogic {
	return &GroupMemberRemoveLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// GroupMemberRemove 用于处理群成员移除逻辑。
// req 是移除请求的参数，包含群ID、用户ID和成员ID等信息。
// 返回值 resp 是移除操作的响应，可能为 nil，err 是操作过程中可能产生的错误。
func (l *GroupMemberRemoveLogic) GroupMemberRemove(req *types.GroupMemberRemoveRequest) (resp *types.GroupMemberRemoveResponse, err error) {
	// 查询数据库，获取待操作的群成员信息。
	var member group_models.GroupMemberModel
	err = l.svcCtx.DB.Take(&member, "group_id = ? and user_id = ?", req.ID, req.UserID).Error
	if err != nil {
		return nil, errors.New("违规操作，请勿删除其他用户")
	}
	// 如果操作者尝试移除自己，检查是否为群主，群主不能退出群组，只能解散群组。
	if req.UserID == req.MemberID {
		if member.Role == 1 {
			return nil, errors.New("群主不能退出群组,只能解散群组")
		}
		// 移除成员并记录退出群组的操作。
		l.svcCtx.DB.Delete(&member)
		l.svcCtx.DB.Create(&group_models.GroupVerifyModel{
			GroupID: member.GroupID,
			UserID:  req.UserID,
			Type:    2, // 退出群组
		})
		return
	}
	// 踢出成员，判断操作者是否有权限移除其他成员。
	if !(member.Role == 1 || member.Role == 2) {
		return nil, errors.New("违规操作，请勿删除其他用户")
	}
	// 获取被移除成员的详细信息，包括消息列表等。
	var member1 group_models.GroupMemberModel
	err = l.svcCtx.DB.Preload("MsgList").Take(&member1, "group_id = ? and user_id = ?", req.ID, req.MemberID).Error
	if err != nil {
		return nil, errors.New("该用户不是群成员")
	}
	// 校验操作权限，确保群主可以移除管理员和普通成员，管理员可以移除普通成员。
	if !(member.Role == 1 && (member1.Role == 2 || member1.Role == 3) || member.Role == 2 && member1.Role == 3) {
		return nil, errors.New("违规操作，请勿删除其他用户")
	}
	// 清除被移除成员的消息记录。
	if len(member1.MsgList) > 0 {
		l.svcCtx.DB.Model(&member1.MsgList).Update("group_member_id", nil)
		logx.Info("清除群消息%d", len(member1.MsgList))
	}
	// 执行移除操作。
	err = l.svcCtx.DB.Delete(&member1).Error
	if err != nil {
		logx.Error(err)
		return nil, errors.New("删除失败")
	}
	return
}
