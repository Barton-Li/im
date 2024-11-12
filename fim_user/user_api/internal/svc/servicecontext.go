package svc

import (
	"fim/common/zrpc_interceptor"
	"fim/core"
	"fim/fim_chat/chat_rpc/chat"
	"fim/fim_chat/chat_rpc/types/chat_rpc"
	"fim/fim_user/user_api/internal/config"
	"fim/fim_user/user_rpc/types/user_rpc"
	"fim/fim_user/user_rpc/users"
	"github.com/go-redis/redis"
	"github.com/zeromicro/go-zero/zrpc"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config   config.Config
	DB       *gorm.DB
	UserRpc  user_rpc.UsersClient
	ChatRpc  chat_rpc.ChatClient
	GroupRpc group_rpc.UsersClient
	Redis    *redis.Client
}

func NewServiceContext(c config.Config) *ServiceContext {
	mysqlDb := core.InitGorm(c.Mysql.DataSource)
	client := core.InitRedis(c.Redis.Addr, c.Redis.Password, c.Redis.DB)
	return &ServiceContext{
		Config:   c,
		DB:       mysqlDb,
		Redis:    client,
		UserRpc:  users.NewUsers(zrpc.MustNewClient(c.UserRpc, zrpc.WithUnaryClientInterceptor(zrpc_interceptor.ClientInfoInterceptor))),
		ChatRpc:  chat.NewChat(zrpc.MustNewClient(c.ChatRpc, zrpc.WithUnaryClientInterceptor(zrpc_interceptor.ClientInfoInterceptor))),
		GroupRpc: users.NewUsers(zrpc.MustNewClient(c.GroupRpc, zrpc.WithUnaryClientInterceptor(zrpc_interceptor.ClientInfoInterceptor))),
	}
}
