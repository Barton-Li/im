package logic

import (
	"context"
	"fim/common/list_query"
	"fim/common/models"
	"fim/fim_user/user_models"
	"fmt"

	"fim/fim_user/user_api/internal/svc"
	"fim/fim_user/user_api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SearchLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSearchLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SearchLogic {
	return &SearchLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SearchLogic) Search(req *types.SearchRequest) (resp *types.SearchResponse, err error) {
	//查询用户信息，数量，是否在线
	users, count, err := list_query.ListQuery(l.svcCtx.DB, user_models.UserConfModel{
		Online: req.Online,
	}, list_query.Option{
		PageInfo: models.PagaInfo{
			Page:  req.Page,
			Limit: req.Limit,
		},
		Preload: []string{"UserModel"},
		Joins:   "left join user_models um on um.id=user_conf.user_id",
		Where:   l.svcCtx.DB.Where("(user_conf_model.search_user <> 0 or user_conf_model.search_user is null) and (user_conf_model.search_user=1 and um.id=?)or (user_conf_model.search_user=2 and(um.id=? or um.nickname like ? ))", req.Key, req.Key, fmt.Sprintf("%%%s%%")),
	})
	var friend user_models.FriendModel
	//查询好友关系
	friends := friend.Friends(l.svcCtx.DB, req.UserID)
	userMap := map[uint]bool{}
	for _, model := range friends {
		if model.SendUserID == req.UserID {
			userMap[model.RevUserID] = true
		} else {
			userMap[model.SendUserID] = true
		}
	}
	list := make([]types.SearchInfo, 0)
	//组装返回数据
	for _, uc := range users {
		list = append(list, types.SearchInfo{
			UserID:   uc.UserID,
			Nickname: uc.UserModel.Nickname,
			Abstract: uc.UserModel.Abstract,
			Avatar:   uc.UserModel.Avatar,
			IsFriend: userMap[uc.UserID],
		})
	}
	return &types.SearchResponse{List: list, Count: count}, nil
}
