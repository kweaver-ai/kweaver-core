package efastret

type CreateMultiLevelDirRsp struct {
	// 创建的多级目录最后一级的gns路径
	DocId string `json:"docid"`

	// 数据变化标识
	Rev string `json:"rev"`

	// 创建时间，UTC时间，此为服务器时间
	Modified int64 `json:"modified"`
}
