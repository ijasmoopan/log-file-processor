package processor

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/ijasmoopan/intucloud-task/log-processor-service/internal/models"
)

// Processor handles the log file processing logic
type Processor struct {
	numWorkers int
	uploadDir  string
}

// ProgressCallback is a function type for reporting progress
type ProgressCallback func(fileName string, progress int, status string, err error)

// NewProcessor creates a new processor with the specified number of workers
func NewProcessor(numWorkers int, uploadDir string) *Processor {
	return &Processor{
		numWorkers: numWorkers,
		uploadDir:  uploadDir,
	}
}

// ProcessLogFile processes a single log file and returns the result
func ProcessLogFile(filePath string, progressCb ProgressCallback) (models.Result, error) {
	fmt.Printf("Opening file: %s\n", filePath)
	file, err := os.Open(filePath)
	if err != nil {
		progressCb(filepath.Base(filePath), 0, "error", err)
		return models.Result{}, fmt.Errorf("error opening file %s: %v", filePath, err)
	}
	defer file.Close()

	result := models.Result{
		FilePath: filePath,
	}

	// Get file size for progress calculation
	fileInfo, err := file.Stat()
	if err != nil {
		progressCb(filepath.Base(filePath), 0, "error", err)
		return result, fmt.Errorf("error getting file info: %v", err)
	}
	fileSize := fileInfo.Size()

	scanner := bufio.NewScanner(file)
	bytesRead := int64(0)
	lastProgress := 0

	for scanner.Scan() {
		line := scanner.Text()
		bytesRead += int64(len(line) + 1) // +1 for newline

		// Calculate progress percentage
		progress := int((float64(bytesRead) / float64(fileSize)) * 100)

		// Only report progress if it has changed by at least 1%
		if progress > lastProgress {
			lastProgress = progress
			progressCb(filepath.Base(filePath), progress, "processing", nil)
		}

		if strings.Contains(line, "ERROR") {
			result.ErrorCount++
		} else if strings.Contains(line, "WARN") {
			result.WarnCount++
		}
	}

	if err := scanner.Err(); err != nil {
		progressCb(filepath.Base(filePath), 0, "error", err)
		return result, fmt.Errorf("error reading file %s: %v", filePath, err)
	}

	// Report 100% completion
	progressCb(filepath.Base(filePath), 100, "completed", nil)

	return result, nil
}

// ProcessFiles processes multiple log files concurrently
func (p *Processor) ProcessFiles(fileNames []string, progressCb ProgressCallback) ([]models.Result, error) {
	filePaths := make([]string, len(fileNames))
	for i, fileName := range fileNames {
		filePaths[i] = filepath.Join(p.uploadDir, fileName)
	}

	var wg sync.WaitGroup
	fileChan := make(chan string, len(filePaths))
	resultChan := make(chan models.Result, len(filePaths))

	// Start workers
	for i := range p.numWorkers {
		fmt.Printf("Starting worker: %d\n", i)
		go p.worker(fileChan, resultChan, &wg, progressCb)
	}

	// Send files to workers
	for _, filePath := range filePaths {
		wg.Add(1)
		fmt.Printf("Adding file to channel: %s\n", filePath)
		fileChan <- filePath
	}
	close(fileChan)

	// Close results channel when all workers are done
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	var results []models.Result
	for result := range resultChan {
		results = append(results, result)
		fmt.Println(result.String())
	}

	return results, nil
}

// worker processes files from the input channel
func (p *Processor) worker(filePaths <-chan string, results chan<- models.Result, wg *sync.WaitGroup, progressCb ProgressCallback) {
	for filePath := range filePaths {
		fmt.Printf("Worker processing file: %s\n", filePath)
		result, err := ProcessLogFile(filePath, progressCb)
		if err != nil {
			fmt.Printf("Error processing file %s: %v\n", filePath, err)
		} else {
			results <- result
		}
		wg.Done()
	}
}
