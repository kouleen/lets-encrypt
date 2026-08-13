package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kouleen/lets-encrypt/internal/api"
	"github.com/kouleen/lets-encrypt/internal/middleware"
	"github.com/kouleen/lets-encrypt/static"
)

func Register(r *gin.Engine) {
	r.StaticFS("/static", http.FS(static.FS))
	publicGroup := r.Group("/acme")
	{
		publicGroup.GET("/sendCode", api.SendCode)
		publicGroup.POST("/login", api.Login)
		publicGroup.GET("/exist", api.Exist)
		publicGroup.POST("/register", api.Register)
		publicGroup.GET("/html", WebHtml)
	}
	domainGroup := publicGroup.Group("/domain")
	domainGroup.Use(middleware.Auth())
	{
		domainGroup.GET("/page", api.PageAcme)
		domainGroup.GET("/refresh", api.RefreshAcme)
		domainGroup.GET("/download", api.DownloadAcme)
		domainGroup.POST("/create", api.CreateAcme)
		domainGroup.POST("/put", api.PutAcme)
		domainGroup.PUT("/updateAuto", api.UpdateAuto)
		domainGroup.PUT("/updateNotice", api.UpdateNotice)
		domainGroup.DELETE("/delete", api.DeleteAcme)
	}

}

func WebHtml(c *gin.Context) {
	// 读取内嵌html文件内容
	data, err := static.FS.ReadFile("index.html")
	if err != nil {
		c.String(http.StatusInternalServerError, "页面加载失败: %v", err)
		return
	}
	// 返回html页面，设置Content-Type
	c.Data(http.StatusOK, "text/html; charset=utf-8", data)
}
