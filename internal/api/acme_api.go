package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kouleen/lets-encrypt/internal/modle"
	"github.com/kouleen/lets-encrypt/internal/service"
	"github.com/kouleen/lets-encrypt/internal/validator"
)

func SendCode(c *gin.Context) {
	ctx := c.Request.Context()
	email := c.Query("email")
	if !validator.ValidateEmail(email) {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid email address"})
		return
	}
	resp, err := service.SendCode(ctx, email)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}
func Exist(c *gin.Context) {
	ctx := c.Request.Context()
	username := c.Query("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Username is required"})
		return
	}
	resp, err := service.ExistAcmeAccount(ctx, username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
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
	var req modle.AcmeAccountLogin
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := validator.ValidateStruct(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := service.Login(ctx, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"access_token": resp})
}

func PageAcme(c *gin.Context) {
	ctx := c.Request.Context()
	var req modle.AcmeEncryptQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	list, total, err := service.PageAcmeEncrypt(ctx, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, modle.ToPage(list, total))
}

func CreateAcme(c *gin.Context) {
	ctx := c.Request.Context()
	var req modle.AcmeEncryptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := validator.ValidateStruct(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	acmeEncrypt := &modle.AcmeEncrypt{
		Encrypt:   req.Encrypt,
		Cipher:    req.Cipher,
		Domain:    req.Domain,
		RemainDay: req.RemainDay,
	}
	resp, err := service.CreateAcmeEncrypt(ctx, acmeEncrypt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func PutAcme(c *gin.Context) {
	ctx := c.Request.Context()
	var req modle.AcmeEncryptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := validator.ValidateStruct(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	acmeEncrypt := &modle.AcmeEncrypt{
		Encrypt:   req.Encrypt,
		Cipher:    req.Cipher,
		Domain:    req.Domain,
		RemainDay: req.RemainDay,
	}
	resp, err := service.UpdateAcmeEncrypt(ctx, acmeEncrypt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func RefreshAcme(c *gin.Context) {
	ctx := c.Request.Context()
	domain := c.Query("domain")
	if domain == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Domain is required"})
		return
	}
	resp, err := service.RefreshAcmeEncrypt(ctx, domain)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}
