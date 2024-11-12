package svc

import (
	"fim/common/zrpc_interceptor"
	"fim/core"
	"fim/fim_file/file_rpc/files"
	"fim/fim_file/file_rpc/types/file_rpc"
	"fim/fim_group/group_api/internal/config"
	"fim/fim_user/user_rpc/types/user_rpc"
	"fim/fim_user/user_rpc/users"
	"github.com/go-redis/redis"
	"github.com/zeromicro/go-zero/zrpc"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config  config.Config
	DB      *gorm.DB
	UserRpc user_rpc.UsersClient
	//GroupRpc        group_rpc.GroupsClient
	FileRpc file_rpc.FilesClient
	Redis   *redis.Client
}

func NewServiceContext(c config.Config) *ServiceContext {
	mysqlDb := core.InitGorm(c.Mysql.DataSource)
	client := core.InitRedis(c.Redis.Addr, c.Redis.Pwd, c.Redis.DB)
	return &ServiceContext{
		Config:  c,
		DB:      mysqlDb,
		UserRpc: users.NewUsers(zrpc.MustNewClient(c.UserRpc, zrpc.WithUnaryClientInterceptor(zrpc_interceptor.ClientInfoInterceptor))),
		//GroupRpc:        groups.NewGroups(zrpc.MustNewClient(c.GroupRpc, zrpc.WithUnaryClientInterceptor(zrpc_interceptor.ClientInfoInterceptor))),
		FileRpc: files.NewFiles(zrpc.MustNewClient(c.FileRpc, zrpc.WithUnaryClientInterceptor(zrpc_interceptor.ClientInfoInterceptor))),
		Redis:   client,
	}
}
