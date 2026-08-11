package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kouleen/lets-encrypt/internal/modle"
	"github.com/kouleen/lets-encrypt/internal/repository"
)

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			accessToken := c.Query("access_token")
			if accessToken == "" {
				c.JSON(http.StatusUnauthorized, gin.H{"success": false, "timestamp": time.Now().UnixMilli(), "data": "Unauthorized"})
				c.Abort()
				return
			}
			authHeader = "Bearer " + accessToken
		}
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "timestamp": time.Now().UnixMilli(), "data": "Invalid Token"})
			c.Abort()
			return
		}
		tokenString := parts[1]
		userHeaderJson, err := repository.GetCacheAuth(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "timestamp": time.Now().UnixMilli(), "data": err.Error()})
			c.Abort()
			return
		}
		if userHeaderJson == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "timestamp": time.Now().UnixMilli(), "data": "Invalid Token"})
			c.Abort()
			return
		}
		if c.Request.Method == "POST" && c.Request.URL.Path == "/logout" {
			if err := repository.DelCacheAuth(tokenString); err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"success": false, "timestamp": time.Now().UnixMilli(), "data": err.Error()})
				c.Abort()
				return
			}
			c.JSON(http.StatusOK, true)
			c.Abort()
			return
		}
		var acmeAccount modle.AcmeAccount
		if err := json.Unmarshal([]byte(userHeaderJson), &acmeAccount); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "timestamp": time.Now().UnixMilli(), "data": "Unmarshal err"})
			c.Abort()
			return
		}
		ctx := context.WithValue(c.Request.Context(), "username", acmeAccount.Username)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
