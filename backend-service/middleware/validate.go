package middleware

import (
	"net/http"
	"path/filepath"
	"strings"

	"mime/multipart"
	"slices"

	"github.com/gin-gonic/gin"
	"github.com/ijasmoopan/intucloud-task/backend-service/config"
)

func ValidateFiles() gin.HandlerFunc {
	return func(c *gin.Context) {
		cfg := config.NewConfig()

		// Get the files from the request
		form, err := c.MultipartForm()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Failed to parse form data",
			})
			c.Abort()
			return
		}

		files := form.File["files"]
		if len(files) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "No files uploaded",
			})
			c.Abort()
			return
		}

		// Validate each file
		validFiles := make([]*multipart.FileHeader, 0)
		for _, file := range files {
			// Check file size
			if file.Size > cfg.MaxFileSize {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "File size exceeds maximum limit",
					"file":  file.Filename,
				})
				c.Abort()
				return
			}

			// Check file extension
			ext := strings.ToLower(filepath.Ext(file.Filename))
			allowed := slices.Contains(cfg.AllowedTypes, ext)

			if !allowed {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "File type not allowed",
					"file":  file.Filename,
				})
				c.Abort()
				return
			}

			validFiles = append(validFiles, file)
		}

		// Store valid files in context for later use
		c.Set("files", validFiles)
		c.Next()
	}
}
