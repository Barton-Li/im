package logic

import (
	"context"
	"errors"
	"fim/fim_group/group_models"

	"fim/fim_group/group_api/internal/svc"
	"fim/fim_group/group_api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GroupValidStatusLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGroupValidStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupValidStatusLogic {
	return &GroupValidStatusLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// GroupValidStatus 处理组验证状态请求。
// req 是包含组验证状态请求信息的结构体。
// 返回值 resp 是包含处理结果的响应结构体，err 是可能的错误信息。
func (l *GroupValidStatusLogic) GroupValidStatus(req *types.GroupValidStatusRequest) (resp *types.GroupValidStatusResponse, err error) {
	// 初始化一个 GroupVerifyModel 结构体，用于存储从数据库查询到的验证请求信息。
	var groupValidModel group_models.GroupVerifyModel
	// 从数据库中根据 ValidID 查询验证请求信息，如果找不到或出现错误，则返回。
	err = l.svcCtx.DB.Take(&groupValidModel, req.ValidID).Error
	if err != nil {
		return nil, errors.New("Invalid ValidID")
	}
	// 根据请求的状态进行不同的操作。
	switch req.Status {
	case 1, 2, 3:
		// 如果验证请求已经被处理过，则不能再次处理。
		if groupValidModel.Status != 0 {
			return nil, errors.New("已经处理过该验证请求了")
		}
	case 4:
		// 只能删除已经被处理过的验证请求。
		if groupValidModel.Status == 0 {
			return nil, errors.New("只能删除处理过的验证请求")
		}
	default:
		// 如果状态非法，则返回错误。
		return nil, errors.New("错误的状态")
	}
	// 查询成员信息，以验证当前用户是否有权限处理该请求。
	var member group_models.GroupMemberModel
	err = l.svcCtx.DB.Take(&member, "group_id =? AND user_id =?", groupValidModel.GroupID, req.UserID).Error
	if err != nil {
		return nil, errors.New("没有处理权限")
	}
	// 如果当前用户不是群主或管理员，则没有权限处理。
	if !(member.Role == 1 || member.Role == 2) {
		return nil, errors.New("没有处理权限")
	}
	// 根据请求的状态执行相应的操作。
	switch req.Status {
	case 0: // 未操作，直接返回。
		return
	case 1: // 通过，创建一个新成员。
		var member1 = group_models.GroupMemberModel{
			GroupID: groupValidModel.GroupID,
			UserID:  groupValidModel.UserID,
			Role:    3, // 普通成员角色。
		}
		l.svcCtx.DB.Create(&member1)
	case 2: // 拒绝，暂无操作。
	case 3: // 忽略，暂无操作。
	case 4: // 删除，删除验证请求。
		l.svcCtx.DB.Delete(&groupValidModel)
		return
	}
	// 更新验证请求的状态。
	l.svcCtx.DB.Model(&groupValidModel).UpdateColumn("status", req.Status)
	return
}
