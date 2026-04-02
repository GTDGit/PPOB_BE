package handler

import "github.com/gin-gonic/gin"

func getGerbangSignature(c *gin.Context) string {
	for _, header := range []string{"X-GTD-Signature", "X-Callback-Signature", "X-Signature"} {
		if value := c.GetHeader(header); value != "" {
			return value
		}
	}
	return ""
}

func getWhatsAppSignature(c *gin.Context) string {
	return c.GetHeader("X-Hub-Signature-256")
}
