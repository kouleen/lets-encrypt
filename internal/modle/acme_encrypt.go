package modle

import "time"

type AcmeEncrypt struct {
	ID         int64      `json:"id,string" gorm:"column:id;primary_key;not null"`
	Username   string     `json:"username" gorm:"column:username;not null"`
	Domain     string     `json:"domain" gorm:"column:domain;not null"`
	ExpireTime *time.Time `json:"expireTime" gorm:"column:expire_time;"`
	Status     uint8      `json:"status" gorm:"column:status;not null;default:1"`
	Remark     string     `json:"remark" gorm:"column:remark;default:''"`
	IsDelete   uint8      `json:"isDelete" gorm:"column:is_delete;not null;default:0"`
	CreatedBy  string     `json:"createdBy" gorm:"column:created_by;not null;"`
	UpdatedBy  string     `json:"updatedBy" gorm:"column:updated_by;not null;"`
	CreateTime *time.Time `json:"createTime" gorm:"column:create_time;default:CURRENT_TIMESTAMP"`
	UpdateTime *time.Time `json:"updateTime" gorm:"column:update_time;default:CURRENT_TIMESTAMP"`
}

func (AcmeEncrypt) TableName() string {
	return "acme_encrypt"
}
