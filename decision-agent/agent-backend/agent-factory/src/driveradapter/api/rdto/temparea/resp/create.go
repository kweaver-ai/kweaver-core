package temparearesp

type CreateResp struct {
	ID      string       `json:"id"`
	Sources []SourceResp `json:"sources"`
}

type SourceResp struct {
	ID      string  `json:"id"`
	Details Details `json:"details"`
}

type Details struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}
