// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package utils

import (
	"fmt"

	jsoniter "github.com/json-iterator/go"
)

// ObjectToJSON 将对象转换为JSON字符串
func ObjectToJSON(obj interface{}) string {
	jsonBytes, _ := jsoniter.Marshal(obj)
	return string(jsonBytes)
}

// ObjectToByte 将对象转换为字节数组
func ObjectToByte(obj interface{}) []byte {
	jsonBytes, _ := jsoniter.Marshal(obj)
	return jsonBytes
}

// SubtractStrings 返回list1中不包含list2元素的差集（保留list1原始顺序）
func SubtractStrings(list1, list2 []string) []string {
	excludeSet := make(map[string]struct{})
	for _, s := range list2 {
		excludeSet[s] = struct{}{}
	}
	// 过滤保留不在排除集合中的元素
	var result []string
	for _, s := range list1 {
		if _, exists := excludeSet[s]; !exists {
			result = append(result, s)
		}
	}
	return result
}

// UniqueStrings 去重函数
func UniqueStrings(input []string) []string {
	// 创建一个映射来存储唯一的字符串
	uniqueMap := make(map[string]struct{})
	var uniqueList []string

	for _, str := range input {
		if _, exists := uniqueMap[str]; !exists {
			if str == "" {
				continue
			}
			uniqueMap[str] = struct{}{}          // 将字符串标记为已存在
			uniqueList = append(uniqueList, str) // 添加到结果列表
		}
	}

	return uniqueList
}

// StringToObject 将JSON字符串转换为对象
func StringToObject(jsonStr string, obj interface{}) error {
	err := jsoniter.Unmarshal([]byte(jsonStr), obj)
	return err
}

// CompareStringSliceLists 比较两个字符串切片列表，返回不同的元素
func CompareStringSliceLists(list1, list2 []string) []string {
	var result []string
	// 创建一个映射来存储list1中的元素
	map1 := make(map[string]bool)
	for _, item := range list1 {
		map1[item] = true
	}
	// 检查list2中的元素是否存在于map1中
	for _, item := range list2 {
		if !map1[item] {
			result = append(result, item)
		}
	}
	return result
}

// FindMissingElements 找出list2中缺少的list1元素（假设list2是list1的子集）
func FindMissingElements(list1, list2 []string) []string {
	// 创建存在映射表
	present := make(map[string]bool)
	for _, item := range list2 {
		present[item] = true
	}

	// 检查并收集缺失元素
	var missing []string
	seen := make(map[string]bool) // 用于结果去重
	for _, item := range list1 {
		if !present[item] && !seen[item] {
			missing = append(missing, item)
			seen[item] = true
		}
	}
	return missing
}

// ToString 将对象转换为字符串
func ToString(value interface{}) string {
	if value == nil {
		return ""
	}
	return fmt.Sprintf("%v", value)
}
