package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"net/smtp"
	"os"
	"strings"
	"time"

	"github.com/kouleen/lets-encrypt/internal/modle"
	"github.com/kouleen/lets-encrypt/internal/repository"
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

func SendCode(ctx context.Context, email string) (any, error) {
	duration := repository.TtlCacheAuth(email)
	if duration > 0 {
		return nil, errors.New("操作太快了，休息会再试吧！")
	}
	var sb strings.Builder
	charLen := big.NewInt(int64(len(chars)))

	for i := 0; i < 6; i++ {
		n, _ := rand.Int(rand.Reader, charLen)
		sb.WriteByte(chars[n.Int64()])
	}
	code := sb.String()
	// 身份认证
	auth := smtp.PlainAuth("", os.Getenv("SEND_EMAIL"), os.Getenv("SEND_PWD"), os.Getenv("SMTP_SERVER"))
	// 邮件头部拼接
	sprintf := fmt.Sprintf(content, code)
	msg := fmt.Sprintf("To: %s\r\nFrom: %s\r\nSubject: %s\r\nContent-Type: text/html; charset=utf-8\r\n\r\n%s",
		email, os.Getenv("SEND_EMAIL"), subject, sprintf)
	// 发送邮件
	if err := smtp.SendMail(os.Getenv("SMTP_SERVER")+":"+os.Getenv("SMTP_PORT"), auth, os.Getenv("SEND_EMAIL"), []string{email}, []byte(msg)); err != nil {
		fmt.Printf("发送失败: %v\n", err)
		return nil, err
	}
	if err := repository.SetCacheAuth(email, "1", time.Duration(600)*time.Second); err != nil {
		return nil, err
	}
	return true, nil
}

func CreateAcmeAccount(ctx context.Context, req *modle.AcmeAccountRegister) (any, error) {
	return true, nil
}
