package logic

import (
	"context"
	"fim/fim_group/group_models"
	"fim/fim_user/user_rpc/types/user_rpc"

	"fim/fim_group/group_api/internal/svc"
	"fim/fim_group/group_api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GroupfriendsListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGroupfriendsListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupfriendsListLogic {
	return &GroupfriendsListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// GroupfriendsList 根据用户的ID获取其好友列表，并筛选出哪些好友也在指定的群组中。
// 它接收一个GroupfriendsListRequest对象作为请求参数，返回一个GroupfriendsListResponse对象和一个错误对象。
// 如果操作成功，返回的好友列表将包含用户的头像、昵称、用户ID以及他们是否在指定群组中的信息。
func (l *GroupfriendsListLogic) GroupfriendsList(req *types.GroupfriendsListRequest) (resp *types.GroupfriendsListResponse, err error) {
	// 调用远程过程调用（RPC）服务获取当前用户的全部好友列表。
	friendResponse, err := l.svcCtx.UserRpc.FriendList(l.ctx, &user_rpc.FriendListRequest{
		User: uint32(req.UserID),
	})
	if err != nil {
		// 如果获取好友列表时发生错误，记录错误并返回。
		logx.Error(err)
		return nil, err
	}

	// 初始化一个空的群组成员列表，用于存储指定群组的所有成员。
	var memberList []group_models.GroupMemberModel
	// 从数据库中查询指定群组的成员。
	l.svcCtx.DB.Find(&memberList, "group_id =?", req.ID)

	// 创建一个映射，用于快速检查用户是否是群组成员。
	var memberMap = map[uint]bool{}
	// 遍历群组成员列表，将用户ID作为键，值为true表示该用户是群组成员。
	for _, model := range memberList {
		memberMap[model.UserID] = true
	}

	// 初始化响应对象。
	resp = new(types.GroupfriendsListResponse)
	// 遍历好友列表，构建最终的响应列表。
	for _, info := range friendResponse.FriendList {
		// 将好友信息添加到响应列表，同时标记该好友是否也在目标群组中。
		resp.List = append(resp.List, types.GroupfriendsResponse{
			UserID:    uint(info.UserId),
			Avatar:    info.Avatar,
			Nickname:  info.NickName,
			IsInGroup: memberMap[uint(info.UserId)],
		})
	}

	// 返回构建好的响应对象和可能的错误。
	return
}
