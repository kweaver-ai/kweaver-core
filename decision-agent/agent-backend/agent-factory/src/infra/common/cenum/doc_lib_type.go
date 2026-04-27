package cenum

import "errors"

type DocLibType string

const (
	DocLibTypeStrPersonal   DocLibType = "user_doc_lib"       // 个人文档库
	DocLibTypeStrDepartment DocLibType = "department_doc_lib" // 部门文档库
	DocLibTypeStrCustom     DocLibType = "custom_doc_lib"     // 自定义文档库
	DocLibTypeStrKnowledge  DocLibType = "knowledge_doc_lib"  // 知识库
)

func (d DocLibType) EnumCheck() error {
	if d != DocLibTypeStrPersonal && d != DocLibTypeStrDepartment && d != DocLibTypeStrCustom && d != DocLibTypeStrKnowledge {
		return errors.New("[DocLibType][EnumCheck]: 无效的文档库类型")
	}

	return nil
}
