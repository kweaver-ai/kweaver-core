package cenum

import (
	"errors"

	"github.com/kweaver-ai/kweaver-go-lib/rest"
)

type AccountType string

func (a *AccountType) String() string {
	return string(*a)
}

const (
	AccountTypeUser      AccountType = "user"
	AccountTypeApp       AccountType = "app"
	AccountTypeAnonymous AccountType = "anonymous"
)

func (a *AccountType) EnumCheck() error {
	if *a != AccountTypeUser && *a != AccountTypeApp && *a != AccountTypeAnonymous {
		return errors.New("[AccountType][EnumCheck]: 无效的账户类型")
	}

	return nil
}

func (a *AccountType) LoadFromMDLVisitorType(vt rest.VisitorType) {
	switch vt {
	case rest.VisitorType_RealName:
		*a = AccountTypeUser
	case rest.VisitorType_User:
		*a = AccountTypeUser
	case rest.VisitorType_Anonymous:
		*a = AccountTypeAnonymous
	case rest.VisitorType_App:
		*a = AccountTypeApp
	default:
		panic("[AccountType][LoadFromMDLVisitorType]: 无效的账户类型")
	}
}

func (a *AccountType) ToMDLVisitorType() rest.VisitorType {
	switch *a {
	case AccountTypeUser:
		return rest.VisitorType_RealName
	case AccountTypeApp:
		return rest.VisitorType_App
	case AccountTypeAnonymous:
		return rest.VisitorType_Anonymous
	default:
		panic("[AccountType][ToMDLVisitorType]: 无效的账户类型")
	}
}
