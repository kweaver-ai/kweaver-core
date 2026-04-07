package mod

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/entity"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/vm"
)

type compare struct{}

func (c *compare) Call(ctx context.Context, name string, nrets int, args ...interface{}) (wait bool, rets []interface{}, err error) {
	op := entity.Operator(name)
	var a, b interface{} = nil, nil

	if len(args) > 0 {
		a = args[0]
	}

	if len(args) > 1 {
		b = args[1]
	}

	rets = append(rets, c.compare(op, a, b))
	return
}

func (c *compare) compare(op entity.Operator, vl interface{}, vr interface{}) bool {
	switch op {
	case entity.OperatorStringIn:
		vls, vrs, ok := c.ParseString(vl, vr)
		if !ok {
			return false
		}
		return strings.Contains(vls, vrs)
	case entity.OperatorStringNotIn:
		vls, vrs, ok := c.ParseString(vl, vr)
		if !ok {
			return false
		}
		return !strings.Contains(vls, vrs)
	case entity.OperateStringEq, entity.OperateWorkflowEq:
		vls, vrs, ok := c.ParseString(vl, vr)
		if !ok {
			return false
		}
		return vrs == vls
	case entity.OperateStringNeq, entity.OperateWorkflowNeq:
		vls, vrs, ok := c.ParseString(vl, vr)
		if !ok {
			return false
		}
		return vrs != vls
	case entity.OperateStringStartWith:
		vls, vrs, ok := c.ParseString(vl, vr)
		if !ok {
			return false
		}
		return strings.HasPrefix(vls, vrs)
	case entity.OperateStringEndWith:
		vls, vrs, ok := c.ParseString(vl, vr)
		if !ok {
			return false
		}
		return strings.HasSuffix(vls, vrs)
	case entity.OperateStringEmpty:
		vls, _, ok := c.ParseString(vl, "")
		if !ok {
			return false
		}
		return vls == ""
	case entity.OperateStringNotEmpty:
		vls, _, ok := c.ParseString(vl, "")
		if !ok {
			return false
		}
		return vls != ""
	case entity.OperateStringMatch:
		vls, vrs, ok := c.ParseString(vl, vr)
		if !ok {
			return false
		}
		matched, err := regexp.MatchString(vrs, vls)
		if err != nil {
			return false
		}
		return matched
	case entity.OperateNumberEq:
		vls, vrs, ok := c.ParseNumber(vl, vr)
		if !ok {
			return false
		}
		return vrs == vls
	case entity.OperateNumberNeq:
		vls, vrs, ok := c.ParseNumber(vl, vr)
		if !ok {
			return false
		}
		return vrs != vls
	case entity.OperateNumberGte:
		vls, vrs, ok := c.ParseNumber(vl, vr)
		if !ok {
			return false
		}
		return vls >= vrs
	case entity.OperateNumberGt:
		vls, vrs, ok := c.ParseNumber(vl, vr)
		if !ok {
			return false
		}
		return vls > vrs
	case entity.OperateNumberLte:
		vls, vrs, ok := c.ParseNumber(vl, vr)
		if !ok {
			return false
		}
		return vls <= vrs
	case entity.OperateNumberLt:
		vls, vrs, ok := c.ParseNumber(vl, vr)
		if !ok {
			return false
		}
		return vls < vrs
	case entity.OperateDateEq:
		vls, vrs, ok := c.ParseDate(vl, vr)
		if !ok {
			return false
		}
		return vls == vrs
	case entity.OperateDateNeq:
		vls, vrs, ok := c.ParseDate(vl, vr)
		if !ok {
			return true
		}
		return vls != vrs
	case entity.OperateDateBefore:
		vls, vrs, ok := c.ParseDate(vl, vr)
		if !ok {
			return false
		}
		return vls < vrs
	case entity.OperateDateAfter:
		vls, vrs, ok := c.ParseDate(vl, vr)
		if !ok {
			return false
		}
		return vls > vrs
	}

	return false
}

// ParseString 变量转换为string类型
func (c *compare) ParseString(vl, vr interface{}) (vls, vrs string, ok bool) {
	vls, ok = vl.(string)
	if !ok {
		return
	}
	vrs, ok = vr.(string)
	if !ok {
		return
	}

	return
}

// ParseNumber 转换数字
func (c *compare) ParseNumber(vl, vr interface{}) (vls, vrs float64, ok bool) {
	switch tl := vl.(type) {
	case string:
		floatvar, err := strconv.ParseFloat(tl, 64)
		if err != nil {
			vls = 0
			break
		}
		vls = floatvar
	case float64:
		vls = tl
	case int64:
		vls = float64(tl)
	}

	switch tr := vr.(type) {
	case string:
		floatvar, err := strconv.ParseFloat(tr, 64)
		if err != nil {
			vls = 0
			break
		}
		vrs = floatvar
	case float64:
		vrs = tr
	case int64:
		vrs = float64(tr)
	}
	return vls, vrs, true
}

// ParseDate 转换日期
func (c *compare) ParseDate(vl, vr interface{}) (vls, vrs int64, ok bool) {
	switch tl := vl.(type) {
	case string:
		v, err := time.Parse(time.RFC3339, tl)
		if err != nil {
			var newNum float64
			_, err := fmt.Sscanf(tl, "%e", &newNum)
			if err != nil {
				vls = 0
				break
			}
			vls = int64(newNum)
			break
		}
		vls = v.UnixNano() / 1e9 // 秒级
	case int64:
		vls = tl
	}

	switch tr := vr.(type) {
	case string:
		v, err := time.Parse(time.RFC3339, tr)
		if err != nil {
			var newNum float64
			_, err := fmt.Sscanf(tr, "%e", &newNum)
			if err != nil {
				vrs = 0
				break
			}
			vrs = int64(newNum)
			break
		}
		vrs = v.UnixNano() / 1e9 // 秒级
	case int64:
		vrs = tr
	}
	return vls, vrs, true
}

var cmpFuncs = &compare{}

var compareFuncs = map[string]vm.Func{
	"@internal/cmp/string-contains":     cmpFuncs,
	"@internal/cmp/string-not-contains": cmpFuncs,
	"@internal/cmp/string-eq":           cmpFuncs,
	"@internal/cmp/string-neq":          cmpFuncs,
	"@internal/cmp/string-start-with":   cmpFuncs,
	"@internal/cmp/string-end-with":     cmpFuncs,
	"@internal/cmp/string-empty":        cmpFuncs,
	"@internal/cmp/string-not-empty":    cmpFuncs,
	"@internal/cmp/string-match":        cmpFuncs,
	"@internal/cmp/number-eq":           cmpFuncs,
	"@internal/cmp/number-neq":          cmpFuncs,
	"@internal/cmp/number-gte":          cmpFuncs,
	"@internal/cmp/number-gt":           cmpFuncs,
	"@internal/cmp/number-lt":           cmpFuncs,
	"@internal/cmp/number-lte":          cmpFuncs,
	"@internal/cmp/date-eq":             cmpFuncs,
	"@internal/cmp/date-neq":            cmpFuncs,
	"@internal/cmp/date-earlier-than":   cmpFuncs,
	"@internal/cmp/date-later-than":     cmpFuncs,
	"@workflow/cmp/approval-eq":         cmpFuncs,
	"@workflow/cmp/approval-neq":        cmpFuncs,
}
