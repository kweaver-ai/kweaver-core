package cutil

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

// 单元测试用例
func TestDifference(t *testing.T) {
	t.Parallel()

	// 测试差集为空的情况
	a := []string{"a", "b", "c"}
	b := []string{}
	expected := []string{"a", "b", "c"}
	result := Difference(a, b)

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Difference failed, expected %v but got %v", expected, result)
	}

	// 测试差集为非空的情况
	a = []string{"a", "b", "c"}
	b = []string{"b", "c"}
	expected = []string{"a"}
	result = Difference(a, b)

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Difference failed, expected %v but got %v", expected, result)
	}
}

// 单元测试用例
func TestIntersection(t *testing.T) {
	t.Parallel()

	// 测试交集为空的情况
	a := []string{"a", "b", "c"}
	b := []string{}
	expected := []string{}
	result := Intersection(a, b)

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Intersection failed, expected %v but got %v", expected, result)
	}

	// 测试交集为非空的情况
	a = []string{"a", "b", "c"}
	b = []string{"b", "c"}
	expected = []string{"b", "c"}
	result = Intersection(a, b)

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Intersection failed, expected %v but got %v", expected, result)
	}

	//	测试a为空的情况
	a = []string{}
	b = []string{"b", "c"}
	expected = []string{}
	result = Difference(a, b)

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Difference failed, expected %v but got %v", expected, result)
	}
}

// 单元测试用例
func TestExists(t *testing.T) {
	t.Parallel()

	// 测试元素存在的情况
	a := []string{"a", "b", "c"}
	v := "b"
	expected := true
	result := Exists(a, v)

	if result != expected {
		t.Errorf("Exists failed, expected %v but got %v", expected, result)
	}

	// 测试元素不存在的情况
	v = "d"
	expected = false
	result = Exists(a, v)

	if result != expected {
		t.Errorf("Exists failed, expected %v but got %v", expected, result)
	}
}

// IsDuplicationGeneric ut
func TestIsDuplicationGeneric(t *testing.T) {
	t.Parallel()

	// 1. 测试有重复元素的情况
	a := []int{1, 2, 3, 1, 2, 3}
	result := IsDuplicationGeneric(a)

	assert.Equal(t, result, true)

	// 2. 测试无重复元素的情况
	a = []int{1, 2, 3}
	result = IsDuplicationGeneric(a)

	assert.Equal(t, result, false)

	// 3. 测试string类型
	a1 := []string{"a", "b", "c", "a", "b", "c"}
	result1 := IsDuplicationGeneric(a1)

	assert.Equal(t, result1, true)

	// 4. 测试int64类型
	a2 := []int64{1, 2, 3, 1, 2, 3}
	result2 := IsDuplicationGeneric(a2)

	assert.Equal(t, result2, true)

	// 5. 测试float64类型
	a3 := []float64{1.1, 2.2, 3.3, 1.1, 2.2, 3.3}
	result3 := IsDuplicationGeneric(a3)

	assert.Equal(t, result3, true)

	// 6. 测试bool类型
	a4 := []bool{true, false, true, false}
	result4 := IsDuplicationGeneric(a4)

	assert.Equal(t, result4, true)

	// 7. 测试空切片
	var a5 []string
	result5 := IsDuplicationGeneric(a5)

	assert.Equal(t, result5, false)

	// 8. 测试nil切片
	var a6 []string

	result6 := IsDuplicationGeneric(a6)

	assert.Equal(t, result6, false)

	// 9. 测试结构体切片
	type Person struct {
		Name string
		Age  int
	}

	a7 := []Person{{"a", 1}, {"b", 2}, {"c", 3}, {"a", 1}, {"b", 2}, {"c", 3}}
	result7 := IsDuplicationGeneric(a7)

	assert.Equal(t, result7, true)
}

// TestDeduplGeneric 测试 DeduplGeneric 函数
func TestDeduplGeneric(t *testing.T) {
	t.Parallel()

	// 测试空切片
	var a1 []int

	expected1 := []int{}
	result1 := DeduplGeneric(a1)

	assert.Equal(t, expected1, result1, "DeduplGeneric failed on empty slice")

	// 测试无重复元素
	a2 := []int{1, 2, 3}
	expected2 := []int{1, 2, 3}
	result2 := DeduplGeneric(a2)

	assert.Equal(t, expected2, result2, "DeduplGeneric failed on no duplicates")

	// 测试所有元素重复
	a3 := []int{1, 1, 1}
	expected3 := []int{1}
	result3 := DeduplGeneric(a3)

	assert.Equal(t, expected3, result3, "DeduplGeneric failed on all elements are the same")

	// 测试部分重复元素
	a4 := []int{1, 2, 2, 3, 1}
	expected4 := []int{1, 2, 3}
	result4 := DeduplGeneric(a4)

	assert.Equal(t, expected4, result4, "DeduplGeneric failed on some duplicates")
}

// TestDeduplPtrGeneric 测试 DeduplPtrGeneric 函数
func TestDeduplPtrGeneric(t *testing.T) {
	t.Parallel()

	// 测试空切片
	var a1 []*int

	expected1 := []*int{}
	result1 := DeduplPtrGeneric(a1)

	assert.Equal(t, expected1, result1, "DeduplPtrGeneric failed on empty slice")

	// 测试无重复元素
	v1, v2, v3 := 1, 2, 3
	a2 := []*int{&v1, &v2, &v3}
	expected2 := []*int{&v1, &v2, &v3}
	result2 := DeduplPtrGeneric(a2)

	assert.Equal(t, expected2, result2, "DeduplPtrGeneric failed on no duplicates")

	// 测试所有元素重复
	v4 := 1
	a3 := []*int{&v4, &v4, &v4}
	expected3 := []*int{&v4}
	result3 := DeduplPtrGeneric(a3)

	assert.Equal(t, expected3, result3, "DeduplPtrGeneric failed on all elements are the same")

	// 测试部分重复元素
	v5, v6, v7, v8, v9 := 1, 2, 2, 3, 1
	a4 := []*int{&v5, &v6, &v7, &v8, &v9}
	expected4 := []*int{&v5, &v6, &v8}
	result4 := DeduplPtrGeneric(a4)

	assert.Equal(t, expected4, result4, "DeduplPtrGeneric failed on some duplicates")

	// 测试结构体指针去重
	t.Run("struct_pointer_deduplication", func(t *testing.T) {
		type Employee struct {
			ID   int
			Name string
		}

		emp1 := &Employee{ID: 1, Name: "张三"}
		emp2 := &Employee{ID: 2, Name: "李四"}
		emp3 := &Employee{ID: 1, Name: "张三"} // 与emp1相同值但不同指针
		emp4 := &Employee{ID: 3, Name: "王五"}
		emp5 := emp2 // 相同指针

		a := []*Employee{emp1, emp2, emp3, emp4, emp5}
		expected := []*Employee{emp1, emp2, emp4} // emp3被去重（与emp1相同值），emp5被去重（与emp2相同指针）
		result := DeduplPtrGeneric(a)

		assert.Equal(t, expected, result, "DeduplPtrGeneric failed for struct pointers")
	})

	t.Run("struct_pointer_all_same", func(t *testing.T) {
		type Config struct {
			Key   string
			Value string
		}

		cfg1 := &Config{Key: "env", Value: "production"}
		cfg2 := &Config{Key: "env", Value: "production"} // 相同值不同指针
		cfg3 := &Config{Key: "env", Value: "production"} // 相同值不同指针

		a := []*Config{cfg1, cfg2, cfg3}
		expected := []*Config{cfg1} // 只保留第一个
		result := DeduplPtrGeneric(a)

		assert.Equal(t, expected, result, "DeduplPtrGeneric failed for all same struct pointers")
	})
}

// TestDifferenceGeneric 测试 DifferenceGeneric 函数
func TestDifferenceGeneric(t *testing.T) {
	t.Parallel()

	a := []int{1, 2, 3, 4}
	b := []int{3, 4}

	expected := []int{1, 2}
	result := DifferenceGeneric(a, b)

	assert.Equal(t, expected, result, "DifferenceGeneric failed")
}

// TestDifferencePtrGeneric 测试 DifferencePtrGeneric 函数
func TestDifferencePtrGeneric(t *testing.T) {
	t.Parallel()

	// 1. 基本功能测试 - 正常差集
	t.Run("basic_difference", func(t *testing.T) {
		v1, v2, v3, v4 := 1, 2, 3, 4
		a := []*int{&v1, &v2, &v3, &v4}
		b := []*int{&v3, &v4}

		expected := []*int{&v1, &v2}
		result := DifferencePtrGeneric(a, b)

		assert.Equal(t, expected, result, "DifferencePtrGeneric failed")
	})

	// 2. 测试空切片 a
	t.Run("empty_slice_a", func(t *testing.T) {
		var a []*int

		v1, v2 := 1, 2
		b := []*int{&v1, &v2}

		expected := []*int{}
		result := DifferencePtrGeneric(a, b)

		assert.Equal(t, expected, result, "DifferencePtrGeneric failed for empty slice a")
	})

	// 3. 测试空切片 b
	t.Run("empty_slice_b", func(t *testing.T) {
		v1, v2, v3 := 1, 2, 3
		a := []*int{&v1, &v2, &v3}

		var b []*int

		expected := []*int{&v1, &v2, &v3}
		result := DifferencePtrGeneric(a, b)

		assert.Equal(t, expected, result, "DifferencePtrGeneric failed for empty slice b")
	})

	// 4. 测试两个空切片
	t.Run("both_empty_slices", func(t *testing.T) {
		var a []*int

		var b []*int

		expected := []*int{}
		result := DifferencePtrGeneric(a, b)

		assert.Equal(t, expected, result, "DifferencePtrGeneric failed for both empty slices")
	})

	// 5. 测试 nil 切片
	t.Run("nil_slices", func(t *testing.T) {
		var a []*int = nil

		var b []*int = nil

		expected := []*int{}
		result := DifferencePtrGeneric(a, b)

		assert.Equal(t, expected, result, "DifferencePtrGeneric failed for nil slices")
	})

	// 6. 测试完全相同的切片
	t.Run("identical_slices", func(t *testing.T) {
		v1, v2, v3 := 1, 2, 3
		a := []*int{&v1, &v2, &v3}
		b := []*int{&v1, &v2, &v3}

		expected := []*int{}
		result := DifferencePtrGeneric(a, b)

		assert.Equal(t, expected, result, "DifferencePtrGeneric failed for identical slices")
	})

	// 7. 测试 a 是 b 的子集
	t.Run("a_subset_of_b", func(t *testing.T) {
		v1, v2, v3, v4 := 1, 2, 3, 4
		a := []*int{&v1, &v2}
		b := []*int{&v1, &v2, &v3, &v4}

		expected := []*int{}
		result := DifferencePtrGeneric(a, b)

		assert.Equal(t, expected, result, "DifferencePtrGeneric failed when a is subset of b")
	})

	// 8. 测试 b 是 a 的子集
	t.Run("b_subset_of_a", func(t *testing.T) {
		v1, v2, v3, v4 := 1, 2, 3, 4
		a := []*int{&v1, &v2, &v3, &v4}
		b := []*int{&v2, &v3}

		expected := []*int{&v1, &v4}
		result := DifferencePtrGeneric(a, b)

		assert.Equal(t, expected, result, "DifferencePtrGeneric failed when b is subset of a")
	})

	// 9. 测试没有交集的切片
	t.Run("no_intersection", func(t *testing.T) {
		v1, v2, v3, v4 := 1, 2, 3, 4
		a := []*int{&v1, &v2}
		b := []*int{&v3, &v4}

		expected := []*int{&v1, &v2}
		result := DifferencePtrGeneric(a, b)

		assert.Equal(t, expected, result, "DifferencePtrGeneric failed for no intersection")
	})

	// 10. 测试重复元素
	t.Run("duplicate_elements", func(t *testing.T) {
		v1, v2, v3 := 1, 2, 3
		a := []*int{&v1, &v2, &v1, &v3} // v1 重复
		b := []*int{&v2}

		expected := []*int{&v1, &v1, &v3} // 保留所有不在 b 中的元素
		result := DifferencePtrGeneric(a, b)

		assert.Equal(t, expected, result, "DifferencePtrGeneric failed for duplicate elements")
	})

	// 11. 测试字符串类型
	t.Run("string_type", func(t *testing.T) {
		s1, s2, s3, s4 := "apple", "banana", "cherry", "date"
		a := []*string{&s1, &s2, &s3, &s4}
		b := []*string{&s2, &s4}

		expected := []*string{&s1, &s3}
		result := DifferencePtrGeneric(a, b)

		assert.Equal(t, expected, result, "DifferencePtrGeneric failed for string type")
	})

	// 12. 测试布尔类型
	t.Run("bool_type", func(t *testing.T) {
		t1, t2, f1, f2 := true, true, false, false
		a := []*bool{&t1, &f1, &t2}
		b := []*bool{&f2} // false

		expected := []*bool{&t1, &t2} // 只保留 true 值
		result := DifferencePtrGeneric(a, b)

		assert.Equal(t, expected, result, "DifferencePtrGeneric failed for bool type")
	})

	// 13. 测试单个元素
	t.Run("single_element", func(t *testing.T) {
		v1, v2 := 42, 43
		a := []*int{&v1}
		b := []*int{&v2}

		expected := []*int{&v1}
		result := DifferencePtrGeneric(a, b)

		assert.Equal(t, expected, result, "DifferencePtrGeneric failed for single element")
	})

	// 14. 测试大量元素
	t.Run("large_slices", func(t *testing.T) {
		// 创建大量元素进行测试
		aVals := make([]int, 1000)
		bVals := make([]int, 500)

		for i := 0; i < 1000; i++ {
			aVals[i] = i
		}

		for i := 0; i < 500; i++ {
			bVals[i] = i * 2 // 偶数
		}

		a := make([]*int, len(aVals))
		b := make([]*int, len(bVals))

		for i := range aVals {
			a[i] = &aVals[i]
		}

		for i := range bVals {
			b[i] = &bVals[i]
		}

		result := DifferencePtrGeneric(a, b)

		// 验证结果中不包含任何偶数（0, 2, 4, ...998）
		expectedLen := 1000 - 500 // 1000个元素减去500个偶数
		assert.Equal(t, expectedLen, len(result), "DifferencePtrGeneric failed for large slices length")

		// 验证结果中都是奇数
		for _, ptr := range result {
			assert.True(t, *ptr%2 == 1, "Result should only contain odd numbers")
		}
	})

	// 15. 测试相同值但不同指针
	t.Run("same_values_different_pointers", func(t *testing.T) {
		// 创建相同值但不同的指针
		a1, a2, a3 := 1, 2, 3
		b1, b2 := 1, 2 // 值相同但是不同的变量

		a := []*int{&a1, &a2, &a3}
		b := []*int{&b1, &b2}

		// 因为是按值比较，所以结果应该只包含 &a3
		expected := []*int{&a3}
		result := DifferencePtrGeneric(a, b)

		assert.Equal(t, expected, result, "DifferencePtrGeneric failed for same values different pointers")
	})

	// 16. 增加结构体指针数组的测试 - 值
	t.Run("struct_pointer_slice", func(t *testing.T) {
		type Person struct {
			Name string
			Age  int
		}

		// 创建测试数据
		p1 := &Person{Name: "Alice", Age: 30}
		p2 := &Person{Name: "Bob", Age: 25}
		p3 := &Person{Name: "Charlie", Age: 35}
		p4 := &Person{Name: "Alice", Age: 30} // 与p1相同的值但不同的指针

		a := []*Person{p1, p2, p3}
		b := []*Person{p4} // 相同值但不同指针

		// 测试结构体指针差集 - p1应该被过滤掉，因为p4与p1有相同的值
		expected := []*Person{p2, p3}
		result := DifferencePtrGeneric(a, b)

		assert.Equal(t, expected, result, "DifferencePtrGeneric failed for struct pointers with same values")
	})

	// 17. 增加结构体指针数组的测试 - 指针
	t.Run("struct_pointer_slice_identical", func(t *testing.T) {
		type Product struct {
			ID    int
			Name  string
			Price float64
		}

		// 创建相同的结构体指针
		prod1 := &Product{ID: 1, Name: "Laptop", Price: 999.99}
		prod2 := &Product{ID: 2, Name: "Mouse", Price: 29.99}
		prod3 := &Product{ID: 3, Name: "Keyboard", Price: 79.99}

		a := []*Product{prod1, prod2, prod3}
		b := []*Product{prod2} // 使用相同的指针

		expected := []*Product{prod1, prod3}
		result := DifferencePtrGeneric(a, b)

		assert.Equal(t, expected, result, "DifferencePtrGeneric failed for struct pointers with identical pointers")
	})

	// 18. 增加结构体指针数组的测试 - 空切片
	t.Run("struct_pointer_slice_empty", func(t *testing.T) {
		type Book struct {
			Title  string
			Author string
		}

		book1 := &Book{Title: "Go语言程序设计", Author: "Alan"}
		book2 := &Book{Title: "数据结构与算法", Author: "Bob"}

		a := []*Book{book1, book2}

		var b []*Book // 空切片

		expected := []*Book{book1, book2}
		result := DifferencePtrGeneric(a, b)

		assert.Equal(t, expected, result, "DifferencePtrGeneric failed for struct pointers with empty b slice")
	})

	// 19. 增加结构体指针数组的测试 - 复杂情况
	t.Run("struct_pointer_slice_complex", func(t *testing.T) {
		type User struct {
			ID       int
			Username string
			Email    string
			Active   bool
		}

		user1 := &User{ID: 1, Username: "alice", Email: "alice@example.com", Active: true}
		user2 := &User{ID: 2, Username: "bob", Email: "bob@example.com", Active: false}
		user3 := &User{ID: 3, Username: "charlie", Email: "charlie@example.com", Active: true}
		user4 := &User{ID: 4, Username: "david", Email: "david@example.com", Active: false}
		user5 := &User{ID: 2, Username: "bob", Email: "bob@example.com", Active: false} // 与user2相同值

		a := []*User{user1, user2, user3, user4}
		b := []*User{user5, user3} // user5与user2值相同，user3是同一个指针

		expected := []*User{user1, user4} // user2被过滤（因为user5有相同值），user3被过滤（相同指针）
		result := DifferencePtrGeneric(a, b)

		assert.Equal(t, expected, result, "DifferencePtrGeneric failed for complex struct pointers")
	})
}

// TestExistsGeneric 测试 ExistsGeneric 函数
func TestExistsGeneric(t *testing.T) {
	t.Parallel()

	// 测试元素存在的情况
	a := []int{1, 2, 3}
	v := 2
	expected := true
	result := ExistsGeneric(a, v)

	assert.Equal(t, expected, result, "ExistsGeneric failed for existing element")

	// 测试元素不存在的情况
	v = 4
	expected = false
	result = ExistsGeneric(a, v)

	assert.Equal(t, expected, result, "ExistsGeneric failed for non-existing element")
}

// TestSliceToPtrSlice 测试 SliceToPtrSlice 函数
func TestSliceToPtrSlice(t *testing.T) {
	t.Parallel()

	a := []int{1, 2, 3}
	expected := []*int{&a[0], &a[1], &a[2]}
	result := SliceToPtrSlice(a)

	assert.Equal(t, expected, result, "SliceToPtrSlice failed")
}

// TestSliceEqualGeneric 测试 IsSliceEqualGeneric 函数
func TestSliceEqualGeneric(t *testing.T) {
	t.Parallel()

	// 1. 测试相等的情况（顺序一样）
	a := []int{1, 2, 3}
	b := []int{1, 2, 3}
	expected := true
	result := IsSliceEqualGeneric(a, b)

	assert.Equal(t, expected, result, "SliceEqual failed for equal slices")

	// 2. 测试相等的情况（顺序不一样）
	b = []int{3, 2, 1}
	expected = true
	result = IsSliceEqualGeneric(a, b)

	assert.Equal(t, expected, result, "SliceEqual failed for equal slices")

	// 3. 测试不相等的情况
	b = []int{1, 2, 4}
	expected = false
	result = IsSliceEqualGeneric(a, b)

	assert.Equal(t, expected, result, "SliceEqual failed for non-equal slices")

	// 4. 测试不相等的情况（长度不一样）
	b = []int{1, 2}
	expected = false
	result = IsSliceEqualGeneric(a, b)

	assert.Equal(t, expected, result, "SliceEqual failed for non-equal slices")

	// 5. 测试空切片
	b = []int{}
	expected = false
	result = IsSliceEqualGeneric(a, b)

	assert.Equal(t, expected, result, "SliceEqual failed for non-equal slices")

	// 6. 测试 nil
	b = nil
	expected = false
	result = IsSliceEqualGeneric(a, b)

	assert.Equal(t, expected, result, "SliceEqual failed for non-equal slices")

	// 7. 测试2个 nil
	var c []int

	var d []int

	expected = true
	result = IsSliceEqualGeneric(c, d)

	assert.Equal(t, expected, result, "SliceEqual failed for non-equal slices")
}

// TestTrimSliceGeneric 测试 TrimSliceGeneric 函数
func TestTrimSliceGeneric(t *testing.T) {
	t.Parallel()

	// 1. 测试空切片
	var a []string

	expected := []string(nil)
	result := TrimSliceGeneric(a, " ")

	assert.Equal(t, expected, result, "TrimSliceGeneric failed for empty slice")

	// 2. 测试无需修剪的情况
	a = []string{"a", "b", "c"}
	expected = []string{"a", "b", "c"}
	result = TrimSliceGeneric(a, " ")

	assert.Equal(t, expected, result, "TrimSliceGeneric failed for no trimming needed")

	// 3. 测试需要修剪空白字符的情况
	a = []string{" a ", " b ", " c "}
	expected = []string{"a", "b", "c"}
	result = TrimSliceGeneric(a, " ")

	assert.Equal(t, expected, result, "TrimSliceGeneric failed for trimming whitespace")

	// 4. 测试修剪后变为空字符串的情况
	a = []string{"a", "  ", "b", "   ", "c"}
	expected = []string{"a", "", "b", "", "c"}
	result = TrimSliceGeneric(a, " ")

	assert.Equal(t, expected, result, "TrimSliceGeneric failed for strings becoming empty after trim")
}

// TestRemoveEmptyStrFromSlice 测试 RemoveEmptyStrFromSlice 函数
func TestRemoveEmptyStrFromSlice(t *testing.T) {
	t.Parallel()

	// 1. 测试空切片
	var a []string

	expected := []string{}
	result := RemoveEmptyStrFromSlice(a)

	assert.Equal(t, expected, result, "RemoveEmptyStrFromSlice failed for empty slice")

	// 2. 测试无空字符串的情况
	a = []string{"a", "b", "c"}
	expected = []string{"a", "b", "c"}
	result = RemoveEmptyStrFromSlice(a)

	assert.Equal(t, expected, result, "RemoveEmptyStrFromSlice failed for no empty strings")

	// 3. 测试有空字符串的情况
	a = []string{"a", "", "b", "", "c"}
	expected = []string{"a", "b", "c"}
	result = RemoveEmptyStrFromSlice(a)

	assert.Equal(t, expected, result, "RemoveEmptyStrFromSlice failed for empty strings")
}
