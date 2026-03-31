package cutil

import (
	"fmt"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

func AddToJSON(jsonStr, jsonPath, value string) (string, error) {
	if !gjson.Valid(jsonStr) {
		return "", fmt.Errorf("[AddToJSON]: Invalid JSON")
	}

	updated, err := sjson.Set(jsonStr, jsonPath, value)
	if err != nil {
		return "", err
	}

	return updated, nil
}

func AddKeyToJSONArray(jsonArrayStr string, key string, value interface{}) (string, error) {
	if !gjson.Valid(jsonArrayStr) {
		return "", fmt.Errorf("[AddKeyToJSONArray]: Invalid JSON")
	}

	// Parse the JSON array string
	array := gjson.Parse(jsonArrayStr).Array()

	// Make a copy of the array
	updatedArray := jsonArrayStr

	// Iterate over each element in the array
	for index := range array {
		// Construct the JSON path for the element's new key
		jsonPath := fmt.Sprintf("%d.%s", index, key)

		// Update the JSON string with the new key-value pair at the specified path
		var err error

		updatedArray, err = sjson.Set(updatedArray, jsonPath, value)
		if err != nil {
			return "", err
		}
	}

	return updatedArray, nil
}

func AddKeyToJSONArrayBys(jsonArrayBys []byte, key string, value interface{}) (newBys []byte, err error) {
	jsonArrayStr := string(jsonArrayBys)

	updatedArray, err := AddKeyToJSONArray(jsonArrayStr, key, value)
	if err != nil {
		return nil, err
	}

	newBys = []byte(updatedArray)

	return
}

func AddJSONArrayToJSON(jsonStr, jsonPath, jsonArrayStr string) (string, error) {
	if !gjson.Valid(jsonStr) {
		return "", fmt.Errorf("Invalid JSON")
	}

	// 解析JSON数组字符串
	jsonArray := gjson.Parse(jsonArrayStr)

	updated, err := sjson.Set(jsonStr, jsonPath, jsonArray.Value())
	if err != nil {
		return "", err
	}

	return updated, nil
}

// 将某个path key从json字符串中删除
func RemoveKeyFromJSON(jsonStr, jsonPath string) (newStr string, err error) {
	if !gjson.Valid(jsonStr) {
		err = fmt.Errorf("Invalid JSON")
		return
	}

	updated, err := sjson.Delete(jsonStr, jsonPath)
	if err != nil {
		return
	}

	newStr = updated

	return
}
