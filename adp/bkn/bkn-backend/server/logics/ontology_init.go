// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package logics

import (
	"context"
	"errors"
	"fmt"

	"github.com/kweaver-ai/kweaver-go-lib/logger"

	"bkn-backend/common"
	"bkn-backend/interfaces"
)

func Init(ctx context.Context, appSetting *common.AppSetting) error {
	logger.Info("Init BKN Dataset Start")

	var vectorDim int = 768 // default dimension

	// Check if small model is enabled
	if appSetting.ServerSetting.DefaultSmallModelEnabled {
		smallModel, err := MFA.GetDefaultModel(ctx)
		if err != nil {
			logger.Errorf("GetDefaultModel err:%v", err)
			return err
		}
		if smallModel == nil {
			logger.Errorf("GetDefaultModel return nil")
			return errors.New("GetDefaultModel return nil")
		}
		vectorDim = smallModel.EmbeddingDim
		logger.Infof("Small model enabled, vector dimension: %d", vectorDim)
	}

	// Get or create catalog
	catalog, err := VBA.GetCatalogByID(ctx, interfaces.BKN_CATALOG_ID)
	if err != nil {
		logger.Errorf("GetCatalogByID err:%v", err)
		return err
	}

	if catalog == nil {
		// Create catalog
		logger.Infof("Catalog %s not found, creating...", interfaces.BKN_CATALOG_NAME)
		catalogReq := &interfaces.CatalogRequest{
			ID:          interfaces.BKN_CATALOG_ID,
			Name:        interfaces.BKN_CATALOG_NAME,
			Description: "BKN的逻辑命名空间",
			Tags:        []string{"BKN", "概念索引"},
			// ConnectorType: "logical",
			// ConnectorCfg:  make(map[string]any),
		}
		catalog, err = VBA.CreateCatalog(ctx, catalogReq)
		if err != nil {
			logger.Errorf("CreateCatalog err:%v", err)
			return err
		}
		logger.Infof("Catalog %s created successfully, ID: %s", interfaces.BKN_CATALOG_NAME, catalog.ID)
	} else {
		logger.Infof("Catalog %s found, ID: %s", interfaces.BKN_CATALOG_NAME, catalog.ID)
	}

	// Get dataset
	dataset, err := VBA.GetResourceByID(ctx, interfaces.BKN_DATASET_ID)
	if err != nil {
		logger.Errorf("GetResourceByID err:%v", err)
		return err
	}

	// Get schema definition
	expectedSchema := interfaces.GetBKNConceptSchemaDefinition(vectorDim, appSetting.ServerSetting.DefaultSmallModelEnabled)

	if dataset == nil {
		// Create dataset
		logger.Infof("Dataset %s not found, creating...", interfaces.BKN_DATASET_NAME)

		dataset = interfaces.BKN_CONCEPT_DATASET
		dataset.SchemaDefinition = expectedSchema
		err = VBA.CreateResource(ctx, dataset)
		if err != nil {
			logger.Errorf("CreateResource err:%v", err)
			return err
		}
		logger.Infof("Dataset %s created successfully, ID: %s", interfaces.BKN_DATASET_NAME, interfaces.BKN_DATASET_ID)
	} else {
		logger.Infof("Dataset %s found, ID: %s", interfaces.BKN_DATASET_NAME, dataset.ID)
		// Deep compare schema
		if !deepCompareSchemas(expectedSchema, dataset.SchemaDefinition) {
			logger.Infof("Schema mismatch detected, deleting and recreating dataset...")
			// Delete dataset
			err = VBA.DeleteResource(ctx, dataset.ID)
			if err != nil {
				logger.Errorf("DeleteResource err:%v", err)
				return err
			}
			// Create dataset again
			dataset = interfaces.BKN_CONCEPT_DATASET
			dataset.SchemaDefinition = expectedSchema
			err = VBA.CreateResource(ctx, dataset)
			if err != nil {
				logger.Errorf("CreateResource err:%v", err)
				return err
			}
			logger.Infof("Dataset %s recreated successfully, ID: %s", interfaces.BKN_DATASET_NAME, interfaces.BKN_DATASET_ID)
		} else {
			logger.Infof("Schema matches, no need to recreate dataset")
		}
	}

	logger.Info("Init BKN Dataset Success")
	return nil
}

// deepCompareSchemas compares two Property arrays deeply
func deepCompareSchemas(schema1, schema2 []*interfaces.Property) bool {
	if len(schema1) != len(schema2) {
		return false
	}

	// Create a map for schema2 for efficient lookup
	schema2Map := make(map[string]*interfaces.Property)
	for _, prop := range schema2 {
		schema2Map[prop.Name] = prop
	}

	// Compare each property in schema1
	for _, prop1 := range schema1 {
		prop2, exists := schema2Map[prop1.Name]
		if !exists {
			return false
		}

		if !compareProperty(prop1, prop2) {
			return false
		}
	}

	return true
}

// compareProperty compares two Property objects
func compareProperty(p1, p2 *interfaces.Property) bool {
	if p1.Name != p2.Name {
		return false
	}
	if p1.Type != p2.Type {
		return false
	}
	if p1.DisplayName != p2.DisplayName {
		return false
	}
	if p1.Description != p2.Description {
		return false
	}

	// Compare Features
	if len(p1.Features) != len(p2.Features) {
		return false
	}

	// Create a map for p2.Features
	features2Map := make(map[string]*interfaces.PropertyFeature)
	for i := range p2.Features {
		features2Map[p2.Features[i].FeatureName] = &p2.Features[i]
	}

	// Compare each feature in p1.Features
	for _, feat1 := range p1.Features {
		feat2, exists := features2Map[feat1.FeatureName]
		if !exists {
			return false
		}

		if !comparePropertyFeature(&feat1, feat2) {
			return false
		}
	}

	return true
}

// comparePropertyFeature compares two PropertyFeature objects
func comparePropertyFeature(f1, f2 *interfaces.PropertyFeature) bool {
	if f1.FeatureName != f2.FeatureName {
		return false
	}
	if f1.FeatureType != f2.FeatureType {
		return false
	}
	if f1.RefProperty != f2.RefProperty {
		return false
	}
	if f1.IsDefault != f2.IsDefault {
		return false
	}
	if f1.IsNative != f2.IsNative {
		return false
	}

	// Compare Config maps
	if len(f1.Config) != len(f2.Config) {
		return false
	}
	for k, v1 := range f1.Config {
		v2, exists := f2.Config[k]
		if !exists {
			return false
		}
		// Simple comparison - for complex types, may need deeper comparison
		if fmt.Sprintf("%v", v1) != fmt.Sprintf("%v", v2) {
			return false
		}
	}

	return true
}
