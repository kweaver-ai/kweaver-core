package ctype

import "github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cenum"

type VisitorInfo struct {
	XAccountID   string            `json:"x_account_id"`
	XAccountType cenum.AccountType `json:"x_account_type"`

	XBusinessDomainID cenum.BizDomainID `json:"x_business_domain_id"`
}
