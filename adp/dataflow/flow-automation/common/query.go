package common

import (
	"fmt"
	"strings"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/utils"
)

var dagFieldQueryLimits = []string{"name", "description", "status", "created_at", "updated_at", "type", "creator", "trigger"}

var FieldQuery = &FieldMapper{FieldMap: map[string]string{}}

func init() {
	FieldQuery.AddFieldMapping("name", "name")
	FieldQuery.AddFieldMapping("description", "description")
	FieldQuery.AddFieldMapping("status", "status")
	FieldQuery.AddFieldMapping("created_at", "createdAt")
	FieldQuery.AddFieldMapping("updated_at", "updatedAt")
	FieldQuery.AddFieldMapping("type", "type")
	FieldQuery.AddFieldMapping("creator", "userid")
	FieldQuery.AddFieldMapping("trigger", "trigger")
}

/*
* 访问者模式
* 后续若有新增的接口按字段查询可复用此函数
* FieldQuery包含所有接口接口查询与数据库字段映射关系
* 通过构造DagQueryVisitor 访问器根据业务自定义需要映射的字段
* 不同业务实现不同的visitor访问器
 */

// FieldMapper 统一存储所有字段映射关系
type FieldMapper struct {
	FieldMap map[string]string // 统一的字段映射表
}

func (f *FieldMapper) AddFieldMapping(apiField string, dbField string) {
	f.FieldMap[apiField] = dbField
}

// Accept 让访问者选择需要的字段
func (f *FieldMapper) Accept(visitor Visitor) ([]string, error) {
	return visitor.Visit(f)
}

// Visitor 接口
type Visitor interface {
	Visit(*FieldMapper) ([]string, error)
}

type DagQueryVisitor struct {
	Fields string
}

func (dv *DagQueryVisitor) Visit(f *FieldMapper) ([]string, error) {
	fieldList := strings.Split(dv.Fields, ",")
	var dbFields []string
	var invalidF []string

	for _, field := range fieldList {
		if dbField, ok := f.FieldMap[strings.TrimSpace(field)]; ok && utils.Contains[string](dagFieldQueryLimits, field) {
			dbFields = append(dbFields, dbField)
		} else {
			invalidF = append(invalidF, field)
		}
	}

	if len(invalidF) > 0 {
		return dbFields, fmt.Errorf("Error: Invalid field %s", strings.Join(invalidF, ","))
	}

	return dbFields, nil
}
