package model

// PaginationParams carries validated page + limit from the request
type PaginationParams struct {
    Page  int
    Limit int
    // Offset is calculated — not from the request
    // Page 1, Limit 10 → Offset 0
    // Page 2, Limit 10 → Offset 10
    // Page 3, Limit 10 → Offset 20
}

func (p *PaginationParams) Offset() int {
    return (p.Page - 1) * p.Limit
}

// PaginatedResult is the standard paginated response shape
type PaginatedResult struct {
    Data       interface{} `json:"data"`
    Total      int64       `json:"total"`
    Page       int         `json:"page"`
    Limit      int         `json:"limit"`
    TotalPages int         `json:"total_pages"`
}

func NewPaginatedResult(
    data interface{},
    total int64,
    params PaginationParams) PaginatedResult {

    totalPages := int(total) / params.Limit
    if int(total)%params.Limit != 0 {
        totalPages++ // round up
    }

    return PaginatedResult{
        Data:       data,
        Total:      total,
        Page:       params.Page,
        Limit:      params.Limit,
        TotalPages: totalPages,
    }
}