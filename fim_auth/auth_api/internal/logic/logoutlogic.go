package logic

import (
	"context"
	"errors"
	"fim/fim_auth/auth_api/internal/svc"
	"fim/utils/jwts"
	"fmt"
	"github.com/zeromicro/go-zero/core/logx"
	"time"
)

type LogoutLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLogoutLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LogoutLogic {
	return &LogoutLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// Logout 实现用户注销功能。
// 它接收一个token作为参数，用于验证用户身份并执行注销操作。
// 如果token有效，它将在Redis中设置一个键，标记该token已注销。
// 参数:
//   token - 用户身份验证的令牌。
// 返回值:
//   resp - 注销操作的结果信息。
//   err - 如果操作失败，返回错误信息。
func (l *LogoutLogic) Logout(token string) (resp string, err error) {
	// 检查是否提供了token，如果没有，返回错误。
	if token == "" {
		err = errors.New("请提供token")
	}

	// 尝试解析token，验证其有效性。
	payload, err := jwts.ParseToken(token, l.svcCtx.Config.Auth.AccessSecret)
	if err != nil {
		err = errors.New("token无效")
		return
	}

	// 获取当前时间，用于计算token的过期时间。
	now := time.Now()
	// 计算token的过期时间与当前时间的间隔。
	expiration := payload.ExpiresAt.Time.Sub(now)

	// 构建Redis键名，用于标记该token已注销。
	key := fmt.Sprintf("logout_%s", token)
	// 在Redis中设置键值对，如果键不存在，则设置成功，表示该token已注销。
	l.svcCtx.Redis.SetNX(key, " ", expiration)

	// 注销成功，返回相应信息。
	resp = "注销成功"
	return
}

