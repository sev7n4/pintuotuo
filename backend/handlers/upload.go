package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	maxFileSize       = 5 << 20
	uploadDir         = "uploads"
	allowedExtensions = ".jpg,.jpeg,.png,.gif,.webp"
)

func init() {
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		fmt.Printf("Warning: failed to create upload directory: %v\n", err)
	}
}

func getUploadSubdir(uploadType string) string {
	switch uploadType {
	case "logo":
		return "logos"
	case "license":
		return "licenses"
	case "idcard":
		return "idcards"
	default:
		return "misc"
	}
}

func isValidExtension(ext string) bool {
	ext = strings.ToLower(ext)
	allowed := strings.Split(allowedExtensions, ",")
	for _, a := range allowed {
		if ext == a {
			return true
		}
	}
	return false
}

func UploadFile(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请选择要上传的文件"})
		return
	}
	defer file.Close()

	if header.Size > maxFileSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "文件大小不能超过5MB"})
		return
	}

	ext := filepath.Ext(header.Filename)
	if !isValidExtension(ext) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "不支持的文件格式，仅支持 JPG、PNG、GIF、WEBP"})
		return
	}

	uploadType := c.DefaultQuery("type", "misc")
	subdir := getUploadSubdir(uploadType)

	fullDir := filepath.Join(uploadDir, subdir)
	if mkdirErr := os.MkdirAll(fullDir, 0755); mkdirErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建目录失败"})
		return
	}

	filename := fmt.Sprintf("%s_%s%s", time.Now().Format("20060102150405"), uuid.New().String()[:8], ext)
	filePath := filepath.Join(fullDir, filename)

	dst, createErr := os.Create(filePath)
	if createErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建文件失败"})
		return
	}
	defer dst.Close()

	if _, copyErr := io.Copy(dst, file); copyErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存文件失败"})
		return
	}

	fileURL := fmt.Sprintf("/uploads/%s/%s", subdir, filename)

	c.JSON(http.StatusOK, gin.H{
		"url":      fileURL,
		"filename": filename,
		"size":     header.Size,
	})
}
