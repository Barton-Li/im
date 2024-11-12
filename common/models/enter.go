package models

import "time"

type Model struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// PagaInfo 定义了分页查询时所需的信息。
// 它包含了当前页码、每页的条数、排序方式和查询关键字。
// 这些字段通过表单参数传递，用于定制化数据的分页和筛选。
type PagaInfo struct {
	Page  int    `form:"page" `  // 当前页码，用于指定返回结果的页码。
	Limit int    `form:"limit" ` // 每页的条数，用于控制每页返回的数据量。
	Sort  string `form:"sort" `  // 排序方式，用于指定返回结果的排序规则。
	Key   string `form:"key" `   // 查询关键字，用于模糊匹配数据。
}
