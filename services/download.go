package services

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// DownloadService handles file downloads
type DownloadService struct {
	client *http.Client
}

// NewDownloadService creates a new download service instance
func NewDownloadService() *DownloadService {
	return &DownloadService{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Download downloads a file from URL and saves it to the specified path
func (ds *DownloadService) Download(url, filePath string) error {
	// Check if file already exists
	if _, err := os.Stat(filePath); err == nil {
		return nil // File already exists, skip download
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Make HTTP request
	resp, err := ds.client.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download from %s: %w", url, err)
	}
	defer resp.Body.Close()

	// Check if request was successful
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Create the file
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", filePath, err)
	}
	defer file.Close()

	// Copy the response body to file
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		// If copy failed, remove the partial file
		os.Remove(filePath)
		return fmt.Errorf("failed to write file %s: %w", filePath, err)
	}

	return nil
}

// DownloadWithRetry downloads a file with retry logic
func (ds *DownloadService) DownloadWithRetry(url, filePath string, maxRetries int) error {
	var lastErr error
	for i := 0; i <= maxRetries; i++ {
		if err := ds.Download(url, filePath); err != nil {
			lastErr = err
			if i < maxRetries {
				// Wait before retrying (exponential backoff)
				waitTime := time.Duration(i+1) * time.Second
				time.Sleep(waitTime)
				continue
			}
		} else {
			return nil // Success
		}
	}
	return lastErr
}
