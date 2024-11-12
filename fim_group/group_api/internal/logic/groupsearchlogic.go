package logic

import (
	"context"
	"fim/common/list_query"
	"fim/common/models"
	"fim/fim_group/group_models"
	"fim/fim_user/user_rpc/types/user_rpc"
	"fim/utils/set"
	"fmt"

	"fim/fim_group/group_api/internal/svc"
	"fim/fim_group/group_api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GroupSearchLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGroupSearchLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupSearchLogic {
	return &GroupSearchLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// GroupSearch 实现了根据用户请求搜索群组的功能。
// 它接收一个 GroupSearchRequest 请求对象，返回一个 GroupSearchListResponse 响应对象和一个错误对象。
// 主要逻辑包括从数据库中查询群组列表，获取在线用户ID列表，并根据这些数据生成响应列表。
func (l *GroupSearchLogic) GroupSearch(req *types.GroupSearchRequest) (resp *types.GroupSearchListResponse, err error) {
	// 使用 ListQuery 方法根据搜索条件查询群组列表和总数
	groups, count, err := list_query.ListQuery(l.svcCtx.DB, group_models.GroupModel{}, list_query.Option{
		PageInfo: models.PagaInfo{
			Page:  req.Page,
			Limit: req.Limit,
		},
		Preload: []string{"MemberList"},
		Where:   l.svcCtx.DB.Where("is_search=1 and (id =? or title like ?)", req.Key, fmt.Sprintf("%%%s%%", req.Key)),
	})
	if err != nil {
		return nil, err // 如果查询出错，返回错误
	}

	// 调用 UserRpc 服务获取在线用户列表
	userOnlineResponse, err := l.svcCtx.UserRpc.UserOnlineList(l.ctx, &user_rpc.UserOnlineListRequest{})
	var userOnlineIDList []uint
	if err == nil {
		// 如果调用成功，转换在线用户ID列表
		for _, u := range userOnlineResponse.UserIdList {
			userOnlineIDList = append(userOnlineIDList, uint(u))
		}
	}

	// 初始化响应对象
	resp = new(types.GroupSearchListResponse)
	// 遍历查询到的群组，生成响应列表
	for _, group := range groups {
		var groupMemberIdList []uint
		var isInGroup bool
		// 遍历群组成员，记录成员ID列表和判断用户是否在群内
		for _, model := range group.MemberList {
			groupMemberIdList = append(groupMemberIdList, model.ID)
			if model.UserID == req.UserID {
				isInGroup = true
			}
		}
		// 计算群组在线人数，并添加到响应列表
		resp.List = append(resp.List, types.GroupSearchResponse{
			GroupID:         group.ID,
			Title:           group.Title,
			Abstract:        group.Abstract,
			Avatar:          group.Avatar,
			UserCount:       len(group.MemberList),
			UserOnlineCount: len(set.Intersect(groupMemberIdList, userOnlineIDList)),
			IsInGroup:       isInGroup,
		})
	}
	// 设置响应列表总数
	resp.Count = int(count)

	return resp, nil
}
