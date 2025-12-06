package utils

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// PaginationParams holds pagination parameters
type PaginationParams struct {
	Page     int
	PageSize int
	Offset   int
}

// DefaultPageSize is the default number of items per page
const DefaultPageSize = 10

// MaxPageSize is the maximum allowed page size
const MaxPageSize = 100

// GetPaginationParams extracts pagination parameters from query string
func GetPaginationParams(c *gin.Context) PaginationParams {
	page := 1
	pageSize := DefaultPageSize

	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	if ps := c.Query("page_size"); ps != "" {
		if parsed, err := strconv.Atoi(ps); err == nil && parsed > 0 {
			pageSize = parsed
			if pageSize > MaxPageSize {
				pageSize = MaxPageSize
			}
		}
	}

	// Also support "limit" and "offset" query params
	if limit := c.Query("limit"); limit != "" {
		if parsed, err := strconv.Atoi(limit); err == nil && parsed > 0 {
			pageSize = parsed
			if pageSize > MaxPageSize {
				pageSize = MaxPageSize
			}
		}
	}

	offset := (page - 1) * pageSize
	if offsetParam := c.Query("offset"); offsetParam != "" {
		if parsed, err := strconv.Atoi(offsetParam); err == nil && parsed >= 0 {
			offset = parsed
			page = (offset / pageSize) + 1
		}
	}

	return PaginationParams{
		Page:     page,
		PageSize: pageSize,
		Offset:   offset,
	}
}

// SortParams holds sorting parameters
type SortParams struct {
	SortBy    string
	SortOrder string
}

// GetSortParams extracts sorting parameters from query string
func GetSortParams(c *gin.Context, allowedFields []string, defaultField string) SortParams {
	sortBy := c.DefaultQuery("sort_by", defaultField)
	sortOrder := c.DefaultQuery("sort_order", "desc")

	// Validate sort field
	validField := false
	for _, field := range allowedFields {
		if field == sortBy {
			validField = true
			break
		}
	}
	if !validField {
		sortBy = defaultField
	}

	// Validate sort order
	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "desc"
	}

	return SortParams{
		SortBy:    sortBy,
		SortOrder: sortOrder,
	}
}

// GetSortClause returns a GORM-compatible sort clause
func (s SortParams) GetSortClause() string {
	return s.SortBy + " " + s.SortOrder
}
