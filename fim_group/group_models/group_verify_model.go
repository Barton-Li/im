package group_models

import (
	"fim/common/models"
	"fim/common/models/ctype"
)

type GroupVerifyModel struct {
	models.Model
	GroupID              uint                        `json:"groupID"`                           // 群
	GroupModel           GroupModel                  `gorm:"foreignKey:GroupID" json:"-"`       // 群
	UserID               uint                        `json:"userID"`                            // 需要加群或者是退群的用户id
	Status               int8                        `json:"status"`                            // 状态 0 未操作 1 同意 2 拒绝 3 忽略
	AdditionalMessages   string                      `gorm:"size:32" json:"additionalMessages"` // 附加消息
	VerificationQuestion *ctype.VerificationQuestion `json:"verificationQuestion"`              // 验证问题  为3和4的时候需要
	Type                 int8                        `json:"type"`                              // 类型 1 加群  2 退群

}
