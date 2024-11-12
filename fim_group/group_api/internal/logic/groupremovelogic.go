package logic

import (
	"context"
	"errors"
	"fim/fim_group/group_models"

	"fim/fim_group/group_api/internal/svc"
	"fim/fim_group/group_api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GroupRemoveLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGroupRemoveLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupRemoveLogic {
	return &GroupRemoveLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// GroupRemove 用于处理群组移除逻辑的函数。
// 参数 req 是一个 GroupRemoveRequest 类型的对象，包含了群组ID和用户ID。
// 返回值 resp 是一个 GroupRemoveResponse 类型的对象，包含了操作结果的信息。
// 返回值 err 是一个错误类型，包含了操作过程中可能出现的错误信息。
func (l *GroupRemoveLogic) GroupRemove(req *types.GroupRemoveRequest) (resp *types.GroupRemoveResponse, err error) {
	// 初始化一个 GroupMemberModel 类型的变量，用于存储群组成员信息。
	var groupMember group_models.GroupMemberModel
	// 从数据库中查询群组成员信息，确保用户在群中且群组存在。
	err = l.svcCtx.DB.Take(&groupMember, "group_id=?and user_id=?", req.ID, req.UserID).Error
	if err != nil {
		// 如果查询错误，返回自定义的错误信息。
		return nil, errors.New("群不存在或用户不在群中")
	}
	// 检查当前用户是否为群主，只有群主才能删除群组。
	if groupMember.Role != 1 {
		return nil, errors.New("只有群主才能删除群")
	}
	// 初始化一个 GroupMsgModel 类型的切片，用于存储群组消息信息。
	var msgList []group_models.GroupMsgModel
	// 查询并删除与群组相关的所有消息。
	l.svcCtx.DB.Find(&msgList, "group_id=?", req.ID).Delete(&msgList)
	// 初始化一个 GroupVerifyModel 类型的切片，用于存储群组验证消息信息。
	var vList []group_models.GroupVerifyModel
	// 查询并删除与群组相关的所有验证消息。
	l.svcCtx.DB.Find(&vList, "group_id=?", req.ID).Delete(&vList)
	// 初始化一个 GroupMemberModel 类型的切片，用于存储群组成员信息。
	var memberList []group_models.GroupMemberModel
	// 查询并删除与群组相关的所有成员。
	l.svcCtx.DB.Find(&memberList, "group_id=?", req.ID).Delete(&memberList)
	// 初始化一个 GroupModel 类型的变量，用于存储群组信息。
	var group group_models.GroupModel
	// 查询并删除指定ID的群组。
	l.svcCtx.DB.Take(&group, "id=?", req.ID).Delete(&group)
	// 记录日志，提供删除操作的相关信息。
	logx.Infof("删除群：%s", group.Title)
	logx.Infof("关联群成员：%d", len(memberList))
	logx.Infof("关联群消息数：%d", len(msgList))
	logx.Infof("关联群验证消息数：%d", len(vList))
	// 返回操作结果，无错误。
	return
}
