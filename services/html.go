package services

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"r34-go/config"
	"r34-go/models"
	"r34-go/utils"
)

const (
	contentURL     = "https://rule34.xxx/index.php?page=post&s=list&tags="
	htmlPageSize   = 42
	rule34BaseURL  = "https://rule34.xxx/"
)

// HTMLService handles HTML parsing and downloading
type HTMLService struct {
	client          *http.Client
	downloadService *DownloadService
}

// NewHTMLService creates a new HTML service instance
func NewHTMLService() *HTMLService {
	return &HTMLService{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		downloadService: NewDownloadService(),
	}
}

// IsSomethingFound checks if there's any content for the specified tags
func (hs *HTMLService) IsSomethingFound(tags string) (bool, error) {
	url := fmt.Sprintf("%s%s", contentURL, tags)
	
	doc, err := hs.loadHTMLDocument(url)
	if err != nil {
		return false, err
	}

	// Check for thumbnail content
	thumbs := doc.Find("div.content span.thumb")
	return thumbs.Length() > 0, nil
}

// GetMaxPid returns the maximum page number for the specified tags
func (hs *HTMLService) GetMaxPid(tags string) (int, error) {
	url := fmt.Sprintf("%s%s", contentURL, tags)
	
	doc, err := hs.loadHTMLDocument(url)
	if err != nil {
		return 0, err
	}

	// Find the last page link
	lastPageLink := doc.Find("div.pagination a[alt='last page']").First()
	if lastPageLink.Length() == 0 {
		return 0, nil // No pagination found
	}

	href, exists := lastPageLink.Attr("href")
	if !exists {
		return 0, fmt.Errorf("last page link has no href attribute")
	}

	// Extract PID from URL
	parts := strings.Split(href, "=")
	if len(parts) < 2 {
		return 0, fmt.Errorf("invalid pagination URL format")
	}

	maxPid, err := strconv.Atoi(parts[len(parts)-1])
	if err != nil {
		return 0, fmt.Errorf("failed to parse max PID: %w", err)
	}

	return maxPid, nil
}

// GetCountContent returns the amount of content on the specified page
func (hs *HTMLService) GetCountContent(tags string, pid int) (int, error) {
	url := fmt.Sprintf("%s%s&pid=%d", contentURL, tags, pid)
	
	doc, err := hs.loadHTMLDocument(url)
	if err != nil {
		return -1, err
	}

	thumbLinks := doc.Find("div.content span.thumb a")
	if thumbLinks.Length() == 0 {
		return -1, nil // No content found
	}

	if pid == 0 {
		return thumbLinks.Length(), nil
	}

	return pid + thumbLinks.Length(), nil
}

// DownloadContent downloads content using HTML parsing method
func (hs *HTMLService) DownloadContent(path, tags string, quantity uint16, progressCallback models.ProgressCallback) (*models.DownloadStats, error) {
	stats := &models.DownloadStats{Total: int(quantity)}
	
	maxPages := int(quantity)
	residue := htmlPageSize

	if quantity < htmlPageSize {
		maxPages = htmlPageSize
		residue = int(quantity)
	}

	for pid := 0; pid < maxPages; pid += htmlPageSize {
		url := fmt.Sprintf("%s%s&pid=%d", contentURL, tags, pid)
		
		doc, err := hs.loadHTMLDocument(url)
		if err != nil {
			return stats, fmt.Errorf("failed to load page at PID %d: %w", pid, err)
		}

		var posts []string
		doc.Find("div.content span.thumb a").Each(func(i int, s *goquery.Selection) {
			href, exists := s.Attr("href")
			if exists && href != "" {
				// Replace &amp; with &
				href = strings.ReplaceAll(href, "&amp;", "&")
				posts = append(posts, href)
			}
		})

		if len(posts) == 0 {
			break // No more posts found
		}

		err = hs.downloadPosts(posts, path, pid, residue, maxPages, stats, progressCallback, int(quantity))
		if err != nil {
			return stats, err
		}
	}

	return stats, nil
}

func (hs *HTMLService) downloadPosts(posts []string, path string, pid, residue, maxPages int, stats *models.DownloadStats, progressCallback models.ProgressCallback, totalQuantity int) error {
	maxPosts := len(posts)
	if maxPages-pid < htmlPageSize {
		maxPosts = maxPages - pid
	} else if maxPages-pid == htmlPageSize {
		maxPosts = residue
	}

	for i := 0; i < maxPosts && i < len(posts); i++ {
		postURL := rule34BaseURL + posts[i]
		
		doc, err := hs.loadHTMLDocument(postURL)
		if err != nil {
			stats.Failed++
			continue
		}

		// Check for video first
		videoSrc, videoExists := doc.Find("video#gelcomVideoPlayer source").Attr("src")
		if videoExists && config.AppSettings.Video {
			filename := hs.extractFilename(videoSrc)
			filePath := filepath.Join(path, "Video", filename)
			
			if err := hs.downloadService.Download(videoSrc, filePath); err != nil {
				stats.Failed++
			} else {
				stats.Videos++
				stats.Downloaded++
			}
		} else {
			// Check for image
			imageSrc, imageExists := doc.Find("div.content img#image").Attr("src")
			if imageExists {
				err := hs.downloadImage(imageSrc, path, stats)
				if err != nil {
					stats.Failed++
				} else {
					stats.Downloaded++
				}
			} else {
				stats.Failed++
			}
		}

		reportStatus := pid + i + 1
		if progressCallback != nil {
			progressCallback(reportStatus, totalQuantity)
		}

		// Small delay to avoid overwhelming the server
		time.Sleep(100 * time.Millisecond)
	}

	return nil
}

func (hs *HTMLService) downloadImage(imageSrc, path string, stats *models.DownloadStats) error {
	// Extract ID from URL query parameters
	id := utils.ExtractIDFromImageURL(imageSrc)
	baseImageURL := strings.Split(imageSrc, "?")[0]
	fileExt := utils.GetFileExtension(baseImageURL)
	filename := utils.SanitizeFilename(id + fileExt)

	fileType := utils.ClassifyFileType(fileExt)
	
	if fileType == "gif" && config.AppSettings.Gif {
		filePath := filepath.Join(path, "Gif", filename)
		err := hs.downloadService.Download(imageSrc, filePath)
		if err == nil {
			stats.Gifs++
		}
		return err
	} else if fileType == "image" && config.AppSettings.Images {
		filePath := filepath.Join(path, "Images", filename)
		err := hs.downloadService.Download(imageSrc, filePath)
		if err == nil {
			stats.Images++
		}
		return err
	}

	return nil // File type disabled in settings
}

func (hs *HTMLService) extractFilename(url string) string {
	return utils.SanitizeFilename(utils.ExtractFilenameFromURL(url))
}

func (hs *HTMLService) loadHTMLDocument(url string) (*goquery.Document, error) {
	resp, err := hs.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status: %s for URL %s", resp.Status, url)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	return doc, nil
}
