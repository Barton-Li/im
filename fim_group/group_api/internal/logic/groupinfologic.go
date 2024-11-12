package logic

import (
	"context"
	"errors"
	"fim/fim_group/group_api/internal/svc"
	"fim/fim_group/group_api/internal/types"
	"fim/fim_group/group_models"
	"fim/fim_user/user_rpc/types/user_rpc"
	"fim/utils/set"

	"github.com/zeromicro/go-zero/core/logx"
)

type GroupInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGroupInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupInfoLogic {
	return &GroupInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// GroupInfo 根据请求参数获取群组信息。
// req: 请求参数，包含群组ID和用户ID。
// 返回值:
//
//	resp: 群组信息的响应对象。
//	err: 错误信息，如果查询过程中出现问题。
func (l *GroupInfoLogic) GroupInfo(req *types.GroupInfoRequest) (resp *types.GroupInfoResponse, err error) {
	// 初始化群组模型对象
	var groupModel group_models.GroupModel
	// 查询数据库，预加载群组的成员列表，根据群组ID获取群组信息
	err = l.svcCtx.DB.Preload("MemberList").Take(&groupModel, req.ID).Error
	if err != nil {
		return nil, errors.New("群不存在")
	}

	// 初始化群组成员模型对象
	var member group_models.GroupMemberModel
	// 再次查询数据库，确认当前用户是否为群组成员
	err = l.svcCtx.DB.Take(&member, "group_id =? AND user_id =?", req.ID, req.UserID).Error
	if err != nil {
		return nil, errors.New("您不是群成员")
	}

	// 构建群组信息的响应对象
	resp = &types.GroupInfoResponse{
		GroupID:         groupModel.ID,
		Title:           groupModel.Title,
		Abstract:        groupModel.Abstract,
		MemberCount:     len(groupModel.MemberList),
		Avatar:          groupModel.Avatar,
		Role:            member.Role,
		IsProhibition:   groupModel.IsProhibition,
		ProhibitionTime: member.GetProhibitionTime(l.svcCtx.Redis, l.svcCtx.DB),
	}

	// 初始化列表，用于存储群组中的用户ID
	var userIDList []uint32
	var userAllIDList []uint32
	// 遍历群组成员列表，筛选出管理员和普通成员的用户ID
	for _, model := range groupModel.MemberList {
		if model.Role == 1 || model.Role == 2 {
			userIDList = append(userIDList, uint32(model.UserID))
		}
		userAllIDList = append(userAllIDList, uint32(model.UserID))
	}

	// 调用用户RPC服务，获取用户详细信息
	userListResponse, err := l.svcCtx.UserRpc.UserListInfo(l.ctx, &user_rpc.UserListInfoRequest{
		UserIdList: userIDList,
	})
	if err != nil {
		return
	}

	// 初始化变量，用于存储群主信息
	var creator types.UserInfo
	// 初始化列表，用于存储管理员信息
	var adminList = make([]types.UserInfo, 0)
	// 调用用户RPC服务，获取在线用户列表
	userOnlineResponse, err := l.svcCtx.UserRpc.UserOnlineList(l.ctx, &user_rpc.UserOnlineListRequest{})
	if err == nil {
		// 计算在线成员数量
		slice := set.Intersect(userOnlineResponse.UserIdList, userAllIDList)
		resp.MemberOnlineCount = len(slice)
	}

	// 遍历群组成员列表，构建群主和管理员的信息
	for _, model := range groupModel.MemberList {
		if model.Role == 3 {
			continue
		}
		userInfo := types.UserInfo{
			UserID:   model.UserID,
			Avatar:   userListResponse.UserInfo[uint32(model.UserID)].Avatar,
			Nickname: userListResponse.UserInfo[uint32(model.UserID)].NickName,
		}
		if model.Role == 1 {
			creator = userInfo
			continue
		}
		if model.Role == 2 {
			adminList = append(adminList, userInfo)
		}
	}

	// 设置响应对象中的群主和管理员信息
	resp.Creator = creator
	resp.AdminList = adminList

	// 如果当前用户是群主或管理员，补充群组的搜索、验证等信息
	if member.Role == 1 || member.Role == 2 {
		resp.IsSearch = groupModel.IsSearch
		resp.Verification = &groupModel.Verification
		resp.IsInvite = &groupModel.IsInvite
		resp.IsTemporarySession = &groupModel.IsTemporarySession
		// 如果群组有验证问题，设置响应对象中的验证问题信息
		if groupModel.VerificationQuestion != nil {
			resp.VerificationQuestion = &types.VerificationQuestion{
				Problem1: groupModel.VerificationQuestion.Problem1,
				Problem2: groupModel.VerificationQuestion.Problem2,
				Problem3: groupModel.VerificationQuestion.Problem3,
				Answer1:  groupModel.VerificationQuestion.Answer1,
				Answer2:  groupModel.VerificationQuestion.Answer2,
				Answer3:  groupModel.VerificationQuestion.Answer3,
			}
		}
	}

	return
}
