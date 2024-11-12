package user_models

import (
	"fim/common/models"
	"fim/common/models/ctype"
)

type UserConfModel struct {
	models.Model
	UserID               uint                        `json:"user_id"`
	UserModel            UserModel                   `gorm:"foreignKey:UserID" json:"-"`
	RecallMessage        *string                     `gorm:"size:32" json:"recall_message"` //撤回消息发提示内容
	FriendOnline         bool                        `json:"friend_online"`                 //好友在线提醒
	Sound                bool                        `json:"sound"`                         //声音提醒
	SecureLink           bool                        `json:"secure_link"`                   //安全链接
	SavePwd              bool                        `json:"save_pwd"`                      //保存密码
	SearchUser           int8                        `json:"search_user"`                   //别人查找你的方式 0不允许别人查找 1通过用户号 2通过昵称
	Verification         int8                        `json:"verification"`                  //好友验证 0不允许任何人添加 1允许任何人添加 2需要验证消息 3需要回答问题 4需要正确回答问题
	VerificationQuestion *ctype.VerificationQuestion `json:"verification_question"`         //验证问题 3 4
	Online               bool                        `json:"online"`                        //在线状态
	CurtailChat          bool                        `json:"curtail_chat"`                  //限制聊天
	CurtailAddUser       bool                        `json:"curtail_add_user"`              //限制添加好友
	CurtailCreateGroup   bool                        `json:"curtail_create_group"`          //限制建群
	CurtailInGroupChat   bool                        `json:"curtail_in_group_chat"`         //限制群聊
}
