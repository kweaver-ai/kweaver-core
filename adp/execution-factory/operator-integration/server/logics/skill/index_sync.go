package skill

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/drivenadapters"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/config"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces/model"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
)

const (
	executionFactoryCatalogID     = "kweaver_execution_factory_catalog"
	executionFactoryCatalogDesc   = "执行工厂的逻辑命名空间"
	executionFactorySkillDataset  = "kweaver_execution_factory_skill_dataset"
	executionFactoryDatasetDesc   = "执行工厂的Skill索引数据集"
	executionFactoryDatasetStatus = "active"
)

type skillIndexSync struct {
	modelManager interfaces.MFModelManager
	modelAPI     interfaces.MFModelAPIClient
	vegaClient   interfaces.VegaBackendClient
	logger       interfaces.Logger
	mu           sync.RWMutex
	initialized  bool
	retryOnce    sync.Once
}

var (
	ssOnce     = sync.Once{}
	ssInstance *skillIndexSync
)

func NewSkillIndexSyncService() interfaces.SkillIndexSyncService {
	ssOnce.Do(func() {
		conf := config.NewConfigLoader()
		ssInstance = &skillIndexSync{
			modelManager: drivenadapters.NewMFModelManager(),
			modelAPI:     drivenadapters.NewMFModelAPIClient(),
			vegaClient:   drivenadapters.NewVegaBackendClient(),
			logger:       conf.GetLogger(),
		}
	})
	return ssInstance
}

func (s *skillIndexSync) EnsureInitialized(ctx context.Context) error {
	if err := s.Init(ctx); err != nil {
		s.retryOnce.Do(func() {
			go s.retryInit()
		})
		return err
	}
	return nil
}

// EnsureDataset 确保Skill索引数据集存在
// 如果不存在，则创建
// 如果存在，则检查是否为最新版本
// 如果不是最新版本，则更新
// 如果是最新版本，则返回成功
func (s *skillIndexSync) Init(ctx context.Context) (err error) {
	// 记录可观测
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)

	initialized := false
	defer func() {
		s.setInitialized(initialized)
	}()
	s.logger.WithContext(ctx).Infof("init skill index dataset, catalog_id=%s, resource_id=%s", executionFactoryCatalogID, executionFactorySkillDataset)
	catalog, err := s.vegaClient.GetCatalogByID(ctx, executionFactoryCatalogID)
	if err != nil {
		s.logger.WithContext(ctx).Errorf("get catalog failed during ensure dataset, catalog_id=%s, err=%v", executionFactoryCatalogID, err)
		return err
	}
	if catalog == nil {
		s.logger.WithContext(ctx).Infof("catalog not found, creating catalog, catalog_id=%s", executionFactoryCatalogID)
		_, err = s.vegaClient.CreateCatalog(ctx, &interfaces.VegaCatalogRequest{
			ID:          executionFactoryCatalogID,
			Name:        executionFactoryCatalogID,
			Tags:        []string{"execution-factory", "索引"},
			Description: executionFactoryCatalogDesc,
		})
		if err != nil {
			s.logger.WithContext(ctx).Errorf("create catalog failed, catalog_id=%s, err=%v", executionFactoryCatalogID, err)
			return err
		}
	}

	resource, err := s.vegaClient.GetResourceByID(ctx, executionFactorySkillDataset)
	if err != nil {
		s.logger.WithContext(ctx).Errorf("get resource failed during ensure dataset, resource_id=%s, err=%v", executionFactorySkillDataset, err)
		return err
	}
	if resource != nil {
		initialized = true
		s.logger.WithContext(ctx).Infof("resource already exists, resource_id=%s", executionFactorySkillDataset)
		return nil
	}
	// 获取嵌入模型
	embeddingModel, err := s.modelManager.GetEmbeddingModel(ctx, interfaces.SmallModelTypeEmbedding, interfaces.SmallModelTypeEmbedding)
	if err != nil {
		s.logger.WithContext(ctx).Errorf("get embedding model failed, resource_id=%s, err=%v", executionFactorySkillDataset, err)
		return err
	}
	s.logger.WithContext(ctx).Infof("creating skill dataset resource, resource_id=%s, dimension=%d", executionFactorySkillDataset, embeddingModel.EmbeddingDim)
	_, err = s.vegaClient.CreateResource(ctx, &interfaces.VegaResourceRequest{
		ID:               executionFactorySkillDataset,
		CatalogID:        executionFactoryCatalogID,
		Name:             executionFactorySkillDataset,
		Tags:             []string{"execution-factory", "skill", "索引"},
		Description:      executionFactoryDatasetDesc,
		Category:         "dataset",
		Status:           executionFactoryDatasetStatus,
		SourceIdentifier: executionFactorySkillDataset,
		SchemaDefinition: buildSkillIndexSchema(embeddingModel.EmbeddingDim),
	})
	if err != nil {
		s.logger.WithContext(ctx).Errorf("create skill dataset resource failed, resource_id=%s, err=%v", executionFactorySkillDataset, err)
		return err
	}
	initialized = true
	return nil
}

func (s *skillIndexSync) UpsertSkill(ctx context.Context, skill *model.SkillRepositoryDB) error {
	log := s.logger
	if !s.isInitialized() {
		log.WithContext(ctx).Warnf("skip skill index upsert because dataset is not initialized, skill_id=%s", skill.SkillID)
		return nil
	}
	document, err := s.buildSkillDocument(ctx, skill)
	if err != nil {
		log.Errorf("build skill index document failed, skill_id=%s, err=%v", skill.SkillID, err)
		return err
	}
	log.Infof("upsert skill index document, skill_id=%s, resource_id=%s", skill.SkillID, executionFactorySkillDataset)
	if err = s.vegaClient.WriteDatasetDocuments(ctx, executionFactorySkillDataset, []map[string]any{document}); err != nil {
		log.Errorf("write skill index document failed, skill_id=%s, err=%v", skill.SkillID, err)
		return err
	}
	return nil
}

func (s *skillIndexSync) UpdateSkill(ctx context.Context, skill *model.SkillRepositoryDB) error {
	log := s.logger
	if !s.isInitialized() {
		log.WithContext(ctx).Warnf("skip skill index update because dataset is not initialized, skill_id=%s", skill.SkillID)
		return nil
	}
	document, err := s.buildSkillDocument(ctx, skill)
	if err != nil {
		log.Errorf("build skill index document failed, skill_id=%s, err=%v", skill.SkillID, err)
		return err
	}
	log.Infof("update skill index document, skill_id=%s, resource_id=%s", skill.SkillID, executionFactorySkillDataset)
	if err = s.vegaClient.UpdateDatasetDocuments(ctx, executionFactorySkillDataset, []map[string]any{document}); err != nil {
		log.Errorf("update skill index document failed, skill_id=%s, err=%v", skill.SkillID, err)
		return err
	}
	return nil
}

func (s *skillIndexSync) DeleteSkill(ctx context.Context, skillID string) error {
	if !s.isInitialized() {
		s.logger.WithContext(ctx).Warnf("skip skill index delete because dataset is not initialized, skill_id=%s", skillID)
		return nil
	}
	s.logger.Infof("delete skill index document, skill_id=%s, resource_id=%s", skillID, executionFactorySkillDataset)
	if err := s.vegaClient.DeleteDatasetDocumentByID(ctx, executionFactorySkillDataset, skillID); err != nil {
		s.logger.WithContext(ctx).Errorf("delete skill index document failed, skill_id=%s, err=%v", skillID, err)
		return err
	}
	return nil
}

func (s *skillIndexSync) isInitialized() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.initialized
}

func (s *skillIndexSync) setInitialized(initialized bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.initialized = initialized
}

func (s *skillIndexSync) retryInit() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	s.logger.Warn("skill index sync service init retry loop started")
	for range ticker.C {
		if err := s.Init(context.Background()); err != nil {
			s.logger.Warnf("retry init skill index sync service failed, error: %v", err)
			continue
		}
		s.logger.Info("skill index sync service init retry succeeded")
		return
	}
}

func (s *skillIndexSync) buildSkillDocument(ctx context.Context, skill *model.SkillRepositoryDB) (map[string]any, error) {
	log := s.logger
	log.Infof("build skill index document, skill_id=%s", skill.SkillID)
	embeddingResp, err := s.modelAPI.Embeddings(ctx, &interfaces.EmbeddingReq{
		Model: interfaces.SmallModelTypeEmbedding,
		Input: []string{buildEmbeddingInput(skill.Name, skill.Description)},
	})
	if err != nil {
		log.Errorf("get skill embedding failed, skill_id=%s, err=%v", skill.SkillID, err)
		return nil, err
	}
	if embeddingResp == nil || len(embeddingResp.Data) == 0 || len(embeddingResp.Data[0].Embedding) == 0 {
		log.Errorf("empty skill embedding result, skill_id=%s", skill.SkillID)
		return nil, fmt.Errorf("embedding result is empty")
	}

	return map[string]any{
		"_id":         skill.SkillID,
		"id":          skill.SkillID,
		"skill_id":    skill.SkillID,
		"name":        skill.Name,
		"description": skill.Description,
		"version":     skill.Version,
		"category":    skill.Category,
		"create_user": skill.CreateUser,
		"create_time": skill.CreateTime,
		"update_user": skill.UpdateUser,
		"update_time": skill.UpdateTime,
		"_vector":     embeddingResp.Data[0].Embedding,
	}, nil
}

func buildEmbeddingInput(name string, description string) string {
	parts := []string{name, description}
	return strings.Join(parts, "\n")
}

func buildSkillIndexSchema(dimension int) []interfaces.VegaProperty {
	return []interfaces.VegaProperty{
		{
			Name:         "skill_id",
			Type:         "string",
			DisplayName:  "skill_id",
			OriginalName: "skill_id",
			Description:  "Skill 业务主键",
			Features: []interfaces.VegaPropertyFeature{{
				Name:        "keyword_skill_id",
				DisplayName: "keyword_skill_id",
				FeatureType: "keyword",
				Description: "Skill ID 精确过滤",
				RefProperty: "skill_id",
				IsDefault:   true,
				IsNative:    false,
				Config:      map[string]any{"ignore_above": 1024},
			}},
		},
		{
			Name:         "name",
			Type:         "text",
			DisplayName:  "name",
			OriginalName: "name",
			Description:  "Skill 名称",
			Features: []interfaces.VegaPropertyFeature{
				{
					Name:        "keyword_name",
					DisplayName: "keyword_name",
					FeatureType: "keyword",
					Description: "Skill 名称的关键词特征",
					RefProperty: "name",
					IsDefault:   true,
					IsNative:    false,
					Config:      map[string]any{"ignore_above": 1024},
				},
				{
					Name:        "fulltext_name",
					DisplayName: "fulltext_name",
					FeatureType: "fulltext",
					Description: "Skill 名称全文检索",
					RefProperty: "name",
					IsDefault:   true,
					IsNative:    false,
					Config:      map[string]any{"analyzer": "standard"},
				}},
		},
		{
			Name:         "description",
			Type:         "text",
			DisplayName:  "description",
			OriginalName: "description",
			Description:  "Skill 描述",
			Features: []interfaces.VegaPropertyFeature{
				{
					Name:        "keyword_description",
					DisplayName: "keyword_description",
					FeatureType: "keyword",
					Description: "Skill 描述的关键词特征",
					RefProperty: "description",
					IsDefault:   true,
					IsNative:    false,
					Config:      map[string]any{"ignore_above": 1024},
				},
				{
					Name:        "fulltext_description",
					DisplayName: "fulltext_description",
					FeatureType: "fulltext",
					Description: "Skill 描述全文检索",
					RefProperty: "description",
					IsDefault:   true,
					IsNative:    false,
					Config:      map[string]any{"analyzer": "standard"},
				}},
		},
		{
			Name:         "version",
			Type:         "string",
			DisplayName:  "version",
			OriginalName: "version",
			Description:  "Skill 版本",
			Features: []interfaces.VegaPropertyFeature{{
				Name:        "keyword_version",
				DisplayName: "keyword_version",
				FeatureType: "keyword",
				Description: "Skill 版本精确过滤",
				RefProperty: "version",
				IsDefault:   true,
				IsNative:    false,
				Config:      map[string]any{"ignore_above": 1024},
			}},
		},
		{
			Name:         "category",
			Type:         "string",
			DisplayName:  "category",
			OriginalName: "category",
			Description:  "Skill 分类",
			Features: []interfaces.VegaPropertyFeature{{
				Name:        "keyword_category",
				DisplayName: "keyword_category",
				FeatureType: "keyword",
				Description: "Skill 分类精确过滤",
				RefProperty: "category",
				IsDefault:   true,
				IsNative:    false,
				Config:      map[string]any{"ignore_above": 1024},
			}},
		},
		{
			Name:         "create_user",
			Type:         "string",
			DisplayName:  "create_user",
			OriginalName: "create_user",
			Description:  "创建人",
			Features: []interfaces.VegaPropertyFeature{{
				Name:        "keyword_create_user",
				DisplayName: "keyword_create_user",
				FeatureType: "keyword",
				Description: "创建人精确过滤",
				RefProperty: "create_user",
				IsDefault:   true,
				IsNative:    false,
				Config:      map[string]any{"ignore_above": 1024},
			}},
		},
		{
			Name:         "create_time",
			Type:         "datetime",
			DisplayName:  "create_time",
			OriginalName: "create_time",
			Description:  "创建时间",
		},
		{
			Name:         "update_user",
			Type:         "string",
			DisplayName:  "update_user",
			OriginalName: "update_user",
			Description:  "更新人",
			Features: []interfaces.VegaPropertyFeature{{
				Name:        "keyword_update_user",
				DisplayName: "keyword_update_user",
				FeatureType: "keyword",
				Description: "更新人精确过滤",
				RefProperty: "update_user",
				IsDefault:   true,
				IsNative:    false,
				Config:      map[string]any{"ignore_above": 1024},
			}},
		},
		{
			Name:         "update_time",
			Type:         "datetime",
			DisplayName:  "update_time",
			OriginalName: "update_time",
			Description:  "更新时间",
		},
		{
			Name:         "_vector",
			Type:         "vector",
			DisplayName:  "_vector",
			OriginalName: "_vector",
			Description:  "Skill 名称与描述向量",
			Features: []interfaces.VegaPropertyFeature{{
				Name:        "vector_skill",
				DisplayName: "vector_skill",
				FeatureType: "vector",
				Description: "Skill 语义检索向量",
				RefProperty: "_vector",
				IsDefault:   true,
				IsNative:    false,
				Config: map[string]any{
					"dimension": dimension,
					"method": map[string]any{
						"name":       "hnsw",
						"space_type": "cosinesimil",
						"engine":     "lucene",
						"parameters": map[string]any{
							"ef_construction": 256,
							"m":               48,
						},
					},
				},
			}},
		},
	}
}
