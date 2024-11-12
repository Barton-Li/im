package handler

import (
	"errors"
	"fim/common/response"
	"fim/fim_file/file_api/internal/svc"
	"fim/fim_file/file_api/internal/types"
	"fim/fim_file/file_model"
	"net/http"
	"os"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// ImageShowHandler 是一个处理图片展示请求的HTTP处理函数。
// 它接收一个svc.ServiceContext指针作为参数，用于获取服务上下文信息。
// 返回的函数符合http.HandlerFunc接口，用于处理HTTP请求。
func ImageShowHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	// 返回的函数实际处理HTTP请求。
	return func(w http.ResponseWriter, r *http.Request) {
		// 解析HTTP请求中的图片展示请求。
		var req types.ImageShowRequest
		if err := httpx.Parse(r, &req); err != nil {
			// 如果请求解析失败，返回错误响应。
			response.Response(r, w, nil, err)
			return
		}

		// 从数据库中查找指定图片信息。
		var fileModel file_model.FileModel
		err := svcCtx.DB.Take(&fileModel, "uid=?", req.ImageName).Error
		if err != nil {
			// 如果图片信息不存在，返回错误响应。
			response.Response(r, w, nil, errors.New("文件不存在"))
			return
		}

		// 读取图片文件内容。
		byteData, err := os.ReadFile(fileModel.FilePath)
		if err != nil {
			// 如果读取文件失败，返回错误响应。
			response.Response(r, w, nil, err)
			return
		}

		// 将图片文件内容写入HTTP响应。
		w.Write(byteData)
	}
}
