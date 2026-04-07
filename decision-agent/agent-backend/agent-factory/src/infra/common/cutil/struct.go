package cutil

import (
	"reflect"

	"github.com/bytedance/sonic"
	"github.com/pkg/errors"
)

var errArgWrong = errors.New("arg wrong")

// CopyStructUseJSON 使用json来复制结构体，会覆盖dst中的原有值
// 【注意】：如果结构体中含有interface{}类型的字段，复制前后interface{}字段的内部类型可能会发生变化
// dst: 目标结构体指针
// src: 源结构体
func CopyStructUseJSON(dst, src interface{}) (err error) {
	// 1. 判断dst, src是否为nil
	if dst == nil || src == nil {
		err = errors.Wrap(errArgWrong, "CopyStructUseJSON: dst and src can not be nil")
		return
	}

	// 2. 判断dst是否为指针
	dstVal := reflect.ValueOf(dst)

	if dstVal.Kind() != reflect.Ptr {
		panic("CopyStructUseJSON: dst reflect kind need to be a reflect.Ptr")
	}

	// 3. 判断dst是否为nil
	if dstVal.IsNil() {
		panic("CopyStructUseJSON: dstVal.IsNil()==true")
	}

	// 4. 判断dst, src底层是否是struct
	dstVal = dstVal.Elem()
	if dstVal.Kind() != reflect.Struct {
		panic("CopyStructUseJSON: dst need to be a struct")
	}

	srcVal := reflect.ValueOf(src)
	if srcVal.Kind() == reflect.Ptr {
		srcVal = srcVal.Elem()
	}

	if srcVal.Kind() != reflect.Struct {
		panic("CopyStructUseJSON: src need to be a struct")
	}

	// 5. 使用json来复制结构体
	// json := JSON()

	bys, err := sonic.Marshal(src)
	if err != nil {
		return
	}

	err = sonic.Unmarshal(bys, dst)

	return
}

func CopyUseJSON(dst, src interface{}) (err error) {
	// 使用json来复制
	// json := JSON()
	bys, err := sonic.Marshal(src)
	if err != nil {
		return
	}

	err = sonic.Unmarshal(bys, dst)

	return
}
