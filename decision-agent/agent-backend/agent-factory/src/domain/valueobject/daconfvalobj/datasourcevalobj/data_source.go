package datasourcevalobj

import (
	"context"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/capierr"
	"github.com/pkg/errors"
)

// RetrieverDataSource 表示检索数据源配置
type RetrieverDataSource struct {
	Kg               []*KgSource               `json:"kg"`                // 图谱类型数据源
	Doc              []*DocSource              `json:"doc"`               // 文档类型数据源
	Metric           []*MetricSource           `json:"metric"`            // 指标类型数据源
	KnEntry          []*KnEntrySource          `json:"kn_entry"`          // 知识库类型数据源
	KnowledgeNetwork []*KnowledgeNetworkSource `json:"knowledge_network"` // 业务知识网络类型数据源
	AdvancedConfig   *RetrieverAdvancedConfig  `json:"advanced_config"`   // 召回高级配置
}

func NewRetrieverDataSource() (r *RetrieverDataSource) {
	return &RetrieverDataSource{
		Kg:               []*KgSource{},
		Doc:              []*DocSource{},
		KnEntry:          []*KnEntrySource{},
		Metric:           []*MetricSource{},
		KnowledgeNetwork: []*KnowledgeNetworkSource{},
		AdvancedConfig:   NewRetrieverAdvancedConfig(),
	}
}

func (r *RetrieverDataSource) IsNotSet() (notSet bool) {
	return len(r.Kg) == 0 && len(r.Doc) == 0 && len(r.Metric) == 0
}

func (r *RetrieverDataSource) GetErrMsgMap() map[string]string {
	// 返回错误信息映射，用于将验证错误转换为用户友好的错误消息
	return map[string]string{}
}

// ValObjCheck 检查检索数据源配置
func (r *RetrieverDataSource) ValObjCheckWithCtx(ctx context.Context) (err error) {
	// 1. 验证每个图谱类型数据源的有效性
	for _, kg := range r.Kg {
		if err = kg.ValObjCheck(); err != nil {
			// 包装错误信息，提供更详细的上下文
			err = errors.Wrap(err, "[RetrieverDataSource]: kg is invalid")
			return
		}
	}

	// 2. 验证每个文档类型数据源的有效性
	for _, doc := range r.Doc {
		if err = doc.ValObjCheck(); err != nil {
			// 包装错误信息，提供更详细的上下文
			err = errors.Wrap(err, "[RetrieverDataSource]: doc is invalid")
			return
		}
	}

	// 3. 验证每个指标类型数据源的有效性
	for _, metric := range r.Metric {
		if err = metric.ValObjCheck(); err != nil {
			err = errors.Wrap(err, "[RetrieverDataSource]: metric is invalid")
			return
		}
	}

	// 4. 验证高级配置
	if len(r.Kg) != 0 || len(r.Doc) != 0 {
		// 4.1 验证高级配置是否为空
		if r.AdvancedConfig == nil {
			err = errors.New("[RetrieverDataSource]: advanced_config is required when kg or doc is not empty")
			return
		}

		// 4.2 验证高级配置中的图谱数据源和文档数据源是否为空
		if len(r.Kg) > 0 && r.AdvancedConfig.KG == nil {
			err = errors.New("[RetrieverDataSource]: advanced_config.kg is required when kg is not empty")
			return
		}

		if len(r.Doc) > 0 && r.AdvancedConfig.Doc == nil {
			err = errors.New("[RetrieverDataSource]: advanced_config.doc is required when doc is not empty")
			return
		}

		// 4.3. 验证高级配置的有效性
		if err = r.AdvancedConfig.ValObjCheck(); err != nil {
			// 包装错误信息，提供更详细的上下文
			err = errors.Wrap(err, "[RetrieverDataSource]: advanced_config is invalid")
			return
		}
	}

	// 4.4 当r.Kg为空时，验证高级配置中的图谱数据源高级配置是否为空
	if len(r.Kg) == 0 {
		if r.AdvancedConfig != nil && r.AdvancedConfig.KG != nil {
			err = errors.New("[RetrieverDataSource]: advanced_config.kg is invalid when data_source.kg is empty")
			return
		}
	}

	// 4.5 当r.Doc为空时，验证高级配置中的文档数据源高级配置是否为空
	if len(r.Doc) == 0 {
		if r.AdvancedConfig != nil && r.AdvancedConfig.Doc != nil {
			err = errors.New("[RetrieverDataSource]: advanced_config.doc is invalid when data_source.doc is empty")
			return
		}
	}

	// 5. 验证知识条目类型数据源
	if len(r.KnEntry) != 0 {
		// 5.1 验证每个知识条目类型数据源的有效性
		for _, knEntry := range r.KnEntry {
			if err = knEntry.ValObjCheck(); err != nil {
				// 包装错误信息，提供更详细的上下文
				err = errors.Wrap(err, "[RetrieverDataSource]: kn_entry is invalid")
				return
			}
		}
	}

	// 5.2 验证知识条目类型数据源是否超过10个
	if len(r.KnEntry) > 10 {
		err = capierr.NewCustom400Err(ctx, capierr.DataAgentConfigRetrieverDataSourceKnEntryExceedLimitSize, "[RetrieverDataSource]: kn_entry exceeds limit size")
		return
	}
	// 6. 验证业务知识网络类型数据源
	if len(r.KnowledgeNetwork) != 0 {
		// 6.1 验证每个业务知识网络类型数据源的有效性
		for _, knowledgeNetwork := range r.KnowledgeNetwork {
			if err = knowledgeNetwork.ValObjCheck(); err != nil {
				err = errors.Wrap(err, "[RetrieverDataSource]: knowledge_network is invalid")
				return
			}
		}
	}

	return
}

func (r *RetrieverDataSource) GetBuiltInDsDocSourceFields() (docFields []*DocSourceField) {
	// for _, doc := range r.Doc {
	// 	if doc.DsID == "0" {
	// 		docFields = append(docFields, doc.Fields...)
	// 		return
	// 	}
	// }
	docSource := r.GetBuiltInDocDataSource()
	if docSource == nil {
		return
	}

	docFields = append(docFields, docSource.Fields...)

	return
}

func (r *RetrieverDataSource) GetBuiltInDocDataSource() (doc *DocSource) {
	for _, _doc := range r.Doc {
		if _doc.DsID == "0" {
			doc = _doc
			return
		}
	}

	return
}

func (r *RetrieverDataSource) GetFirstDocDatasetId() string {
	if len(r.Doc) == 0 {
		return ""
	}

	docSource := r.GetBuiltInDocDataSource()
	if docSource == nil {
		return ""
	}

	return docSource.GetFirstDatasetId()
}
