package category

import (
	"context"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/dbaccess"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/cache"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/config"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/validator"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces/model"
)

// categoryManager 分类管理器
type categoryManager struct {
	logger     interfaces.Logger
	DBTx       model.DBTx
	DBCategory model.DBCategory
	Validator  interfaces.Validator
	Cache      interfaces.Cache
}

// NewCategoryManager 创建分类管理器
func NewCategoryManager() interfaces.CategoryManager {
	c := &categoryManager{
		logger:     config.NewConfigLoader().GetLogger(),
		DBTx:       dbaccess.NewBaseTx(),
		DBCategory: dbaccess.NewCategoryDBSingleton(),
		Validator:  validator.NewValidator(),
		Cache:      cache.NewInMemoryCache(),
	}
	// 从数据库中加载分类信息到缓存中
	categoryDBList, err := c.DBCategory.SelectList(context.Background(), nil)
	if err != nil {
		c.logger.Errorf("load category from db failed, err: %v", err)
		return nil
	}
	for _, categoryDB := range categoryDBList {
		c.Cache.Set(categoryDB.CategoryID, categoryDB.CategoryName)
	}
	return c
}
