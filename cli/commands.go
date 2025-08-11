package cli

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"

	"r34-go/config"
	"r34-go/models"
	"r34-go/services"
)

var (
	tags      string
	quantity  uint16
	outputDir string
	useAPI    bool
	images    bool
	gifs      bool
	videos    bool
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "r34downloader",
	Short: "Rule34.xxx content downloader",
	Long: `A CLI tool to download content from rule34.xxx using tags.

Examples:
  # Download 50 images with specific tags
  r34downloader -t "cat_girl solo" -q 50 -o ./my_downloads

  # Download only videos using API method
  r34downloader -t "animated" -q 20 --videos --no-images --no-gifs --api

  # Download everything (images, gifs, videos) using HTML parsing
  r34downloader -t "pokemon" -q 100 --no-api`,
	Run: runDownload,
}

// ConfigCmd shows current configuration
var ConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Show current configuration",
	Long:  `Display the current configuration settings`,
	Run:   showConfig,
}

// CheckCmd checks if content exists for given tags
var CheckCmd = &cobra.Command{
	Use:   "check",
	Short: "Check if content exists for given tags",
	Long:  `Check if there's any content available for the specified tags without downloading`,
	Run:   checkContent,
}

func init() {
	// Initialize configuration
	config.Init()

	// Root command flags
	RootCmd.Flags().StringVarP(&tags, "tags", "t", "", "Tags to search for (required)")
	RootCmd.Flags().Uint16VarP(&quantity, "quantity", "q", 100, "Number of items to download")
	RootCmd.Flags().StringVarP(&outputDir, "output", "o", "./downloads", "Output directory")
	RootCmd.Flags().BoolVarP(&useAPI, "api", "a", config.AppSettings.IsAPI, "Use API method (faster) instead of HTML parsing")
	RootCmd.Flags().BoolVar(&images, "images", config.AppSettings.Images, "Download images")
	RootCmd.Flags().BoolVar(&gifs, "gifs", config.AppSettings.Gif, "Download GIFs")
	RootCmd.Flags().BoolVar(&videos, "videos", config.AppSettings.Video, "Download videos")

	// Convenience flags for disabling file types
	RootCmd.Flags().BoolVar(&images, "no-images", !config.AppSettings.Images, "Don't download images")
	RootCmd.Flags().BoolVar(&gifs, "no-gifs", !config.AppSettings.Gif, "Don't download GIFs")
	RootCmd.Flags().BoolVar(&videos, "no-videos", !config.AppSettings.Video, "Don't download videos")

	// Mark required flags
	RootCmd.MarkFlagRequired("tags")

	// Add subcommands
	RootCmd.AddCommand(ConfigCmd)
	RootCmd.AddCommand(CheckCmd)

	// Check command flags
	CheckCmd.Flags().StringVarP(&tags, "tags", "t", "", "Tags to search for (required)")
	CheckCmd.Flags().BoolVarP(&useAPI, "api", "a", config.AppSettings.IsAPI, "Use API method to check")
	CheckCmd.MarkFlagRequired("tags")
}

func runDownload(cmd *cobra.Command, args []string) {
	// Handle negative flags
	if cmd.Flag("no-images").Changed {
		images = false
	}
	if cmd.Flag("no-gifs").Changed {
		gifs = false
	}
	if cmd.Flag("no-videos").Changed {
		videos = false
	}

	// Update config with command line flags
	config.AppSettings.Limit = quantity
	config.AppSettings.Images = images
	config.AppSettings.Gif = gifs
	config.AppSettings.Video = videos
	config.AppSettings.IsAPI = useAPI

	// Validate that at least one file type is enabled
	if !images && !gifs && !videos {
		log.Fatal("Error: At least one file type must be enabled (images, gifs, or videos)")
	}

	// Create output directory
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	fmt.Printf("Downloading %d items for tags: %s\n", quantity, tags)
	fmt.Printf("Output directory: %s\n", outputDir)
	fmt.Printf("Method: %s\n", getMethodName())
	fmt.Printf("File types: %s\n", getEnabledFileTypes())
	fmt.Println()

	// Create progress bar
	bar := progressbar.NewOptions(int(quantity),
		progressbar.OptionSetDescription("Downloading..."),
		progressbar.OptionSetWidth(50),
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
		progressbar.OptionSetRenderBlankState(true),
	)

	// Progress callback function
	progressCallback := func(current, total int) {
		bar.Set(current)
	}

	var stats *models.DownloadStats
	var err error

	if useAPI {
		// Use API method
		apiService := services.NewAPIService()

		// Check if content exists
		count, err := apiService.GetContentCount(tags)
		if err != nil {
			log.Fatalf("Failed to get content count: %v", err)
		}

		if count == 0 {
			fmt.Println("No content found for the specified tags.")
			return
		}

		fmt.Printf("Found %d total items available.\n", count)

		if quantity > uint16(count) {
			fmt.Printf("Warning: Requested %d items but only %d available. Downloading all available items.\n", quantity, count)
			quantity = uint16(count)
		}

		stats, err = apiService.DownloadContent(outputDir, tags, quantity, progressCallback)
	} else {
		// Use HTML parsing method
		htmlService := services.NewHTMLService()

		// Check if content exists
		found, err := htmlService.IsSomethingFound(tags)
		if err != nil {
			log.Fatalf("Failed to check for content: %v", err)
		}

		if !found {
			fmt.Println("No content found for the specified tags.")
			return
		}

		stats, err = htmlService.DownloadContent(outputDir, tags, quantity, progressCallback)
	}

	if err != nil {
		log.Fatalf("Download failed: %v", err)
	}

	// Finish progress bar
	bar.Finish()

	// Print final statistics
	printDownloadSummary(stats, outputDir)
}

func showConfig(cmd *cobra.Command, args []string) {
	fmt.Println("Current Configuration:")
	fmt.Printf("  Limit: %d\n", config.AppSettings.Limit)
	fmt.Printf("  Download Images: %t\n", config.AppSettings.Images)
	fmt.Printf("  Download GIFs: %t\n", config.AppSettings.Gif)
	fmt.Printf("  Download Videos: %t\n", config.AppSettings.Video)
	fmt.Printf("  Use API: %t\n", config.AppSettings.IsAPI)
}

func checkContent(cmd *cobra.Command, args []string) {
	fmt.Printf("Checking content for tags: %s\n", tags)
	fmt.Printf("Method: %s\n", getMethodName())

	if useAPI {
		apiService := services.NewAPIService()
		count, err := apiService.GetContentCount(tags)
		if err != nil {
			log.Fatalf("Failed to check content: %v", err)
		}

		if count > 0 {
			fmt.Printf("✓ Found %d items available\n", count)
		} else {
			fmt.Println("✗ No content found for the specified tags")
		}
	} else {
		htmlService := services.NewHTMLService()
		found, err := htmlService.IsSomethingFound(tags)
		if err != nil {
			log.Fatalf("Failed to check content: %v", err)
		}

		if found {
			// Try to get more detailed info
			maxPid, err := htmlService.GetMaxPid(tags)
			if err == nil && maxPid > 0 {
				fmt.Printf("✓ Content found (up to page %d)\n", maxPid)
			} else {
				fmt.Println("✓ Content found")
			}
		} else {
			fmt.Println("✗ No content found for the specified tags")
		}
	}
}

func getMethodName() string {
	if useAPI {
		return "API (faster)"
	}
	return "HTML parsing"
}

func getEnabledFileTypes() string {
	var types []string
	if images {
		types = append(types, "images")
	}
	if gifs {
		types = append(types, "gifs")
	}
	if videos {
		types = append(types, "videos")
	}

	if len(types) == 0 {
		return "none"
	}

	return strings.Join(types, ", ")
}

func printDownloadSummary(stats *models.DownloadStats, outputDir string) {
	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("Download Summary:")
	fmt.Printf("Total requested: %d\n", stats.Total)
	fmt.Printf("Successfully downloaded: %d\n", stats.Downloaded)

	if stats.Failed > 0 {
		fmt.Printf("Failed: %d\n", stats.Failed)
	}
	if stats.Skipped > 0 {
		fmt.Printf("Skipped (already exists): %d\n", stats.Skipped)
	}

	if stats.Images > 0 || stats.Gifs > 0 || stats.Videos > 0 {
		fmt.Println("\nBy file type:")
		if stats.Images > 0 {
			fmt.Printf("  Images: %d\n", stats.Images)
		}
		if stats.Gifs > 0 {
			fmt.Printf("  GIFs: %d\n", stats.Gifs)
		}
		if stats.Videos > 0 {
			fmt.Printf("  Videos: %d\n", stats.Videos)
		}
	}

	fmt.Printf("\nFiles saved to: %s\n", outputDir)

	// Show folder structure
	fmt.Println("\nFolder structure:")
	if config.AppSettings.Images {
		fmt.Printf("  %s/Images/\n", outputDir)
	}
	if config.AppSettings.Gif {
		fmt.Printf("  %s/Gif/\n", outputDir)
	}
	if config.AppSettings.Video {
		fmt.Printf("  %s/Video/\n", outputDir)
	}
}
