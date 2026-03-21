package handlers

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func createTestImage() []byte {
	return []byte{
		0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46,
		0x49, 0x46, 0x00, 0x01, 0x01, 0x00, 0x00, 0x01,
		0x00, 0x01, 0x00, 0x00, 0xFF, 0xD9,
	}
}

func TestUploadAvatar_ValidImage_ReturnsURL(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	
	part, err := writer.CreateFormFile("avatar", "test.jpg")
	assert.NoError(t, err)
	
	_, err = part.Write(createTestImage())
	assert.NoError(t, err)
	
	err = writer.Close()
	assert.NoError(t, err)
	
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	
	c.Request = httptest.NewRequest(http.MethodPost, "/api/users/avatar", body)
	c.Request.Header.Set("Content-Type", writer.FormDataContentType())
	c.Set("user_id", 1)
	
	UploadAvatar(c)
	
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "url")
}

func TestUploadAvatar_InvalidFormat_ReturnsError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	
	part, err := writer.CreateFormFile("avatar", "test.txt")
	assert.NoError(t, err)
	
	_, err = part.Write([]byte("not an image"))
	assert.NoError(t, err)
	
	err = writer.Close()
	assert.NoError(t, err)
	
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	
	c.Request = httptest.NewRequest(http.MethodPost, "/api/users/avatar", body)
	c.Request.Header.Set("Content-Type", writer.FormDataContentType())
	c.Set("user_id", 1)
	
	UploadAvatar(c)
	
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid file type")
}

func TestUploadAvatar_TooLarge_ReturnsError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	
	part, err := writer.CreateFormFile("avatar", "large.jpg")
	assert.NoError(t, err)
	
	largeContent := make([]byte, 3*1024*1024)
	_, err = part.Write(largeContent)
	assert.NoError(t, err)
	
	err = writer.Close()
	assert.NoError(t, err)
	
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	
	c.Request = httptest.NewRequest(http.MethodPost, "/api/users/avatar", body)
	c.Request.Header.Set("Content-Type", writer.FormDataContentType())
	c.Set("user_id", 1)
	
	UploadAvatar(c)
	
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "file too large")
}

func TestUploadAvatar_NoAuth_ReturnsError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	
	part, err := writer.CreateFormFile("avatar", "test.jpg")
	assert.NoError(t, err)
	
	_, err = part.Write([]byte("fake image content"))
	assert.NoError(t, err)
	
	err = writer.Close()
	assert.NoError(t, err)
	
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	
	c.Request = httptest.NewRequest(http.MethodPost, "/api/users/avatar", body)
	c.Request.Header.Set("Content-Type", writer.FormDataContentType())
	
	UploadAvatar(c)
	
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestValidateImageType_JPG_ReturnsTrue(t *testing.T) {
	valid := validateImageType("image/jpeg")
	assert.True(t, valid)
}

func TestValidateImageType_PNG_ReturnsTrue(t *testing.T) {
	valid := validateImageType("image/png")
	assert.True(t, valid)
}

func TestValidateImageType_GIF_ReturnsTrue(t *testing.T) {
	valid := validateImageType("image/gif")
	assert.True(t, valid)
}

func TestValidateImageType_Invalid_ReturnsFalse(t *testing.T) {
	valid := validateImageType("text/plain")
	assert.False(t, valid)
}

func TestGetUploadDir(t *testing.T) {
	dir := getUploadDir()
	assert.NotEmpty(t, dir)
	assert.Contains(t, dir, "uploads")
}

func TestEnsureUploadDir(t *testing.T) {
	testDir := "/tmp/test_uploads_" + "test"
	defer os.RemoveAll(testDir)
	
	err := ensureUploadDir(testDir)
	assert.NoError(t, err)
	
	_, err = os.Stat(testDir)
	assert.NoError(t, err)
}
