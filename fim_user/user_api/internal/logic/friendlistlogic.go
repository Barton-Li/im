package logic

import (
	"context"
	"fim/common/list_query"
	"fim/common/models"
	"fim/fim_user/user_models"
	"strconv"

	"fim/fim_user/user_api/internal/svc"
	"fim/fim_user/user_api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type FriendListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewFriendListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FriendListLogic {
	return &FriendListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// FriendList 获取用户的朋友列表
// @param req 包含页码和每页数量的请求参数
// @return resp 包含朋友信息的响应对象
// @return err 可能发生的错误
func (l *FriendListLogic) FriendList(req *types.FriendListRequest) (resp *types.FriendListResponse, err error) {
	// 使用ListQuery查询朋友信息，包括分页和预加载用户信息
	friends, count, _ := list_query.ListQuery(l.svcCtx.DB, user_models.FriendModel{}, list_query.Option{
		PageInfo: models.PagaInfo{
			Page:  req.Page,
			Limit: req.Limit,
		},
		Where:   l.svcCtx.DB.Where("send_user_id=? or recv_user_id=?", req.UserID, req.UserID),
		Preload: []string{"SendUserModel", "RecvUserModel"},
	})

	// 从Redis获取在线用户列表
	// 查询用户在线
	onlineMap := l.svcCtx.Redis.HGetAll("online").Val()
	var onlineUserMap = map[uint]bool{}
	for key, _ := range onlineMap {
		val, err1 := strconv.Atoi(key)
		if err1 != nil {
			logx.Error(err1)
			continue
		}
		onlineUserMap[uint(val)] = true
	}

	// 处理朋友信息，构建FriendInfoResponse列表
	var list []types.FriendInfoResponse
	for _, friend := range friends {
		info := types.FriendInfoResponse{}
		if friend.SendUserID == req.UserID {
			info = types.FriendInfoResponse{
				UserID:   friend.RevUserID,
				Nickname: friend.RevUserModel.Nickname,
				Abstract: friend.RevUserModel.Abstract,
				Avatar:   friend.RevUserModel.Avatar,
				Notice:   friend.SenUserNotice,
				IsOnline: onlineUserMap[friend.RevUserID],
			}
		}
		if friend.RevUserID == req.UserID {
			info = types.FriendInfoResponse{
				UserID:   friend.SendUserID,
				Nickname: friend.SendUserModel.Nickname,
				Abstract: friend.SendUserModel.Abstract,
				Avatar:   friend.SendUserModel.Avatar,
				Notice:   friend.RevUserNotice,
				IsOnline: onlineUserMap[friend.SendUserID],
			}
		}
		list = append(list, info)
	}

	// 构建并返回朋友列表响应
	return &types.FriendListResponse{
		Count: int(count),
		List:  list,
	}, nil
}
