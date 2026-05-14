// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package resource

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"

	"vega-backend/common"
	"vega-backend/drivenadapters/entityextension"
	"vega-backend/interfaces"
)

func resourceExtCol(params interfaces.ResourcesQueryParams, col string) string {
	if len(params.ExtensionKeys) > 0 {
		return "t_resource." + col
	}
	return col
}

func applyResourceExtensionJoins(builder sq.SelectBuilder, params interfaces.ResourcesQueryParams) sq.SelectBuilder {
	if len(params.ExtensionKeys) == 0 {
		return builder
	}
	return entityextension.ApplyJoinsForResource(builder, params.ExtensionKeys, params.ExtensionValues)
}

func resourceListOrderExpr(params interfaces.ResourcesQueryParams) string {
	col := params.Sort
	if col == "" {
		col = "f_update_time"
	}
	prefix := ""
	if len(params.ExtensionKeys) > 0 {
		prefix = "t_resource."
	}
	return fmt.Sprintf("%s%s %s", prefix, col, params.Direction)
}

func attachResourceExtensions(ctx context.Context, app *common.AppSetting, params interfaces.ResourcesQueryParams, resources []*interfaces.Resource) error {
	if len(resources) == 0 {
		return nil
	}
	if !params.IncludeExtensions {
		for _, r := range resources {
			r.Extensions = nil
		}
		return nil
	}
	ids := make([]string, 0, len(resources))
	for _, r := range resources {
		ids = append(ids, r.ID)
	}
	st := entityextension.NewStore(app)
	m, err := st.GetByEntityIDs(ctx, entityextension.KindResource, ids)
	if err != nil {
		return err
	}
	for _, r := range resources {
		kv := m[r.ID]
		if kv == nil {
			kv = map[string]string{}
		}
		if params.IncludeExtensionKeys != "" {
			kv = entityextension.FilterKeys(kv, params.IncludeExtensionKeys)
		}
		r.Extensions = kv
	}
	return nil
}

func attachSingleResourceExtensions(ctx context.Context, app *common.AppSetting, r *interfaces.Resource) error {
	if r == nil {
		return nil
	}
	st := entityextension.NewStore(app)
	kv, err := st.GetByEntityID(ctx, entityextension.KindResource, r.ID)
	if err != nil {
		return err
	}
	if kv == nil {
		kv = map[string]string{}
	}
	r.Extensions = kv
	return nil
}
