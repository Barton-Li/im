package handler

import (
	"errors"
	"fim/common/response"
	"fim/fim_file/file_api/internal/logic"
	"fim/fim_file/file_api/internal/svc"
	"fim/fim_file/file_api/internal/types"
	"fim/fim_file/file_model"
	"fim/utils"
	"fmt"
	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
)

// ImageHandler 处理图像上传请求。
// svcCtx: 服务上下文，包含配置信息和数据库访问等服务。
func ImageHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 解析请求体中的图像请求信息。
		var req types.ImageRequest
		if err := httpx.Parse(r, &req); err != nil {
			response.Response(r, w, nil, err)
			return
		}

		// 从请求参数中获取图像类型。
		imageType := r.FormValue("imageType")
		// 校验图像类型是否有效。
		switch imageType {
		case "avatar", "group_avatar", "chat":
		default:
			response.Response(r, w, nil, errors.New("invalid imageType,只能为 avatar,group_avatar,chat"))
			return
		}

		// 从请求中获取上传的图像文件。
		file, fileHead, err := r.FormFile("image")
		if err != nil {
			response.Response(r, w, nil, err)
			return
		}
		// 检查文件大小是否超出限制。
		// 文件大小限制
		mSize := float64(fileHead.Size) / float64(1024) / float64(1024)
		if mSize > svcCtx.Config.FileSize {
			response.Response(r, w, nil, fmt.Errorf("文件大小超过限制,最大%vM", svcCtx.Config.FileSize))
			return
		}

		// 检查文件类型是否在允许的白名单中。
		// 文件类型限制
		nameList := strings.Split(fileHead.Filename, ".")
		var suffix string
		if len(nameList) > 1 {
			suffix = nameList[len(nameList)-1]
		}
		if !utils.InList(svcCtx.Config.WhiteList, suffix) {
			response.Response(r, w, nil, errors.New("文件类型不允许上传"))
			return
		}

		// 读取文件数据并计算MD5用于唯一标识。
		// 计算hash
		imageData, _ := io.ReadAll(file)
		imageHash := utils.MD5(imageData)

		// 使用图像逻辑类处理图像业务逻辑。
		l := logic.NewImageLogic(r.Context(), svcCtx)
		resp, err := l.Image(&req)

		// 检查数据库中是否已存在相同的图像文件。
		var fileModel file_model.FileModel
		err = svcCtx.DB.Take(&fileModel, "hash = ?", imageHash).Error
		if err == nil {
			// 如果已存在，则使用已有的文件路径。
			resp.Url = fileModel.WebPath()
			logx.Infof("文件%s hash已存在", fileHead.Filename)
			response.Response(r, w, resp, err)
			return
		}

		// 构建图像文件在服务器上的存储路径。
		// 拼路径 /uploads/imageType/{uid}.{后缀}
		dirPath := path.Join(svcCtx.Config.UploadDir, imageType)
		// 如果目录不存在，则创建目录。
		_, err = os.ReadDir(dirPath)
		if err != nil {
			os.MkdirAll(dirPath, 0666)
		}

		// 生成新的文件模型，包括文件名、哈希值、大小等信息。
		fileName := fileHead.Filename
		newFileModel := file_model.FileModel{
			UserID:   req.UserID,
			FileName: fileName,
			Hash:     utils.MD5(imageData),
			Size:     fileHead.Size,
			Uid:      uuid.New(),
		}
		// 设置文件的存储路径。
		newFileModel.FilePath = path.Join(dirPath, fmt.Sprintf("%s.%s", newFileModel.Uid, suffix))

		// 将文件数据写入到服务器存储路径。
		err = os.WriteFile(newFileModel.FilePath, imageData, 0666)
		if err != nil {
			logx.Errorf("文件写入失败:%s", err.Error())
			response.Response(r, w, nil, err)
			return
		}

		// 将新的文件模型插入到数据库。
		err = svcCtx.DB.Create(&newFileModel).Error
		if err != nil {
			logx.Errorf("文件写入失败:%s", err.Error())
			response.Response(r, w, nil, err)
			return
		}

		// 构建并返回图像的访问URL。
		resp.Url = newFileModel.WebPath()
		response.Response(r, w, resp, err)
	}
}
