package util

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"
)

// SaveRSAKeyToFile 将RSA私钥写入PEM文件
func SaveRSAKeyToFile(key *rsa.PrivateKey, path string) error {
	// 序列化私钥为DER字节
	derKey := x509.MarshalPKCS1PrivateKey(key)
	// PEM封装
	pemBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: derKey,
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(f)
	return pem.Encode(f, pemBlock)
}

// LoadRSAKeyFromFile 从PEM文件加载RSA私钥，文件不存在返回nil
func LoadRSAKeyFromFile(path string) (*rsa.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // 文件不存在，外部新建密钥
		}
		return nil, err
	}
	block, _ := pem.Decode(data)
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		return nil, errors.New("invalid pem key file")
	}
	priKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return priKey, nil
}
