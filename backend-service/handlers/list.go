package handlers

import (
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ijasmoopan/intucloud-task/backend-service/config"
	"github.com/ijasmoopan/intucloud-task/backend-service/pagination"
)

type FileInfo struct {
	FileName     string    `json:"file_name"`
	Size         int64     `json:"size"`
	Path         string    `json:"path"`
	UploadTime   time.Time `json:"upload_time"`
	LastModified time.Time `json:"last_modified"`
}

type ListResponse struct {
	Files      []FileInfo `json:"files"`
	Page       int        `json:"page"`
	PageSize   int        `json:"page_size"`
	TotalItems int        `json:"total_items"`
	TotalPages int        `json:"total_pages"`
	HasNext    bool       `json:"has_next"`
	HasPrev    bool       `json:"has_prev"`
}

func ListFiles(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get pagination parameters
		params, err := pagination.GetPaginationParams(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		files, err := os.ReadDir(cfg.UploadDir)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to read upload directory",
			})
			return
		}

		// First pass: collect all valid files
		allFiles := make([]FileInfo, 0)
		for _, file := range files {
			if file.IsDir() {
				continue
			}

			info, err := file.Info()
			if err != nil {
				continue
			}

			filePath := filepath.Join(cfg.UploadDir, file.Name())
			allFiles = append(allFiles, FileInfo{
				FileName:     file.Name(),
				Size:         info.Size(),
				Path:         filePath,
				UploadTime:   info.ModTime(), // Using ModTime as upload time for now
				LastModified: info.ModTime(),
			})
		}

		// Calculate pagination
		paginationInfo := pagination.CalculatePagination(len(allFiles), params)
		startIndex, endIndex := pagination.GetPageIndices(paginationInfo)

		// Get paginated files
		paginatedFiles := allFiles[startIndex:endIndex]

		c.JSON(http.StatusOK, ListResponse{
			Files:      paginatedFiles,
			Page:       paginationInfo.CurrentPage,
			PageSize:   paginationInfo.PageSize,
			TotalItems: paginationInfo.TotalItems,
			TotalPages: paginationInfo.TotalPages,
			HasNext:    paginationInfo.HasNext,
			HasPrev:    paginationInfo.HasPrev,
		})
	}
}
