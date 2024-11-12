package logic

import (
	"context"
	"errors"
	"fim/fim_file/file_model"
	"strings"

	"fim/fim_file/file_rpc/internal/svc"
	"fim/fim_file/file_rpc/types/file_rpc"

	"github.com/zeromicro/go-zero/core/logx"
)

type FileInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewFileInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FileInfoLogic {
	return &FileInfoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// FileInfo 提供文件信息查询功能。
// 它接收一个 FileInfoRequest 对象，返回一个 FileInfoResponse 对象及可能的错误。
// 主要功能包括：根据UID查询文件模型，判断文件是否存在，提取文件类型，
// 并返回文件名、文件Hash、文件大小和文件类型等信息。
func (l *FileInfoLogic) FileInfo(in *file_rpc.FileInfoRequest) (*file_rpc.FileInfoResponse, error) {
	// 初始化文件模型变量
	var fileModel file_model.FileModel
	// 使用GORM框架从数据库中查询文件信息，条件是文件的UID等于请求中的FileId
	err := l.svcCtx.DB.Take(&fileModel, "uid=?", in.FileId).Error
	// 如果查询错误，则返回错误信息
	if err != nil {
		return nil, errors.New("文件不存在")
	}

	// 初始化文件类型变量
	var tp string
	// 分割文件名以获取文件类型（后缀名）
	nameList := strings.Split(fileModel.FileName, ".")
	// 如果文件名包含后缀，则提取文件类型
	if len(nameList) > 1 {
		tp = nameList[len(nameList)-1]
	}

	// 构建并返回文件信息响应对象
	return &file_rpc.FileInfoResponse{
		FileName: fileModel.FileName,
		FileHash: fileModel.Hash,
		FileSize: fileModel.Size,
		FileType: tp,
	}, nil
}
