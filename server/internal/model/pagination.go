package model

// Paginated 分页响应
type Paginated struct {
	List     []any `json:"list"`
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
}

// PageQuery 分页查询参数
type PageQuery struct {
	Page     int
	PageSize int
}

// Normalize 规范化分页参数
func (p *PageQuery) Normalize() {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PageSize < 1 {
		p.PageSize = 10
	}
	if p.PageSize > 50 {
		p.PageSize = 50
	}
}

// Offset 计算偏移量
func (p *PageQuery) Offset() int {
	return (p.Page - 1) * p.PageSize
}
