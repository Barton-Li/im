package config

import (
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	rest.RestConf
	Etcd  string
	Mysql struct {
		DataSource string
	}
	UserRpc  zrpc.RpcClientConf
	ChatRpc  zrpc.RpcClientConf
	GroupRpc zrpc.RpcClientConf
	Redis    struct {
		Addr     string
		Password string
		DB       int
	}
}
