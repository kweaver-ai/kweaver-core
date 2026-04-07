package authzhttpres

import (
	"time"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdapmsenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/httpaccess/authzhttp/authzhttpreq"
)

// 策略管理相关的响应结构

// ListPolicyRes 查询策略列表响应
type ListPolicyRes struct {
	Entries    []*PolicyEntry `json:"entries"`
	TotalCount int            `json:"total_count"`
}

// PolicyEntry 策略条目
type PolicyEntry struct {
	ID        string                       `json:"id"`
	ExpiresAt string                       `json:"expires_at"`
	Resource  *authzhttpreq.PolicyResource `json:"resource"`
	Accessor  *PolicyAccessor              `json:"accessor"`
	Operation *PolicyOperation             `json:"operation"`
	Condition string                       `json:"condition"`
}

// FilterByExpiresAt 过滤已过期的策略条目
// expires_at 格式为 "1970-01-01T08:00:00+08:00"，对应 unix 时间戳为 0 表示永不过期
// 过滤掉 expires_at 对应的 unix 时间戳小于等于当前时间戳的数据
func (l *ListPolicyRes) FilterByExpiresAt() (err error) {
	if l == nil || len(l.Entries) == 0 {
		return
	}

	now := time.Now().Unix()
	filteredEntries := make([]*PolicyEntry, 0, len(l.Entries))

	for _, entry := range l.Entries {
		if entry == nil {
			continue
		}

		// 解析过期时间
		var expiresAt time.Time

		expiresAt, err = time.Parse(time.RFC3339, entry.ExpiresAt)
		if err != nil {
			// 如果解析失败，跳过该条目
			return
		}

		expiresAtUnix := expiresAt.Unix()

		// unix 时间戳为 0 表示永不过期
		// 或者过期时间大于当前时间，则保留
		if expiresAtUnix == 0 || expiresAtUnix > now {
			filteredEntries = append(filteredEntries, entry)
		}
	}

	// 更新结果
	l.Entries = filteredEntries
	l.TotalCount = len(filteredEntries)

	return
}

// FilterByOperation 根据指定的操作权限过滤策略条目
//
// 该方法实现基于权限控制的策略过滤逻辑：
// 1. 只保留在允许列表(Allow)中包含指定操作的策略条目
// 2. 拒绝列表(Deny)具有更高优先级，如果操作同时存在于Allow和Deny中，则该条目会被过滤掉
// 3. 过滤后会自动更新TotalCount字段
func (l *ListPolicyRes) FilterByOperation(operation cdapmsenum.Operator) (err error) {
	// 空值检查：如果列表为空或nil，直接返回
	if l == nil || len(l.Entries) == 0 {
		return
	}

	// 预分配过滤后的条目切片，提高性能
	filteredEntries := make([]*PolicyEntry, 0, len(l.Entries))

	// 遍历所有策略条目进行权限过滤
	for _, entry := range l.Entries {
		// 跳过nil条目
		if entry == nil {
			continue
		}

		// 1. 基础验证：过滤掉操作权限配置不完整的条目
		// 如果Operation为nil或Allow列表为nil，说明该条目没有有效的权限配置
		if entry.Operation == nil || entry.Operation.Allow == nil {
			continue
		}

		// 2. 拒绝列表检查：deny 优于 allow
		// 如果指定的操作在deny列表中，则无论allow列表如何都要过滤掉该条目
		isDenied := false

		if entry.Operation.Deny != nil {
			for _, op := range entry.Operation.Deny {
				if op.ID == operation {
					isDenied = true
					break
				}
			}
		}
		// 如果操作被拒绝，跳过该条目
		if isDenied {
			continue
		}

		// 3. 允许列表检查：只有在allow列表中找到匹配操作的条目才会被保留
		// 遍历允许操作列表，寻找匹配的操作
		for _, op := range entry.Operation.Allow {
			if op.ID == operation {
				// 找到匹配的操作，将该条目加入过滤结果
				filteredEntries = append(filteredEntries, entry)
				break // 找到匹配项后立即跳出循环，避免重复添加
			}
		}
	}

	// 更新过滤结果：用过滤后的条目替换原有条目，并更新总数
	l.Entries = filteredEntries
	l.TotalCount = len(filteredEntries)

	return
}
