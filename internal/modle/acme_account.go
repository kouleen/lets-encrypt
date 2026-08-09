package modle

import "time"

type AcmeAccountRegister struct {
	Username   string `json:"username" validate:"required"`
	Password   string `json:"password" validate:"required"`
	Cloudflare string `json:"cloudflare" validate:"required"`
	Code       string `json:"code" validate:"required"`
}

type AcmeAccountLogin struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type AcmeAccount struct {
	ID         int64      `json:"id,string" gorm:"column:id;primary_key;not null"`
	Username   string     `json:"username" gorm:"column:username;unique;not null"`
	Password   string     `json:"password" gorm:"column:password;not null"`
	Cloudflare string     `json:"cloudflare" gorm:"column:cloudflare;not null"`
	PrivateKey string     `json:"privateKey" gorm:"column:private_key;not null"`
	Remark     string     `json:"remark" gorm:"column:remark;default:''"`
	IsDelete   uint8      `json:"isDelete" gorm:"column:is_delete;not null;default:0"`
	CreatedBy  string     `json:"createdBy" gorm:"column:created_by;not null;"`
	UpdatedBy  string     `json:"updatedBy" gorm:"column:updated_by;not null;"`
	CreateTime *time.Time `json:"createTime" gorm:"column:create_time;default:CURRENT_TIMESTAMP"`
	UpdateTime *time.Time `json:"updateTime" gorm:"column:update_time;default:CURRENT_TIMESTAMP"`
}

func (AcmeAccount) TableName() string {
	return "acme_account"
}
