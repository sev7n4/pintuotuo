package services

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const maxFileSize = 25 << 20

var allowedImageMIMETypes = map[string]bool{
	"image/png":  true,
	"image/jpeg": true,
	"image/webp": true,
}

var allowedAudioMIMETypes = map[string]bool{
	"audio/mpeg": true,
	"audio/wav":  true,
	"audio/mp4":  true,
	"audio/webm": true,
	"audio/flac": true,
	"audio/ogg":  true,
}

func ParseFileField(c *gin.Context, fieldName string) ([]byte, string, error) {
	file, header, err := c.Request.FormFile(fieldName)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read file field %q: %w", fieldName, err)
	}
	defer file.Close()

	if header.Size > maxFileSize {
		return nil, "", fmt.Errorf("file size %d exceeds maximum allowed size %d bytes", header.Size, maxFileSize)
	}

	buf := make([]byte, header.Size)
	_, err = file.Read(buf)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read file content: %w", err)
	}

	return buf, header.Filename, nil
}

func ParseFileFieldWithMIMEValidation(c *gin.Context, fieldName string, category string) ([]byte, string, error) {
	data, filename, err := ParseFileField(c, fieldName)
	if err != nil {
		return nil, "", err
	}

	contentType := detectMIMEType(filename)
	var allowed map[string]bool
	switch category {
	case "image":
		allowed = allowedImageMIMETypes
	case "audio":
		allowed = allowedAudioMIMETypes
	default:
		return data, filename, nil
	}

	if !allowed[contentType] {
		allowedTypes := make([]string, 0, len(allowed))
		for t := range allowed {
			allowedTypes = append(allowedTypes, t)
		}
		return nil, "", fmt.Errorf("file type %q not allowed for %s, allowed types: %s", contentType, category, strings.Join(allowedTypes, ", "))
	}

	return data, filename, nil
}

func ParseFormField(c *gin.Context, fieldName string) (string, error) {
	value := c.PostForm(fieldName)
	return value, nil
}

func ParseFormInt(c *gin.Context, fieldName string, defaultValue int) int {
	value := c.PostForm(fieldName)
	if value == "" {
		return defaultValue
	}
	var result int
	fmt.Sscanf(value, "%d", &result)
	if result <= 0 {
		return defaultValue
	}
	return result
}

func detectMIMEType(filename string) string {
	lower := strings.ToLower(filename)
	if strings.HasSuffix(lower, ".png") {
		return "image/png"
	}
	if strings.HasSuffix(lower, ".jpg") || strings.HasSuffix(lower, ".jpeg") {
		return "image/jpeg"
	}
	if strings.HasSuffix(lower, ".webp") {
		return "image/webp"
	}
	if strings.HasSuffix(lower, ".mp3") || strings.HasSuffix(lower, ".mpeg") {
		return "audio/mpeg"
	}
	if strings.HasSuffix(lower, ".wav") {
		return "audio/wav"
	}
	if strings.HasSuffix(lower, ".m4a") || strings.HasSuffix(lower, ".mp4") {
		return "audio/mp4"
	}
	if strings.HasSuffix(lower, ".webm") {
		return "audio/webm"
	}
	if strings.HasSuffix(lower, ".flac") {
		return "audio/flac"
	}
	if strings.HasSuffix(lower, ".ogg") {
		return "audio/ogg"
	}
	return http.DetectContentType(nil)
}
