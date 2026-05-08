package services

import (
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func StreamBinaryResponse(c *gin.Context, resp *http.Response, contentType string) error {
	if contentType != "" {
		c.Header("Content-Type", contentType)
	} else if ct := resp.Header.Get("Content-Type"); ct != "" {
		c.Header("Content-Type", ct)
	}

	if cl := resp.Header.Get("Content-Length"); cl != "" {
		c.Header("Content-Length", cl)
	}

	c.Status(resp.StatusCode)

	_, err := io.Copy(c.Writer, resp.Body)
	if err != nil {
		log.Printf("binary stream error: %v", err)
		return err
	}

	return nil
}
