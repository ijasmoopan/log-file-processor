package models

import (
	"gorm.io/gorm"
)

// FileResult represents a file processing result in the database
type FileResult struct {
	gorm.Model
	FileName   string `gorm:"not null"`
	ClientID   string `gorm:"not null"`
	Status     string `gorm:"not null"`
	WarnCount  *int   `gorm:"default:null"`
	ErrorCount *int   `gorm:"default:null"`
	Error      string `gorm:"type:text"`
}

func (FileResult) TableName() string {
	return "log_stats"
}
