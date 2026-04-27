package agentrespvo

type DocRetrievalField struct {
	Text  string     `json:"text"`
	Cites []*CiteDoc `json:"cites,omitempty"`
}

type GraphRetrievalField struct {
	KGID     string      `json:"kg_id,omitempty"`
	KGName   string      `json:"kg_name,omitempty"`
	Subgraph interface{} `json:"subgraph,omitempty"`

	Text  string      `json:"text"`
	Cites interface{} `json:"cites,omitempty"`
}

type CiteDoc struct {
	Content    string     `json:"content"`      // 文档内容
	ExtType    string     `json:"ext_type"`     // 文档扩展名 带.
	DocID      string     `json:"doc_id"`       // 文档ID，gns:// 开头
	DocName    string     `json:"doc_name"`     // 文档名称
	ObjectID   string     `json:"object_id"`    // 文档对象ID gns最后一段
	ParentPath string     `json:"parent_path"`  // 文档父路径
	Size       int64      `json:"size"`         // 文档大小，单位：字节
	Type       string     `json:"type"`         // 类型，document
	Slices     []*V1Slice `json:"slices"`       // 文档切片
	SpaceID    string     `json:"space_id"`     // 空间ID
	DocLibType string     `json:"doc_lib_type"` // 文档库类型，knowledge_doc_lib等
}

// V1Slice 切片文档
type V1Slice struct {
	ID      string `json:"id"`
	No      int    `json:"no"`
	Content string `json:"content"`
	Pages   []int  `json:"pages"`
}
