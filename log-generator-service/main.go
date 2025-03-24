package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"sync"
	"time"
)

const expectedFileSize = 5000 * 1024 * 1024 // 5GB

func getLogMessage() string {
	var logLevels = []string{"DEBUG", "INFO", "WARN", "ERROR", "FATAL"}

	messages := map[string]string{
		"INFO":  "User authentication successful",
		"FATAL": "Database connection failed",
		"WARN":  "API request failed: timeout",
		"ERROR": "Invalid request format detected",
		"DEBUG": "Payment process started",
	}
	level := logLevels[rand.Intn(len(logLevels))]
	return fmt.Sprintf("%s %s", level, messages[level])
}

func randomJSONPayload() string {
	// 30% chance to include JSON payload
	if rand.Intn(100) > 70 {
		payload := map[string]any{
			"userId": rand.Intn(1000),
			"ip":     fmt.Sprintf("192.168.%d.%d", rand.Intn(256), rand.Intn(256)),
		}
		data, _ := json.Marshal(payload)
		return fmt.Sprintf(" %s", string(data))
	}
	return ""
}

func generateLogEntry() string {
	timestamp := time.Now().Format(time.RFC3339)

	return fmt.Sprintf("[%s] %s%s\n", timestamp, getLogMessage(), randomJSONPayload())
}

func writeLogToFile(wg *sync.WaitGroup, logFilePath string, expectedFileSize int64) {
	defer wg.Done()

	fmt.Printf("Creating file: %s\n", logFilePath)
	file, err := os.Create(logFilePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	var buffer bytes.Buffer
	var currentFileSize int64

	const bufferThreshold = 64 * 1024 // Batch write size: 64KB

	for currentFileSize < expectedFileSize {
		buffer.WriteString(generateLogEntry())

		if buffer.Len() >= bufferThreshold {
			n, err := file.Write(buffer.Bytes())
			if err != nil {
				panic(err)
			}
			currentFileSize += int64(n)
			buffer.Reset()
		}

		if buffer.Len() > 0 {
			n, err := file.Write(buffer.Bytes())
			if err != nil {
				panic(err)
			}
			currentFileSize += int64(n)
		}
	}

	fmt.Printf("Log file '%s' generated successfully with size ~ 5 GB\n", logFilePath)
}

func main() {
	var wg sync.WaitGroup

	logFilePaths := []string{
		"../uploads/app6.log",
		// "../uploads/app2.log",
		// "../uploads/app3.log",
		// "../uploads/app4.log",
		// "../uploads/app5.log",
	}

	for _, logFilePath := range logFilePaths {
		wg.Add(1)
		go writeLogToFile(&wg, logFilePath, expectedFileSize)
	}
	wg.Wait()
}
