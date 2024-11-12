package logic

import (
	"context"
	"errors"
	"fim/common/list_query"
	"fim/common/models"
	"fim/fim_chat/chat_models"
	"fim/fim_user/user_rpc/types/user_rpc"
	"fmt"

	"fim/fim_chat/chat_api/internal/svc"
	"fim/fim_chat/chat_api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ChatSessionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}
type Date struct {
	SU         uint   `gorm:"column:sU"`
	RU         uint   `gorm:"column:rU"`
	MaxDate    string `gorm:"column:maxDate"`
	MaxPreview string `gorm:"column:maxPreview"`
	IsTop      bool   `gorm:"column:isTop"`
}

func NewChatSessionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ChatSessionLogic {
	return &ChatSessionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// ChatSession 会话列表逻辑处理
// 根据请求获取用户的会话列表，包括会话的基本信息、最新消息预览和是否置顶等。
// 同时，还会标记会话中对方用户是否在线。
//
// 参数:
//
//	req - 包含用户ID、页码和每页数量的请求结构体。
//
// 返回值:
//
//	resp - 包含会话列表响应结构体。
//	err - 错误信息，如果操作失败。
func (l *ChatSessionLogic) ChatSession(req *types.ChatSessionRequest) (resp *types.ChatSessionResponse, err error) {
	// 构造查询条件，判断用户是否有置顶会话
	column := fmt.Sprintf("if ((select 1 from top_user_models where user_id =%d and (top_user_id=sU or top_user_id=rU)limit 1) ,1,0)as isTop", req.UserID)

	// 获取用户的好友列表
	var friendIDList []uint
	friendRes, err := l.svcCtx.UserRpc.FriendList(l.ctx, &user_rpc.FriendListRequest{
		User: uint32(req.UserID),
	})
	if err != nil {
		logx.Error(err)
		return nil, errors.New("获取好友列表失败")

	}
	for _, info := range friendRes.FriendList {
		friendIDList = append(friendIDList, uint(info.UserId))
	}

	// 查询会话列表，包括每页的数据和总数
	chatList, count, _ := list_query.ListQuery(l.svcCtx.DB, Date{}, list_query.Option{
		PageInfo: models.PagaInfo{
			Page:  req.Page,
			Limit: req.Limit,
			Sort:  "isTop desc,maxDate desc",
		},
		Table: func() (string, any) {
			return "(?) as u", l.svcCtx.DB.Model(&chat_models.ChatModel{}).
				Select("least(send_user_id,rev_user_id)as sU",
					"greatest(send_user-id,rev_user_id)as rU",
					"max(date)as maxDate",
					fmt.Sprintf("(select msg_preview from chat_models where((send_user_id=sU and rev_user_id=rUor (send_user_id=rU and rev_user_id=sU))and id not in (select chat_id from user_chat_delete_models where user_id=%d)order by created_at desc limit 1))as maxPreview", req.UserID),
					column).
				Where("(send_user_id=? or rev_user_id =?)and id not in (select chat_id from user_chat_delete_models where user_id= ?)and (send_user_id=?and rev_user_id in ?)or (rev_user_id=?and send_user_id in ?)",
					req.UserID, req.UserID, req.UserID, req.UserID, friendIDList, req.UserID, friendIDList).
				Group("least(send_user_id,rev_user_id)").
				Group("greatest(send_user_id,rev_user_id)")

		},
	})

	// 收集会话中涉及的用户ID，为后续获取用户信息做准备
	var userIDList []uint32
	for _, data := range chatList {
		if data.RU != req.UserID {
			userIDList = append(userIDList, uint32(data.RU))
		}
		if data.SU != req.UserID {
			userIDList = append(userIDList, uint32(data.SU))
		}
		if data.SU == req.UserID && req.UserID == data.RU {
			userIDList = append(userIDList, uint32(req.UserID))
		}
	}

	// 根据用户ID列表获取用户基本信息
	response, err := l.svcCtx.UserRpc.UserListInfo(l.ctx, &user_rpc.UserListInfoRequest{
		UserIdList: userIDList,
	})
	if err != nil {
		logx.Error(err)
		return nil, errors.New("获取用户信息失败")
	}

	// 获取在线用户列表
	userOnlineRes, err := l.svcCtx.UserRpc.UserOnlineList(l.ctx, &user_rpc.UserOnlineListRequest{})
	if err != nil {
		logx.Error(err)
		return nil, errors.New("获取在线用户失败")
	}

	// 构建在线用户映射
	var onlineUserMap = map[uint]bool{}
	for _, u := range userOnlineRes.UserIdList {
		onlineUserMap[uint(u)] = true
	}

	// 构建最终的会话响应列表
	var list = make([]types.ChatSession, 0)
	for _, data := range chatList {
		s := types.ChatSession{
			CreatedAt:  data.MaxDate,
			MsgPreview: data.MaxPreview,
			IsTop:      data.IsTop,
		}
		if data.RU != req.UserID {
			s.UserID = data.RU
			s.Avatar = response.UserInfo[uint32(s.UserID)].Avatar
			s.Nickname = response.UserInfo[uint32(s.UserID)].NickName
		}
		if data.SU != req.UserID {
			s.UserID = data.SU
			s.Avatar = response.UserInfo[uint32(s.UserID)].Avatar
			s.Nickname = response.UserInfo[uint32(s.UserID)].NickName
		}
		if data.SU == req.UserID && data.RU == req.UserID {
			s.UserID = data.SU
			s.Avatar = response.UserInfo[uint32(s.UserID)].Avatar
			s.Nickname = response.UserInfo[uint32(s.UserID)].NickName
		}
		s.IsOnline = onlineUserMap[s.UserID]
		list = append(list, s)
	}

	// 返回会话列表响应
	return &types.ChatSessionResponse{
		Count: count,
		List:  list,
	}, nil
}
