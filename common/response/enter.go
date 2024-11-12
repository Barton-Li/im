package response

import (
	"github.com/zeromicro/go-zero/rest/httpx"
	"net/http"
)

type Body struct {
	Code uint32      `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func Response(r *http.Request, w http.ResponseWriter, resp interface{}, err error) {
	if err == nil {
		r := &Body{
			Code: 0,
			Msg:  "成功",
			Data: resp,
		}
		httpx.WriteJson(w, http.StatusOK, r)
		return
	}
	errCode := uint32(7)
	httpx.WriteJson(w, http.StatusOK, &Body{
		Code: errCode,
		Msg:  err.Error(),
		Data: nil,
	})

}
