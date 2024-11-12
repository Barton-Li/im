package zrpc_interceptor

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// ClientInfoInterceptor 是一个gRPC客户端拦截器，用于在客户端调用服务时添加客户端IP和UserID信息到元数据中。
// 这样做可以在服务端获取到这些信息，便于实现更细粒度的权限控制、审计等功能。
// 参数:
// - ctx: 上下文，用于传递请求范围内的值
// - method: 被调用的方法名
// - req: 请求消息
// - reply: 响应消息
// - cc: 客户端连接
// - invoker: 实际调用方法的调用器
// - opts: 调用选项
// 返回值:
// - error: 调用过程中可能产生的错误
func ClientInfoInterceptor(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	// 初始化clientIP和userID变量，用于存储从上下文中获取的值
	var clientIP, userID string

	// 从上下文中获取clientIP值，如果存在则赋值给clientIP变量
	cl := ctx.Value("clientIP")
	if cl != nil {
		clientIP = cl.(string)
	}

	// 从上下文中获取userID值，如果存在则赋值给userID变量
	ui := ctx.Value("userID")
	if ui != nil {
		userID = ui.(string)
	}

	// 创建一个新的元数据，将clientIP和userID添加进去
	md := metadata.New(map[string]string{"clientIP": clientIP, "userID": userID})

	// 使用新创建的元数据创建一个新的上下文，用于后续的调用
	ctx = metadata.NewOutgoingContext(context.Background(), md)

	// 调用实际的调用器进行方法调用，并返回可能的错误
	err := invoker(ctx, method, req, reply, cc, opts...)
	return err
}
