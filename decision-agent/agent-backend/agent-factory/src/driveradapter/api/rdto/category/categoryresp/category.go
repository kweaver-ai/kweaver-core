package categoryresp

type ListResp []CategoryResp

type CategoryResp struct {
	ID          string `json:"category_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}
