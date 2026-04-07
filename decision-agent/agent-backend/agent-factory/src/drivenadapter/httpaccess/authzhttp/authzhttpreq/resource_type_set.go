package authzhttpreq

// ResourceTypeSetReq 设置资源类型请求
type ResourceTypeSetReq struct {
	Name        string                       `json:"name"`
	Description string                       `json:"description,omitempty"`
	InstanceURL string                       `json:"instance_url,omitempty"`
	DataStruct  string                       `json:"data_struct"`
	Operation   []*ResourceTypeOperationItem `json:"operation"`
	Hidden      bool                         `json:"hidden,omitempty"`
}

// ResourceTypeOperationItem 资源类型操作项
type ResourceTypeOperationItem struct {
	ID          string                       `json:"id"`
	Name        []*ResourceTypeOperationName `json:"name"`
	Description string                       `json:"description,omitempty"`
	Scope       []string                     `json:"scope"`
}

// ResourceTypeOperationName 资源类型操作名称
type ResourceTypeOperationName struct {
	Language string `json:"language"`
	Value    string `json:"value"`
}
