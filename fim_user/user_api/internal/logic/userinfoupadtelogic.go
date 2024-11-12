package logic

import (
	"context"
	"errors"
	"fim/common/models/ctype"
	"fim/fim_user/user_models"
	"fim/utils/maps"

	"fim/fim_user/user_api/internal/svc"
	"fim/fim_user/user_api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserInfoUpadteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserInfoUpadteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserInfoUpadteLogic {
	return &UserInfoUpadteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// UserInfoUpadte 用于更新用户信息和用户配置信息。
// req 包含需要更新的用户信息和用户配置信息。
// 返回值 resp 是更新后的响应信息，err 是更新过程中可能出现的错误。
func (l *UserInfoUpadteLogic) UserInfoUpadte(req *types.UserInfoUpdateRequest) (resp *types.UserInfoUpdateResponse, err error) {
	// 将请求中的用户信息转换为 map，用于后续更新数据库中的用户信息。
	userMaps := maps.RefToMap(*req, "user")
	// 如果用户信息不为空，则尝试更新用户信息。
	if len(userMaps) != 0 {
		// 根据用户ID查询用户信息，确保用户存在。
		var user user_models.UserModel
		err = l.svcCtx.DB.Take(&user, req.UserID).Error
		if err != nil {
			return nil, errors.New("用户不存在")
		}
		// 更新用户信息。
		err = l.svcCtx.DB.Model(&user).Updates(userMaps).Error
		if err != nil {
			logx.Error(err)
			logx.Error(userMaps)
			return nil, errors.New("更新失败")
		}
	}

	// 将请求中的用户配置信息转换为 map，用于后续更新数据库中的用户配置信息。
	userConfMaps := maps.RefToMap(*req, "user_conf")
	// 如果用户配置信息不为空，则尝试更新用户配置信息。
	if len(userConfMaps) != 0 {
		// 根据用户ID查询用户配置信息，确保用户配置存在。
		var userConf user_models.UserConfModel
		err = l.svcCtx.DB.Take(&userConf, "user_id = ?", req.UserID).Error
		if err != nil {
			return nil, errors.New("用户配置不存在")
		}
		// 如果更新请求中包含验证问题，特殊处理并更新验证问题。
		verificationQuestion, ok := userConfMaps["verification_question"]
		if ok {
			delete(userConfMaps, "verification_question")
			data := ctype.VerificationQuestion{}
			maps.MapToStruct(verificationQuestion.(map[string]any), &data)
			// 更新用户配置中的验证问题。
			l.svcCtx.DB.Model(&userConf).Updates(&user_models.UserConfModel{
				VerificationQuestion: &data,
			})
		}
		// 更新剩余的用户配置信息。
		err = l.svcCtx.DB.Model(&userConf).Updates(userConfMaps).Error
		if err != nil {
			logx.Error(err)
			logx.Error(userConfMaps)
			return nil, errors.New("用户信息更新失败")
		}
	}

	// 返回更新后的响应信息。
	return
}
