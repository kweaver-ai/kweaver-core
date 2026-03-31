package dapo

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/entity/dolphintpleo"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/daenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/daconfvalobj"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/daconfvalobj/datasourcevalobj"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
	"github.com/pkg/errors"
)

type DataAgentPo struct {
	ID   string `json:"id" db:"f_id"`
	Name string `json:"name" db:"f_name"`
	Key  string `json:"key" db:"f_key"`

	Profile *string `json:"profile" db:"f_profile"`

	ProductKey string `json:"product_key" db:"f_product_key"`

	AvatarType cdaenum.AvatarType `json:"avatar_type" db:"f_avatar_type"`
	Avatar     string             `json:"avatar" db:"f_avatar"`

	Status cdaenum.Status `json:"status" db:"f_status"`

	IsBuiltIn *cdaenum.BuiltIn `json:"is_built_in" db:"f_is_built_in"`

	IsSystemAgent *cenum.YesNoInt8 `json:"is_system_agent" db:"f_is_system_agent"`

	CreatedAt int64 `json:"created_at" db:"f_created_at"`
	UpdatedAt int64 `json:"updated_at" db:"f_updated_at"`

	CreatedBy string `json:"created_by" db:"f_created_by"`
	UpdatedBy string `json:"updated_by" db:"f_updated_by"`

	DeletedAt int64  `json:"deleted_at" db:"f_deleted_at"`
	DeletedBy string `json:"deleted_by" db:"f_deleted_by"`

	// PublishedAt int64  `json:"published_at" db:"f_published_at"`
	// PublishedBy string `json:"published_by" db:"f_published_by"`

	Config string `json:"config" db:"f_config"`

	CreatedType daenum.AgentCreatedType `json:"created_type" db:"f_created_type"`

	CreateFrom string `json:"create_from" db:"f_create_from"`
}

func (p *DataAgentPo) TableName() string {
	return "t_data_agent_config"
}

func (p *DataAgentPo) IsBuiltInBool() bool {
	if p.IsBuiltIn == nil {
		return false
	}

	return *p.IsBuiltIn == cdaenum.BuiltInYes
}

func (p *DataAgentPo) SetIsBuiltIn(builtIn cdaenum.BuiltIn) {
	p.IsBuiltIn = &builtIn
}

func (p *DataAgentPo) GetConfigStruct() (conf *daconfvalobj.Config, err error) {
	conf = &daconfvalobj.Config{}

	err = cutil.JSON().UnmarshalFromString(p.Config, conf)
	if err != nil {
		err = errors.Wrapf(err, "get config struct error")
		return
	}

	return
}

func (p *DataAgentPo) GetProfileStr() (profile string) {
	if p.Profile == nil {
		return ""
	}

	return *p.Profile
}

func (p *DataAgentPo) ResetForImport() {
	p.CreatedAt = 0
	p.UpdatedAt = 0

	p.CreatedBy = ""
	p.UpdatedBy = ""

	p.DeletedAt = 0
	p.DeletedBy = ""

	p.CreatedType = daenum.AgentCreatedTypeImport
	p.CreateFrom = ""

	p.Status = cdaenum.StatusUnpublished

	p.IsBuiltIn = nil
	// p.IsSystemAgent = nil

	p.ID = ""
}

func (p *DataAgentPo) SetConfigStruct(conf *daconfvalobj.Config) (err error) {
	confStr, err := cutil.JSON().MarshalToString(conf)
	if err != nil {
		err = errors.Wrapf(err, "set config struct error")
		return
	}

	p.Config = confStr

	return
}

func (p *DataAgentPo) RemoveDataSourceFromConfig(isRemoveFromDolphin bool) (err error) {
	// 1. 获取配置
	confS, err := p.GetConfigStruct()
	if err != nil {
		return
	}

	// 2. 清除数据源
	confS.DataSource = datasourcevalobj.NewRetrieverDataSource()

	if isRemoveFromDolphin {
		dolTplMapStruct := dolphintpleo.NewDolphinTplMapStruct()

		dolTplMapStruct.LoadFromConfig(confS, "", false)

		contextOrganizeEo := dolTplMapStruct.ContextOrganize.ToDolphinTplEo()

		err = confS.RemoveDataSourceFromPreDolphin(contextOrganizeEo.Value)
		if err != nil {
			return
		}
	}

	// 3. 设置配置
	err = p.SetConfigStruct(confS)
	if err != nil {
		return
	}

	return
}
