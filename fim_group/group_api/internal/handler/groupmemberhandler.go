package handler

import (
	"fim/common/response"
	"fim/fim_group/group_api/internal/logic"
	"fim/fim_group/group_api/internal/svc"
	"fim/fim_group/group_api/internal/types"
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func groupMemberHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GroupMemberRequest
		if err := httpx.Parse(r, &req); err != nil {
			response.Response(r, w, nil, err)
			return
		}

		l := logic.NewGroupMemberLogic(r.Context(), svcCtx)
		resp, err := l.GroupMember(&req)
		response.Response(r, w, resp, err)

	}
}
