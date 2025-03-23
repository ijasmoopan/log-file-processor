package pagination

import (
	"errors"
	"math"
	"strconv"

	"github.com/gin-gonic/gin"
)

const (
	defaultPageSize = 10
	maxPageSize     = 100
)

// PaginationParams holds the pagination parameters
type PaginationParams struct {
	Page     int
	PageSize int
}

// PaginationInfo holds the pagination metadata
type PaginationInfo struct {
	CurrentPage int
	PageSize    int
	TotalItems  int
	TotalPages  int
	HasNext     bool
	HasPrev     bool
}

// GetPaginationParams extracts and validates pagination parameters from the request
func GetPaginationParams(c *gin.Context) (*PaginationParams, error) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil {
		return nil, errors.New("invalid page parameter")
	}

	pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", strconv.Itoa(defaultPageSize)))
	if err != nil {
		return nil, errors.New("invalid page_size parameter")
	}

	// Validate pagination parameters
	if page < 1 {
		return nil, errors.New("page must be greater than 0")
	}
	if pageSize < 1 || pageSize > maxPageSize {
		return nil, errors.New("page_size must be between 1 and 100")
	}

	return &PaginationParams{
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// CalculatePagination calculates pagination metadata
func CalculatePagination(totalItems int, params *PaginationParams) *PaginationInfo {
	if totalItems <= 0 {
		return &PaginationInfo{
			CurrentPage: 1,
			PageSize:    params.PageSize,
			TotalItems:  0,
			TotalPages:  0,
			HasNext:     false,
			HasPrev:     false,
		}
	}

	totalPages := int(math.Ceil(float64(totalItems) / float64(params.PageSize)))

	// Adjust page if it exceeds total pages
	if params.Page > totalPages {
		params.Page = totalPages
	}

	return &PaginationInfo{
		CurrentPage: params.Page,
		PageSize:    params.PageSize,
		TotalItems:  totalItems,
		TotalPages:  totalPages,
		HasNext:     params.Page < totalPages,
		HasPrev:     params.Page > 1,
	}
}

// GetPageIndices returns the start and end indices for the current page
func GetPageIndices(pagination *PaginationInfo) (startIndex, endIndex int) {
	startIndex = (pagination.CurrentPage - 1) * pagination.PageSize
	endIndex = startIndex + pagination.PageSize
	if endIndex > pagination.TotalItems {
		endIndex = pagination.TotalItems
	}
	return startIndex, endIndex
}
