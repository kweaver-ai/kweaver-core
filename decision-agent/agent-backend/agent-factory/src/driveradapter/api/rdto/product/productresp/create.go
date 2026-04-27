package productresp

type CreateRes struct {
	Key string `json:"key" example:"smart-customer-service"` // 产品Key
	ID  int64  `json:"id" example:"12345"`                   // 产品ID
}
