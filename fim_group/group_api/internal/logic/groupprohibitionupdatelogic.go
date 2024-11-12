package logic

import (
	"context"
	"errors"
	"fim/fim_group/group_models"
	"fmt"
	"time"

	"fim/fim_group/group_api/internal/svc"
	"fim/fim_group/group_api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GroupProhibitionUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGroupProhibitionUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupProhibitionUpdateLogic {
	return &GroupProhibitionUpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// GroupProhibitionUpdate 群组禁言设置接口
// 该函数用于更新群组中某成员的禁言状态。只有群主和管理员有权限对其他成员进行禁言操作。
// 参数:
//
//	req - 请求体，包含群组ID、用户ID、成员ID和禁言时间等信息。
//
// 返回值:
//
//	resp - 响应体，包含操作结果等信息。
//	err - 错误信息，操作失败时返回具体错误原因。
func (l *GroupProhibitionUpdateLogic) GroupProhibitionUpdate(req *types.GroupProhibitionUpdateRequest) (resp *types.GroupProhibitionUpdateResponse, err error) {
	// 查询当前用户在群组中的角色
	var member group_models.GroupMemberModel
	err = l.svcCtx.DB.Take(&member, "group_id =? AND user_id =?", req.GroupID, req.UserID).Error
	if err != nil {
		return nil, errors.New("当前用户错误")
	}
	// 判断当前用户是否为群主或管理员
	if !(member.Role == 1 || member.Role == 2) {
		return nil, errors.New("当前用户不是群主或管理员")
	}
	// 查询被操作成员在群组中的角色
	var member1 group_models.GroupMemberModel
	err = l.svcCtx.DB.Take(&member1, "group_id =? and user_id =?", req.GroupID, req.MemberID).Error
	if err != nil {
		return nil, errors.New("用户不是群成员")
	}
	// 校验当前用户是否对被操作成员有禁言权限
	if !((member.Role == 1 && (member1.Role == 2 || member1.Role == 3)) || (member.Role == 2 && member1.Role == 3)) {
		return nil, errors.New("当前用户不是群主或管理员")
	}
	// 更新被操作成员的禁言时间
	l.svcCtx.DB.Model(&member1).Update("prohibition_time", req.ProhibitionTime)
	// 根据被操作成员的ID生成禁言状态的Redis键名
	key := fmt.Sprintf("prohibition:%d", member1.ID)
	// 根据禁言时间设置或删除Redis中的禁言状态
	if req.ProhibitionTime != nil {
		l.svcCtx.Redis.Set(key, "1", time.Duration(*req.ProhibitionTime)*time.Minute)
	} else {
		l.svcCtx.Redis.Del(key)
	}
	return
}
