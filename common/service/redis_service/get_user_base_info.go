package redis_service

import (
	"context"
	"encoding/json"
	"fim/common/models/ctype"
	"fim/fim_user/user_rpc/types/user_rpc"
	"fmt"
	"github.com/go-redis/redis"
	"time"
)

// GetUserBaseInfo 通过Redis缓存和用户RPC服务获取用户基本信息。
// 该函数首先尝试从Redis缓存中获取指定userID的用户信息。如果缓存中不存在，
// 则通过用户RPC服务调用UserBaseInfo接口获取用户基本信息。获取到信息后，
// 将信息写入Redis缓存，并返回用户基本信息和可能的错误。
//
// 参数:
// - client: Redis客户端，用于与Redis服务器进行通信。
// - userRpc: 用户RPC服务客户端，用于调用用户相关的RPC接口。
// - userID: 用户ID，用于标识需要获取信息的用户。
//
// 返回值:
// - userInfo: 用户基本信息结构体，包含用户ID、头像、昵称等信息。
// - err: 可能发生的错误，如果执行过程中无错误发生，则err为nil。
func GetUserBaseInfo(client *redis.Client, userRpc user_rpc.UsersClient, userID uint) (userInfo ctype.UserInfo, err error) {
	// 根据用户ID生成Redis键名
	key := fmt.Sprintf("fim_server_uers_%d", userID)
	// 尝试从Redis缓存中获取用户信息字符串
	str, err := client.Get(key).Result()
	// 如果缓存获取失败（可能是用户信息不存在于缓存中）
	if err != nil {
		// 通过RPC调用获取用户基本信息
		userBaseResponse, err1 := userRpc.UserBaseInfo(context.Background(), &user_rpc.UserBaseInfoRequest{
			UserId: uint32(userID),
		})
		// 检查RPC调用是否出错
		if err1 != nil {
			err = err1
			return
		}
		// 如果RPC调用成功，清空错误状态
		err = nil
		// 将获取到的用户信息填充到userInfo结构体中
		userInfo.ID = userID
		userInfo.Avatar = userBaseResponse.Avatar
		userInfo.NickName = userBaseResponse.NickName
		// 将用户信息序列化为JSON字符串
		byteData, _ := json.Marshal(userInfo)
		// 将用户信息写入Redis缓存，设置过期时间为1小时
		client.Set(key, string(byteData), time.Hour)
		// 返回用户信息和错误状态
		return
	}
	// 如果缓存获取成功，反序列化字符串到userInfo结构体
	err = json.Unmarshal([]byte(str), &userInfo)
	// 检查反序列化是否出错
	if err != nil {
		return
	}
	// 返回用户信息和错误状态
	return
}
