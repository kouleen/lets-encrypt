package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"time"
)

type CertInfo struct {
	ExpireTime time.Time // 证书过期时间
	RemainDay  int       // 剩余天数
	NeedRenew  bool      // 是否需要续期（剩余≤30 天）
}

// checkRemoteCertExpire 远程检测域名证书 domain 域名
func checkRemoteCertExpire(domain string) (*CertInfo, error) {
	conn, err := tls.Dial("tcp", domain+":443", &tls.Config{InsecureSkipVerify: true})
	if err != nil {
		return nil, err
	}
	defer func(conn *tls.Conn) {
		_ = conn.Close()
	}(conn)
	cert := conn.ConnectionState().PeerCertificates[0]
	now := time.Now()
	localExpire := cert.NotAfter.Local()
	remainDur := localExpire.Sub(now)
	remainDay := int(remainDur.Hours() / 24)
	return &CertInfo{
		ExpireTime: localExpire,
		RemainDay:  remainDay,
		NeedRenew:  remainDay <= 3,
	}, nil
}

// checkLocalCertExpire 读取pem证书文件，判断是否过期 certPemPath pem证书文件地址
func checkLocalCertExpire(certPemPath string) (*CertInfo, error) {
	// 读取证书文件
	data, err := os.ReadFile(certPemPath)
	if err != nil {
		return nil, fmt.Errorf("读取证书失败: %w", err)
	}
	// 解析PEM块
	block, _ := pem.Decode(data)
	if block == nil || block.Type != "CERTIFICATE" {
		return nil, fmt.Errorf("无效的证书pem文件")
	}
	// 解析x509证书
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("解析证书失败: %w", err)
	}
	localExpire := cert.NotAfter.Local()
	now := time.Now()
	remainDur := localExpire.Sub(now)
	remainDay := int(remainDur.Hours() / 24)

	return &CertInfo{
		ExpireTime: localExpire,
		RemainDay:  remainDay,
		NeedRenew:  remainDay <= 30, // 小于等于30天自动续期
	}, nil
}
