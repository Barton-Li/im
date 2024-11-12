package logic

import (
	"context"
	"errors"
	"fim/common/list_query"
	"fim/common/models"
	"fim/common/models/ctype"
	"fim/fim_group/group_models"
	"fim/fim_user/user_rpc/types/user_rpc"
	"fim/utils"
	"time"

	"fim/fim_group/group_api/internal/svc"
	"fim/fim_group/group_api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GroupHistoryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGroupHistoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupHistoryLogic {
	return &GroupHistoryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

type HistoryResponse struct {
	GroupID        uint          `json:"group_id"`
	UserID         uint          `json:"user_id"`
	UserNickname   string        `json:"user_nickname"`
	UserAvatar     string        `json:"user_avatar"`
	Msg            ctype.Msg     `json:"msg"`
	MsgPreview     string        `json:"msg_preview"`
	ID             uint          `json:"id"`
	MsgType        ctype.MsgType `json:"msg_type"`
	CreatedAt      string        `json:"created_at"`
	IsMe           bool          `json:"is_me"`
	MemberNickname string        `json:"member_nickname"`
	ShowDate       bool          `json:"show_date"`
}
type HistoryListResponse struct {
	List  []HistoryResponse `json:"list"`
	Count int64             `json:"count"`
}

// GroupHistory 查询群组历史消息
// req: 请求参数，包含群组ID、用户ID、页码和每页数量
// resp: 响应数据，包含历史消息列表和总数
// err: 错误信息，如果查询失败返回具体错误原因
func (l *GroupHistoryLogic) GroupHistory(req *types.GroupHistoryRequest) (resp *HistoryListResponse, err error) {
	// 检查用户是否为群组成员
	var member group_models.GroupMemberModel
	err = l.svcCtx.DB.Where(&member, "group_id =? and user_id =?", req.ID, req.UserID).Error
	if err != nil {
		return nil, errors.New("该用户不是该群成员")
	}

	// 查询用户已删除的消息ID列表
	var msgIDList []uint
	l.svcCtx.DB.Model(group_models.GroupUserMsgDeleteModel{}).
		Where("group_id =? and user_id =?", req.ID, req.UserID).
		Select("msg_id").Scan(&msgIDList)

	// 初始化查询对象
	var query = l.svcCtx.DB.Where("")
	// 如果消息ID列表不为空，添加过滤条件
	if len(msgIDList) > 0 {
		query.Where("id not in ?", msgIDList)
	}

	// 执行分页查询，获取消息列表和总数
	groupMsgList, count, err := list_query.ListQuery(l.svcCtx.DB, group_models.GroupMsgModel{GroupID: req.ID}, list_query.Option{
		PageInfo: models.PagaInfo{
			Page:  req.Page,
			Limit: req.Limit,
			Sort:  "created_at desc",
		},
		Where:   query,
		Preload: []string{"GroupMemberModel"},
	})
	if err != nil {
		return nil, err
	}

	// 提取发送用户ID列表
	var userIDList []uint32
	for _, model := range groupMsgList {
		userIDList = append(userIDList, uint32(model.SendUserID))
	}
	userIDList = utils.DeduplicationList(userIDList)

	// 批量查询用户信息
	userListResponse, err1 := l.svcCtx.UserRpc.UserListInfo(l.ctx, &user_rpc.UserListInfoRequest{
		UserIdList: userIDList,
	})

	// 反转消息列表，使时间顺序从旧到新
	utils.ReverseAny(groupMsgList)

	// 构建响应数据
	var list = make([]HistoryResponse, 0)
	for index, model := range groupMsgList {
		info := HistoryResponse{
			GroupID:   model.GroupID,
			UserID:    model.SendUserID,
			Msg:       model.Msg,
			ID:        model.ID,
			MsgType:   model.MsgType,
			CreatedAt: model.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		// 标记是否显示日期，第一条消息或与前一条消息时间间隔超过1小时
		if index == 0 {
			info.ShowDate = true
		} else {
			sub := model.CreatedAt.Sub(groupMsgList[index-1].CreatedAt)
			if sub > time.Hour {
				info.ShowDate = true
			}
		}
		// 设置发送用户昵称
		if model.GroupMemberModel != nil {
			info.MemberNickname = model.GroupMemberModel.MemberNickname
		}
		// 设置用户昵称和头像
		if err1 == nil {
			info.UserNickname = userListResponse.UserInfo[uint32(info.UserID)].NickName
			info.UserAvatar = userListResponse.UserInfo[uint32(info.UserID)].Avatar
		}
		// 标记是否为本人发送
		if req.UserID == info.UserID {
			info.IsMe = true
		}
		list = append(list, info)
	}

	// 构建返回对象
	resp = new(HistoryListResponse)
	resp.List = list
	resp.Count = count
	return
}
