package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	apperrors "github.com/pintuotuo/backend/errors"
	"github.com/pintuotuo/backend/middleware"
)

const (
	maxAvatarSize = 2 * 1024 * 1024
	uploadDir     = "uploads/avatars"
)

var allowedImageTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/gif":  true,
}

func UploadAvatar(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}

	file, header, err := c.Request.FormFile("avatar")
	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"INVALID_FILE",
			"No file provided",
			http.StatusBadRequest,
			err,
		))
		return
	}
	defer file.Close()

	if header.Size > maxAvatarSize {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"FILE_TOO_LARGE",
			"file too large: maximum size is 2MB",
			http.StatusBadRequest,
			nil,
		))
		return
	}

	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil && err != io.EOF {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"FILE_READ_ERROR",
			"Failed to read file",
			http.StatusBadRequest,
			err,
		))
		return
	}

	contentType := http.DetectContentType(buffer)
	if !validateImageType(contentType) {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"INVALID_FILE_TYPE",
			"invalid file type: only jpg, png, gif are allowed",
			http.StatusBadRequest,
			nil,
		))
		return
	}

	_, err = file.Seek(0, 0)
	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"FILE_SEEK_ERROR",
			"Failed to process file",
			http.StatusInternalServerError,
			err,
		))
		return
	}

	if err := ensureUploadDir(uploadDir); err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"UPLOAD_DIR_ERROR",
			"Failed to create upload directory",
			http.StatusInternalServerError,
			err,
		))
		return
	}

	ext := filepath.Ext(header.Filename)
	if ext == "" {
		ext = ".jpg"
	}
	filename := fmt.Sprintf("%d%s", userID, ext)
	filepath := filepath.Join(uploadDir, filename)

	dst, err := os.Create(filepath)
	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"FILE_CREATE_ERROR",
			"Failed to create file",
			http.StatusInternalServerError,
			err,
		))
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"FILE_WRITE_ERROR",
			"Failed to save file",
			http.StatusInternalServerError,
			err,
		))
		return
	}

	avatarURL := fmt.Sprintf("/%s/%s", uploadDir, filename)

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"url": avatarURL,
		},
	})
}

func validateImageType(contentType string) bool {
	return allowedImageTypes[contentType]
}

func getUploadDir() string {
	return uploadDir
}

func ensureUploadDir(dir string) error {
	return os.MkdirAll(dir, 0755)
}

func init() {
	if err := ensureUploadDir(uploadDir); err != nil {
		fmt.Printf("Warning: failed to create upload directory: %v\n", err)
	}
}
