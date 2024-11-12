package file_model

import (
	"fim/common/models"
	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
	"os"
)

type FileModel struct {
	models.Model
	Uid      uuid.UUID `json:"uid"`      //文件ID/api/file/{uuid}
	UserID   uint      `json:"userID"`   //用户ID
	FileName string    `json:"fileName"` //文件名
	FilePath string    `json:"filePath"` //文件路径
	Size     int64     `json:"size"`     //文件大小
	Hash     string    `json:"hash"`     //文件hash
}

// WebPath 返回文件在Web上的访问路径。
// 该路径由固定的API路径和文件的唯一标识符UID组成。
// 参数:
//   * 无
// 返回值:
//   * string: 文件的Web访问路径
func (file *FileModel) WebPath() string {
	return "/api/file/" + file.Uid.String()
}

// BeforeDelete 在删除文件模型之前执行的操作。
// 该方法主要用于在数据库操作之前，实际删除文件系统中的文件。
// 参数:
//   * tx: GORM事务对象，用于数据库操作
// 返回值:
//   * error: 如果删除文件失败，返回相应的错误；否则返回nil
func (file *FileModel) BeforeDelete(tx *gorm.DB) (err error) {
	logx.Infof("删除文件的名称 %s", file.FileName)
	if file.FilePath != "" {
		err1 := os.Remove(file.FilePath)
		if err1 != nil {
			logx.Error(err1)
		} else {
			logx.Infof("文件源地址删除%s", file.FilePath)
		}
	}
	return
}
