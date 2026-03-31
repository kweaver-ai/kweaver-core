package umret

type AppInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type AppListResDto struct {
	Entries    []*AppInfo `json:"entries"`
	TotalCount int64      `json:"total_count"`
}
