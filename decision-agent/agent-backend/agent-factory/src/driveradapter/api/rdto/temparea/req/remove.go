package tempareareq

type RemoveReq struct {
	TempAreaID string   `json:"-"`         // 路径参数中的id
	UserID     string   `json:"-"`         // 用户ID
	SourceIDs  []string `json:"source_id"` // query中传递的source_id
}
