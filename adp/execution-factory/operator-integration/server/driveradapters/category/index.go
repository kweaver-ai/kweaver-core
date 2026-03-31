// Package category 算子分类
package category

import (
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/config"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	lcategory "github.com/kweaver-ai/adp/execution-factory/operator-integration/server/logics/category"
)

type CategoryHandler interface {
	CategoryList(c *gin.Context)
	CategoryUpdate(c *gin.Context)
	CategoryCreate(c *gin.Context)
	CategoryDelete(c *gin.Context)
}

var (
	once sync.Once
	h    CategoryHandler
)

type categoryHandler struct {
	Logger          interfaces.Logger
	CategoryManager interfaces.CategoryManager
}

func NewCategoryHandler() CategoryHandler {
	once.Do(func() {
		confLoader := config.NewConfigLoader()
		handler := &categoryHandler{
			Logger:          confLoader.GetLogger(),
			CategoryManager: lcategory.NewCategoryManager(),
		}
		handler.initData()
		h = handler
	})
	return h
}
