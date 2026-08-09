package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/kouleen/lets-encrypt/internal/api"
	"github.com/kouleen/lets-encrypt/internal/middleware"
)

func Register(r *gin.Engine) {
	publicGroup := r.Group("/acme")
	{
		publicGroup.GET("/sendCode", api.SendCode)
		publicGroup.POST("/login", api.Login)
		publicGroup.POST("/register", api.Register)
	}
	privateGroup := publicGroup.Group("/")
	privateGroup.Use(middleware.Auth())
	{
		privateGroup.POST("/")
	}
}
