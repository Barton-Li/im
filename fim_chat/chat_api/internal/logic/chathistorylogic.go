package logic

import (
	"context"
	"errors"
	"fim/common/list_query"
	"fim/common/models"
	"fim/common/models/ctype"
	"fim/fim_chat/chat_models"
	"fim/fim_user/user_rpc/types/user_rpc"
	"fim/utils"
	"time"

	"fim/fim_chat/chat_api/internal/svc"
	"fim/fim_chat/chat_api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ChatHistoryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}
type ChatHistory struct {
	ID        uint             `json:"id"`
	SendUser  ctype.UserInfo   `json:"sendUser"`
	RecvUser  ctype.UserInfo   `json:"recvUser"`
	IsMe      bool             `json:"isMe"`
	CreatedAt string           `json:"createdAt"`
	Msg       ctype.Msg        `json:"msg"`
	SystemMsg *ctype.SystemMsg `json:"systemMsg"`
	ShowDate  bool             `json:"showDate"`
}
type ChatHistoryResponse struct {
	List  []ChatHistory `json:"list"`
	Count int64         `json:"count"`
}

func NewChatHistoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ChatHistoryLogic {
	return &ChatHistoryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// ChatHistory 查询聊天记录
// 如果请求的用户ID和朋友ID不一致，会先检查他们是否是好友关系。
// 根据请求的页码和每页数量，从数据库查询聊天记录，并去重用户ID，以获取用户信息。
// 最后，根据查询到的聊天记录模型，构造聊天记录响应列表。
func (l *ChatHistoryLogic) ChatHistory(req *types.ChatHisoryRequest) (resp *ChatHistoryResponse, err error) {
	// 检查用户和朋友是否是好友关系
	if req.UserID != req.FriendID {
		res, err := l.svcCtx.UserRpc.IsFriend(l.ctx, &user_rpc.IsFriendRequest{
			User1: uint32(req.UserID),
			User2: uint32(req.FriendID),
		})
		if err != nil {
			return
		}
		if !res.IsFriend {
			return nil, errors.New("你们还不是好友")
		}
	}

	// 从数据库查询聊天记录列表和总数
	chatList, count, _ := list_query.ListQuery(l.svcCtx.DB, chat_models.ChatModel{}, list_query.Option{
		PageInfo: models.PagaInfo{
			Page:  req.Page,
			Limit: req.Limit,
			Sort:  "created_at desc",
		},
		Where: l.svcCtx.DB.Where("((send_user_id = ? and rev_user_id = ?)or(send_user_id = ? and rev_user_id = ?))and id not in (select chat_id from user_chat_delete_models where user_id = ?)",
			req.UserID, req.FriendID, req.FriendID, req.UserID, req.UserID),
	})

	// 构建用户ID列表，用于后续获取用户信息
	var userIDList []uint32
	for _, model := range chatList {
		userIDList = append(userIDList, uint32(model.SendUserID))
		userIDList = append(userIDList, uint32(model.RevUserID))
	}
	// 去重用户ID列表
	// 去重
	userIDList = utils.DeduplicationList(userIDList)

	// 通过用户ID列表获取用户信息
	// 获取用户信息
	response, err := l.svcCtx.UserRpc.UserListInfo(l.ctx, &user_rpc.UserListInfoRequest{
		UserIdList: userIDList,
	})
	if err != nil {
		logx.Error(err)
		return nil, errors.New("获取用户信息失败")
	}

	// 构建聊天记录响应列表
	var list = make([]ChatHistory, 0)
	utils.ReverseAny(chatList)
	for index, model := range chatList {
		sendUser := ctype.UserInfo{
			ID:       model.SendUserID,
			NickName: response.UserInfo[uint32(model.SendUserID)].NickName,
			Avatar:   response.UserInfo[uint32(model.SendUserID)].Avatar,
		}
		revUser := ctype.UserInfo{
			ID:       model.RevUserID,
			NickName: response.UserInfo[uint32(model.RevUserID)].NickName,
			Avatar:   response.UserInfo[uint32(model.RevUserID)].Avatar,
		}
		info := ChatHistory{
			ID:        model.ID,
			CreatedAt: model.CreatedAt.Format("2006-01-02 15:04:05"),
			SendUser:  sendUser,
			RecvUser:  revUser,
			Msg:       model.Msg,
			SystemMsg: model.SystemMsg,
		}
		// 根据时间间隔决定是否显示日期
		if index == 0 {
			info.ShowDate = true
		} else {
			sub := model.CreatedAt.Sub(chatList[index-1].CreatedAt)
			if sub > time.Hour {
				info.ShowDate = true
			}
		}
		// 标记消息是否为自己发送
		if info.SendUser.ID == req.UserID {
			info.IsMe = true
		}
		list = append(list, info)
	}

	// 构造并返回聊天记录响应
	resp = &ChatHistoryResponse{
		List:  list,
		Count: count,
	}
	return
}
