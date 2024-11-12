package svc

// 这是一个服务层的实现，负责业务逻辑的处理
// 导入必要的依赖包
import (
	"fim/core"
	"fim/fim_auth/auth_api/internal/config"
	"fim/fim_user/user_rpc/types/user_rpc"
	"fim/fim_user/user_rpc/users"
	"github.com/go-redis/redis"
	"github.com/zeromicro/go-zero/zrpc"
	"gorm.io/gorm"
)

// ServiceContext 定义了服务上下文结构体，包含了服务运行所需的各种依赖
// 这些依赖包括配置信息、数据库连接、Redis客户端以及用户RPC服务客户端
type ServiceContext struct {
	Config  config.Config        // 配置信息，包含了数据库、Redis和用户RPC服务的配置
	DB      *gorm.DB             // 数据库连接
	Redis   *redis.Client        // Redis客户端
	UserRpc user_rpc.UsersClient // 用户RPC服务客户端
}

// NewServiceContext 根据配置信息初始化服务上下文
// 该函数负责建立数据库连接、初始化Redis客户端以及创建用户RPC服务客户端
// 参数c是配置信息，包含了服务运行所需的各项配置
// 返回一个指向ServiceContext的指针，包含了初始化后的各种依赖
func NewServiceContext(c config.Config) *ServiceContext {
	// 初始化MySQL数据库连接
	mysqlDb := core.InitGorm(c.Mysql.DataSource)
	// 初始化Redis客户端
	client := core.InitRedis(c.Redis.Addr, c.Redis.Password, c.Redis.DB)
	// 创建用户RPC服务客户端
	userRpc := users.NewUsers(zrpc.MustNewClient(c.UserRpc))
	// 返回初始化后的服务上下文
	return &ServiceContext{
		Config:  c,
		DB:      mysqlDb,
		Redis:   client,
		UserRpc: userRpc,
	}
}
