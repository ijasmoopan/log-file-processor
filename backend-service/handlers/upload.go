package handlers

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ijasmoopan/intucloud-task/backend-service/config"
)

type UploadFileInfo struct {
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
	Path     string `json:"path"`
}

func UploadFile(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the files from context (set by middleware)
		files, exists := c.Get("files")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Files not found in context",
			})
			return
		}

		uploadedFiles := files.([]*multipart.FileHeader)
		uploadedFileInfos := make([]UploadFileInfo, 0)

		// Process each file
		for _, file := range uploadedFiles {
			// Generate unique filename
			ext := filepath.Ext(file.Filename)
			nameWithoutExt := strings.TrimSuffix(file.Filename, ext)
			filename := fmt.Sprintf("%s_%s%s", nameWithoutExt, time.Now().Format("20060102150405"), ext)
			dst := filepath.Join(cfg.UploadDir, filename)

			// Save the file
			if err := c.SaveUploadedFile(file, dst); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Failed to save file",
					"file":  file.Filename,
				})
				return
			}

			uploadedFileInfos = append(uploadedFileInfos, UploadFileInfo{
				Filename: filename,
				Size:     file.Size,
				Path:     dst,
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Files uploaded successfully",
			"files":   uploadedFileInfos,
		})
	}
}
