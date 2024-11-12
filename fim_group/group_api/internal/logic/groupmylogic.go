package logic

import (
	"context"
	"fim/common/list_query"
	"fim/common/models"
	"fim/fim_group/group_api/internal/svc"
	"fim/fim_group/group_api/internal/types"
	"fim/fim_group/group_models"

	"github.com/zeromicro/go-zero/core/logx"
)

type GroupMyLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGroupMyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupMyLogic {
	return &GroupMyLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// GroupMy 根据用户ID查询用户所在的群组列表，并根据模式筛选群组角色。
// req为请求参数，包含用户ID、模式等信息。
// 返回值resp为群组列表的响应数据，包含列表和总数。
// err为执行过程中可能的错误。
func (l *GroupMyLogic) GroupMy(req *types.GroupMyRequest) (resp *types.GroupMyListResponse, err error) {
	// 初始化群组ID列表
	var groupIDList []uint
	// 构建查询：根据用户ID筛选群组成员
	query := l.svcCtx.DB.Model(&group_models.GroupMemberModel{}).Where("user_id=?", req.UserID)
	// 如果模式为1，进一步筛选角色为1的群组
	if req.Mode == 1 {
		query.Where("role=?", 1)
	}
	// 选择查询结果为群组ID，并将结果扫描到groupIDList中
	query.Select("group_id").Scan(&groupIDList)
	// 根据群组ID列表查询群组列表和总数
	groupList, count, _ := list_query.ListQuery(l.svcCtx.DB, group_models.GroupModel{}, list_query.Option{
		PageInfo: models.PagaInfo{
			Page:  req.Page,
			Limit: req.Limit,
		},
		Preload: []string{"MemberList"},
		Where:   l.svcCtx.DB.Where("id in ?", groupIDList),
	})
	// 初始化响应数据
	resp = new(types.GroupMyListResponse)
	// 遍历群组列表，处理每个群组
	for _, model := range groupList {
		// 初始化角色
		var role int8
		// 遍历群组成员，查找目标用户的角色
		for _, memberModel := range model.MemberList {
			if memberModel.UserID == req.UserID {
				role = memberModel.Role
			}
		}
		// 构建群组响应数据，并添加到列表中
		resp.List = append(resp.List, types.GroupMyResponse{
			GroupID:     model.ID,
			GroupTitle:  model.Title,
			GroupAvatar: model.Avatar,
			Role:        role,
			Mode:        req.Mode,
		})
	}
	// 设置响应数据的总数
	resp.Count = int(count)
	// 返回群组列表响应数据
	return
}
