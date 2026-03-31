package example

import (
	"context"
	"fmt"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/common/ormhelper"
)

// MCPServerConfigExample 示例结构体，模拟实际的MCPServerConfigDB
type MCPServerConfigExample struct {
	ID          string `json:"f_id" db:"f_id"`
	CreateUser  string `json:"f_create_user" db:"f_create_user"`
	CreateTime  int64  `json:"f_create_time" db:"f_create_time"`
	UpdateUser  string `json:"f_update_user" db:"f_update_user"`
	UpdateTime  int64  `json:"f_update_time" db:"f_update_time"`
	Name        string `json:"f_name" db:"f_name"`
	Description string `json:"f_description" db:"f_description"`
	Mode        string `json:"f_mode" db:"f_mode"`
	URL         string `json:"f_url" db:"f_url"`
	Headers     string `json:"f_headers" db:"f_headers"`
	Command     string `json:"f_command" db:"f_command"`
	Env         string `json:"f_env" db:"f_env"`
	Args        string `json:"f_args" db:"f_args"`
	Status      string `json:"f_status" db:"f_status"`
	Category    string `json:"f_category" db:"f_category"`
	Source      string `json:"f_source" db:"f_source"`
}

// DebugFieldMappingExample 演示字段映射调试
func DebugFieldMappingExample() {
	config := &MCPServerConfigExample{}

	fmt.Println("=== ORM Helper 字段映射调试示例 ===")

	// 调试结构体字段映射
	ormhelper.DebugFieldMapping(config)

	// 模拟数据库返回的列顺序（可能与结构体字段顺序不同）
	columns := []string{
		"f_id", "f_name", "f_description", "f_mode", "f_url", "f_headers",
		"f_command", "f_env", "f_args", "f_status", "f_category", "f_source",
		"f_create_user", "f_create_time", "f_update_user", "f_update_time",
	}

	// 调试列映射
	ormhelper.DebugColumnMapping(config, columns)
}

// ExampleSelectByIDWithDebug 在SelectByID方法中使用调试功能的示例
func ExampleSelectByIDWithDebug(ctx context.Context, orm *ormhelper.DB, id string) (*MCPServerConfigExample, error) {
	config := &MCPServerConfigExample{}

	// 调试：打印字段映射信息
	fmt.Println("调试SelectByID方法:")
	ormhelper.DebugFieldMapping(config)

	err := orm.Select().From("t_mcp_server_config").WhereEq("f_id", id).First(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("查询失败: %w", err)
	}

	return config, nil
}

// DemonstrateFixedVsUnfixed 演示修复前后的差异
func DemonstrateFixedVsUnfixed() {
	fmt.Println("=== 修复前后对比 ===")

	config := &MCPServerConfigExample{}

	fmt.Println("修复前的问题:")
	fmt.Println("- ScanOne方法按结构体字段顺序扫描")
	fmt.Println("- 不考虑数据库列顺序")
	fmt.Println("- 导致字段类型不匹配错误")
	fmt.Println()

	fmt.Println("修复后的改进:")
	fmt.Println("- First方法使用QueryContext获取列信息")
	fmt.Println("- 通过db标签建立字段映射")
	fmt.Println("- 按数据库列顺序正确映射到结构体字段")
	fmt.Println()

	// 展示字段映射
	ormhelper.DebugFieldMapping(config)

	// 模拟问题场景：数据库列顺序与结构体字段顺序不同
	dbColumns := []string{"f_id", "f_name", "f_description", "f_create_time"}
	fmt.Printf("数据库列顺序: %v\n", dbColumns)
	fmt.Printf("结构体字段顺序: ID[0], CreateUser[1], CreateTime[2], UpdateUser[3]...\n")
	fmt.Println("修复前: f_description(string) -> CreateTime(int64) ❌ 类型错误")
	fmt.Println("修复后: f_description(string) -> Description(string) ✅ 正确映射")
}
