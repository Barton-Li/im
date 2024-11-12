package svc

import (
	"fim/core"
	"fim/fim_file/file_api/internal/config"
	"fim/fim_user/user_rpc/types/user_rpc"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config  config.Config
	UserRpc user_rpc.UsersClient
	DB      *gorm.DB
}

func NewServiceContext(c config.Config) *ServiceContext {
	mysqlDb := core.InitGorm(c.Mysql.DataSource)
	return &ServiceContext{
		Config: c,
		//UserRpc: users.NewUsers(zrpc.MustNewClient(c.UserRpc, z)),
		DB: mysqlDb,
	}
}
