package svc

import (
	"fim/core"
	"fim/fim_file/file_rpc/internal/config"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config config.Config
	DB     *gorm.DB
}

func NewServiceContext(c config.Config) *ServiceContext {
	mysqlDb := core.InitGorm(c.Mysql.DataSource)
	return &ServiceContext{
		Config: c,
		DB:     mysqlDb,
	}
}
