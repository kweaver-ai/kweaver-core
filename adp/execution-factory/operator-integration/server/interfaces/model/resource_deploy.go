package model

import (
	"context"
	"database/sql"
)

type ResourceDeployDB struct {
	ID          int64  `json:"f_id" db:"f_id"`
	ResourceID  string `json:"f_resource_id" db:"f_resource_id"`
	Type        string `json:"f_type" db:"f_type"`
	Version     int    `json:"f_version" db:"f_version"`
	Name        string `json:"f_name" db:"f_name"`
	Description string `json:"f_description" db:"f_description"`
	Config      string `json:"f_config" db:"f_config"`
	Status      string `json:"f_status" db:"f_status"`
	CreateUser  string `json:"f_create_user" db:"f_create_user"`
	CreateTime  int64  `json:"f_create_time" db:"f_create_time"`
	UpdateUser  string `json:"f_update_user" db:"f_update_user"`
	UpdateTime  int64  `json:"f_update_time" db:"f_update_time"`
}

type DBResourceDeploy interface {
	Insert(ctx context.Context, tx *sql.Tx, resourceDeploy *ResourceDeployDB) (ID string, err error)
	Update(ctx context.Context, tx *sql.Tx, resourceDeploy *ResourceDeployDB) error
	Delete(ctx context.Context, tx *sql.Tx, resourceVersion int, resourceType, resourceID string) error
	SelectList(ctx context.Context, tx *sql.Tx, resourceDeploy *ResourceDeployDB) (list []*ResourceDeployDB, err error)
	DeleteByResourceID(ctx context.Context, tx *sql.Tx, resourceID string) error
	SelectListByResourceID(ctx context.Context, resourceID string) (list []*ResourceDeployDB, err error)
	Exists(ctx context.Context, resourceID string, version int) (exists bool, err error)
}
