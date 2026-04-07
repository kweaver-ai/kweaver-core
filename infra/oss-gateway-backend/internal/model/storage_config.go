package model

import (
	"time"
)

type StorageConfig struct {
	StorageID        string    `gorm:"column:f_storage_id;type:varchar(50);primaryKey" json:"storage_id"`
	StorageName      string    `gorm:"column:f_storage_name;type:varchar(128);not null" json:"storage_name"`
	VendorType       string    `gorm:"column:f_vendor_type;type:varchar(32);not null" json:"vendor_type"`
	Endpoint         string    `gorm:"column:f_endpoint;type:varchar(256);not null" json:"endpoint"`
	BucketName       string    `gorm:"column:f_bucket_name;type:varchar(128);not null" json:"bucket_name"`
	AccessKeyID      string    `gorm:"column:f_access_key_id;type:varchar(256);not null" json:"access_key_id,omitempty"` // 改为 omitempty，缓存时包含
	AccessKey        string    `gorm:"column:f_access_key;type:varchar(512);not null" json:"access_key,omitempty"`       // 改为 omitempty，缓存时包含
	Region           string    `gorm:"column:f_region;type:varchar(64);default:''" json:"region"`
	IsDefault        bool      `gorm:"column:f_is_default;type:int;default:0" json:"is_default"`
	IsEnabled        bool      `gorm:"column:f_is_enabled;type:int;default:1" json:"is_enabled"`
	InternalEndpoint string    `gorm:"column:f_internal_endpoint;type:varchar(256);default:''" json:"internal_endpoint"`
	SiteID           string    `gorm:"column:f_site_id;type:varchar(64);default:''" json:"site_id"`
	CreatedAt        time.Time `gorm:"column:f_created_at;type:datetime(6);autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time `gorm:"column:f_updated_at;type:datetime(6);autoUpdateTime" json:"updated_at"`
}

func (StorageConfig) TableName() string {
	return "t_storage_config"
}
