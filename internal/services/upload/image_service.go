package upload

import (
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

type ImageService struct {
	uploadDir string
	baseURL   string
	logger    *log.Logger
}

func NewImageService(uploadDir, baseURL string) *ImageService {
	// Create upload directory if it doesn't exist
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		log.Printf("Warning: Could not create upload directory: %v", err)
	}

	return &ImageService{
		uploadDir: uploadDir,
		baseURL:   baseURL,
		logger:    log.New(log.Writer(), "[IMAGE_UPLOAD] ", log.LstdFlags|log.Lshortfile),
	}
}

type UploadResult struct {
	Filename string `json:"filename"`
	URL      string `json:"url"`
	Size     int64  `json:"size"`
}

func (s *ImageService) UploadProductImage(file *multipart.FileHeader) (*UploadResult, error) {
	s.logger.Printf("Uploading product image: %s", file.Filename)

	// Validate file
	if err := s.validateImageFile(file); err != nil {
		s.logger.Printf("Image validation failed: %v", err)
		return nil, err
	}

	// Generate unique filename
	filename := s.generateFilename(file.Filename)
	filePath := filepath.Join(s.uploadDir, "products", filename)

	// Create products subdirectory if it doesn't exist
	productsDir := filepath.Join(s.uploadDir, "products")
	if err := os.MkdirAll(productsDir, 0755); err != nil {
		s.logger.Printf("Error creating products directory: %v", err)
		return nil, fmt.Errorf("failed to create upload directory: %w", err)
	}

	// Open uploaded file
	src, err := file.Open()
	if err != nil {
		s.logger.Printf("Error opening uploaded file: %v", err)
		return nil, fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	// Create destination file
	dst, err := os.Create(filePath)
	if err != nil {
		s.logger.Printf("Error creating destination file: %v", err)
		return nil, fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dst.Close()

	// Copy file content
	size, err := io.Copy(dst, src)
	if err != nil {
		s.logger.Printf("Error copying file content: %v", err)
		return nil, fmt.Errorf("failed to save file: %w", err)
	}

	result := &UploadResult{
		Filename: filename,
		URL:      fmt.Sprintf("%s/uploads/products/%s", s.baseURL, filename),
		Size:     size,
	}

	s.logger.Printf("Successfully uploaded image: %s (size: %d bytes)", filename, size)
	return result, nil
}

func (s *ImageService) DeleteProductImage(filename string) error {
	if filename == "" {
		return errors.New("filename is required")
	}

	// Extract filename from URL if full URL is provided
	if strings.Contains(filename, "/") {
		parts := strings.Split(filename, "/")
		filename = parts[len(parts)-1]
	}

	filePath := filepath.Join(s.uploadDir, "products", filename)

	if err := os.Remove(filePath); err != nil {
		if os.IsNotExist(err) {
			s.logger.Printf("File not found for deletion: %s", filename)
			return nil // File doesn't exist, consider it deleted
		}
		s.logger.Printf("Error deleting file %s: %v", filename, err)
		return fmt.Errorf("failed to delete file: %w", err)
	}

	s.logger.Printf("Successfully deleted image: %s", filename)
	return nil
}

func (s *ImageService) validateImageFile(file *multipart.FileHeader) error {
	// Check file size (max 5MB)
	const maxSize = 5 * 1024 * 1024 // 5MB
	if file.Size > maxSize {
		return fmt.Errorf("file size too large: %d bytes (max: %d bytes)", file.Size, maxSize)
	}

	// Check file extension
	allowedExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".webp": true,
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowedExts[ext] {
		return fmt.Errorf("unsupported file type: %s (allowed: jpg, jpeg, png, gif, webp)", ext)
	}

	return nil
}

func (s *ImageService) generateFilename(originalFilename string) string {
	ext := filepath.Ext(originalFilename)
	uuid := uuid.New().String()
	timestamp := time.Now().Unix()
	return fmt.Sprintf("%d_%s%s", timestamp, uuid, ext)
}
