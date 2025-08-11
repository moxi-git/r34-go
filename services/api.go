package services

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"r34-go/config"
	"r34-go/models"
)

const (
	apiURL   = "https://rule34.xxx/index.php?page=dapi&s=post&q=index"
	pageSize = 100
)

// APIService handles Rule34 API interactions
type APIService struct {
	client          *http.Client
	downloadService *DownloadService
}

// NewAPIService creates a new API service instance
func NewAPIService() *APIService {
	return &APIService{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		downloadService: NewDownloadService(),
	}
}

// GetContentCount returns the total number of posts for given tags
func (as *APIService) GetContentCount(tags string) (int, error) {
	url := fmt.Sprintf("%s&tags=%s", apiURL, tags)
	
	resp, err := as.client.Get(url)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch content count: %w", err)
	}
	defer resp.Body.Close()

	var apiResp models.APIResponse
	if err := xml.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return 0, fmt.Errorf("failed to decode XML response: %w", err)
	}

	return apiResp.Count, nil
}

// DownloadContent downloads posts using the API method
func (as *APIService) DownloadContent(path, tags string, quantity uint16, progressCallback models.ProgressCallback) (*models.DownloadStats, error) {
	stats := &models.DownloadStats{Total: int(quantity)}
	
	maxPid := as.calculateMaxPid(quantity)
	
	for pid := 0; pid <= maxPid; pid++ {
		url := fmt.Sprintf("%s&tags=%s&pid=%d", apiURL, tags, pid)
		
		resp, err := as.client.Get(url)
		if err != nil {
			return stats, fmt.Errorf("failed to fetch page %d: %w", pid, err)
		}

		var apiResp models.APIResponse
		if err := xml.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
			resp.Body.Close()
			return stats, fmt.Errorf("failed to decode XML for page %d: %w", pid, err)
		}
		resp.Body.Close()

		postCount := as.calculatePostCount(quantity, pid)
		if postCount > len(apiResp.Posts) {
			postCount = len(apiResp.Posts)
		}

		for i := 0; i < postCount; i++ {
			post := apiResp.Posts[i]
			
			if err := as.downloadPost(post, path); err != nil {
				stats.Failed++
				// Continue with next post on error
				continue
			}
			
			stats.Downloaded++
			reportStatus := pid*pageSize + i + 1
			if progressCallback != nil {
				progressCallback(reportStatus, int(quantity))
			}
			
			// Small delay to avoid overwhelming the server
			time.Sleep(100 * time.Millisecond)
		}
	}

	return stats, nil
}

func (as *APIService) downloadPost(post models.Post, basePath string) error {
	if post.FileURL == "" {
		return fmt.Errorf("empty file URL for post %s", post.ID)
	}

	fileExt := strings.ToLower(filepath.Ext(post.FileURL))
	filename := post.ID + fileExt

	switch fileExt {
	case ".mp4", ".webm":
		if !config.AppSettings.Video {
			return nil // Skip videos if disabled
		}
		
		// Use sample URL if available for videos
		downloadURL := post.SampleURL
		if downloadURL == "" {
			downloadURL = post.FileURL
		}
		
		filePath := filepath.Join(basePath, "Video", filename)
		return as.downloadService.Download(downloadURL, filePath)
		
	case ".gif":
		if !config.AppSettings.Gif {
			return nil // Skip gifs if disabled
		}
		
		filePath := filepath.Join(basePath, "Gif", filename)
		return as.downloadService.Download(post.FileURL, filePath)
		
	default:
		if !config.AppSettings.Images {
			return nil // Skip images if disabled
		}
		
		filePath := filepath.Join(basePath, "Images", filename)
		return as.downloadService.Download(post.FileURL, filePath)
	}
}

func (as *APIService) calculateMaxPid(quantity uint16) int {
	if quantity <= pageSize {
		return 0
	}
	if quantity%pageSize == 0 {
		return int(quantity/pageSize) - 1
	}
	return int(quantity / pageSize)
}

func (as *APIService) calculatePostCount(quantity uint16, pid int) int {
	remaining := int(quantity) - pid*pageSize
	if remaining < pageSize {
		return remaining
	}
	return pageSize
}
