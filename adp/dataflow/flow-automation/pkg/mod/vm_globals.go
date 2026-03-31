package mod

import (
	"fmt"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/entity"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/vm"
)

const (
	KeyGlobalDagInsID      = "__g_dagInsID"
	KeyGloablAuthorization = "__g_authorization"
)

type global struct {
	dagIns *entity.DagInstance
}

func (g *global) Get(vm *vm.VM, name string, path []interface{}) interface{} {
	switch name {
	case KeyGlobalDagInsID:
		if g.dagIns != nil {
			return g.dagIns.ID
		}
		return ""
	case KeyGloablAuthorization:
		if g.dagIns != nil && g.dagIns.UserID != "" && g.dagIns.Status == entity.DagInstanceStatusRunning {
			tokenMgnt := NewTokenMgnt(g.dagIns.UserID)
			token, _ := tokenMgnt.GetUserToken("", g.dagIns.UserID)
			if token != nil {
				return fmt.Sprintf("Bearer %s", token.Token)
			}
		}

		return "********"
	default:
		return nil
	}
}

func NewGlobals(dagIns *entity.DagInstance) map[string]vm.Global {
	g := &global{dagIns}

	return map[string]vm.Global{
		KeyGlobalDagInsID:      g,
		KeyGloablAuthorization: g,
	}
}

var _ vm.Global = (*global)(nil)
