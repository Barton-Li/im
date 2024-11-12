package logic

import (
	"context"
	"fim/common/list_query"
	"fim/common/models"
	"fim/fim_group/group_models"
	"fim/fim_user/user_rpc/types/user_rpc"

	"fim/fim_group/group_api/internal/svc"
	"fim/fim_group/group_api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GroupValidListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGroupValidListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupValidListLogic {
	return &GroupValidListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GroupValidListLogic) GroupValidList(req *types.GroupValidListRequest) (resp *types.GroupValidListResponse, err error) {
	var groupIDList []uint //我管理的群
	l.svcCtx.DB.Model(group_models.GroupMemberModel{}).Where("user_id =? and (role=1 or role=2)", req.UserID).Select("group_id").Scan(&groupIDList)
	var groupMap = map[uint]bool{}
	for _, u := range groupIDList {
		groupMap[u] = true
	}
	groups, count, err := list_query.ListQuery(l.svcCtx.DB, group_models.GroupVerifyModel{}, list_query.Option{
		PageInfo: models.PagaInfo{
			Page:  req.Page,
			Limit: req.Limit,
		},
		Preload: []string{"GroupModel"},
		Where:   l.svcCtx.DB.Where("group_id IN? or user_id=? ", groupIDList, req.UserID),
	})
	var userIDList []uint32
	for _, group := range groups {
		userIDList = append(userIDList, uint32(group.UserID))
	}
	userList, err1 := l.svcCtx.UserRpc.UserListInfo(l.ctx, &user_rpc.UserListInfoRequest{
		UserIdList: userIDList,
	})
	resp = new(types.GroupValidListResponse)
	resp.Count = int(count)
	for _, groupVerify := range groups {
		info := types.GroupValidInfoResponse{
			ID:                groupVerify.ID,
			GroupID:           groupVerify.GroupID,
			UserID:            groupVerify.UserID,
			Status:            groupVerify.Status,
			AddtionalMessages: groupVerify.AdditionalMessages,
			Title:             groupVerify.GroupModel.Title,
			CreatedAt:         groupVerify.CreatedAt.String(),
			Type:              groupVerify.Type,
			Avator:            groupVerify.GroupModel.Avatar,
			Flage:             "send", // 我是发送着
		}
		if groupVerify.VerificationQuestion != nil {
			info.VerificationQuestion = &types.VerificationQuestion{
				Problem1: groupVerify.VerificationQuestion.Problem1,
				Problem2: groupVerify.VerificationQuestion.Problem2,
				Problem3: groupVerify.VerificationQuestion.Problem3,
				Answer1:  groupVerify.VerificationQuestion.Answer1,
				Answer2:  groupVerify.VerificationQuestion.Answer2,
				Answer3:  groupVerify.VerificationQuestion.Answer3,
			}
		}
		// 怎么判断我是加群方还是验证方呢？
		// 只需要判断 groupVerify.GroupID
		if groupMap[groupVerify.GroupID] {
			info.Flage = "rev" // 我是接收者
		}
		if err1 != nil {
			info.UserNickname = userList.UserInfo[uint32(info.UserID)].NickName
			info.Avator = userList.UserInfo[uint32(info.UserID)].Avatar
		}
		resp.List = append(resp.List, info)
	}

	return
}
