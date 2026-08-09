package util

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"log"
	"os"
)

var userKeyMap map[string]crypto.PrivateKey
var UserKey crypto.PrivateKey

func GetUserKeyMap(email string) crypto.PrivateKey {
	key, ok := userKeyMap[email]
	if !ok {
		return nil
	}
	return key
}

func init() {
	const keyPath = "acme_account.key"
	// 优先加载持久化密钥
	userKey, err := LoadRSAKeyFromFile(keyPath)
	if err != nil {
		log.Fatalf("load account key failed: %v", err)
	}
	if userKey != nil {
		log.Println("load account success key to", keyPath)
		UserKey = userKey
		return
	}
	// 文件不存在则生成新密钥并持久化保存
	log.Println("account key not found, generate new RSA key")
	userKey, err = rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatalf("generate rsa key failed: %v", err)
	}
	// 保存到磁盘
	err = SaveRSAKeyToFile(userKey, keyPath)
	if err != nil {
		log.Fatalf("save account key failed: %v", err)
	}
	log.Println("new account key saved to", keyPath)
	UserKey = userKey
}

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
