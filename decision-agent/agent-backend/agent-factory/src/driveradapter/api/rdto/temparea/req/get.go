package tempareareq

type GetReq struct {
	TempAreaID string `json:"-"` // 路径参数中的id
}
