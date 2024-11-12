package logic

import (
	"context"
	"fim/fim_chat/chat_api/internal/svc"
	"fim/fim_chat/chat_api/internal/types"
	"fim/fim_chat/chat_models"

	"github.com/zeromicro/go-zero/core/logx"
)

type ChatDeleteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewChatDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ChatDeleteLogic {
	return &ChatDeleteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// ChatDelete 逻辑层的删除聊天记录方法
// 根据请求中的用户ID和聊天记录ID列表，删除相应的聊天记录。
// 这里实现了逻辑判断，确保只删除用户自己的聊天记录，并且避免重复删除。
func (l *ChatDeleteLogic) ChatDelete(req *types.ChatDeleteRequest) (resp *types.ChatDeleteResponse, err error) {
	// 查询所有指定ID的聊天记录
	var chatList []chat_models.ChatModel
	l.svcCtx.DB.Find(&chatList, "id in ?", req.IdList)

	// 查询所有已标记为删除的聊天记录
	var useDeleteChatList []chat_models.UserChatDeleteModel
	l.svcCtx.DB.Find(&useDeleteChatList, "chat_id in ?", req.IdList)

	// 使用映射存储已标记为删除的聊天记录ID，以提高查找效率
	chatDeleteMap := map[uint]struct{}{}
	for _, model := range useDeleteChatList {
		chatDeleteMap[model.ChatID] = struct{}{}
	}

	// 准备待删除的聊天记录列表
	var deleteChatList []chat_models.UserChatDeleteModel
	if len(chatList) > 0 {
		for _, model := range chatList {
			// 判断聊天记录是否属于当前用户
			if !(model.SendUserID == req.UserID || model.RevUserID == req.UserID) {
				continue // 不是自己的聊天记录，跳过
			}
			// 检查是否已经标记为删除
			//删除聊天记录
			_, ok := chatDeleteMap[model.ID]
			if ok {
				continue // 已经删除过的聊天记录，跳过
			}
			// 添加到待删除列表
			deleteChatList = append(deleteChatList, chat_models.UserChatDeleteModel{
				ChatID: model.ID,
				UserID: req.UserID,
			})
		}
	}

	// 如果有聊天记录需要删除，则执行删除操作
	if len(deleteChatList) > 0 {
		l.svcCtx.DB.Create(&deleteChatList)
	}

	// 记录删除的聊天记录数量
	logx.Infof("已删除聊天记录%d条", len(deleteChatList))
	return
}
