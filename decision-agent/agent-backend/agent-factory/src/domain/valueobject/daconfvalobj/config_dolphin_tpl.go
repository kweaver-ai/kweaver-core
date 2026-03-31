package daconfvalobj

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
)

func (p *Config) GetDolphinTplLength() (length int) {
	length = len(p.PreDolphin) + len(p.PostDolphin)
	return
}

// 根据config中的pre_dolphin和post_dolphin，判断一个dolphin tpl是否被用户禁用
func (p *Config) IsOneDolphinTplDisabled(key cdaenum.DolphinTplKey) (disabled bool) {
	var tpls []*DolphinTpl
	tpls = append(tpls, p.PreDolphin...)
	tpls = append(tpls, p.PostDolphin...)

	for _, tpl := range tpls {
		if tpl.Key == key && !tpl.Enabled {
			return true
		}
	}

	return false
}

func (p *Config) IsOneDolphinTplEdited(key cdaenum.DolphinTplKey) (edited bool) {
	var tpls []*DolphinTpl
	tpls = append(tpls, p.PreDolphin...)
	tpls = append(tpls, p.PostDolphin...)

	for _, tpl := range tpls {
		if tpl.Key == key && tpl.Edited {
			return true
		}
	}

	return false
}

func (p *Config) RemoveDataSourceFromPreDolphin(contextOrganizeValue string) (err error) {
	if p.DataSource != nil && !p.DataSource.IsNotSet() {
		panic("call this func when config has no data source")
	}

	isDocRetrieveEdited := p.IsOneDolphinTplEdited(cdaenum.DolphinTplKeyDocRetrieve)
	isGraphRetrieveEdited := p.IsOneDolphinTplEdited(cdaenum.DolphinTplKeyGraphRetrieve)
	isContextOrganizeEdited := p.IsOneDolphinTplEdited(cdaenum.DolphinTplKeyContextOrganize)

	// 1. 构建新的pre_dolphin
	newPreDolphin := make([]*DolphinTpl, 0)

	for _, tpl := range p.PreDolphin {
		switch tpl.Key {
		case cdaenum.DolphinTplKeyDocRetrieve:
			// 如果用户没有编辑过文档召回dolphin tpl，则跳过（不加到newPreDolphin中）
			if !isDocRetrieveEdited {
				continue
			}
		case cdaenum.DolphinTplKeyGraphRetrieve:
			// 如果用户没有编辑过业务知识网络召回dolphin tpl，则跳过（不加到newPreDolphin中）
			if !isGraphRetrieveEdited {
				continue
			}
		case cdaenum.DolphinTplKeyContextOrganize:
			// 如果用户没有编辑过上下文组织dolphin tpl，使用新的contextOrganizeValue替换原有值
			if !isContextOrganizeEdited {
				tpl.Value = contextOrganizeValue
			}
		}

		newPreDolphin = append(newPreDolphin, tpl)
	}

	// 2. 更新pre_dolphin
	p.PreDolphin = newPreDolphin

	return
}
