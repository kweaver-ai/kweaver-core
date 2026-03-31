// Package policy logics 服务开关控制
package policy

import (
	"sync"

	commonLog "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/log"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/mod"
)

//go:generate mockgen -package mock_logics -source ../../logics/policy/policy.go -destination ../../tests/mock_logics/policy_mock.go

// Handler method interfaces
type Handler interface {
	CheckStatus() (bool, error)
	SetStatus(status bool) error
}

var (
	pOnce sync.Once
	p     Handler
)

type policy struct {
	log   commonLog.Logger
	mongo mod.Store
}

// NewPolicy new policy handler
func NewPolicy() Handler {
	pOnce.Do(func() {
		p = &policy{
			log:   commonLog.NewLogger(),
			mongo: mod.GetStore(),
		}
	})

	return p
}

func (p *policy) CheckStatus() (bool, error) {
	status, err := p.mongo.GetSwitchStatus()
	if err != nil {
		p.log.Errorf("[logic.CheckStatus] CheckStatus err, detail: %s", err.Error())
	}
	return status, err
}

func (p *policy) SetStatus(status bool) error {
	err := p.mongo.SetSwitchStatus(status)
	if err != nil {
		p.log.Errorf("[logic.SetStatus] SetStatus err, detail: %s", err.Error())
	}
	return err
}
