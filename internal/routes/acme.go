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
		publicGroup.GET("/exist", api.Exist)
		publicGroup.POST("/register", api.Register)
	}
	domainGroup := publicGroup.Group("/domain")
	domainGroup.Use(middleware.Auth())
	{
		domainGroup.GET("/page", api.PageAcme)
		domainGroup.GET("/refresh", api.RefreshAcme)
		domainGroup.POST("/create", api.CreateAcme)
		domainGroup.POST("/put", api.PutAcme)
	}

}
