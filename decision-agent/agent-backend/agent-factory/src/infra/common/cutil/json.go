package cutil

import (
	"bytes"
	"encoding/json"
	"fmt"

	jsoniter "github.com/json-iterator/go"
)

func JSON() (json jsoniter.API) {
	json = jsoniter.ConfigCompatibleWithStandardLibrary
	return
}

func JSONObjectToArray(bys []byte) (newBys []byte) {
	var buffer bytes.Buffer

	buffer.WriteByte('[')
	buffer.Write(bys)
	buffer.WriteByte(']')
	newBys = buffer.Bytes()

	return
}

// FormatJSONString takes a JSON string as input and returns a formatted JSON string.
// if input is empty, return ""
func FormatJSONString(input string) (string, error) {
	if input == "" {
		return "", nil
	}

	var jsonObj interface{}

	err := json.Unmarshal([]byte(input), &jsonObj)
	if err != nil {
		return "", err
	}

	formattedJSON, err := json.MarshalIndent(jsonObj, "", "  ")
	if err != nil {
		return "", err
	}

	return string(formattedJSON), nil
}

// FormatJSON takes an interface{} as input and returns a formatted JSON string.
func FormatJSON(input interface{}) (string, error) {
	formattedJSON, err := json.MarshalIndent(input, "", "  ")
	if err != nil {
		return "", err
	}

	return string(formattedJSON), nil
}

func ToMapByJSON(obj interface{}) (m map[string]interface{}, err error) {
	m = make(map[string]interface{})

	var tmpBys []byte

	tmpBys, err = JSON().Marshal(obj)
	if err != nil {
		return
	}

	err = JSON().Unmarshal(tmpBys, &m)

	return
}

func PrintFormatJSONString(str, prefix string) (err error) {
	formattedLogJSON, err := FormatJSONString(str)
	if err != nil {
		return
	}

	fmt.Printf("%s:\n%s\n", prefix, formattedLogJSON)

	return
}

func PrintFormatJSON(obj interface{}, prefix string) (err error) {
	formattedLogJSON, err := FormatJSON(obj)
	if err != nil {
		return
	}

	fmt.Printf("%s:\n%s\n", prefix, formattedLogJSON)

	return
}
