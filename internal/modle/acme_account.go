package modle

import "time"

type AcmeAccountRegister struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
	Code     string `json:"code" validate:"required"`
}

type AcmeAccountLogin struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type AcmeAccount struct {
	ID         int64      `json:"id,string" gorm:"column:id;primary_key;not null"`
	Username   string     `json:"username" gorm:"column:username;unique;not null"`
	Password   string     `gorm:"column:password;not null"`
	PrivateKey string     `gorm:"column:private_key;not null"`
	Remark     string     `json:"remark" gorm:"column:remark;default:''"`
	CreateTime *time.Time `json:"createTime" gorm:"column:create_time;default:CURRENT_TIMESTAMP"`
	UpdateTime *time.Time `json:"updateTime" gorm:"column:update_time;default:CURRENT_TIMESTAMP"`
}

func (AcmeAccount) TableName() string {
	return "acme_account"
}
