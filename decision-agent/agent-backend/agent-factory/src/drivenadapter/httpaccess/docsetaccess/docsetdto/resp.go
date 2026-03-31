package docsetdto

type FullTextRsp struct {
	DocID   string `json:"doc_id"`
	Version string `json:"version"`
	Status  string `json:"status"`
	Url     string `json:"url"`
}
