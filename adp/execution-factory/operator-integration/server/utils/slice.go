package utils

// SliceToInterface 将任意类型的切片转换为接口切片
func SliceToInterface[T any](slice []T) []interface{} {
	interfaces := make([]interface{}, len(slice))
	for i, item := range slice {
		interfaces[i] = item
	}
	return interfaces
}

// RemoveStringFromSlice 剔除字符串列表中的某个元素
func RemoveStringFromSlice(strings []string, target string) []string {
	var result []string
	for _, str := range strings {
		if str != target {
			result = append(result, str)
		}
	}
	return result
}

// CalculateIntersection 计算两个字符串切片的交集，优化性能
func CalculateIntersection(list1, list2 []string) []string {
	// 如果任一列表为空，直接返回空列表
	if len(list1) == 0 || len(list2) == 0 {
		return []string{}
	}
	
	// 优化：使用较小的列表构建映射，减少内存使用
	if len(list1) > len(list2) {
		list1, list2 = list2, list1
	}

	// 构建较小列表的映射
	mapping := make(map[string]struct{}, len(list1))
	for _, id := range list1 {
		mapping[id] = struct{}{}
	}

	// 预分配结果切片，最大可能为较小列表的长度
	result := make([]string, 0, len(list1))
	for _, id := range list2 {
		if _, exists := mapping[id]; exists {
			result = append(result, id)
		}
	}

	return result
}
