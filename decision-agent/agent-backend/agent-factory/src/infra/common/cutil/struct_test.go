package cutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type Person2 struct {
	Name string
	Age  int
}

type Student struct {
	Name   string
	Age    int8
	School string
}

// CopyStructUseJSON 使用json序列化反序列化复制结构体，类型不同的字段可能会复制，比如这里的age
func TestCopyStructUseJson(t *testing.T) {
	t.Parallel()

	p1 := Person2{
		Name: "tom",
		Age:  19,
	}
	s1 := Student{}

	err := CopyStructUseJSON(&s1, &p1)
	if err != nil {
		t.Error(err)
	}

	assert.True(t, s1.Name == "tom")
	assert.True(t, s1.Age == 19)
	assert.True(t, s1.School == "")
}

type Inner struct {
	Name string
	Age  int
}

type Outer struct {
	Inner
	School string
}

// CopyStructUseJSON 复制结构体，结构体中嵌套结构体
func TestCopyStructUseJson_compositeStruct(t *testing.T) {
	t.Parallel()

	p1 := Person2{
		Name: "tom",
		Age:  19,
	}
	s1 := Outer{School: "school1"}

	err := CopyStructUseJSON(&s1, &p1)
	if err != nil {
		t.Error(err)
	}
	// fmt.Printf("%#v\n", s1)
	assert.True(t, s1.Name == "tom")
	assert.True(t, s1.Age == 19)
	assert.True(t, s1.School == "school1")
}

func TestCopyStructUseJson_Error(t *testing.T) {
	t.Parallel()

	p1 := Person2{
		Name: "tom",
		Age:  19,
	}
	s1 := Student{}

	err := CopyStructUseJSON(nil, nil)
	assert.Equal(t, err.Error(), "CopyStructUseJSON: dst and src can not be nil: arg wrong")

	err = CopyStructUseJSON(nil, &p1)
	assert.Equal(t, err.Error(), "CopyStructUseJSON: dst and src can not be nil: arg wrong")

	err = CopyStructUseJSON(&s1, nil)
	assert.Equal(t, err.Error(), "CopyStructUseJSON: dst and src can not be nil: arg wrong")

	assert.PanicsWithValue(t, "CopyStructUseJSON: dst reflect kind need to be a reflect.Ptr", func() {
		_ = CopyStructUseJSON(s1, &p1)
	})

	var s2 *Student

	assert.PanicsWithValue(t, "CopyStructUseJSON: dstVal.IsNil()==true", func() {
		_ = CopyStructUseJSON(s2, &p1)
	})
}
