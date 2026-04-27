package metadata

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/dbaccess"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/config"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/errors"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/telemetry"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces/model"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/logics/parsers"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/utils"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
)

// metadataService 统一元数据管理服务
type metadataService struct {
	Logger             interfaces.Logger
	APIMetadataDB      model.IAPIMetadataDB
	FuncMetadataDB     model.IFunctionMetadataDB
	OperatorRegisterDB model.IOperatorRegisterDB
	ParserRegistry     *parsers.Registry
}

var (
	mOnce    sync.Once
	mManager interfaces.IMetadataService
)

// NewMetadataService 创建统一元数据管理模块
func NewMetadataService() interfaces.IMetadataService {
	mOnce.Do(func() {
		mManager = &metadataService{
			Logger:             config.NewConfigLoader().GetLogger(),
			APIMetadataDB:      dbaccess.NewAPIMetadataDB(),
			FuncMetadataDB:     dbaccess.NewFunctionMetadataDB(),
			OperatorRegisterDB: dbaccess.NewOperatorManagerDB(),
			ParserRegistry:     parsers.NewRegistry(),
		}
	})
	return mManager
}

// GetMetadataBySource 根据SourceID、SourceType查询元数据
func (m *metadataService) GetMetadataBySource(ctx context.Context, sourceID string, sourceType model.SourceType) (has bool, metadata interfaces.IMetadataDB, err error) {
	// 记录可观测
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	telemetry.SetSpanAttributes(ctx, map[string]interface{}{
		"source_id":   sourceID,
		"source_type": string(sourceType),
	})
	// 根据SourceType查询元数据
	switch sourceType {
	case model.SourceTypeOpenAPI:
		has, metadata, err = m.APIMetadataDB.SelectByVersion(ctx, sourceID)
	case model.SourceTypeFunction:
		has, metadata, err = m.FuncMetadataDB.SelectByVersion(ctx, sourceID)
	case model.SourceTypeOperator:
		var operatorDB *model.OperatorRegisterDB
		has, operatorDB, err = m.OperatorRegisterDB.SelectByOperatorID(ctx, nil, sourceID)
		if err == nil && has {
			has, metadata, err = m.GetMetadataBySource(ctx, operatorDB.MetadataVersion, model.SourceType(operatorDB.MetadataType))
		}
	default:
		err = fmt.Errorf("unsupported source type: %s", sourceType)
	}
	if err != nil {
		m.Logger.WithContext(ctx).Errorf("get metadata failed, err: %v", err)
		err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
	}
	return
}

func (m *metadataService) BatchGetMetadataBySourceIDs(ctx context.Context, sourceMap map[model.SourceType][]string) (sourceIDToMetadata map[string]interfaces.IMetadataDB, err error) {
	// 记录可观测
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	sourceIDToMetadata = map[string]interfaces.IMetadataDB{}
	if len(sourceMap) == 0 {
		return
	}
	var wg sync.WaitGroup
	// 添加停止标志
	var stopFlag int32
	// 使用线程安全的映射
	resultMutex := sync.Mutex{}
	errorsMutex := sync.Mutex{}
	var errList []error
	for sourceType, sourceIDs := range sourceMap {
		if len(sourceIDs) == 0 {
			continue
		}
		wg.Add(1)
		go func(st model.SourceType, sourceIDList []string) {
			defer wg.Done()

			// 检查是否已经需要停止
			if atomic.LoadInt32(&stopFlag) == 1 {
				return
			}

			var localErr error
			sourceIDList = utils.UniqueStrings(sourceIDList)
			switch st {
			case model.SourceTypeOpenAPI:
				var metadataList []*model.APIMetadataDB
				metadataList, localErr = utils.BatchQueryWithContext(ctx, sourceIDList,
					interfaces.DefaultBatchSize, m.APIMetadataDB.SelectListByVersion)
				if localErr != nil {
					m.Logger.WithContext(ctx).Errorf("batch query api metadata failed, err: %v", localErr)
					localErr = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, localErr.Error())
				} else {
					resultMutex.Lock()
					for _, metadata := range metadataList {
						sourceIDToMetadata[metadata.Version] = metadata
					}
					resultMutex.Unlock()
				}
			case model.SourceTypeFunction:
				var metadataList []*model.FunctionMetadataDB
				metadataList, localErr = utils.BatchQueryWithContext(ctx, sourceIDList,
					interfaces.DefaultBatchSize, m.FuncMetadataDB.SelectListByVersion)
				if localErr != nil {
					m.Logger.WithContext(ctx).Errorf("batch query function metadata failed, err: %v", localErr)
					localErr = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, localErr.Error())
				} else {
					resultMutex.Lock()
					for _, metadata := range metadataList {
						sourceIDToMetadata[metadata.Version] = metadata
					}
					resultMutex.Unlock()
				}
			case model.SourceTypeOperator:
				var operatorList []*model.OperatorRegisterDB
				operatorList, localErr = utils.BatchQueryWithContext(ctx, sourceIDList,
					interfaces.DefaultBatchSize, m.OperatorRegisterDB.SelectByOperatorIDs)
				if localErr != nil {
					m.Logger.WithContext(ctx).Errorf("batch query operator metadata failed, err: %v", localErr)
					localErr = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, localErr.Error())
				} else {
					operatorSourceMap := map[model.SourceType][]string{}
					for _, operator := range operatorList {
						operatorSourceMap[model.SourceType(operator.MetadataType)] = append(operatorSourceMap[model.SourceType(operator.MetadataType)],
							operator.MetadataVersion)
					}
					var operatorSourceIDToMetadata map[string]interfaces.IMetadataDB
					operatorSourceIDToMetadata, localErr = m.BatchGetMetadataBySourceIDs(ctx, operatorSourceMap)
					if localErr != nil {
						m.Logger.WithContext(ctx).Errorf("batch query operator metadata failed, err: %v", localErr)
						localErr = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, localErr.Error())
					} else {
						resultMutex.Lock()
						for _, operatorDB := range operatorList {
							sourceIDToMetadata[operatorDB.OperatorID] = operatorSourceIDToMetadata[operatorDB.MetadataVersion]
						}
						resultMutex.Unlock()
					}
				}
			}
			// 处理错误
			if localErr != nil {
				errorsMutex.Lock()
				errList = append(errList, localErr)
				errorsMutex.Unlock()
				// 设置停止标志，但不取消上下文
				atomic.StoreInt32(&stopFlag, 1)
			}
		}(sourceType, sourceIDs)
	}

	wg.Wait()
	// 处理错误
	if len(errList) > 0 {
		// 返回第一个错误作为主要错误
		err = errList[0]
		if len(errList) > 1 {
			m.Logger.WithContext(ctx).Warnf("multiple errors occurred during batch get metadata: %v", errList)
		}
	}
	return sourceIDToMetadata, err
}

// ParseMetadata 解析元数据
func (m *metadataService) ParseMetadata(ctx context.Context, metadataType interfaces.MetadataType, input any) ([]interfaces.IMetadataDB, error) {
	parser, err := m.ParserRegistry.Get(metadataType)
	if err != nil {
		return nil, err
	}
	return parser.Parse(ctx, input)
}

// ParseRawContent 获取解析后的原始内容
func (m *metadataService) ParseRawContent(ctx context.Context, metadataType interfaces.MetadataType, input any) (content any, err error) {
	parser, err := m.ParserRegistry.Get(metadataType)
	if err != nil {
		return nil, err
	}
	// 解析原始数据为目标结构
	content, err = parser.GetAllContent(ctx, input)
	if err != nil {
		return nil, err
	}
	return
}

// RegisterMetadata 注册单个元数据
func (m *metadataService) RegisterMetadata(ctx context.Context, tx *sql.Tx, metadata interfaces.IMetadataDB) (version string, err error) {
	// 验证元数据
	err = m.ValidateMetadata(ctx, metadata)
	if err != nil {
		return
	}

	// 根据类型存储到对应的表
	switch metadata.GetType() {
	case string(model.SourceTypeOpenAPI):
		apiMetadata, ok := metadata.(*model.APIMetadataDB)
		if !ok {
			err = fmt.Errorf("invalid metadata type for API: %T", metadata)
			return
		}
		if apiMetadata.Version == "" {
			apiMetadata.Version = uuid.New().String()
		}
		version, err = m.APIMetadataDB.InsertAPIMetadata(ctx, tx, apiMetadata)
		if err != nil {
			m.Logger.WithContext(ctx).Errorf("insert API metadata failed, err: %v", err)
			return
		}
	case string(model.SourceTypeFunction):
		funcMetadata, ok := metadata.(*model.FunctionMetadataDB)
		if !ok {
			err = fmt.Errorf("invalid metadata type for Function: %T", metadata)
			return
		}
		if funcMetadata.Version == "" {
			funcMetadata.Version = uuid.New().String()
		}
		funcMetadata.Path = interfaces.SetAOIFuncExecPath(funcMetadata.Version)
		version, err = m.FuncMetadataDB.InsertFuncMetadata(ctx, tx, funcMetadata)
		if err != nil {
			m.Logger.WithContext(ctx).Errorf("insert Function metadata failed, err: %v", err)
			return
		}
	default:
		err = fmt.Errorf("unsupported metadata type: %s", metadata.GetType())
		return
	}
	return
}

// BatchRegisterMetadata 批量注册元数据
func (m *metadataService) BatchRegisterMetadata(ctx context.Context, tx *sql.Tx, metadatas []interfaces.IMetadataDB) (versions []string, err error) {
	if len(metadatas) == 0 {
		return []string{}, nil
	}

	// 按类型分组
	apiMetadatas := make([]*model.APIMetadataDB, 0)
	funcMetadatas := make([]*model.FunctionMetadataDB, 0)

	for _, metadata := range metadatas {
		err = m.ValidateMetadata(ctx, metadata)
		if err != nil {
			return
		}

		switch metadata.GetType() {
		case string(model.SourceTypeOpenAPI):
			apiMetadata, ok := metadata.(*model.APIMetadataDB)
			if !ok {
				err = fmt.Errorf("invalid metadata type for API: %T", metadata)
				return
			}
			if apiMetadata.Version == "" {
				apiMetadata.Version = uuid.New().String()
			}
			apiMetadatas = append(apiMetadatas, apiMetadata)
		case string(model.SourceTypeFunction):
			funcMetadata, ok := metadata.(*model.FunctionMetadataDB)
			if !ok {
				err = fmt.Errorf("invalid metadata type for Function: %T", metadata)
				return
			}
			if funcMetadata.Version == "" {
				funcMetadata.Version = uuid.New().String()
			}
			funcMetadata.Path = interfaces.SetAOIFuncExecPath(funcMetadata.Version)
			funcMetadatas = append(funcMetadatas, funcMetadata)
		default:
			err = fmt.Errorf("unsupported metadata type: %s", metadata.GetType())
			return
		}
	}

	versions = make([]string, 0, len(metadatas))

	// 批量插入API元数据
	if len(apiMetadatas) > 0 {
		apiVersions, err := m.APIMetadataDB.InsertAPIMetadatas(ctx, tx, apiMetadatas)
		if err != nil {
			m.Logger.WithContext(ctx).Errorf("batch insert API metadata failed, err: %v", err)
			return nil, err
		}
		versions = append(versions, apiVersions...)
	}

	// 批量插入Function元数据
	if len(funcMetadatas) > 0 {
		funcVersions, err := m.FuncMetadataDB.InsertFuncMetadatas(ctx, tx, funcMetadatas)
		if err != nil {
			m.Logger.WithContext(ctx).Errorf("batch insert Function metadata failed, err: %v", err)
			return nil, err
		}
		versions = append(versions, funcVersions...)
	}

	return versions, nil
}

// CheckMetadataExists 检查元数据是否存在
func (m *metadataService) CheckMetadataExists(ctx context.Context, metadataType interfaces.MetadataType, version string) (exists bool,
	metadata interfaces.IMetadataDB, err error) {
	switch metadataType {
	case interfaces.MetadataTypeAPI:
		exists, metadata, err = m.APIMetadataDB.SelectByVersion(ctx, version)
		if err != nil {
			err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, fmt.Sprintf("select API metadata by version failed, err: %v", err))
			m.Logger.WithContext(ctx).Errorf("select API metadata by version failed, err: %v", err)
			return
		}
	case interfaces.MetadataTypeFunc:
		exists, metadata, err = m.FuncMetadataDB.SelectByVersion(ctx, version)
		if err != nil {
			err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, fmt.Sprintf("select Function metadata by version failed, err: %v", err))
			m.Logger.WithContext(ctx).Errorf("select Function metadata by version failed, err: %v", err)
			return
		}
	default:
		m.Logger.WithContext(ctx).Warnf("unsupported metadata type: %s", metadataType)
		err = errors.DefaultHTTPError(ctx, http.StatusBadRequest, fmt.Sprintf("unsupported metadata type: %s", metadataType))
	}
	return
}

// GetMetadataByVersion 根据版本查询元数据
func (m *metadataService) GetMetadataByVersion(ctx context.Context, metadataType interfaces.MetadataType, version string) (interfaces.IMetadataDB, error) {
	switch metadataType {
	case interfaces.MetadataTypeAPI:
		exist, metadata, err := m.APIMetadataDB.SelectByVersion(ctx, version)
		if err != nil {
			err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, fmt.Sprintf("select API metadata by version failed, err: %v", err))
			m.Logger.WithContext(ctx).Errorf("select API metadata by version failed, err: %v", err)
			return nil, err
		}
		if !exist {
			return nil, errors.NewHTTPError(ctx, http.StatusNotFound, errors.ErrExtMetadataNotFound, fmt.Sprintf("API metadata version %s not found", version))
		}
		return metadata, nil
	case interfaces.MetadataTypeFunc:
		exist, metadata, err := m.FuncMetadataDB.SelectByVersion(ctx, version)
		if err != nil {
			err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, fmt.Sprintf("select Function metadata by version failed, err: %v", err))
			m.Logger.WithContext(ctx).Errorf("select Function metadata by version failed, err: %v", err)
			return nil, err
		}
		if !exist {
			return nil, errors.NewHTTPError(ctx, http.StatusNotFound, errors.ErrExtMetadataNotFound, fmt.Sprintf("Function metadata version %s not found", version))
		}
		return metadata, nil
	default:
		return nil, errors.DefaultHTTPError(ctx, http.StatusBadRequest, fmt.Sprintf("unsupported metadata type: %s", metadataType))
	}
}

// BatchGetMetadata 批量查询元数据
func (m *metadataService) BatchGetMetadata(ctx context.Context, apiVersions, funcVersions []string) (result []interfaces.IMetadataDB, err error) {
	// 并发查询API元数据
	var apiMetadatas []*model.APIMetadataDB
	var funcMetadatas []*model.FunctionMetadataDB
	var apiErr, funcErr error
	result = []interfaces.IMetadataDB{}
	var wg sync.WaitGroup

	// 查询OpenAPI元数据
	if len(apiVersions) > 0 {
		apiVersions = utils.UniqueStrings(apiVersions)
		wg.Add(1)
		go func() {
			defer wg.Done()
			apiMetadatas, apiErr = utils.BatchQueryWithContext[*model.APIMetadataDB, string](
				ctx, apiVersions, interfaces.DefaultBatchSize, m.APIMetadataDB.SelectListByVersion)
		}()
	}
	// 查询Function元数据
	if len(funcVersions) > 0 {
		funcVersions = utils.UniqueStrings(funcVersions)
		wg.Add(1)
		go func() {
			defer wg.Done()
			funcMetadatas, funcErr = utils.BatchQueryWithContext[*model.FunctionMetadataDB, string](
				ctx, funcVersions, interfaces.DefaultBatchSize, m.FuncMetadataDB.SelectListByVersion)
		}()
	}

	// 等待所有查询完成
	wg.Wait()

	// 错误处理
	if apiErr != nil || funcErr != nil {
		err = fmt.Errorf("batch get metadata failed, apiErr: %v, funcErr: %v", apiErr, funcErr)
		return
	}

	// 合并结果
	for _, metadata := range apiMetadatas {
		result = append(result, metadata)
	}
	for _, metadata := range funcMetadatas {
		result = append(result, metadata)
	}
	return result, nil
}

// UpdateMetadata 更新元数据
func (m *metadataService) UpdateMetadata(ctx context.Context, tx *sql.Tx, metadata interfaces.IMetadataDB) error {
	// 验证元数据
	err := m.ValidateMetadata(ctx, metadata)
	if err != nil {
		return err
	}

	// 根据类型更新对应的表
	switch metadata.GetType() {
	case string(model.SourceTypeOpenAPI):
		apiMetadata, ok := metadata.(*model.APIMetadataDB)
		if !ok {
			return fmt.Errorf("invalid metadata type for API: %T", metadata)
		}
		now := time.Now().UnixNano()
		apiMetadata.UpdateTime = now
		err = m.APIMetadataDB.UpdateByVersion(ctx, tx, apiMetadata.Version, apiMetadata)
		if err != nil {
			m.Logger.WithContext(ctx).Errorf("update API metadata failed, err: %v", err)
			return err
		}
	case string(model.SourceTypeFunction):
		funcMetadata, ok := metadata.(*model.FunctionMetadataDB)
		if !ok {
			return fmt.Errorf("invalid metadata type for Function: %T", metadata)
		}
		now := time.Now().UnixNano()
		funcMetadata.UpdateTime = now
		funcMetadata.Path = interfaces.SetAOIFuncExecPath(funcMetadata.Version)
		err = m.FuncMetadataDB.UpdateByVersion(ctx, tx, funcMetadata)
		if err != nil {
			m.Logger.WithContext(ctx).Errorf("update Function metadata failed, err: %v", err)
			return err
		}
	default:
		return fmt.Errorf("unsupported metadata type: %s", metadata.GetType())
	}
	return nil
}

// DeleteMetadata 删除元数据
func (m *metadataService) DeleteMetadata(ctx context.Context, tx *sql.Tx, metadataType interfaces.MetadataType, version string) error {
	switch metadataType {
	case interfaces.MetadataTypeAPI:
		err := m.APIMetadataDB.DeleteByVersion(ctx, tx, version)
		if err != nil {
			m.Logger.WithContext(ctx).Errorf("delete API metadata failed, err: %v", err)
			return err
		}
	case interfaces.MetadataTypeFunc:
		err := m.FuncMetadataDB.DeleteByVersion(ctx, tx, version)
		if err != nil {
			m.Logger.WithContext(ctx).Errorf("delete Function metadata failed, err: %v", err)
			return err
		}
	default:
		return fmt.Errorf("unsupported metadata type: %s", metadataType)
	}
	return nil
}

// BatchDeleteMetadata 批量删除元数据
func (m *metadataService) BatchDeleteMetadata(ctx context.Context, tx *sql.Tx, metadataType interfaces.MetadataType, versions []string) error {
	if len(versions) == 0 {
		return nil
	}
	versions = utils.UniqueStrings(versions)
	switch metadataType {
	case interfaces.MetadataTypeAPI:
		err := m.APIMetadataDB.DeleteByVersions(ctx, tx, versions)
		if err != nil {
			m.Logger.WithContext(ctx).Errorf("batch delete API metadata failed, err: %v", err)
			err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, fmt.Sprintf("batch delete API metadata failed, err: %v", err))
			return err
		}
	case interfaces.MetadataTypeFunc:
		err := m.FuncMetadataDB.DeleteByVersions(ctx, tx, versions)
		if err != nil {
			m.Logger.WithContext(ctx).Errorf("batch delete Function metadata failed, err: %v", err)
			err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, fmt.Sprintf("batch delete Function metadata failed, err: %v", err))
			return err
		}
	default:
		return errors.DefaultHTTPError(ctx, http.StatusInternalServerError, fmt.Sprintf("unsupported metadata type: %s", metadataType))
	}
	return nil
}

// ValidateMetadata 验证元数据格式
func (m *metadataService) ValidateMetadata(ctx context.Context, metadata interfaces.IMetadataDB) error {
	if metadata == nil {
		return fmt.Errorf("metadata is nil")
	}
	return metadata.Validate(ctx)
}
