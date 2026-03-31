package valueobject

type TempFile struct {
	ID      string      `json:"id"`
	Type    string      `json:"type"`
	Name    string      `json:"name"`
	Details interface{} `json:"details"`
}
