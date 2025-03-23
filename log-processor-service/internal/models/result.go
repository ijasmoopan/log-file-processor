package models

import "fmt"

// Result represents the processing result for a single log file
type Result struct {
	FilePath   string
	ErrorCount int
	WarnCount  int
}

// Add combines two results
func (r *Result) Add(other Result) {
	r.ErrorCount += other.ErrorCount
	r.WarnCount += other.WarnCount
}

// String returns a formatted string representation of the result
func (r *Result) String() string {
	return fmt.Sprintf("File: %s, Error Count: %d, Warn Count: %d", 
		r.FilePath, r.ErrorCount, r.WarnCount)
}
