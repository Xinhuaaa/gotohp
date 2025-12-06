package backend

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
)

// MediaBrowser handles media browsing operations
type MediaBrowser struct{}

// GetMediaList retrieves a paginated list of media items
func (m *MediaBrowser) GetMediaList(pageToken string, limit int) (*MediaListResult, error) {
	api, err := NewApi()
	if err != nil {
		return nil, fmt.Errorf("failed to create API client: %w", err)
	}

	result, err := api.GetMediaList(pageToken, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get media list: %w", err)
	}

	return result, nil
}

// GetThumbnail retrieves a thumbnail for a media item and returns it as base64
func (m *MediaBrowser) GetThumbnail(mediaKey string, size string) (string, error) {
	api, err := NewApi()
	if err != nil {
		return "", fmt.Errorf("failed to create API client: %w", err)
	}

	// Parse size to width/height
	var width, height int
	switch size {
	case "small":
		width, height = 200, 200
	case "medium":
		width, height = 400, 400
	case "large":
		width, height = 800, 800
	default:
		width, height = 400, 400 // default to medium
	}

	thumbnailData, err := api.GetThumbnail(mediaKey, width, height, false, 0, false)
	if err != nil {
		return "", fmt.Errorf("failed to get thumbnail: %w", err)
	}

	// Convert to base64
	base64Data := base64.StdEncoding.EncodeToString(thumbnailData)
	return base64Data, nil
}

// DownloadMedia downloads a media item to the user's Downloads folder
func (m *MediaBrowser) DownloadMedia(mediaKey string) (string, error) {
	api, err := NewApi()
	if err != nil {
		return "", fmt.Errorf("failed to create API client: %w", err)
	}

	// Get media info to determine filename
	mediaInfo, err := api.GetMediaInfo(mediaKey)
	if err != nil {
		return "", fmt.Errorf("failed to get media info: %w", err)
	}

	// Get download URLs
	downloadURLs, err := api.GetDownloadURLs(mediaKey)
	if err != nil {
		return "", fmt.Errorf("failed to get download URLs: %w", err)
	}

	// Use original URL if available, otherwise use edited URL
	downloadURL := downloadURLs.EditedURL
	if downloadURLs.OriginalURL != "" {
		downloadURL = downloadURLs.OriginalURL
	}

	// Get user's Downloads folder
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	downloadsDir := filepath.Join(homeDir, "Downloads", "gotohp")
	err = os.MkdirAll(downloadsDir, 0755)
	if err != nil {
		return "", fmt.Errorf("failed to create downloads directory: %w", err)
	}

	// Determine output path
	filename := mediaInfo.Filename
	if filename == "" {
		filename = fmt.Sprintf("%s.jpg", mediaKey[:10])
	}
	outputPath := filepath.Join(downloadsDir, filename)

	// Download the file
	err = api.DownloadFile(downloadURL, outputPath)
	if err != nil {
		return "", fmt.Errorf("failed to download file: %w", err)
	}

	return outputPath, nil
}


