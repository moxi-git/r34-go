package models

// Post represents a Rule34 post
type Post struct {
	ID          string `xml:"id,attr"`
	FileURL     string `xml:"file_url,attr"`
	SampleURL   string `xml:"sample_url,attr"`
	PreviewURL  string `xml:"preview_url,attr"`
	Tags        string `xml:"tags,attr"`
	Score       int    `xml:"score,attr"`
	Rating      string `xml:"rating,attr"`
	Width       int    `xml:"width,attr"`
	Height      int    `xml:"height,attr"`
	MD5         string `xml:"md5,attr"`
	CreatedAt   string `xml:"created_at,attr"`
}

// APIResponse represents the XML response from Rule34 API
type APIResponse struct {
	Count int    `xml:"count,attr"`
	Posts []Post `xml:"post"`
}

// ProgressCallback is a function type for progress reporting
type ProgressCallback func(current, total int)

// DownloadStats holds download statistics
type DownloadStats struct {
	Total       int
	Downloaded  int
	Skipped     int
	Failed      int
	Images      int
	Gifs        int
	Videos      int
}
