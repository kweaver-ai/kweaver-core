// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package catalog

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"

	"vega-backend/common"
	"vega-backend/drivenadapters/entityextension"
	"vega-backend/interfaces"
)

// catalogExtCol 列表/计数在带 extensions JOIN 时为列名加 t_catalog. 前缀，避免歧义
func catalogExtCol(params interfaces.CatalogsQueryParams, col string) string {
	if len(params.ExtensionKeys) > 0 {
		return "t_catalog." + col
	}
	return col
}

func applyCatalogExtensionJoins(builder sq.SelectBuilder, params interfaces.CatalogsQueryParams) sq.SelectBuilder {
	if len(params.ExtensionKeys) == 0 {
		return builder
	}
	return entityextension.ApplyJoinsForCatalog(builder, params.ExtensionKeys, params.ExtensionValues)
}

func catalogListOrderExpr(params interfaces.CatalogsQueryParams) string {
	col := params.Sort
	if col == "" {
		col = "f_update_time"
	}
	prefix := ""
	if len(params.ExtensionKeys) > 0 {
		prefix = "t_catalog."
	}
	return fmt.Sprintf("%s%s %s", prefix, col, params.Direction)
}

func attachCatalogExtensions(ctx context.Context, app *common.AppSetting, params interfaces.CatalogsQueryParams, catalogs []*interfaces.Catalog) error {
	if len(catalogs) == 0 {
		return nil
	}
	if !params.IncludeExtensions {
		for _, c := range catalogs {
			c.Extensions = nil
		}
		return nil
	}
	ids := make([]string, 0, len(catalogs))
	for _, c := range catalogs {
		ids = append(ids, c.ID)
	}
	st := entityextension.NewStore(app)
	m, err := st.GetByEntityIDs(ctx, ids)
	if err != nil {
		return err
	}
	for _, c := range catalogs {
		kv := m[c.ID]
		if kv == nil {
			kv = map[string]string{}
		}
		if params.IncludeExtensionKeys != "" {
			kv = entityextension.FilterKeys(kv, params.IncludeExtensionKeys)
		}
		c.Extensions = kv
	}
	return nil
}

func attachSingleCatalogExtensions(ctx context.Context, app *common.AppSetting, c *interfaces.Catalog) error {
	if c == nil {
		return nil
	}
	st := entityextension.NewStore(app)
	kv, err := st.GetByEntityID(ctx, c.ID)
	if err != nil {
		return err
	}
	if kv == nil {
		kv = map[string]string{}
	}
	c.Extensions = kv
	return nil
}
