package validation

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type URLValidator struct {
	client *http.Client
}

func NewURLValidator() *URLValidator {
	return &URLValidator{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// ValidateImageURL validates if the URL is a valid image URL
func (v *URLValidator) ValidateImageURL(imageURL string) error {
	if imageURL == "" {
		return nil // Empty URL is allowed
	}

	// Parse URL
	parsedURL, err := url.Parse(imageURL)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	// Check if URL has valid scheme
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("URL must use http or https scheme")
	}

	// Check if URL has valid host
	if parsedURL.Host == "" {
		return fmt.Errorf("URL must have a valid host")
	}

	// Check if URL points to an image (basic check by extension)
	if !v.isImageURL(imageURL) {
		return fmt.Errorf("URL does not appear to be an image")
	}

	return nil
}

// ValidateImageURLWithHTTPCheck validates URL and checks if it's accessible
func (v *URLValidator) ValidateImageURLWithHTTPCheck(imageURL string) error {
	if err := v.ValidateImageURL(imageURL); err != nil {
		return err
	}

	if imageURL == "" {
		return nil
	}

	// Make HEAD request to check if URL is accessible
	resp, err := v.client.Head(imageURL)
	if err != nil {
		return fmt.Errorf("unable to access URL: %w", err)
	}
	defer resp.Body.Close()

	// Check if response is successful
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("URL returned status code: %d", resp.StatusCode)
	}

	// Check content type
	contentType := resp.Header.Get("Content-Type")
	if !v.isImageContentType(contentType) {
		return fmt.Errorf("URL does not serve an image (content-type: %s)", contentType)
	}

	return nil
}

// isImageURL checks if URL appears to be an image based on extension or known image hosting patterns
func (v *URLValidator) isImageURL(imageURL string) bool {
	// Convert to lowercase for case-insensitive comparison
	lowerURL := strings.ToLower(imageURL)

	// Check common image extensions
	imageExtensions := []string{".jpg", ".jpeg", ".png", ".gif", ".webp", ".bmp", ".svg"}
	for _, ext := range imageExtensions {
		if strings.Contains(lowerURL, ext) {
			return true
		}
	}

	// Check known image hosting services
	imageHosts := []string{
		"cloudinary.com",
		"imgur.com",
		"unsplash.com",
		"pexels.com",
		"pixabay.com",
		"amazonaws.com", // S3
		"googleusercontent.com",
		"firebase.com",
		"imagekit.io",
	}

	for _, host := range imageHosts {
		if strings.Contains(lowerURL, host) {
			return true
		}
	}

	return false
}

// isImageContentType checks if the content type is an image
func (v *URLValidator) isImageContentType(contentType string) bool {
	if contentType == "" {
		return false
	}

	imageTypes := []string{
		"image/jpeg",
		"image/jpg",
		"image/png",
		"image/gif",
		"image/webp",
		"image/bmp",
		"image/svg+xml",
	}

	contentType = strings.ToLower(strings.Split(contentType, ";")[0])
	for _, imageType := range imageTypes {
		if contentType == imageType {
			return true
		}
	}

	return false
}