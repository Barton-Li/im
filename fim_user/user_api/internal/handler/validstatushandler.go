package handler

import (
	"fim/common/response"
	"fim/fim_user/user_api/internal/logic"
	"fim/fim_user/user_api/internal/svc"
	"fim/fim_user/user_api/internal/types"
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func validStatusHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.FriendValidStatusRequest
		if err := httpx.Parse(r, &req); err != nil {
			response.Response(r, w, nil, err)
			return
		}

		l := logic.NewValidStatusLogic(r.Context(), svcCtx)
		resp, err := l.ValidStatus(&req)
		response.Response(r, w, resp, err)

	}
}
