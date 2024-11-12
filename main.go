package main

import (
	"fim/core"
	"fim/fim_chat/chat_models"
	"fim/fim_file/file_model"
	"fim/fim_group/group_models"
	"fim/fim_user/user_models"
	"flag"
	"fmt"
)

type Options struct {
	DB bool
}

func main() {

	var opt Options
	// 此段代码用于设置和解析命令行参数中的数据库选项
	// 参数:
	// - opt.DB: 一个布尔值，指明是否使用数据库选项
	// - "db": 命令行参数中的键名，用于识别数据库选项
	// - false: 参数的默认值，表示在未明确指定时，默认不使用数据库
	// - "db": 对命令行参数的描述信息
	flag.BoolVar(&opt.DB, "db", false, "db")
	flag.Parse() // 解析命令行参数

	if opt.DB {
		db := core.InitGorm("root:root@tcp(localhost:3306)/fim_db?charset=utf8mb4&parseTime=True&loc=Local")
		// AutoMigrate将根据当前数据库连接自动迁移数据库结构。
		// 它会检测模型结构的变化，并自动在数据库中创建或更新相应的表。
		// 注意：这可能会导致现有数据的丢失，因此在生产环境中应谨慎使用。
		//
		// 参数:
		// db - 数据库连接实例，用于执行数据库迁移操作。
		//
		// 返回值:
		// err - 执行迁移过程中遇到的任何错误。

		err := db.AutoMigrate(
			&user_models.UserModel{},                // 用户表
			&user_models.FriendModel{},              // 好友表
			&user_models.FriendVerifyModel{},        // 好友验证表
			&user_models.UserConfModel{},            // 用户配置表
			&chat_models.ChatModel{},                // 对话表
			&chat_models.TopUserModel{},             // 置顶用户表
			&chat_models.UserChatDeleteModel{},      // 用户删除聊天记录表
			&group_models.GroupModel{},              // 群组表
			&group_models.GroupMsgModel{},           // 群消息表
			&group_models.GroupVerifyModel{},        // 群验证表
			&group_models.GroupMemberModel{},        // 群成员表
			&group_models.GroupUserMsgDeleteModel{}, // 用户删除聊天记录表
			&group_models.GroupUserTopModel{},       // 用户置顶群聊表
			&file_model.FileModel{},                 // 文件表
			//&logs_model.LogModel{},                  // 日志表
			//&settings_model.SettingsModel{},         // 系统表
		)
		if err != nil {
			fmt.Println("表结构生成失败", err)
			return
		}
		fmt.Println("表结构生成成功！")

	}

}
