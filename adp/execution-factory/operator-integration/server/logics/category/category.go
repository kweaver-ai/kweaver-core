// Package category implements the interfaces.CategoryManager interface.
package category

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/common"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/errors"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/localize"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces/model"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
)

// GetCategoryName 获取分类名称
func (c *categoryManager) GetCategoryName(ctx context.Context, category interfaces.BizCategory) (categoryName string) {
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, nil)
	if category == "" {
		return
	}
	switch category {
	case interfaces.CategoryTypeOther:
		return c.getCategoryOther(ctx).CategoryName
	case interfaces.CategoryTypeSystem:
		return c.getCategorySystem(ctx).CategoryName
	default:
		// 从缓存中获取分类名称
		value, ok := c.Cache.Get(category.String())
		if ok {
			categoryName = value.(string)
			return
		}
		var categoryDB *model.CategoryDB
		categoryDB, err := c.DBCategory.SelectListByCategoryID(ctx, nil, category.String())
		if err != nil {
			c.logger.Errorf("get category name failed, err: %v", err)
			return ""
		}
		if categoryDB == nil {
			return ""
		}
		categoryName = categoryDB.CategoryName
		// 将分类名称存入缓存中
		c.Cache.Set(category.String(), categoryName)
		return
	}
}

// CheckCategory 检查分类是否存在
func (c *categoryManager) CheckCategory(category interfaces.BizCategory) (isExist bool) {
	if category == interfaces.CategoryTypeOther || category == interfaces.CategoryTypeSystem {
		isExist = true
		return
	}
	var categoryDB *model.CategoryDB
	categoryDB, err := c.DBCategory.SelectListByCategoryID(context.Background(), nil, category.String())
	if err != nil {
		c.logger.Errorf("check category failed, err: %v", err)
		return false
	}
	isExist = categoryDB != nil
	return
}

// GetCategoryList
func (c *categoryManager) GetCategoryList(ctx context.Context) (categoryList []*interfaces.CategoryInfo, err error) {
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	categoryDBList, err := c.DBCategory.SelectList(ctx, nil)
	if err != nil {
		return
	}

	// 默认添加内置的"其他"分类
	categoryList = append(categoryList, c.getCategoryOther(ctx), c.getCategorySystem(ctx))

	for _, categoryDB := range categoryDBList {
		categoryList = append(categoryList, &interfaces.CategoryInfo{
			CategoryType: interfaces.BizCategory(categoryDB.CategoryID),
			CategoryName: categoryDB.CategoryName,
		})
	}
	return
}

func (c *categoryManager) getCategoryOther(ctx context.Context) *interfaces.CategoryInfo {
	language := common.GetLanguageByCtx(ctx)
	tr := localize.NewI18nTranslator(language)
	categoryName := tr.Trans("category." + interfaces.CategoryTypeOther.String())
	return &interfaces.CategoryInfo{
		CategoryType: interfaces.CategoryTypeOther,
		CategoryName: categoryName,
	}
}

func (c *categoryManager) getCategorySystem(ctx context.Context) *interfaces.CategoryInfo {
	language := common.GetLanguageByCtx(ctx)
	tr := localize.NewI18nTranslator(language)
	categoryName := tr.Trans("category." + interfaces.CategoryTypeSystem.String())
	return &interfaces.CategoryInfo{
		CategoryType: interfaces.CategoryTypeSystem,
		CategoryName: categoryName,
	}
}

// UpdateCategory 更新分类
func (c *categoryManager) UpdateCategory(ctx context.Context, req *interfaces.UpdateCategoryReq) (resp *interfaces.UpdateCategoryResp, err error) {
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	// 校验分类名称
	err = c.Validator.ValidatorCategoryName(ctx, req.CategoryName)
	if err != nil {
		return
	}
	// 检查默认分类是否存在
	err = c.checkDefaultCategory(ctx, req.CategoryType.String(), req.CategoryName)
	if err != nil {
		return
	}
	// 检查分类是否存在， 类型ID或类型名称不能重复
	categoryList, err := c.DBCategory.SelectListByCategoryIDOrName(ctx, nil, req.CategoryType.String(), req.CategoryName)
	if err != nil {
		return
	}

	if len(categoryList) == 0 {
		// 分类不存在，报错资源不存在
		return nil, errors.NewHTTPError(ctx, http.StatusNotFound, errors.ErrExtCategoryNotFound, "category_type: "+req.CategoryType.String()+" not found")
	} else if len(categoryList) > 1 {
		// 分类存在多个，报错资源名称重复
		return nil, errors.NewHTTPError(ctx, http.StatusBadRequest, errors.ErrExtCategoryNameExist, "name: "+req.CategoryName+"  already exists")
	} else if categoryList[0].CategoryID != req.CategoryType.String() {
		// 分类ID不匹配，报错资源不存在
		return nil, errors.NewHTTPError(ctx, http.StatusNotFound, errors.ErrExtCategoryNotFound, "category_type: "+req.CategoryType.String()+" not found")
	}

	category := &model.CategoryDB{
		CategoryID:   req.CategoryType.String(),
		CategoryName: req.CategoryName,
		UpdateUser:   req.UserID,
	}

	err = c.DBCategory.UpdateByID(ctx, nil, category)
	if err != nil {
		return
	}
	resp = &interfaces.UpdateCategoryResp{
		CategoryType: req.CategoryType,
		CategoryName: req.CategoryName,
	}
	// 更新缓存中的分类名称
	c.Cache.Set(req.CategoryType.String(), req.CategoryName)
	return
}

// CreateCategory 创建分类
func (c *categoryManager) CreateCategory(ctx context.Context, req *interfaces.CreateCategoryReq) (resp *interfaces.CreateCategoryResp, err error) {
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	// 校验分类名称
	err = c.Validator.ValidatorCategoryName(ctx, req.CategoryName)
	if err != nil {
		return
	}
	// 检查默认分类是否存在
	err = c.checkDefaultCategory(ctx, req.CategoryType.String(), req.CategoryName)
	if err != nil {
		return
	}
	// 检查分类是否存在， 类型ID或类型名称不能重复
	err = c.checkDuplicatedCategory(ctx, req.CategoryType.String(), req.CategoryName)
	if err != nil {
		return
	}
	resp, err = c.insertCategory(ctx, nil, req)
	if err != nil {
		return
	}
	// 将分类名称存入缓存中
	c.Cache.Set(req.CategoryType.String(), req.CategoryName)
	return
}

// BatchCreateCategory 批量创建分类
func (c *categoryManager) BatchCreateCategory(ctx context.Context, req []*interfaces.CreateCategoryReq) (resp []*interfaces.CreateCategoryResp, err error) {
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	tx, err := c.DBTx.GetTx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			_ = tx.Commit()
		}
	}()
	resp = make([]*interfaces.CreateCategoryResp, 0, len(req))
	for _, req := range req {
		// 检查分类是否存在， 类型ID或类型名称不能重复, 如果存在，就跳过
		categoryList, err := c.DBCategory.SelectListByCategoryIDOrName(ctx, nil, req.CategoryType.String(), req.CategoryName)
		if err != nil {
			return nil, err
		}
		if len(categoryList) > 0 {
			continue
		}
		respItem, err := c.insertCategory(ctx, tx, req)
		if err != nil {
			return nil, err
		}
		// 将分类名称存入缓存中
		c.Cache.Set(req.CategoryType.String(), req.CategoryName)
		resp = append(resp, respItem)
	}
	return
}

// DeleteCategory 删除分类
func (c *categoryManager) DeleteCategory(ctx context.Context, req *interfaces.DeleteCategoryReq) (err error) {
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	if string(req.CategoryType) == interfaces.CategoryTypeOther.String() {
		return errors.DefaultHTTPError(ctx, http.StatusForbidden, "category_type: "+interfaces.CategoryTypeOther.String()+" is a built-in system category and cannot be deleted")
	}
	var categoryDB *model.CategoryDB
	categoryDB, err = c.DBCategory.SelectListByCategoryID(ctx, nil, string(req.CategoryType))
	if err != nil {
		c.logger.Errorf("[DeleteCategory] select list by category id failed, err: %v", err)
		return err
	}
	if categoryDB == nil {
		return errors.NewHTTPError(ctx, http.StatusNotFound, errors.ErrExtCategoryNotFound, "category_type: "+string(req.CategoryType)+" not found")
	}
	err = c.DBCategory.DeleteByCategoryID(ctx, nil, string(req.CategoryType))
	if err != nil {
		c.logger.Errorf("[DeleteCategory] delete by category id failed, err: %v", err)
		return
	}
	// 删除缓存中的分类名称
	c.Cache.Delete(string(req.CategoryType))
	return
}

func (c *categoryManager) insertCategory(ctx context.Context, tx *sql.Tx, req *interfaces.CreateCategoryReq) (resp *interfaces.CreateCategoryResp, err error) {
	category := &model.CategoryDB{
		CategoryID:   req.CategoryType.String(),
		CategoryName: req.CategoryName,
		CreateUser:   req.UserID,
		CreateTime:   time.Now().UnixNano(),
	}

	categoryID, err := c.DBCategory.Insert(ctx, tx, category)
	if err != nil {
		c.logger.WithContext(ctx).Errorf("[insertCategory] insert failed, err: %v", err)
		return
	}
	resp = &interfaces.CreateCategoryResp{
		CategoryType: interfaces.BizCategory(categoryID),
		CategoryName: req.CategoryName,
	}
	return
}

// checkDuplicatedCategory 检查分类是否存在， 类型ID或类型名称不能重复
func (c *categoryManager) checkDuplicatedCategory(ctx context.Context, categoryID, categoryName string) (err error) {
	categoryList, err := c.DBCategory.SelectListByCategoryIDOrName(ctx, nil, categoryID, categoryName)
	if err != nil {
		return
	}
	if len(categoryList) > 0 {
		for _, categoryItem := range categoryList {
			if categoryItem.CategoryID == categoryID {
				err = errors.NewHTTPError(ctx, http.StatusBadRequest, errors.ErrExtCategoryNameExist, "category_type: "+categoryID+" name already exists")
				return
			}
			if categoryItem.CategoryName == categoryName {
				err = errors.NewHTTPError(ctx, http.StatusBadRequest, errors.ErrExtCategoryNameExist, "name: "+categoryName+" already exists")
				return
			}
		}
	}
	return nil
}

// checkDefaultCategory 检查默认分类是否存在
func (c *categoryManager) checkDefaultCategory(ctx context.Context, categoryID, categoryName string) (err error) {
	otherCategory := c.getCategoryOther(ctx)
	if otherCategory.CategoryType.String() == categoryID || otherCategory.CategoryName == categoryName {
		// 其他分类是系统内置分类，不允许创建
		return errors.DefaultHTTPError(ctx, http.StatusBadRequest, "category_type: "+interfaces.CategoryTypeOther.String()+" is a built-in system category and cannot be created or modified")
	}
	systemCategory := c.getCategorySystem(ctx)
	if systemCategory.CategoryType.String() == categoryID || systemCategory.CategoryName == categoryName {
		// 系统内置分类是系统内置分类，不允许创建
		return errors.DefaultHTTPError(ctx, http.StatusBadRequest, "category_type: "+interfaces.CategoryTypeSystem.String()+" is a built-in system category and cannot be created or modified")
	}
	return nil
}
