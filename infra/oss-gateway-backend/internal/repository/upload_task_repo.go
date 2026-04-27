package repository

import (
	"context"
	"oss-gateway/internal/model"
	"time"

	"gorm.io/gorm"
)

type UploadTaskRepository interface {
	Create(ctx context.Context, task *model.MultipartUploadTask) error
	Update(ctx context.Context, task *model.MultipartUploadTask) error
	GetByUploadID(ctx context.Context, storageID, objectKey, uploadID string) (*model.MultipartUploadTask, error)
	Delete(ctx context.Context, taskID string) error
	DeleteExpired(ctx context.Context, before time.Time) error
}

type uploadTaskRepository struct {
	db *gorm.DB
}

func NewUploadTaskRepository(db *gorm.DB) UploadTaskRepository {
	return &uploadTaskRepository{db: db}
}

func (r *uploadTaskRepository) Create(ctx context.Context, task *model.MultipartUploadTask) error {
	return r.db.WithContext(ctx).Create(task).Error
}

func (r *uploadTaskRepository) Update(ctx context.Context, task *model.MultipartUploadTask) error {
	return r.db.WithContext(ctx).Save(task).Error
}

func (r *uploadTaskRepository) GetByUploadID(ctx context.Context, storageID, objectKey, uploadID string) (*model.MultipartUploadTask, error) {
	var task model.MultipartUploadTask
	err := r.db.WithContext(ctx).
		Where("f_storage_id = ? AND f_object_key = ? AND f_upload_id = ?", storageID, objectKey, uploadID).
		First(&task).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *uploadTaskRepository) Delete(ctx context.Context, taskID string) error {
	return r.db.WithContext(ctx).
		Where("f_id = ?", taskID).
		Delete(&model.MultipartUploadTask{}).Error
}

// DeleteExpired 删除创建时间超过7天且未完成的任务
func (r *uploadTaskRepository) DeleteExpired(ctx context.Context, before time.Time) error {
	return r.db.WithContext(ctx).
		Where("f_created_at < ? AND f_status = ?", before, model.UploadStatusInProgress).
		Delete(&model.MultipartUploadTask{}).Error
}
