package logic

import (
	"context"
	"fim/common/list_query"
	"fim/common/models"
	"fim/fim_user/user_models"

	"fim/fim_user/user_api/internal/svc"
	"fim/fim_user/user_api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserValidListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserValidListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserValidListLogic {
	return &UserValidListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserValidListLogic) UserValidList(req *types.FriendValidRequest) (resp *types.FriendValidResponse, err error) {
	fvs, count, _ := list_query.ListQuery(l.svcCtx.DB, user_models.FriendVerifyModel{}, list_query.Option{
		PageInfo: models.PagaInfo{
			Page:  req.Page,
			Limit: req.Limit,
			Sort:  "created_at desc",
		},
		Where: l.svcCtx.DB.Where("(send_user_id =? and rev_user_id  <> ? and send_status <> 4) or(rev_user_id =? and send_user_id <> ?and rev_user_id <> 4)or(send_user_id =? and rev_user_id =? and not (send_status = 4 or rev_status=4 ))",
			req.UserID, req.UserID, req.UserID, req.UserID, req.UserID, req.UserID),
		Preload: []string{"RevUserModel.UserConfModel", "SendUserModel.UserConfModel"},
	})
	var list []types.FriendValidInfo
	for _, fv := range fvs {
		info := types.FriendValidInfo{
			AdditionalMessages: fv.AdditionalMessages,
			ID:                 fv.ID,
			CreatedAt:          fv.CreatedAt.String(),
			SendStatus:         fv.SendStatus,
			RevStatus:          fv.RevStatus,
		}
		if fv.SendUserID == req.UserID {
			info.UserID = fv.RevUserID
			info.Nickname = fv.RevUserModel.Nickname
			info.Avatar = fv.RevUserModel.Avatar
			info.Verification = fv.RevUserModel.UserConfModel.Verification
			info.Status = fv.SendStatus
			info.Flag = "send"
		}
		if fv.RevUserID == req.UserID {
			info.UserID = fv.SendUserID
			info.Nickname = fv.SendUserModel.Nickname
			info.Avatar = fv.SendUserModel.Avatar
			info.Verification = fv.SendUserModel.UserConfModel.Verification
			info.Status = fv.RevStatus
			info.Flag = "rev"
		}
		if fv.VerificationQuestion != nil {
			info.VerificationQuestion = &types.VerificationQuestion{
				Answer1:  fv.VerificationQuestion.Answer1,
				Answer2:  fv.VerificationQuestion.Answer2,
				Answer3:  fv.VerificationQuestion.Answer3,
				Problem1: fv.VerificationQuestion.Problem1,
				Problem2: fv.VerificationQuestion.Problem2,
				Problem3: fv.VerificationQuestion.Problem3,
			}
		}
		list = append(list, info)
	}

	return &types.FriendValidResponse{
		List:  list,
		Count: count,
	}, nil

}
