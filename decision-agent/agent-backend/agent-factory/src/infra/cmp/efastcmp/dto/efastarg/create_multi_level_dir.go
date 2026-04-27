package efastarg

type CreateMultiLevelDirReq struct {
	// 待创建多级目录的父目录gns路径
	DocId string `json:"docid"`

	// 多级目录名称，UTF8编码
	Path string `json:"path"`
}
