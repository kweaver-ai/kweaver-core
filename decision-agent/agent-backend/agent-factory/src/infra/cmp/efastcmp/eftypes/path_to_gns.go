package eftypes

type Path2GnsReq struct {
	Namepath string `json:"namepath"` // 名字路径，由顶级入口（个人文档/文档库等）开始的对象全路径，以”/”分隔
}

type Path2GnsResponse struct {
	// ClientMtime
	// 如果是文件，返回由客户端设置的文件本地修改时间
	// 若未设置，返回modified的值
	ClientMtime int `json:"client_mtime"`

	// Updated 目录修改时间/文件上传时间，UTC时间，此为文件上传到服务器时间
	Modified int `json:"modified"`

	// DocId 文件/目录的gns路径
	DocId string `json:"docid"`

	// Name 文件/目录的名称，UTF8编码
	Name string `json:"name"`

	// Rev 文件版本号或目录数据变化标识
	Rev string `json:"rev"`

	// Size 文件的大小，目录大小为-1
	Size int `json:"size"`
}
