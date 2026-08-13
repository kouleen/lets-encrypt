package service

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/kouleen/lets-encrypt/internal/modle"
	"github.com/kouleen/lets-encrypt/internal/repository"
	"github.com/kouleen/lets-encrypt/pkg/util"
	"github.com/kouleen/lets-encrypt/static"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const chars = "0123456789"

var tpl *template.Template

func init() {
	emailTpl, err := template.ParseFS(static.FS, "code.html")
	if err != nil {
		fmt.Printf("读取模板失败: %v\n", err)
		return
	}
	fmt.Printf("Email template initialized successfully")
	tpl = emailTpl
}

func SendCode(ctx context.Context, username string) (any, error) {
	durationTtl := repository.TtlCacheAuth(username + "_ttl")
	if durationTtl > 0 {
		return nil, errors.New("操作太快了，休息会再试吧！")
	}
	// 记录60秒
	if err := repository.SetCacheAuth(username+"_ttl", "1", time.Duration(60)*time.Second); err != nil {
		return nil, err
	}
	var sb strings.Builder
	charLen := big.NewInt(int64(len(chars)))

	for i := 0; i < 6; i++ {
		n, _ := rand.Int(rand.Reader, charLen)
		sb.WriteByte(chars[n.Int64()])
	}
	code := sb.String()
	go func() {
		var buf bytes.Buffer
		if err := tpl.Execute(&buf, map[string]any{"Captcha": code}); err != nil {
			fmt.Printf("解析模板失败: %v\n", err)
			return
		}
		util.SendMail(ctx, username, buf.String())
		if err := repository.SetCacheAuth(username, code, time.Duration(600)*time.Second); err != nil {
			fmt.Printf("保存失败: %v\n", err)
		}
	}()
	return true, nil
}

func ExistAcmeAccount(ctx context.Context, username string) (any, error) {
	byUsername, err := repository.GetAcmeAccountByUsername(ctx, username)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	return byUsername != nil, nil
}

func CreateAcmeAccount(ctx context.Context, req *modle.AcmeAccountRegister) (any, error) {
	code, err := repository.GetCacheAuth(req.Username)
	if err != nil {
		return nil, err
	}
	if req.Code != code {
		return false, errors.New("验证码不正确！")
	}
	if err := repository.DelCacheAuth(req.Username); err != nil {
		return nil, err
	}
	resp, err := repository.GetAcmeAccountByUsername(ctx, req.Username)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if resp != nil {
		return nil, errors.New("该用户名已存在")
	}
	hashPwd, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	password := string(hashPwd)
	node, err := snowflake.NewNode(1)
	if err != nil {
		return nil, err
	}
	id := node.Generate()
	privateKeyPath := req.Username + ".key"
	userKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	if err = util.SaveRSAKeyToFile(userKey, privateKeyPath); err != nil {
		return nil, err
	}
	acmeAccount := &modle.AcmeAccount{
		ID:         id.Int64(),
		Username:   req.Username,
		Password:   password,
		PrivateKey: privateKeyPath,
		Remark:     "SUCCESS",
	}
	return repository.CreateAcmeAccount(ctx, acmeAccount)
}

func Login(ctx context.Context, req *modle.AcmeAccountLogin) (any, error) {
	resp, err := repository.GetAcmeAccountByUsername(ctx, req.Username)
	if err != nil {
		return nil, err
	}
	if err = bcrypt.CompareHashAndPassword([]byte(resp.Password), []byte(req.Password)); err != nil {
		return nil, errors.New("用户名或密码错误")
	}
	marshal, err := json.Marshal(resp)
	if err != nil {
		return nil, err
	}
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80

	token := fmt.Sprintf("%x-%x-%x-%x-%x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
	if err = repository.SetCacheAuth(token, string(marshal), time.Duration(24)*time.Hour); err != nil {
		return nil, err
	}
	return token, nil
}

func PageAcmeEncrypt(ctx context.Context, req *modle.AcmeEncryptQuery) (any, int64, error) {
	if req.Current < 1 || req.Size < 1 {
		req.Current = 1
		req.Size = 20
	}
	return repository.PageAcmeEncrypt(ctx, req)
}

func ListAcmeEncrypt(ctx context.Context, req *modle.AcmeEncryptQuery) ([]*modle.AcmeEncrypt, error) {
	return repository.ListAcmeEncrypt(ctx, req)
}

func CreateAcmeEncrypt(ctx context.Context, acmeEncrypt *modle.AcmeEncrypt) (any, error) {
	username, ok := ctx.Value("username").(string)
	if !ok {
		return nil, errors.New("invalid Token")
	}
	acmeEncrypt.Username = username
	resp, err := repository.GetAcmeEncryptByDomain(ctx, acmeEncrypt.Domain)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if resp != nil {
		return nil, errors.New("该记录已存在")
	}
	acmeAccount, err := repository.GetAcmeAccountByUsername(ctx, acmeEncrypt.Username)
	if err != nil {
		return nil, err
	}
	privateKey, err := util.LoadRSAKeyFromFile(acmeAccount.PrivateKey)
	if err != nil {
		return nil, err
	}
	acmeUser := &modle.AcmeUser{
		Username:   acmeEncrypt.Username,
		PrivateKey: privateKey,
	}
	node, err := snowflake.NewNode(1)
	if err != nil {
		return nil, err
	}
	id := node.Generate()
	acmeEncrypt.ID = id.Int64()
	encrypt, err := repository.CreateAcmeEncrypt(ctx, acmeEncrypt)
	if err != nil {
		return nil, err
	}
	go generate(acmeUser, encrypt)
	return encrypt, nil
}

func generate(acmeUser *modle.AcmeUser, acmeEncrypt *modle.AcmeEncrypt) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	acmeEncrypt.Remark = "SUCCESS"
	if acmeEncrypt.Encrypt != "" {
		if err := acmeUser.LetsEncryptGenerate(ctx, acmeEncrypt, true); err != nil {
			fmt.Printf("生成证书文件失败: %v\n", err)
			e := uint8(0)
			acmeEncrypt.Status = &e
			acmeEncrypt.Remark = err.Error()
			_, err = repository.UpdateAcmeEncrypt(ctx, acmeEncrypt)
			if err != nil {
				fmt.Printf("保存失败: %v\n", err)
				return
			}
		}
		certInfo, err := util.GetLocalCertExpire(acmeEncrypt.Encrypt+"/"+acmeEncrypt.Domain+"_bundle"+".pem", acmeEncrypt.RemainDay)
		if err != nil {
			fmt.Printf("读取pem证书文件，获取过期时间失败: %v\n", err)
			e := uint8(0)
			acmeEncrypt.Status = &e
			acmeEncrypt.Remark = err.Error()
			_, err = repository.UpdateAcmeEncrypt(ctx, acmeEncrypt)
			if err != nil {
				fmt.Printf("保存失败: %v\n", err)
				return
			}
			return
		}
		acmeEncrypt.ExpireTime = &certInfo.ExpireTime
		if err := repository.ReloadConfig(ctx, "nginx"); err != nil {
			fmt.Printf("刷新证书容器nginx失败: %v\n", err)
			e := uint8(0)
			acmeEncrypt.Status = &e
			acmeEncrypt.Remark = err.Error()
		}
		_, err = repository.UpdateAcmeEncrypt(ctx, acmeEncrypt)
		if err != nil {
			fmt.Printf("保存失败: %v\n", err)
			return
		}
		return
	}
	certInfo, err := util.GetRemoteCertExpire(acmeEncrypt.Domain, acmeEncrypt.RemainDay)
	if err != nil {
		fmt.Printf("远程检测域名证书 domain 域名失败: %v\n", err)
		e := uint8(0)
		acmeEncrypt.Status = &e
		acmeEncrypt.Remark = err.Error()
	} else {
		acmeEncrypt.ExpireTime = &certInfo.ExpireTime
	}
	_, err = repository.UpdateAcmeEncrypt(ctx, acmeEncrypt)
	if err != nil {
		fmt.Printf("保存失败: %v\n", err)
	}
}

func UpdateAcmeEncrypt(ctx context.Context, acmeEncrypt *modle.AcmeEncrypt) (any, error) {
	resp, err := repository.GetAcmeEncryptByDomain(ctx, acmeEncrypt.Domain)
	if err != nil {
		return nil, err
	}
	acmeAccount, err := repository.GetAcmeAccountByUsername(ctx, resp.Username)
	if err != nil {
		return nil, err
	}
	privateKey, err := util.LoadRSAKeyFromFile(acmeAccount.PrivateKey)
	if err != nil {
		return nil, err
	}
	acmeUser := &modle.AcmeUser{
		Username:   resp.Username,
		PrivateKey: privateKey,
	}
	now := time.Now()
	resp.UpdateTime = &now
	resp.Encrypt = acmeEncrypt.Encrypt
	resp.Cipher = acmeEncrypt.Cipher
	resp.RemainDay = acmeEncrypt.RemainDay
	encrypt, err := repository.UpdateAcmeEncrypt(ctx, resp)
	if err != nil {
		return nil, err
	}
	go generate(acmeUser, encrypt)
	return encrypt, nil
}

func UpdateAuto(ctx context.Context, req *modle.AcmeEncryptAuto) (any, error) {
	username, ok := ctx.Value("username").(string)
	if !ok {
		return nil, errors.New("invalid Token")
	}
	resp, err := repository.GetAcmeEncryptByDomain(ctx, req.Domain)
	if err != nil {
		return nil, err
	}
	if resp.Username != username {
		return nil, errors.New("无权操作此记录")
	}
	if err = repository.UpdateAuto(ctx, req.Domain, req.Auto); err != nil {
		return nil, err
	}
	return true, nil
}

func UpdateNotice(ctx context.Context, req *modle.AcmeEncryptNotice) (any, error) {
	username, ok := ctx.Value("username").(string)
	if !ok {
		return nil, errors.New("invalid Token")
	}
	resp, err := repository.GetAcmeEncryptByDomain(ctx, req.Domain)
	if err != nil {
		return nil, err
	}
	if resp.Username != username {
		return nil, errors.New("无权操作此记录")
	}
	if err = repository.UpdateNotice(ctx, req.Domain, req.Notice); err != nil {
		return nil, err
	}
	return true, nil
}

func DeleteAcmeEncrypt(ctx context.Context, domain string) (any, error) {
	username, ok := ctx.Value("username").(string)
	if !ok {
		return nil, errors.New("invalid Token")
	}
	resp, err := repository.GetAcmeEncryptByDomain(ctx, domain)
	if err != nil {
		return nil, err
	}
	if resp.Username != username {
		return nil, errors.New("无权操作此记录")
	}
	if err := repository.DeleteAcmeEncryptByDomain(ctx, domain); err != nil {
		return nil, err
	}
	return true, nil
}

func RefreshAcmeEncrypt(ctx context.Context, domain string) (any, error) {
	resp, err := repository.GetAcmeEncryptByDomain(ctx, domain)
	if err != nil {
		return nil, err
	}
	if resp.Encrypt == "" {
		return nil, errors.New("证书路径为空，刷新失败")
	}
	acmeAccount, err := repository.GetAcmeAccountByUsername(ctx, resp.Username)
	if err != nil {
		return nil, err
	}
	privateKey, err := util.LoadRSAKeyFromFile(acmeAccount.PrivateKey)
	if err != nil {
		return nil, err
	}
	acmeUser := &modle.AcmeUser{
		Username:   resp.Username,
		PrivateKey: privateKey,
	}
	now := time.Now()
	resp.UpdateTime = &now
	encrypt, err := repository.UpdateAcmeEncrypt(ctx, resp)
	if err != nil {
		return nil, err
	}
	go generate(acmeUser, encrypt)
	return resp, nil
}

func DownloadAcmeEncrypt(ctx context.Context, domain string) ([]byte, string, error) {
	username, ok := ctx.Value("username").(string)
	if !ok {
		return nil, "", errors.New("invalid Token")
	}
	resp, err := repository.GetAcmeEncryptByDomain(ctx, domain)
	if err != nil {
		return nil, "", err
	}
	if resp.Username != username {
		return nil, "", errors.New("无权操作此记录")
	}
	if resp.Encrypt == "" {
		return nil, "", errors.New("证书路径为空，无法下载")
	}

	bundlePath := filepath.Join(resp.Encrypt, resp.Domain+"_bundle.pem")
	keyPath := filepath.Join(resp.Encrypt, resp.Domain+".key")

	bundleData, err := os.ReadFile(bundlePath)
	if err != nil {
		return nil, "", fmt.Errorf("读取证书链文件失败: %v", err)
	}
	keyData, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, "", fmt.Errorf("读取私钥文件失败: %v", err)
	}

	zipBuf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(zipBuf)

	fw, err := zipWriter.Create(resp.Domain + "_bundle.pem")
	if err != nil {
		return nil, "", err
	}
	if _, err = fw.Write(bundleData); err != nil {
		return nil, "", err
	}

	fw, err = zipWriter.Create(resp.Domain + ".key")
	if err != nil {
		return nil, "", err
	}
	if _, err = fw.Write(keyData); err != nil {
		return nil, "", err
	}

	if err := zipWriter.Close(); err != nil {
		return nil, "", err
	}

	filename := resp.Domain + "_certs.zip"
	return zipBuf.Bytes(), filename, nil
}
