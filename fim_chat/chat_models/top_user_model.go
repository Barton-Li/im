package chat_models

import "fim/common/models"

type TopUserModel struct {
	models.Model
	UserID    uint `json:"user_id"`
	TopUserID uint `json:"top_user_id"`
}
