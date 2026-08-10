package main

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"log"
	"os"

	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/providers/dns/cloudflare"
	"github.com/go-acme/lego/v4/registration"
)

type MyUser struct {
	Email        string
	Registration *registration.Resource
	key          crypto.PrivateKey
}

func (u *MyUser) GetEmail() string                        { return u.Email }
func (u *MyUser) GetRegistration() *registration.Resource { return u.Registration }
func (u *MyUser) GetPrivateKey() crypto.PrivateKey        { return u.key }

func letsEncrypt() {
	// ========== 配置 ==========
	email := "kouleen.china@gmail.com"
	domains := []string{"doc.kouleen.cn"}
	outFullchain := "doc.kouleen.cn.pem"
	outPrivkey := "doc.kouleen.cn.key"

	// 1. 生成账户密钥（生产建议持久化保存，不要每次重新生成）
	userKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatal(err)
	}
	user := &MyUser{Email: email, key: userKey}

	// 2. ACME 客户端配置【本地调试先用测试环境，避免触发限流】
	cfg := lego.NewConfig(user)
	// 测试环境（证书不被浏览器信任，调试用）
	cfg.CADirURL = lego.LEDirectoryStaging
	// 调试通过后切换生产环境
	//cfg.CADirURL = lego.LEDirectoryProduction
	cfg.Certificate.KeyType = certcrypto.RSA2048

	client, err := lego.NewClient(cfg)
	if err != nil {
		log.Fatal(err)
	}

	// 3. 配置 Cloudflare DNS 验证
	cfConfig := cloudflare.NewDefaultConfig()
	// 从环境变量读取 Token，也可以直接写 cfConfig.AuthToken = "xxx"
	cfConfig.AuthToken = os.Getenv("CLOUDFLARE_ACCESS_TOKEN")
	dnsProvider, err := cloudflare.NewDNSProviderConfig(cfConfig)
	if err != nil {
		log.Fatalf("初始化DNS验证器失败: %v", err)
	}
	err = client.Challenge.SetDNS01Provider(dnsProvider)
	if err != nil {
		log.Fatal(err)
	}

	// 4. 注册 ACME 账户
	//reg, err := client.Registration.ResolveAccountByKey()
	reg, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	if err != nil {
		log.Printf("账户注册提示（通常是已存在）: %v", err)
	} else {
		user.Registration = reg
		log.Println("ACME 账户注册成功")
	}

	// 5. 申请证书
	request := certificate.ObtainRequest{
		Domains: domains,
		Bundle:  true,
	}
	cert, err := client.Certificate.Obtain(request)
	if err != nil {
		log.Fatalf("证书申请失败: %v", err)
	}
	log.Println("证书申请成功")

	// 6. 保存证书文件
	if err := os.WriteFile(outFullchain, cert.Certificate, 0644); err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile(outPrivkey, cert.PrivateKey, 0600); err != nil {
		log.Fatal(err)
	}
	log.Printf("证书已保存到当前目录: %s, %s", outFullchain, outPrivkey)
}
