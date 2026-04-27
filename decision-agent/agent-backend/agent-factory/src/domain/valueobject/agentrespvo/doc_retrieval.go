package agentrespvo

type DocRetrievalRes struct {
	Answer      interface{}        `json:"answer"`
	BlockAnswer DocRetrievalAnswer `json:"block_answer"`
}
type DocRetrievalAnswer struct {
	Result     string     `json:"result"`
	FullResult FullResult `json:"full_result"`
}
type FullResult struct {
	Text       string       `json:"text"`
	References []*Reference `json:"references"`
}
type AnswerCite struct {
	Content  string      `json:"content"`
	CiteType string      `json:"type"`
	Meta     interface{} `json:"meta"`
	Score    float64     `json:"score"`
}

func (r *DocRetrievalRes) AnswerAndCites() (answer string, cites []*AnswerCite) {
	answer = r.Answer.(DocRetrievalAnswer).FullResult.Text

	for _, reference := range r.Answer.(DocRetrievalAnswer).FullResult.References {
		cites = append(cites, &AnswerCite{
			Content:  reference.Content,
			Meta:     reference.Meta,
			CiteType: reference.RetrieveSourceType,
			Score:    reference.Score,
		})
	}

	return
}

// Reference 表示单个参考信息
type Reference struct {
	Content            string  `json:"content"`
	RetrieveSourceType string  `json:"retrieve_source_type"`
	Score              float64 `json:"score"`
	Meta               *Meta   `json:"meta"`
}

// Meta 表示参考信息的元数据
type Meta struct {
	DocLibType string   `json:"doc_lib_type"`
	ObjectID   string   `json:"object_id"`
	DocName    string   `json:"doc_name"`
	ExtType    string   `json:"ext_type"`
	ParentPath string   `json:"parent_path"`
	Size       int      `json:"size"`
	DocID      string   `json:"doc_id"`
	Slices     []*Slice `json:"slices"`
	DataSource string   `json:"data_source"`
}

// Slice 表示文档的切片信息
type Slice struct {
	Score   float64 `json:"score"`
	ID      string  `json:"id"`
	No      int     `json:"no"`
	Content string  `json:"content"`
	Pages   []int   `json:"pages"`
}
