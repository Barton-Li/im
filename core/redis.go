package core

import (
	"context"
	"github.com/go-redis/redis"
	"time"
)

func InitRedis(addr string, pwd string, db int) (client *redis.Client) {
	client = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: pwd, //无密码
		DB:       db,  //数据库
		PoolSize: 100, //连接池大小
	})
	// 创建一个带有超时的上下文，用于控制Ping操作的时间限制。
	_, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	// 尝试与Redis服务器建立连接并验证。
	_, err := client.Ping().Result()
	if err != nil {
		panic(err)
		return
	}
	return
}
