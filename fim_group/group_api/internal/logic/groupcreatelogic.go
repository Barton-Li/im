package logic

import (
	"context"
	"encoding/json"
	"errors"
	"fim/fim_group/group_api/internal/svc"
	"fim/fim_group/group_api/internal/types"
	"fim/fim_group/group_models"
	"fim/fim_user/user_models"
	"fim/fim_user/user_rpc/types/user_rpc"
	"fim/utils"
	"fim/utils/set"
	"fmt"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

type GroupCreateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGroupCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupCreateLogic {
	return &GroupCreateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// GroupCreate 实现了群组创建逻辑。
// req 是群组创建请求的参数，包括群组的详细信息和创建者信息。
// 返回值 resp 是群组创建成功后的响应，包括群组ID和其他信息。
// 如果创建过程中出现错误，会返回错误信息 err。
func (l *GroupCreateLogic) GroupCreate(req *types.GroupCreateRequest) (resp *types.GroupCreateResponse, err error) {
	// 初始化群组模型，设置默认值和从请求中获取的值。
	var groupModel = group_models.GroupModel{
		Creator:      req.UserID,
		Abstract:     fmt.Sprintf("本群创建于%s:  群主很懒,什么都没有留下", time.Now().Format("2006-01-02")),
		IsSearch:     false,
		Size:         50,
		Verification: 2,
	}

	// 调用用户RPC服务获取创建者的详细信息。
	userInfo, err := l.svcCtx.UserRpc.UserInfo(l.ctx, &user_rpc.UserInfoRequest{
		UserId: uint32(req.UserID),
	})
	if err != nil {
		logx.Error(err)
		return nil, errors.New("获取用户信息失败")
	}

	// 将RPC返回的用户信息转换为用户模型。
	var userInfoModel user_models.UserModel
	json.Unmarshal(userInfo.Data, &userInfoModel)

	// 检查用户是否被允许创建群组。
	if userInfoModel.UserConfModel.CurtailCreateGroup {
		return nil, errors.New("您已被禁止创建群聊")
	}

	// 初始化群成员列表，最初只包含创建者。
	var groupUserList = []uint{req.UserID}

	// 根据请求中的模式创建群组。
	switch req.Mode {
	case 1: // 直接创建
		// 检查群名称和群大小是否符合要求。
		if req.Name == "" {
			return nil, errors.New("群名称不能为空")
		}
		if req.Size >= 1000 {
			return nil, errors.New("群人数不能超过1000人")
		}
		// 设置群组信息。
		groupModel.Title = req.Name
		groupModel.Size = req.Size
		groupModel.IsSearch = req.IsSearch
	case 2: // 选人创建
		// 检查是否选择了成员。
		if len(req.UserIDList) == 0 {
			return nil, errors.New("请选择创建群聊的人")
		}
		// 处理选中的成员列表，添加到群成员列表中。
		var userIDList = []uint32{uint32(req.UserID)}
		for _, u := range req.UserIDList {
			userIDList = append(userIDList, uint32(u))
			groupUserList = append(groupUserList, u)
		}
		// 确保群成员列表不包含重复的用户ID。
		groupUserList = utils.DeduplicationList(groupUserList)

		// 获取创建者的好友列表，用于验证选中的成员是否都是好友。
		userFriendResponse, err := l.svcCtx.UserRpc.FriendList(l.ctx, &user_rpc.FriendListRequest{
			User: uint32(req.UserID),
		})
		if err != nil {
			logx.Error(err)
			return nil, errors.New("获取好友列表失败")
		}

		// 检查选中的成员是否都在创建者的好友列表中。
		var friendIDList []uint
		for _, i2 := range userFriendResponse.FriendList {
			friendIDList = append(friendIDList, uint(i2.UserId))
		}
		slice := set.Difference(req.UserIDList, friendIDList)
		if len(slice) != 0 {
			return nil, errors.New("选择的好友列表中有人不是好友")
		}

		// 获取选中成员的详细信息，用于设置群名称。
		userListResponse, err1 := l.svcCtx.UserRpc.UserListInfo(l.ctx, &user_rpc.UserListInfoRequest{
			UserIdList: userIDList,
		})
		if err1 != nil {
			logx.Error(err1)
			return nil, errors.New("获取用户信息失败")
		}

		// 根据选中成员的昵称生成群名称。
		var nameList []string
		for _, info := range userListResponse.UserInfo {
			if len(strings.Join(nameList, "、")) >= 29 {
				break
			}
			nameList = append(nameList, info.NickName)
		}
		groupModel.Title = strings.Join(nameList, "、") + "的群聊"

	default:
		return nil, errors.New("创建群聊模式错误")
	}

	// 设置群头像，使用群名称的首字母。
	groupModel.Avatar = string([]rune(groupModel.Title)[0])

	// 将群组信息保存到数据库。
	err = l.svcCtx.DB.Create(&groupModel).Error
	if err != nil {
		logx.Error(err)
		return nil, errors.New("创建群聊失败")
	}

	// 创建群成员记录。
	var members []group_models.GroupMemberModel
	for i, u := range groupUserList {
		memberModel := group_models.GroupMemberModel{
			GroupID: groupModel.ID,
			UserID:  u,
			Role:    3,
		}
		// 设置创建者为群主。
		if i == 0 {
			memberModel.Role = 1
		}
		members = append(members, memberModel)
	}

	// 将群成员信息保存到数据库。
	err = l.svcCtx.DB.Create(&members).Error
	if err != nil {
		logx.Error(err)
		return nil, errors.New("添加群成员失败")
	}

	return
}
