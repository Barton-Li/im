package core

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// InitGorm 初始化Gorm数据库连接。
// 该函数不接受参数，返回一个*gorm.DB类型的数据库连接实例。
func InitGorm(MysqlDataSource string) *gorm.DB {
	db, err := gorm.Open(mysql.Open(MysqlDataSource), &gorm.Config{})
	if err != nil {
		panic("failed to connect database" + err.Error())
	} else {
		fmt.Println("connect database success")
	}

	return db
}
