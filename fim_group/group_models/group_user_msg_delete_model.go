package group_models

import "fim/common/models"

type GroupUserMsgDeleteModel struct {
	models.Model
	UserID  uint `json:"userID"`  // 用户id
	MsgID   uint `json:"msgID"`   // 群聊天记录的id
	GroupID uint `json:"groupID"` // 群id
}
