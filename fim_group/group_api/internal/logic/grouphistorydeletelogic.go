package logic

import (
	"context"
	"errors"
	"fim/fim_group/group_models"
	"fim/utils/set"

	"fim/fim_group/group_api/internal/svc"
	"fim/fim_group/group_api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GroupHistoryDeleteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGroupHistoryDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupHistoryDeleteLogic {
	return &GroupHistoryDeleteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// GroupHistoryDelete 用于删除指定群组的历史消息记录。
// req 是删除请求的参数，包含群ID、用户ID和需要删除的消息ID列表。
// 返回值 resp 是删除操作的结果，包括删除成功的消息ID列表。
// err 是操作过程中可能遇到的错误，例如用户不是群成员或消息一致性错误。
func (l *GroupHistoryDeleteLogic) GroupHistoryDelete(req *types.GroupHistoryDeleteRequest) (resp *types.GroupHistoryDeleteResponse, err error) {
	// 查询用户是否为群成员，以验证操作权限。
	var member group_models.GroupMemberModel
	err = l.svcCtx.DB.Take(&member, "group_id =? AND user_id =?", req.ID, req.UserID).Error
	if err != nil {
		return nil, errors.New("用户不是群成员")
	}

	// 获取用户在该群组中已标记删除的消息ID列表，用于后续对比。
	var msgIDList []uint
	l.svcCtx.DB.Model(group_models.GroupUserMsgDeleteModel{}).
		Where("group_id =? AND user_id =?", req.ID, req.UserID).
		Select("msg_id").Scan(&msgIDList)

	// 计算需要新增删除标记的消息ID列表。
	addMSgIDList := set.Difference(req.MsgIDList, msgIDList)
	logx.Infof("删除聊天记录的id列表%v", addMSgIDList)

	// 如果没有新的消息需要删除标记，则直接返回。
	if len(addMSgIDList) == 0 {
		return
	}

	// 校验待删除消息的有效性，确保消息存在。
	var msgIDFindList []uint
	l.svcCtx.DB.Model(group_models.GroupMsgModel{}).
		Where("id IN (?)", addMSgIDList).
		Select("id").Scan(&msgIDFindList)
	if len(msgIDFindList) != len(addMSgIDList) {
		return nil, errors.New("消息一致性错误")
	}

	// 准备删除记录列表，用于批量创建删除标记。
	var list []group_models.GroupUserMsgDeleteModel
	for _, i2 := range addMSgIDList {
		list = append(list, group_models.GroupUserMsgDeleteModel{
			MsgID:   i2,
			UserID:  req.UserID,
			GroupID: req.ID,
		})
	}

	// 执行批量创建删除标记的操作。
	err = l.svcCtx.DB.Create(&list).Error
	if err != nil {
		return
	}

	return
}
