package model

import "time"

const (
	UploadStatusInProgress = 0
	UploadStatusCompleted  = 1
	UploadStatusCancelled  = 2
)

type MultipartUploadTask struct {
	ID         string    `gorm:"column:f_id;type:varchar(50);primaryKey" json:"id"`
	StorageID  string    `gorm:"column:f_storage_id;type:varchar(50);not null" json:"storage_id"`
	ObjectKey  string    `gorm:"column:f_object_key;type:varchar(512);not null" json:"object_key"`
	UploadID   string    `gorm:"column:f_upload_id;type:varchar(256);not null" json:"upload_id"`
	TotalSize  int64     `gorm:"column:f_total_size;type:bigint;not null" json:"total_size"`
	PartSize   int       `gorm:"column:f_part_size;type:int;not null" json:"part_size"`
	TotalParts int       `gorm:"column:f_total_parts;type:int;not null" json:"total_parts"`
	Status     int       `gorm:"column:f_status;type:smallint;default:0" json:"status"`
	CreatedAt  time.Time `gorm:"column:f_created_at;type:datetime(6);autoCreateTime" json:"created_at"`
	ExpiresAt  time.Time `gorm:"column:f_expires_at;type:datetime(6);not null" json:"expires_at"`
}

func (MultipartUploadTask) TableName() string {
	return "t_multipart_upload_task"
}
