package cutil

import (
	"bytes"
)

// DeepCopyMap 复制给定的原始map，并返回其深拷贝。
// 原始map中的任何改变不会影响到新的深拷贝map。
// originalMap: 原始的map[string]interface{}，
// 返回值是深拷贝的map以及可能遇到的错误信息。
func DeepCopyMap(originalMap map[string]interface{}) (map[string]interface{}, error) {
	var buf bytes.Buffer
	// 使用JSON编码器将原始map编码到缓冲区中。
	if err := JSON().NewEncoder(&buf).Encode(originalMap); err != nil {
		// 如果编码过程中产生错误，返回nil和错误信息。
		return nil, err
	}

	var copiedMap map[string]interface{}
	// 从缓冲区中解码到新的map中，实现深拷贝。
	if err := JSON().NewDecoder(&buf).Decode(&copiedMap); err != nil {
		// 如果解码过程中产生错误，返回nil和错误信息。
		return nil, err
	}

	// 返回深拷贝的map和nil错误信息。
	return copiedMap, nil
}
