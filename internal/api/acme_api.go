package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-acme/lego/v4/log"
	"github.com/kouleen/lets-encrypt/internal/modle"
	"github.com/kouleen/lets-encrypt/internal/service"
	"github.com/kouleen/lets-encrypt/internal/validator"
)

func SendCode(c *gin.Context) {
	_ = c.Request.Context()
	email := c.Query("email")
	if !validator.ValidateEmail(email) {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid email address"})
	}
}

func Register(c *gin.Context) {
	ctx := c.Request.Context()
	var req modle.AcmeAccountRegister
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := validator.ValidateStruct(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := service.CreateAcmeAccount(ctx, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func Login(c *gin.Context) {
	ctx := c.Request.Context()
	log.Infof("Login success %v", ctx)
	c.JSON(http.StatusOK, gin.H{"message": "login success"})
}
