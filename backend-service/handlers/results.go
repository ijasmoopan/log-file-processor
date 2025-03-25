package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ijasmoopan/intucloud-task/backend-service/models"
	"github.com/ijasmoopan/intucloud-task/backend-service/pagination"
	"gorm.io/gorm"
)

type ResultResponse struct {
	Results    []models.FileResult `json:"results"`
	Page       int                 `json:"page"`
	PageSize   int                 `json:"page_size"`
	TotalItems int64               `json:"total_items"`
	TotalPages int                 `json:"total_pages"`
	HasNext    bool                `json:"has_next"`
	HasPrev    bool                `json:"has_prev"`
}

// GetResults returns a paginated list of file processing results
func GetResults(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get pagination parameters
		params, err := pagination.GetPaginationParams(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		var results []models.FileResult
		var totalItems int64

		// Get total count
		if err := db.Model(&models.FileResult{}).Count(&totalItems).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to count results",
			})
			return
		}

		// Calculate pagination
		offset := (params.Page - 1) * params.PageSize

		// Get paginated results
		if err := db.Offset(offset).Limit(params.PageSize).Order("created_at DESC").Find(&results).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to fetch results",
			})
			return
		}

		// Calculate total pages
		totalPages := (int(totalItems) + params.PageSize - 1) / params.PageSize

		c.JSON(http.StatusOK, ResultResponse{
			Results:    results,
			Page:       params.Page,
			PageSize:   params.PageSize,
			TotalItems: totalItems,
			TotalPages: totalPages,
			HasNext:    params.Page < totalPages,
			HasPrev:    params.Page > 1,
		})
	}
}

// GetResultByID returns a single file processing result by ID
func GetResultByID(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		var result models.FileResult
		if err := db.First(&result, id).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{
					"error": "Result not found",
				})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to fetch result",
			})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

// GetResultByFilename returns a single file processing result by filename
func GetResultByFilename(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		filename := c.Param("filename")

		var result models.FileResult
		if err := db.Where("file_name = ?", filename).Order("updated_at DESC").First(&result).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{
					"error": "Result not found",
				})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to fetch result",
			})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}
