package user_models

import (
	"fim/common/models"
	"gorm.io/gorm"
)

type FriendModel struct {
	models.Model
	SendUserID    uint      `json:"sendUserID"`                     // 发起验证方
	SendUserModel UserModel `gorm:"foreignKey:SendUserID" json:"-"` // 发起验证方
	RevUserID     uint      `json:"revUserID"`                      // 接受验证方
	RevUserModel  UserModel `gorm:"foreignKey:RevUserID" json:"-"`  // 接受验证方
	SenUserNotice string    `gorm:"size:128" json:"senUserNotice"`  // 发送方备注
	RevUserNotice string    `gorm:"size:128" json:"revUserNotice"`  // 接收方备注

}

// IsFriend 检查A和B是否互为好友
// 使用GORM的Take方法尝试获取好友模型，如果存在则返回true，否则返回false
// 参数:
//
//	db: GORM数据库实例
//	A: 用户A的ID
//	B: 用户B的ID
//
// 返回值:
//
//	bool: A和B是否互为好友
func (f *FriendModel) IsFriend(db *gorm.DB, A, B uint) bool {
	// 尝试根据用户ID组合查询是否存在好友关系
	err := db.Take(&f, "(send_user_id =? AND rev_user_id =?)or (send_user_id =? AND rev_user_id =?)", A, B, B, A)

	// 如果没有错误，即找到了相关记录，说明A和B是好友
	if err == nil {
		return true
	}
	// 如果有错误，说明没有找到相关记录，A和B不是好友
	return false
}

// Friends 获取用户的所有好友列表
// 使用GORM的Find方法查询所有发送者或接收者为指定userID的好友模型
// 参数:
//
//	db: GORM数据库实例
//	userID: 查询用户的ID
//
// 返回值:
//
//	[]FriendModel: 好友模型列表
func (f *FriendModel) Friends(db *gorm.DB, userID uint) (list []FriendModel) {
	// 根据userID查询所有发送者或接收者为userID的好友记录
	db.Find(&list, "send_user_id =? or rev_user_id =?", userID, userID)

	// 返回查询结果
	return
}

// GetUserNotice 根据userID获取对应的用户备注
// 如果userID等于SendUserID，返回SenUserNotice
// 如果userID等于RevUserID，返回RevUserNotice
// 否则返回空字符串
// 参数:
//
//	userID: 查询的用户ID
//
// 返回值:
//
//	string: 用户通知字符串
func (f *FriendModel) GetUserNotice(userID uint) string {
	// 检查userID是否等于SendUserID
	if userID == f.SendUserID {
		return f.SenUserNotice
	}
	// 检查userID是否等于RevUserID
	if userID == f.RevUserID {
		return f.RevUserNotice
	}
	// 如果都不匹配，返回空字符串
	return ""
}
