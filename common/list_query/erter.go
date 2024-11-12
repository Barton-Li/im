package list_query

import (
	"fim/common/models"
	"fmt"
	"gorm.io/gorm"
)

type Option struct {
	PageInfo models.PagaInfo
	Where    *gorm.DB
	Debug    bool
	Joins    string
	Likes    []string
	Preload  []string
	Table    func() (string, any)
	Group    []string
}

// ListQuery 根据提供的选项查询数据库，并返回查询结果列表、总数和可能的错误。
// 使用 GORM 提供的数据库操作接口，根据选项进行灵活的查询。
// 参数:
//
//	db *gorm.DB: 数据库连接实例。
//	model T: 查询结果的模型类型。
//	option Option: 查询选项，包含分页、筛选、排序等信息。
//
// 返回值:
//
//	[]T: 查询结果的列表，类型与模型相同。
//	int64: 查询结果的总数。
//	error: 查询过程中可能发生的错误。
func ListQuery[T any](db *gorm.DB, model T, option Option) (list []T, count int64, err error) {
	// 开启调试模式，如果选项中设置了。
	if option.Debug {
		db = db.Debug()
	}

	// 使用模型定义查询的基本结构。
	query := db.Model(model)

	// 如果提供了搜索关键字和喜欢的列，构建包含这些条件的查询。
	if option.PageInfo.Key != "" && len(option.Likes) > 0 {
		likeQuery := db.Where("")
		for index, column := range option.Likes {
			// 对于第一个条件使用Where，后续条件使用Or来连接。
			if index == 0 {
				likeQuery.Where(fmt.Sprintf("%s like ?"), fmt.Sprintf("%%%s%%", option.PageInfo.Key))
			} else {
				likeQuery.Or(fmt.Sprintf("%s like ?", column), fmt.Sprintf("%%%s%%", option.PageInfo.Key))
			}
		}
		query.Where(likeQuery)
	}

	// 如果提供了自定义的表名和数据，使用它们来重定义查询的表。
	if option.Table != nil {
		table, data := option.Table()
		query = query.Table(table, data)
	}

	// 如果提供了JOIN条件，添加到查询中。
	if option.Joins != "" {
		query = query.Joins(option.Joins)
	}

	// 如果提供了WHERE条件，添加到查询中。
	if option.Where != nil {
		query = query.Where(option.Where)
	}

	// 如果提供了GROUP BY条件，添加到查询中。
	if len(option.Group) > 0 {
		for _, group := range option.Group {
			query = query.Group(group)
		}
	}

	// 计算查询结果的总数。
	query.Model(model).Count(&count)

	// 预加载关联数据，如果在选项中指定了。
	for _, s := range option.Preload {
		query = query.Preload(s)
	}

	// 处理分页选项，确保页码和限制值是有效的。
	if option.PageInfo.Page <= 0 {
		option.PageInfo.Page = 1
	}
	if option.PageInfo.Limit != -1 {
		if option.PageInfo.Limit <= 0 {
			option.PageInfo.Limit = 10
		}
	}

	// 计算查询的偏移量。
	offset := (option.PageInfo.Page - 1) * option.PageInfo.Limit

	// 如果提供了排序条件，添加到查询中。
	if option.PageInfo.Sort != " " {
		query.Order(option.PageInfo.Sort)
	}

	// 执行查询，并将结果存储到列表中。
	err = query.Limit(option.PageInfo.Limit).Offset(offset).Find(&list).Error
	return
}
