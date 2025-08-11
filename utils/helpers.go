package utils

import (
	"fmt"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// ValidateTags validates and cleans up tag input
func ValidateTags(tags string) (string, error) {
	if strings.TrimSpace(tags) == "" {
		return "", fmt.Errorf("tags cannot be empty")
	}
	
	// Clean up tags: remove extra spaces, normalize
	cleaned := strings.TrimSpace(tags)
	cleaned = regexp.MustCompile(`\s+`).ReplaceAllString(cleaned, " ")
	
	return cleaned, nil
}

// URLEncodeTags properly encodes tags for URL usage
func URLEncodeTags(tags string) string {
	return url.QueryEscape(tags)
}

// ExtractFilenameFromURL extracts filename from URL, handling query parameters
func ExtractFilenameFromURL(rawURL string) string {
	filename := filepath.Base(rawURL)
	
	// Remove query parameters if present
	if questionIndex := strings.Index(filename, "?"); questionIndex > 0 {
		filename = filename[:questionIndex]
	}
	
	// Handle edge case where filename might be empty after query removal
	if filename == "" || filename == "." {
		// Generate a filename based on timestamp
		filename = fmt.Sprintf("file_%d", time.Now().Unix())
	}
	
	return filename
}

// ExtractIDFromImageURL extracts ID from Rule34 image URLs
func ExtractIDFromImageURL(imageURL string) string {
	parts := strings.Split(imageURL, "?")
	if len(parts) >= 2 {
		return parts[1]
	}
	
	// Fallback: try to extract from filename
	filename := filepath.Base(imageURL)
	if dotIndex := strings.LastIndex(filename, "."); dotIndex > 0 {
		return filename[:dotIndex]
	}
	
	return fmt.Sprintf("unknown_%d", time.Now().Unix())
}

// GetFileExtension safely gets file extension from URL
func GetFileExtension(rawURL string) string {
	// Remove query parameters first
	if questionIndex := strings.Index(rawURL, "?"); questionIndex > 0 {
		rawURL = rawURL[:questionIndex]
	}
	
	ext := strings.ToLower(filepath.Ext(rawURL))
	
	// Handle cases where extension might be missing
	if ext == "" {
		// Try to determine from URL pattern or return default
		return ".jpg" // Default fallback
	}
	
	return ext
}

// IsValidFileExtension checks if the file extension is supported
func IsValidFileExtension(ext string) bool {
	supportedExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".webp": true,
		".bmp":  true,
		".mp4":  true,
		".webm": true,
		".avi":  true,
		".mov":  true,
	}
	
	return supportedExts[strings.ToLower(ext)]
}

// ClassifyFileType determines file type category from extension
func ClassifyFileType(ext string) string {
	ext = strings.ToLower(ext)
	
	switch ext {
	case ".mp4", ".webm", ".avi", ".mov":
		return "video"
	case ".gif":
		return "gif"
	case ".jpg", ".jpeg", ".png", ".webp", ".bmp":
		return "image"
	default:
		return "unknown"
	}
}

// FormatFileSize formats bytes into human readable format
func FormatFileSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// FormatDuration formats duration into human readable format
func FormatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	
	if d < time.Hour {
		return fmt.Sprintf("%.1fm", d.Minutes())
	}
	
	return fmt.Sprintf("%.1fh", d.Hours())
}

// SanitizeFilename removes or replaces invalid characters for filenames
func SanitizeFilename(filename string) string {
	// Replace invalid characters with underscores
	invalidChars := regexp.MustCompile(`[<>:"/\\|?*]`)
	sanitized := invalidChars.ReplaceAllString(filename, "_")
	
	// Remove leading/trailing spaces and dots
	sanitized = strings.Trim(sanitized, " .")
	
	// Ensure filename isn't too long (Windows has 255 char limit)
	if len(sanitized) > 200 {
		ext := filepath.Ext(sanitized)
		name := strings.TrimSuffix(sanitized, ext)
		if len(name) > 200-len(ext) {
			name = name[:200-len(ext)]
		}
		sanitized = name + ext
	}
	
	// Ensure filename isn't empty after sanitization
	if sanitized == "" {
		sanitized = fmt.Sprintf("file_%d", time.Now().Unix())
	}
	
	return sanitized
}

// CalculateETA calculates estimated time of arrival
func CalculateETA(current, total int, elapsed time.Duration) time.Duration {
	if current == 0 {
		return 0
	}
	
	rate := float64(current) / elapsed.Seconds()
	remaining := float64(total - current)
	
	if rate <= 0 {
		return 0
	}
	
	return time.Duration(remaining/rate) * time.Second
}

// TruncateString truncates a string to specified length with ellipsis
func TruncateString(s string, maxLength int) string {
	if len(s) <= maxLength {
		return s
	}
	
	if maxLength <= 3 {
		return s[:maxLength]
	}
	
	return s[:maxLength-3] + "..."
}

// ParsePageFromURL extracts page number from Rule34 URLs
func ParsePageFromURL(rawURL string) int {
	// Extract PID parameter from URL
	if strings.Contains(rawURL, "pid=") {
		parts := strings.Split(rawURL, "pid=")
		if len(parts) >= 2 {
			pidPart := strings.Split(parts[1], "&")[0]
			var pid int
			fmt.Sscanf(pidPart, "%d", &pid)
			return pid
		}
	}
	return 0
}

// BuildContentURL builds the content URL for Rule34
func BuildContentURL(baseURL, tags string, pid int) string {
	encodedTags := URLEncodeTags(tags)
	if pid > 0 {
		return fmt.Sprintf("%s%s&pid=%d", baseURL, encodedTags, pid)
	}
	return fmt.Sprintf("%s%s", baseURL, encodedTags)
}

// RetryWithBackoff executes a function with exponential backoff
func RetryWithBackoff(maxRetries int, operation func() error) error {
	var lastErr error
	
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if err := operation(); err != nil {
			lastErr = err
			if attempt < maxRetries {
				// Exponential backoff: 1s, 2s, 4s, 8s, etc.
				waitTime := time.Duration(1<<uint(attempt)) * time.Second
				time.Sleep(waitTime)
				continue
			}
		} else {
			return nil // Success
		}
	}
	
	return lastErr
}

// IsSupportedImageFormat checks if the file extension is a supported image format
func IsSupportedImageFormat(ext string) bool {
	imageFormats := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".webp": true,
		".bmp":  true,
	}
	
	return imageFormats[strings.ToLower(ext)]
}

// IsSupportedVideoFormat checks if the file extension is a supported video format
func IsSupportedVideoFormat(ext string) bool {
	videoFormats := map[string]bool{
		".mp4":  true,
		".webm": true,
		".avi":  true,
		".mov":  true,
	}
	
	return videoFormats[strings.ToLower(ext)]
}

// IsGifFormat checks if the file extension is GIF
func IsGifFormat(ext string) bool {
	return strings.ToLower(ext) == ".gif"
}
