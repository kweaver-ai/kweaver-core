package agentresp

type FileCheckResp struct {
	Progress    int    `json:"progress"`
	ProcessInfo []Info `json:"process_info"`
}

type Info struct {
	ID     string `json:"id"`
	Status string `json:"status"` // completed processing failed
}
