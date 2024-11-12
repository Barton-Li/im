package etcd

import (
	"context"
	"fim/core"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/netx"
	"strings"
)

// DeliveryAddress 将服务地址注册到ETCD中。
// 参数:
// etcdAddr: ETCD的地址，用于初始化ETCD客户端。
// serviceName: 服务的名称，用于在ETCD中标识服务。
// addr: 服务的地址，格式为"IP:端口"。
// 该函数首先校验地址格式，然后将绑定到所有网络接口的地址替换为内网IP，
// 最后将服务地址注册到ETCD。
func DeliveryAddress(etcdAddr string, serviceName string, addr string) {
	// 将地址按冒号分割以提取IP和端口。
	list := strings.Split(addr, ":")
	// 检查地址格式是否正确。
	if len(list) != 2 {
		logx.Errorf("地址错误 %s", addr)
		return
	}
	// 如果IP是0.0.0.0，则替换为内网IP。
	if list[0] == "0.0.0.0" {
		ip := netx.InternalIp()
		addr = strings.ReplaceAll(addr, "0.0.0.0", ip)
	}
	// 初始化ETCD客户端。
	client := core.InitEtcd(etcdAddr)
	// 将服务地址注册到ETCD。
	_, err := client.Put(context.Background(), serviceName, addr)
	// 如果注册失败，记录错误信息。
	if err != nil {
		logx.Errorf("地址上发送失败%s", err.Error())
		return
	}
	// 如果注册成功，记录日志。
	logx.Infof("地址上发送成功 %s", serviceName, addr)
}

// GetAddress 通过服务名称从etcd获取服务地址。
// 参数:
//   etcdAddr: etcd的地址，用于初始化etcd客户端。
//   serviceName: 需要查询的服务名称。
// 返回值:
//   addr: 查询到的服务地址，如果未查询到则为空字符串。
func GetAddress(etcdAddr string, serviceName string) (addr string) {
	// 初始化etcd客户端
	client := core.InitEtcd(etcdAddr)

	// 使用context.Background()作为上下文，调用etcd客户端的Get方法查询serviceName对应的服务地址
	res, err := client.Get(context.Background(), serviceName)

	// 如果查询成功且结果集不为空，则返回查询结果的第一个服务地址
	if err == nil && len(res.Kvs) > 0 {
		return string(res.Kvs[0].Value)
	}

	// 如果查询失败或结果集为空，则返回空字符串
	return ""
}

