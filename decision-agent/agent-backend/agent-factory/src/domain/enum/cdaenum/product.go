package cdaenum

import "github.com/pkg/errors"

type Product string

const (
	ProductDIP      Product = "DIP"
	ProductAnyShare Product = "AnyShare"
	// 智能问数
	ProductChatBI Product = "ChatBI"
)

func (p Product) EnumCheck() (err error) {
	if p != ProductDIP && p != ProductAnyShare && p != ProductChatBI {
		err = errors.New("[Product]: invalid product")
		return
	}

	return
}
