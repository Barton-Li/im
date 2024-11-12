package logic

import (
	"context"
	"fim/common/list_query"
	"fim/common/models"
	"fim/fim_group/group_models"
	"fmt"

	"fim/fim_group/group_api/internal/svc"
	"fim/fim_group/group_api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GroupSessionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGroupSessionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupSessionLogic {
	return &GroupSessionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

type SessionData struct {
	GroupID       uint   `gorm:"column:group_id"`
	NewMsgDate    string `gorm:"column:new_msg_date"`
	NewMsgPreview string `gorm:"column:new_msg_preview"`
	IsTop         bool   `gorm:"column:is_top"`
}

// GroupSession 根据用户ID获取群组会话列表。
// req: 请求参数，包含UserID、Page和Limit。
// 返回值: resp为响应数据，包含会话列表和总数；err为错误信息。
func (l *GroupSessionLogic) GroupSession(req *types.GroupSessionRequest) (resp *types.GroupSessionListResponse, err error) {
	// 初始化用户群组ID列表
	var userGroupIDList []uint
	// 查询用户所在的群组ID
	l.svcCtx.DB.Model(group_models.GroupMemberModel{}).Where("user_id =?", req.UserID).Select("group_id").Scan(&userGroupIDList)
	// 构建是否置顶的查询列
	column := fmt.Sprintf("(if ((select 1 from group_user_top_models where user_id=%d and group_user_top_model.group_id=group_msg_models.group_id),1,0))as isTop", req.UserID)
	// 初始化删除的消息ID列表
	var msgDeleteIDList []uint
	// 查询被该用户删除的消息ID
	l.svcCtx.DB.Model(group_models.GroupUserMsgDeleteModel{}).Where("group_id in ?", userGroupIDList).Select("msg_id").Scan(&msgDeleteIDList)
	// 初始化查询
	query := l.svcCtx.DB.Where("group_id in (?)", userGroupIDList)
	// 如果有删除的消息，则排除这些消息
	if len(msgDeleteIDList) > 0 {
		query.Where("id not in ?", msgDeleteIDList)
	}
	// 执行分页查询，获取会话列表和总数
	sessionList, count, _ := list_query.ListQuery(l.svcCtx.DB, SessionData{}, list_query.Option{
		PageInfo: models.PagaInfo{
			Page:  req.Page,
			Limit: req.Limit,
			Sort:  "isTop desc,newMSgDate desc",
		},
		Debug: true,
		Table: func() (string, any) {
			return "(?) as u", l.svcCtx.DB.Model(&group_models.GroupMsgModel{}).
				Select("group_id as g_id",
					"max(created_at)as newMSgDate",
					column,
					"(select msg_preview from group_msg_models as g where g.group_id=g_id order by g.created_at desc limit 1)as newMsgPreview").
				Where(query).Group("group_id")
		},
	})
	// 初始化群组ID列表
	var groupIDList []uint
	// 提取会话列表中的群组ID
	for _, data := range sessionList {
		groupIDList = append(groupIDList, data.GroupID)
	}
	// 查询群组详细信息
	var groupListModel []group_models.GroupModel
	l.svcCtx.DB.Find(&groupListModel, "id in ?", groupIDList)
	// 构建群组ID到群组信息的映射
	var groupMap = map[uint]group_models.GroupModel{}
	for _, model := range groupListModel {
		groupMap[model.ID] = model
	}
	// 初始化响应数据
	resp = new(types.GroupSessionListResponse)
	// 构建最终的会话列表
	for _, data := range sessionList {
		resp.List = append(resp.List, types.GroupSessionResponse{
			GroupID:       data.GroupID,
			Title:         groupMap[data.GroupID].Title,
			Avatar:        groupMap[data.GroupID].Avatar,
			NewMsgDate:    data.NewMsgDate,
			NewMsgPreview: data.NewMsgPreview,
			IsTop:         data.IsTop,
		})
	}
	// 设置总条数
	resp.Count = int(count)

	return
}
