package group_models

import (
	"fim/common/models"
	"fmt"
	"github.com/go-redis/redis"
	"gorm.io/gorm"
	"time"
)

// GroupMemberModel 群成员表
type GroupMemberModel struct {
	models.Model
	GroupID         uint            `json:"groupID"`                           // 群id
	GroupModel      GroupModel      `gorm:"foreignKey:GroupID" json:"-"`       // 群
	UserID          uint            `json:"userID"`                            // 用户id
	MemberNickname  string          `gorm:"size:32" json:"memberNickname"`     // 群成员昵称
	Role            int8            `json:"role"`                              // 1 群主 2 管理员  3 普通成员
	ProhibitionTime *int            `json:"prohibitionTime"`                   // 禁言时间 单位分钟
	MsgList         []GroupMsgModel `json:"-" gorm:"foreignKey:GroupMemberID"` // 这个用户发的消息
}

// GetProhibitionTime 获取禁言时间
// 该函数尝试从Redis中获取群成员的禁言时间，并将其转换为分钟。
// 如果Redis中没有设置禁言时间或者禁言时间已过期，则尝试从数据库中更新禁言时间。
// 参数client用于连接Redis，db用于连接数据库。
// 返回值是一个指针，指向计算出的禁言时间（单位为分钟），如果没有有效的禁言时间则返回nil。
func (gm GroupMemberModel) GetProhibitionTime(client *redis.Client, db *gorm.DB) *int {
	// 如果当前对象的禁言时间已经是nil，直接返回nil
	if gm.ProhibitionTime == nil {
		return nil
	}
	// 尝试从Redis中获取禁言时间的剩余时间
	t, err := client.TTL(fmt.Sprintf("prohibition:%d", gm.ID)).Result()
	// 如果获取过程中发生错误，将数据库中的禁言时间设置为nil，并返回nil
	if err != nil {
		db.Model(&gm).Update("prohibition_time", nil)
		return nil
	}
	// 如果Redis中禁言时间已过期（-2表示过期），同样将数据库中的禁言时间设置为nil，并返回nil
	if t == -2*time.Second {
		db.Model(&gm).Update("prohibition_time", nil)
		return nil
	}
	// 将禁言时间从秒转换为分钟
	res := int(t / time.Minute)
	// 返回转换后的禁言时间
	return &res
}
