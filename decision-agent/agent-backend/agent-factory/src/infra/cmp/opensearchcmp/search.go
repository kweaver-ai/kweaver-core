package opensearchcmp

import (
	"encoding/json"
	"log"
)

// ToJSON 将接口转换为JSON字符串
func ToJSON(v interface{}) string {
	data, err := json.Marshal(v)
	if err != nil {
		log.Printf("marshal to json failed: %v", err)
		return "{}"
	}

	return string(data)
}
