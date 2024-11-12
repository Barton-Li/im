package logic

import (
	"context"
	"errors"
	"fim/common/list_query"
	"fim/common/models"
	"fim/common/models/ctype"
	"fim/fim_group/group_api/internal/svc"
	"fim/fim_group/group_api/internal/types"
	"fim/fim_group/group_models"
	"fim/fim_user/user_rpc/types/user_rpc"
	"fmt"
	"github.com/zeromicro/go-zero/core/logx"
)

type GroupMemberLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGroupMemberLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupMemberLogic {
	return &GroupMemberLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

type Data struct {
	GroupID         uint   `gorm:"column:group_id"`
	UserID          uint   `gorm:"column:user_id"`
	Role            int8   `gorm:"column:role"`
	CreatedAt       string `gorm:"column:create_at"`
	MemberNickname  string `gorm:"column:member_nickname"`
	NewMsgDate      string `gorm:"column:new_msg_date"`
	ProhibitionTime *int   `gorm:"column:prohibition_time"`
}

// GroupMember 根据请求中的排序方式和群组ID获取群成员信息。
// 参数 req: GroupMemberRequest对象，包含排序方式、页码、每页限制和群组ID等信息。
// 返回值 resp: GroupMemberResponse对象，包含成员列表和总数等信息。
// 返回值 err: 错误信息，如果执行过程中发生错误，会返回具体的错误信息。
func (l *GroupMemberLogic) GroupMember(req *types.GroupMemberRequest) (resp *types.GroupMemberResponse, err error) {
	// 根据请求的排序方式进行合法性校验
	switch req.Sort {
	case "new_msg_date desc", "new_msg_date asc":
	case "role asc":
	case "create_at desc", "create_at asc":
	default:
		// 如果排序方式不支持，则返回错误
		return nil, errors.New("不支持的排序模式")
	}
	// 构造新消息日期的SQL子查询字符串
	column := fmt.Sprintf(fmt.Sprintf("(select group_msg_models.created_at from group_msg_models where group_msg_models.group_id=%d and group_msg_models.send_user_id=user_id order by created_at desc limit 1)as new_msg_date", req.ID))
	// 执行列表查询并获取成员列表和总数
	memberList, count, _ := list_query.ListQuery(l.svcCtx.DB, Data{}, list_query.Option{
		PageInfo: models.PagaInfo{
			Page:  req.Page,
			Limit: req.Limit,
			Sort:  req.Sort,
		},
		Where: l.svcCtx.DB.Where("group_id=?", req.ID),
		Table: func() (string, any) {
			return "(?)as u", l.svcCtx.DB.Model(&group_models.GroupMemberModel{GroupID: req.ID}).
				Select("group_id",
					"user_id",
					"role",
					"created_at",
					"member_nickname",
					"prohibition_time",
					column)
		},
	})
	// 初始化用户ID列表
	var userIDList []uint32
	// 从成员列表中提取用户ID
	for _, data := range memberList {
		userIDList = append(userIDList, uint32(data.UserID))
	}
	// 初始化用户信息映射
	var userInfoMap = map[uint]ctype.UserInfo{}
	// 调用用户RPC服务获取用户详细信息
	userListResponse, err := l.svcCtx.UserRpc.UserListInfo(l.ctx, &user_rpc.UserListInfoRequest{
		UserIdList: userIDList,
	})
	if err == nil {
		// 将用户详细信息填充到映射中
		for u, info := range userListResponse.UserInfo {
			userInfoMap[uint(u)] = ctype.UserInfo{
				ID:       uint(u),
				NickName: info.NickName,
				Avatar:   info.Avatar,
			}
		}
	} else {
		logx.Error(err)
	}
	// 初始化好友映射
	var friendMap = map[uint]bool{}
	// 调用用户RPC服务获取好友列表
	friendResponse, err := l.svcCtx.UserRpc.FriendList(l.ctx, &user_rpc.FriendListRequest{
		User: uint32(req.UserID),
	})
	if err != nil {
		logx.Error(err)
		return nil, errors.New("获取好友列表失败")
	}
	// 将好友关系填充到映射中
	for _, info := range friendResponse.FriendList {
		friendMap[uint(info.UserId)] = true
	}
	// 初始化用户在线状态映射
	var userOnlineMap = map[uint]bool{}
	// 调用用户RPC服务获取用户在线列表
	userOnlineResponse, err := l.svcCtx.UserRpc.UserOnlineList(l.ctx, &user_rpc.UserOnlineListRequest{})
	if err == nil {
		// 将用户在线状态填充到映射中
		for _, u := range userOnlineResponse.UserIdList {
			userOnlineMap[uint(u)] = true
		}
	} else {
		logx.Error(err)
	}
	// 初始化响应对象
	resp = new(types.GroupMemberResponse)
	// 填充成员信息到响应列表中
	for _, data := range memberList {
		resp.List = append(resp.List, types.GroupMemberInfo{
			UserID:          data.UserID,
			UserNickname:    userInfoMap[data.UserID].NickName,
			Avatar:          userInfoMap[data.UserID].Avatar,
			IsOnline:        userOnlineMap[data.UserID],
			Role:            data.Role,
			MemberNickname:  data.MemberNickname,
			CreatedAt:       data.CreatedAt,
			NewMsgDate:      data.NewMsgDate,
			IsFriend:        friendMap[data.UserID],
			ProhibitionTime: data.ProhibitionTime,
		})
	}
	// 设置响应的成员总数
	resp.Count = int(count)
	return
}
