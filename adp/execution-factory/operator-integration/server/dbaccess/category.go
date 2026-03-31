package dbaccess

import (
	"context"
	"database/sql"
	"sync"
	"time"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/common/ormhelper"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/config"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/db"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces/model"
	"github.com/kweaver-ai/proton-rds-sdk-go/sqlx"
	"github.com/google/uuid"
)

type categoryDB struct {
	dbPool *sqlx.DB
	logger interfaces.Logger
	dbName string
	orm    *ormhelper.DB
}

var (
	categoryOnce sync.Once
	category     model.DBCategory
)

const (
	tbCategory = "t_category"
)

// NewCategoryDBSingleton 创建分类数据库访问对象单例
func NewCategoryDBSingleton() model.DBCategory {
	confLoader := config.NewConfigLoader()
	dbPool := db.NewDBPool()
	dbName := confLoader.GetDBName()
	logger := confLoader.GetLogger()

	categoryOnce.Do(func() {
		orm := ormhelper.New(dbPool, dbName)
		category = &categoryDB{
			dbPool: dbPool,
			logger: logger,
			dbName: dbName,
			orm:    orm,
		}
	})
	return category
}

// Insert 插入分类
func (c *categoryDB) Insert(ctx context.Context, tx *sql.Tx, category *model.CategoryDB) (categoryID string, err error) {
	now := time.Now().UnixNano()

	if category.CategoryID == "" {
		category.CategoryID = uuid.New().String()
	}
	category.CreateTime = now
	category.UpdateTime = now

	orm := c.orm
	if tx != nil {
		orm = c.orm.WithTx(tx)
	}

	// 使用ORM Helper插入数据
	_, err = orm.Insert().Into(tbCategory).Values(map[string]interface{}{
		"f_category_id":   category.CategoryID,
		"f_category_name": category.CategoryName,
		"f_create_user":   category.CreateUser,
		"f_create_time":   category.CreateTime,
		"f_update_user":   category.UpdateUser,
		"f_update_time":   category.UpdateTime,
	}).Execute(ctx)

	if err != nil {
		return "", err
	}

	return category.CategoryID, nil
}

// UpdateByID 更新分类
func (c *categoryDB) UpdateByID(ctx context.Context, tx *sql.Tx, category *model.CategoryDB) error {
	now := time.Now().UnixNano()
	category.UpdateTime = now

	orm := c.orm
	if tx != nil {
		orm = c.orm.WithTx(tx)
	}

	_, err := orm.Update(tbCategory).SetData(map[string]interface{}{
		"f_category_name": category.CategoryName,
		"f_update_user":   category.UpdateUser,
		"f_update_time":   category.UpdateTime,
	}).WhereEq("f_category_id", category.CategoryID).Execute(ctx)

	return err
}

// SelectList 查询分类列表
func (c *categoryDB) SelectList(ctx context.Context, tx *sql.Tx) (categoryList []*model.CategoryDB, err error) {
	orm := c.orm
	if tx != nil {
		orm = c.orm.WithTx(tx)
	}

	query := orm.Select().From(tbCategory).Sort(&ormhelper.SortParams{
		Fields: []ormhelper.SortField{
			{Field: "f_update_time", Order: ormhelper.SortOrderDesc},
		},
	})

	categoryList = []*model.CategoryDB{}
	err = query.Get(ctx, &categoryList)
	return categoryList, err
}

// SelectListByCategoryIDOrName 根据分类ID或名称查询分类列表
func (c *categoryDB) SelectListByCategoryIDOrName(ctx context.Context, tx *sql.Tx, categoryID, categoryName string) (categoryList []*model.CategoryDB, err error) {
	orm := c.orm
	if tx != nil {
		orm = c.orm.WithTx(tx)
	}

	query := orm.Select().From(tbCategory).Or(func(w *ormhelper.WhereBuilder) {
		if categoryID != "" {
			w.Eq("f_category_id", categoryID)
		}
		if categoryName != "" {
			w.Eq("f_category_name", categoryName)
		}
	})

	categoryList = []*model.CategoryDB{}
	err = query.Get(ctx, &categoryList)
	return categoryList, err
}

// SelectListByCategoryID 根据分类ID查询分类列表
func (c *categoryDB) SelectListByCategoryID(ctx context.Context, tx *sql.Tx, categoryID string) (category *model.CategoryDB, err error) {
	orm := c.orm
	if tx != nil {
		orm = c.orm.WithTx(tx)
	}

	category = &model.CategoryDB{}

	query := orm.Select().From(tbCategory).WhereEq("f_category_id", categoryID)
	err = query.First(ctx, category)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return category, nil
}

// DeleteByCategoryID 根据分类ID删除分类
func (c *categoryDB) DeleteByCategoryID(ctx context.Context, tx *sql.Tx, categoryID string) error {
	orm := c.orm
	if tx != nil {
		orm = c.orm.WithTx(tx)
	}

	_, err := orm.Delete().From(tbCategory).WhereEq("f_category_id", categoryID).Execute(ctx)
	return err
}
