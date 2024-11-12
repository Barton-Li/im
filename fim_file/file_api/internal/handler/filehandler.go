package handler

import (
	"context"
	"errors"
	"fim/common/response"
	"fim/fim_file/file_api/internal/logic"
	"fim/fim_file/file_api/internal/svc"
	"fim/fim_file/file_api/internal/types"
	"fim/fim_file/file_model"
	"fim/fim_user/user_rpc/types/user_rpc"
	"fim/utils"
	"fmt"
	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
	"io"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// FileHandler 是一个HTTP处理器，用于处理文件上传请求。
// 它从服务上下文中使用依赖项来执行文件处理和数据库操作等任务。
func FileHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 初始化文件请求对象。
		var req types.FileRequest
		// 解析请求，如果发生错误，则返回错误。
		if err := httpx.Parse(r, &req); err != nil {
			response.Response(r, w, nil, err)
			return
		}
		// 尝试从请求中获取文件，如果发生错误，则返回错误。
		file, fileHeader, err := r.FormFile("file")
		if err != nil {
			response.Response(r, w, nil, err)
			return
		}
		// 将文件头信息中的文件名按"."分割成字符串切片，用于获取文件名和扩展名。
		nameList := strings.Split(fileHeader.Filename, ".")
		var suffix string
		if len(nameList) > 1 {
			suffix = nameList[len(nameList)-1]
		}
		// 检查文件后缀是否在黑名单中，如果是，则返回错误。
		if utils.InList(svcCtx.Config.BlackList, suffix) {
			response.Response(r, w, nil, errors.New("图片非法"))
			return
		}
		// 读取文件数据并计算MD5哈希。
		fileData, _ := io.ReadAll(file)
		fileHash := utils.MD5(fileData)
		// 创建文件逻辑实例。
		l := logic.NewFileLogic(r.Context(), svcCtx)
		// 执行文件逻辑操作。
		resp, err := l.File(&req)

		// 查询数据库中是否存在相同哈希值的文件。
		var fileModel file_model.FileModel
		err = svcCtx.DB.Take(&fileModel, "hash=?", fileHash).Error
		if err == nil {
			// 如果找到重复的文件，更新响应中的源路径。
			resp.Src = fileModel.WebPath()
			logx.Infof("文件 %s hash重复", fileHeader.Filename)
			response.Response(r, w, resp, err)
			return
		}
		// 获取用户信息。
		userResponse, err := svcCtx.UserRpc.UserListInfo(context.Background(), &user_rpc.UserListInfoRequest{
			UserIdList: []uint32{uint32(req.UserID)},
		})
		if err != nil {
			response.Response(r, w, nil, err)
			return
		}
		// 构建文件目录名称。
		dirName := fmt.Sprintf("%d_%s", req.UserID, userResponse.UserInfo[uint32(req.UserID)].NickName)
		dirPath := path.Join(svcCtx.Config.UploadDir, "file", dirName)
		// 检查目录是否存在，如果不存在则创建。
		_, err = os.ReadDir(dirPath)
		if err != nil {
			os.MkdirAll(dirPath, 0666)
		}
		// 创建新的文件模型。
		newFileModel := file_model.FileModel{
			UserID:   req.UserID,
			FileName: fileHeader.Filename,
			Size:     fileHeader.Size,
			Hash:     fileHash,
			Uid:      uuid.New(),
		}
		// 设置文件路径。
		newFileModel.FilePath = path.Join(dirPath, fmt.Sprintf("%s.%s", newFileModel.Uid, suffix))
		// 写入文件数据。
		err = os.WriteFile(newFileModel.FilePath, fileData, 0666)
		if err != nil {
			response.Response(r, w, nil, err)
			return
		}
		// 将新文件模型保存到数据库。
		err = svcCtx.DB.Create(&newFileModel).Error
		if err != nil {
			response.Response(r, w, nil, err)
			return
		}
		// 更新响应中的源路径。
		resp.Src = newFileModel.WebPath()
		// 返回最终的响应。
		response.Response(r, w, resp, err)
	}
}
