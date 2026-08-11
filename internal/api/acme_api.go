package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kouleen/lets-encrypt/internal/modle"
	"github.com/kouleen/lets-encrypt/internal/service"
	"github.com/kouleen/lets-encrypt/internal/validator"
)

func SendCode(c *gin.Context) {
	ctx := c.Request.Context()
	email := c.Query("email")
	if !validator.ValidateEmail(email) {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "timestamp": time.Now().UnixMilli(), "data": "Invalid email address"})
		return
	}
	resp, err := service.SendCode(ctx, email)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "timestamp": time.Now().UnixMilli(), "data": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "timestamp": time.Now().UnixMilli(), "data": resp})
}

func Exist(c *gin.Context) {
	ctx := c.Request.Context()
	username := c.Query("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "timestamp": time.Now().UnixMilli(), "data": "Username is required"})
		return
	}
	resp, err := service.ExistAcmeAccount(ctx, username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "timestamp": time.Now().UnixMilli(), "data": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "timestamp": time.Now().UnixMilli(), "data": resp})
}

func Register(c *gin.Context) {
	ctx := c.Request.Context()
	var req modle.AcmeAccountRegister
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "timestamp": time.Now().UnixMilli(), "data": err.Error()})
		return
	}
	if err := validator.ValidateStruct(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "timestamp": time.Now().UnixMilli(), "data": err.Error()})
		return
	}
	resp, err := service.CreateAcmeAccount(ctx, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "timestamp": time.Now().UnixMilli(), "data": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "timestamp": time.Now().UnixMilli(), "data": resp})
}

func Login(c *gin.Context) {
	ctx := c.Request.Context()
	var req modle.AcmeAccountLogin
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "timestamp": time.Now().UnixMilli(), "data": err.Error()})
		return
	}
	if err := validator.ValidateStruct(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "timestamp": time.Now().UnixMilli(), "data": err.Error()})
		return
	}
	resp, err := service.Login(ctx, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "timestamp": time.Now().UnixMilli(), "data": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "timestamp": time.Now().UnixMilli(), "data": resp})
}

func PageAcme(c *gin.Context) {
	ctx := c.Request.Context()
	var req modle.AcmeEncryptQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "timestamp": time.Now().UnixMilli(), "data": err.Error()})
		return
	}
	list, total, err := service.PageAcmeEncrypt(ctx, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "timestamp": time.Now().UnixMilli(), "data": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "timestamp": time.Now().UnixMilli(), "data": modle.ToPage(list, total)})
}

func CreateAcme(c *gin.Context) {
	ctx := c.Request.Context()
	var req modle.AcmeEncryptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "timestamp": time.Now().UnixMilli(), "data": err.Error()})
		return
	}
	if err := validator.ValidateStruct(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "timestamp": time.Now().UnixMilli(), "data": err.Error()})
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
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "timestamp": time.Now().UnixMilli(), "data": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "timestamp": time.Now().UnixMilli(), "data": resp})
}

func PutAcme(c *gin.Context) {
	ctx := c.Request.Context()
	var req modle.AcmeEncryptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "timestamp": time.Now().UnixMilli(), "data": err.Error()})
		return
	}
	if err := validator.ValidateStruct(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "timestamp": time.Now().UnixMilli(), "data": err.Error()})
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
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "timestamp": time.Now().UnixMilli(), "data": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "timestamp": time.Now().UnixMilli(), "data": resp})
}

func RefreshAcme(c *gin.Context) {
	ctx := c.Request.Context()
	domain := c.Query("domain")

	if domain == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "timestamp": time.Now().UnixMilli(), "data": "Domain is required"})
		return
	}
	resp, err := service.RefreshAcmeEncrypt(ctx, domain)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "timestamp": time.Now().UnixMilli(), "data": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "timestamp": time.Now().UnixMilli(), "data": resp})
}

func DeleteAcme(c *gin.Context) {
	ctx := c.Request.Context()
	domain := c.Query("domain")
	if domain == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "timestamp": time.Now().UnixMilli(), "data": "Domain is required"})
		return
	}
	resp, err := service.DeleteAcmeEncrypt(ctx, domain)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "timestamp": time.Now().UnixMilli(), "data": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "timestamp": time.Now().UnixMilli(), "data": resp})
}
