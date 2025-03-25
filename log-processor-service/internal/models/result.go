package models

import "fmt"

// Result represents the processing result for a single log file
type Result struct {
	ClientID   string `json:"client_id,omitempty"`
	FilePath   string `json:"file_path"`
	ErrorCount int    `json:"error_count"`
	WarnCount  int    `json:"warn_count"`
}

// Add combines two results
func (r *Result) Add(result Result) {
	r.ErrorCount += result.ErrorCount
	r.WarnCount += result.WarnCount
}

// String returns a formatted string representation of the result
func (r *Result) String() string {
	return fmt.Sprintf("File: %s, Error Count: %d, Warn Count: %d",
		r.FilePath, r.ErrorCount, r.WarnCount)
}
