package model

// PaginationParams carries validated page + limit from the request.
type PaginationParams struct {
	Page  int
	Limit int
}

func (p *PaginationParams) Offset() int {
	return (p.Page - 1) * p.Limit
}

// PaginatedResult is generic — T is the item type.
// Caller specifies: PaginatedResult[*URL] or PaginatedResult[*User]
type PaginatedResult[T any] struct {
	Data       []T   `json:"data"`
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalPages int   `json:"total_pages"`
}

// NewPaginatedResult constructs a result — T inferred from data slice.
func NewPaginatedResult[T any](
	data []T,
	total int64,
	params PaginationParams) PaginatedResult[T] {

	totalPages := int(total) / params.Limit
	if int(total)%params.Limit != 0 {
		totalPages++
	}

	return PaginatedResult[T]{
		Data:       data,
		Total:      total,
		Page:       params.Page,
		Limit:      params.Limit,
		TotalPages: totalPages,
	}
}
