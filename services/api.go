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
	
	downloaded := 0
	pid := 0
	
	// Keep fetching pages until we have enough content or run out of pages
	for downloaded < int(quantity) {
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

		// If no posts found, we've reached the end
		if len(apiResp.Posts) == 0 {
			break
		}

		// Process posts on this page
		remaining := int(quantity) - downloaded
		postsToProcess := len(apiResp.Posts)
		if remaining < postsToProcess {
			postsToProcess = remaining
		}

		for i := 0; i < postsToProcess; i++ {
			post := apiResp.Posts[i]
			
			downloadResult := as.downloadPost(post, path, stats)
			if downloadResult == "downloaded" {
				stats.Downloaded++
				downloaded++
			} else if downloadResult == "skipped" {
				stats.Skipped++
				downloaded++ // Count skipped as processed
			} else if downloadResult == "failed" {
				stats.Failed++
			}
			// If disabled file type, don't count towards downloaded but continue
			
			if progressCallback != nil {
				progressCallback(downloaded, int(quantity))
			}
			
			// Break if we've downloaded enough
			if downloaded >= int(quantity) {
				break
			}
			
			// Small delay to avoid overwhelming the server
			time.Sleep(100 * time.Millisecond)
		}
		
		// Move to next page
		pid++
		
		// If we processed fewer posts than available on this page and we're done, break
		if downloaded >= int(quantity) {
			break
		}
	}

	return stats, nil
}

func (as *APIService) downloadPost(post models.Post, basePath string, stats *models.DownloadStats) string {
	if post.FileURL == "" {
		return "failed"
	}

	fileExt := strings.ToLower(filepath.Ext(post.FileURL))
	filename := post.ID + fileExt

	var filePath string
	var shouldDownload bool

	switch fileExt {
	case ".mp4", ".webm":
		if !config.AppSettings.Video {
			return "disabled" // File type disabled
		}
		shouldDownload = true
		
		// Use sample URL if available for videos
		downloadURL := post.SampleURL
		if downloadURL == "" {
			downloadURL = post.FileURL
		}
		
		filePath = filepath.Join(basePath, "Video", filename)
		err := as.downloadService.Download(downloadURL, filePath)
		if err != nil {
			if err.Error() == "file already exists" {
				return "skipped"
			}
			return "failed"
		}
		stats.Videos++
		
	case ".gif":
		if !config.AppSettings.Gif {
			return "disabled" // File type disabled
		}
		shouldDownload = true
		
		filePath = filepath.Join(basePath, "Gif", filename)
		err := as.downloadService.Download(post.FileURL, filePath)
		if err != nil {
			if err.Error() == "file already exists" {
				return "skipped"
			}
			return "failed"
		}
		stats.Gifs++
		
	default:
		if !config.AppSettings.Images {
			return "disabled" // File type disabled
		}
		shouldDownload = true
		
		filePath = filepath.Join(basePath, "Images", filename)
		err := as.downloadService.Download(post.FileURL, filePath)
		if err != nil {
			if err.Error() == "file already exists" {
				return "skipped"
			}
			return "failed"
		}
		stats.Images++
	}

	if shouldDownload {
		return "downloaded"
	}
	
	return "skipped"
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
