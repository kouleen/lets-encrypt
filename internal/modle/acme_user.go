package modle

import (
	"context"
	"crypto"
	"log"
	"os"
	"path/filepath"

	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/providers/dns/cloudflare"
	"github.com/go-acme/lego/v4/registration"
)

type Encrypt interface {
	GetDomain() string
	GetCipher() string
	GetEncrypt() string
}
type AcmeUser struct {
	Username     string
	Registration *registration.Resource
	PrivateKey   crypto.PrivateKey
}

func (u *AcmeUser) GetEmail() string                        { return u.Username }
func (u *AcmeUser) GetRegistration() *registration.Resource { return u.Registration }
func (u *AcmeUser) GetPrivateKey() crypto.PrivateKey        { return u.PrivateKey }

func (u *AcmeUser) LetsEncryptGenerate(ctx context.Context, encrypt Encrypt, register bool) error {
	cfg := lego.NewConfig(u)
	// 调试通过后切换生产环境
	cfg.CADirURL = lego.LEDirectoryProduction
	if os.Getenv("ENV") == "dev" {
		// 测试环境（证书不被浏览器信任，调试用）
		cfg.CADirURL = lego.LEDirectoryStaging
	}
	cfg.Certificate.KeyType = certcrypto.RSA2048
	client, err := lego.NewClient(cfg)
	if err != nil {
		return err
	}
	// 1. 配置 Cloudflare DNS 验证
	cfConfig := cloudflare.NewDefaultConfig()
	// 从环境变量读取 Token，也可以直接写 cfConfig.AuthToken = "xxx"
	cfConfig.AuthToken = encrypt.GetCipher()
	dnsProvider, err := cloudflare.NewDNSProviderConfig(cfConfig)
	if err != nil {
		return err
	}
	err = client.Challenge.SetDNS01Provider(dnsProvider)
	if err != nil {
		return err
	}
	if register {
		reg, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
		if err != nil {
			return err
		}
		u.Registration = reg
	} else {
		reg, err := client.Registration.ResolveAccountByKey()
		if err != nil {
			return err
		}
		u.Registration = reg
	}
	request := certificate.ObtainRequest{
		Domains: []string{encrypt.GetDomain()},
		Bundle:  true,
	}
	cert, err := client.Certificate.Obtain(request)
	if err != nil {
		return err
	}
	log.Println("证书申请成功")
	if err := os.MkdirAll(encrypt.GetEncrypt(), 0700); err != nil {
		return err
	}
	// 证书写入安全写法
	certPath := filepath.Join(encrypt.GetEncrypt(), encrypt.GetDomain()+"_bundle"+".pem")
	keyPath := filepath.Join(encrypt.GetEncrypt(), encrypt.GetDomain()+".key")

	if err := safeWriteFile(certPath, cert.Certificate, 0644); err != nil {
		return err
	}
	if err := safeWriteFile(keyPath, cert.PrivateKey, 0600); err != nil {
		return err
	}
	log.Printf("证书已保存到当前目录: %s, %s", encrypt.GetEncrypt()+encrypt.GetDomain()+"_bundle"+".pem", encrypt.GetEncrypt()+encrypt.GetDomain()+".key")
	return nil
}

func safeWriteFile(dst string, data []byte, perm os.FileMode) error {
	tmp := dst + ".tmp"
	// 写入临时文件
	if err := os.WriteFile(tmp, data, perm); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	// 跨平台：如果目标存在先删掉
	if _, err := os.Stat(dst); err == nil {
		if err := os.Remove(dst); err != nil {
			_ = os.Remove(tmp)
			return err
		}
	}
	err := os.Rename(tmp, dst)
	if err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return nil
}
