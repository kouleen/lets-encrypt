package util

import (
	"context"
	"fmt"
	"net/smtp"
	"os"
)

const subject = "LetsEncrypt证书系统"

func SendMail(ctx context.Context, targetEmail, context string) {
	// 身份认证
	sendEmail := os.Getenv("SEND_EMAIL")
	smtpServerHost := os.Getenv("SMTP_SERVER")
	sendPassword := os.Getenv("SEND_PWD")
	auth := smtp.PlainAuth("", sendEmail, sendPassword, smtpServerHost)
	// 邮件头部拼接
	msg := fmt.Sprintf("To: %s\r\nFrom: %s\r\nSubject: %s\r\nContent-Type: text/html; charset=utf-8\r\n\r\n%s", targetEmail, sendEmail, subject, context)
	// 发送邮件
	smtpPort := os.Getenv("SMTP_PORT")
	if err := smtp.SendMail(smtpServerHost+":"+smtpPort, auth, sendEmail, []string{targetEmail}, []byte(msg)); err != nil {
		fmt.Printf("发送失败: %v\n", err)
		return
	}
}
