package daconfvalobj

import (
	"fmt"
	"strings"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
	"github.com/pkg/errors"
)

type Fields []*Field

func (p *Fields) ValObjCheck() (err error) {
	var fileFields []*Field

	// 1. 验证每个字段的有效性
	for _, field := range *p {
		if err = field.ValObjCheck(); err != nil {
			// 包装错误信息，提供更详细的上下文
			err = errors.Wrap(err, "[Input]: field is invalid")
			return
		}

		if field.Type == cdaenum.InputFieldTypeFile {
			fileFields = append(fileFields, field)
		}
	}

	// 2. 检查是否有多个文件类型字段
	if len(fileFields) > 1 {
		err = errors.New("[Input]: multiple file type fields are not allowed")
		return
	}

	// 3. 检查是否有重名的字段
	if p.IsFieldNameRepeat() {
		err = errors.New("[Input]: field names must be unique")
		return
	}

	return
}

func (p *Fields) IsEnabledTempZone() bool {
	for _, field := range *p {
		if field.Type == cdaenum.InputFieldTypeFile {
			return true
		}
	}

	return false
}

func (p *Fields) IsFieldNameRepeat() bool {
	nameMap := make(map[string]bool)
	for _, field := range *p {
		if _, ok := nameMap[field.Name]; ok {
			return true
		}

		nameMap[field.Name] = true
	}

	return false
}

// 生成非文件类型字段的dolphin字符串（详情见：https://confluence.aishu.cn/pages/viewpage.action?pageId=272343001）
func (p *Fields) GenNotFileDolphinStr() (dolphinStr string) {
	sb := strings.Builder{}

	skipFields := []string{"history", "tool", "header", "self_config"}

	for _, field := range *p {
		// 文件类型字段不参与
		if field.Type == cdaenum.InputFieldTypeFile {
			continue
		}

		// 跳过
		if cutil.ExistsGeneric(skipFields, field.Name) {
			continue
		}

		str := fmt.Sprintf(`"%s: " + $%s + "\n" + `, field.Name, field.Name)

		sb.WriteString(str)
	}

	dolphinStr = sb.String()

	// 去掉最后一个 "\n" +
	if dolphinStr != "" {
		dolphinStr = strings.TrimSuffix(dolphinStr, ` + "\n" + `)
	}

	dolphinStr += " -> all_inputs \n"

	return
}

// 生成文件类型字段的dolphin字符串（详情见：https://confluence.aishu.cn/pages/viewpage.action?pageId=272343001）
func (p *Fields) GenFileDolphinStr() (dolphinStr, filedName string) {
	sb := strings.Builder{}

	for _, field := range *p {
		// 文件类型字段参与
		if field.Type != cdaenum.InputFieldTypeFile {
			continue
		}

		str := fmt.Sprintf(`@process_file_intelligent(query=$query, file_infos=$%s) -> %s`, field.Name, field.Name)

		sb.WriteString(str)

		filedName = field.Name

		// 文件类型的字段只有1个
		break
	}

	dolphinStr = sb.String()

	return
}
