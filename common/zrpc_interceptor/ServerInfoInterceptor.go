package zrpc_interceptor

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// ServerUnaryInterceptor 是一个GRPC服务端的单向拦截器。
// 它的作用是在实际处理请求之前，从传入的元数据中提取客户端IP和UserID，
// 并将这些值存入上下文中，以便后续处理可以访问到这些信息。
// 这对于日志记录、权限验证等功能非常有用。
//
// 参数:
//
//	ctx: 请求的上下文，用于传递值、取消请求等
//	req: 客户端发送的原始请求
//	info: 包含即将处理的请求的方法信息
//	handler: 实际处理请求的GRPC handler函数
//
// 返回值:
//
//	resp: 处理请求后返回的响应
//	err: 如果处理过程中发生错误，会返回错误信息
func ServerUnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	// 从上下文中提取客户端IP
	clientIP := metadata.ValueFromIncomingContext(ctx, "clientIP")
	// 从上下文中提取UserID
	userID := metadata.ValueFromIncomingContext(ctx, "userID")
	// 如果客户端IP存在，将其存入上下文中
	if len(clientIP) > 0 {
		ctx = context.WithValue(ctx, "clientIP", clientIP[0])
	}
	// 如果UserID存在，将其存入上下文中
	if len(userID) > 0 {
		ctx = context.WithValue(ctx, "userID", userID[0])
	}
	// 调用实际的请求处理函数，并返回其结果
	return handler(ctx, req)
}
