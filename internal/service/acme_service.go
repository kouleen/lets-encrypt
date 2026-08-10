package service

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/smtp"
	"os"
	"strings"
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/kouleen/lets-encrypt/internal/modle"
	"github.com/kouleen/lets-encrypt/internal/repository"
	"github.com/kouleen/lets-encrypt/pkg/util"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const (
	chars   = "0123456789"
	subject = "LetsEncrypt证书系统"
	content = `<!doctype html>
<html>
<head>
<meta charset="utf-8">
<meta http-equiv="X-UA-Compatible" content="IE=edge">
<meta name="description" content="email code">
<meta name="viewport" content="width=device-width, initial-scale=1">
</head>
<body>
<div style="background-color:#ECECEC; padding: 35px;">
<table cellpadding="0" align="center" style="width: 800px;height: 100%%; margin: 0px auto; text-align: left; position: relative; border-top-left-radius: 5px; border-top-right-radius: 5px; border-bottom-right-radius: 5px; border-bottom-left-radius: 5px; font-size: 14px; font-family:微软雅黑, 黑体; line-height: 1.5; box-shadow: rgb(153, 153, 153) 0px 0px 5px; border-collapse: collapse; background-position: initial initial; background-repeat: initial initial;background:#fff;">
<tbody>
<tr>
<th valign="middle" style="height: 25px; line-height: 25px; padding: 15px 35px; border-bottom-width: 1px; border-bottom-style: solid; border-bottom-color: RGB(148,0,211); background-color: RGB(148,0,211); border-top-left-radius: 5px; border-top-right-radius: 5px; border-bottom-right-radius: 0px; border-bottom-left-radius: 0px;"><font id="sentence" face="微软雅黑" size="5" style="color: rgb(255, 255, 255); ">每日一句 Happy every day </font></th>
</tr>
<tr>
<td style="word-break:break-all">
<div style="padding:25px 35px 40px; background-color:#fff;opacity:0.8;">
<h2 style="margin: 5px 0px; "><font color="#333333" style="line-height: 20px; "> <font style="line-height: 22px; " size="4"> 尊敬的用户：</font> </font></h2>
<p>您好！您正在进行邮箱验证，本次请求的验证码为：<font id="captcha" color="#ff8c00">%s</font>，有效期10分钟，请尽快填写验证码完成验证！</p><br>
<div style="width:100%%;margin:0 auto;">
<div style="padding:10px 10px 0;border-top:1px solid #ccc;color:#747474;margin-bottom:20px;line-height:1.3em;font-size:12px;text-align:right">
<p>此为系统邮件，请勿回复<br> Please do not reply to this system email</p>
</div>
</div>
</div></td>
</tr>
</tbody>
</div>
</div>
</body>
</html>`
)

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
		// 身份认证
		auth := smtp.PlainAuth("", os.Getenv("SEND_EMAIL"), os.Getenv("SEND_PWD"), os.Getenv("SMTP_SERVER"))
		// 邮件头部拼接
		sprintf := fmt.Sprintf(content, code)
		msg := fmt.Sprintf("To: %s\r\nFrom: %s\r\nSubject: %s\r\nContent-Type: text/html; charset=utf-8\r\n\r\n%s", username, os.Getenv("SEND_EMAIL"), subject, sprintf)
		// 发送邮件
		if err := smtp.SendMail(os.Getenv("SMTP_SERVER")+":"+os.Getenv("SMTP_PORT"), auth, os.Getenv("SEND_EMAIL"), []string{username}, []byte(msg)); err != nil {
			fmt.Printf("发送失败: %v\n", err)
			return
		}
		if err := repository.SetCacheAuth(username, code, time.Duration(600)*time.Second); err != nil {
			fmt.Printf("保存失败: %v\n", err)
		}
	}()
	return true, nil
}

func ExistAcmeAccount(ctx context.Context, username string) (any, error) {
	byUsername, err := repository.GetAcmeAccountByUsername(ctx, username)
	if err != nil {
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
	bytes, err := json.Marshal(resp)
	if err != nil {
		return nil, err
	}
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80

	token := fmt.Sprintf("%x-%x-%x-%x-%x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
	if err = repository.SetCacheAuth(token, string(bytes), time.Duration(24)*time.Hour); err != nil {
		return nil, err
	}
	return token, nil
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
	go func() {
		ctxAsync, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		acmeEncrypt.Status = 1
		if acmeEncrypt.Encrypt != "" {
			if err := acmeUser.LetsEncryptGenerate(ctxAsync, acmeEncrypt, false); err != nil {
				acmeEncrypt.Status = 0
				acmeEncrypt.Remark = err.Error()
				_, err = repository.CreateAcmeEncrypt(ctxAsync, acmeEncrypt)
				if err != nil {
					fmt.Printf("保存失败: %v\n", err)
					return
				}
			}
			certInfo, err := util.GetLocalCertExpire(acmeEncrypt.Encrypt+"/"+acmeEncrypt.Domain+".pem", acmeEncrypt.RemainDay)
			if err != nil {
				fmt.Printf("读取pem证书文件，判断是否过期失败: %v\n", err)
				return
			}
			acmeEncrypt.ExpireTime = &certInfo.ExpireTime
			_, err = repository.CreateAcmeEncrypt(ctxAsync, acmeEncrypt)
			if err != nil {
				fmt.Printf("保存失败: %v\n", err)
				return
			}
			return
		}
		certInfo, err := util.GetRemoteCertExpire(acmeEncrypt.Domain, acmeEncrypt.RemainDay)
		if err != nil {
			fmt.Printf("远程检测域名证书 domain 域名失败: %v\n", err)
			return
		}
		acmeEncrypt.ExpireTime = &certInfo.ExpireTime
		_, err = repository.CreateAcmeEncrypt(ctxAsync, acmeEncrypt)
		if err != nil {
			fmt.Printf("保存失败: %v\n", err)
			return
		}
	}()
	return acmeEncrypt, nil
}
