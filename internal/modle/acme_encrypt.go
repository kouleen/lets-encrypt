package modle

import "time"

type AcmeEncryptQuery struct {
	Current int    `form:"current"`
	Size    int    `form:"size"`
	Domain  string `form:"domain"`
	Status  *uint8 `form:"status"`
	Auto    *uint8 `form:"auto"`
	Notice  *uint8 `form:"notice"`
}

type AcmeEncryptAuto struct {
	Domain string `json:"domain" validate:"required"` // 证书域名
	Auto   *uint8 `json:"auto" validate:"required"`
}

type AcmeEncryptNotice struct {
	Domain string `json:"domain" validate:"required"`
	Notice *uint8 `json:"notice" validate:"required"`
}

type AcmeEncryptRequest struct {
	Domain    string `json:"domain" validate:"required"`    // 证书域名
	Encrypt   string `json:"encrypt"`                       // 证书路径
	Cipher    string `json:"cipher" validate:"required"`    // 加密API
	RemainDay int    `json:"remainDay" validate:"required"` // 剩余时间
}
type AcmeEncrypt struct {
	ID         int64      `json:"id,string" gorm:"column:id;primary_key;not null"`
	Username   string     `json:"username" gorm:"column:username;not null;index:idx_acme_encrypt_username"` // 用户名称
	Domain     string     `json:"domain" gorm:"column:domain;not null;uniqueIndex:uq_acme_encrypt_domain"`  // 证书域名
	Cipher     string     `json:"-" gorm:"column:cipher;not null"`                                          // 加密API
	Encrypt    string     `json:"encrypt" gorm:"column:encrypt"`                                            // 证书路径
	RemainDay  int        `json:"remainDay" gorm:"column:remain_day;not null"`                              // 续期阈值
	ExpireTime *time.Time `json:"expireTime" gorm:"column:expire_time;"`                                    // 剩余时间
	Status     *uint8     `json:"status" gorm:"column:status;not null;default:1"`                           // 状态
	Auto       *uint8     `json:"auto" gorm:"column:auto;not null;default:0"`                               // 自动配置 默认关闭：0
	Notice     *uint8     `json:"notice" gorm:"column:notice;not null;default:1"`                           // 开启邮件通知 默认开启：1
	Remark     string     `json:"remark" gorm:"column:remark;default:''"`
	CreateTime *time.Time `json:"createTime" gorm:"column:create_time;default:CURRENT_TIMESTAMP"`
	UpdateTime *time.Time `json:"updateTime" gorm:"column:update_time;default:CURRENT_TIMESTAMP"`
}

func (p *AcmeEncrypt) TableName() string {
	return "acme_encrypt"
}

func (p *AcmeEncrypt) GetDomain() string {
	return p.Domain
}

func (p *AcmeEncrypt) GetEncrypt() string {
	return p.Encrypt
}

func (p *AcmeEncrypt) GetCipher() string {
	return p.Cipher
}
