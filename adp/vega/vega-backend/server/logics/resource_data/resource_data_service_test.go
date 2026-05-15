// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package resource_data

import (
	"reflect"
	"testing"

	"vega-backend/interfaces"
)

func TestPrepareOutputFieldsParams_FiltersUndefinedFields(t *testing.T) {
	rds := &resourceDataService{}
	resource := &interfaces.Resource{
		Category: interfaces.ResourceCategoryTable,
		SchemaDefinition: []*interfaces.Property{
			{Name: "name"},
			{Name: "age"},
		},
	}
	params := &interfaces.ResourceDataQueryParams{
		OutputFields: []string{"name", "missing", "age"},
	}

	rds.prepareOutputFieldsParams(resource, params)

	expected := []string{"name", "age"}
	if !reflect.DeepEqual(params.OutputFields, expected) {
		t.Fatalf("expected output fields %v, got %v", expected, params.OutputFields)
	}
}

func TestPrepareOutputFieldsParams_IndexKeepsScore(t *testing.T) {
	rds := &resourceDataService{}
	resource := &interfaces.Resource{
		Category: interfaces.ResourceCategoryIndex,
		SchemaDefinition: []*interfaces.Property{
			{Name: "name"},
		},
	}
	params := &interfaces.ResourceDataQueryParams{
		OutputFields: []string{"name", "_score", "missing"},
	}

	rds.prepareOutputFieldsParams(resource, params)

	expected := []string{"name", "_score"}
	if !reflect.DeepEqual(params.OutputFields, expected) {
		t.Fatalf("expected output fields %v, got %v", expected, params.OutputFields)
	}
}
