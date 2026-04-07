package value

import (
	"fmt"
	"strings"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MapValue map[string]interface{}

type Setter func(v interface{})

type MapValueCallback func(walkContext *WalkContext, v interface{}) error

type WalkContext struct {
	path   []string
	Setter Setter
}

func NewWalkContext() *WalkContext {
	return &WalkContext{
		path:   nil,
		Setter: nil,
	}
}

func (c *WalkContext) pushPath(path string) {
	c.path = append(c.path, path)
}

func (c *WalkContext) Path() string {
	var b strings.Builder
	for i, s := range c.path {
		if !strings.HasPrefix(s, "[") && i != 0 {
			b.WriteString(".")
		}
		b.WriteString(s)
	}
	return b.String()
}

func (c *WalkContext) reset() {
	c.Setter = nil
	c.path = c.path[:len(c.path)-1]
}

func (val MapValue) Walk(callback MapValueCallback) error {
	return val.walk(NewWalkContext(), callback)
}

func (val MapValue) walk(walkContext *WalkContext, callback MapValueCallback) error {
	for k, v := range val {
		err := func() error {
			walkContext.pushPath(k)
			defer walkContext.reset()

			setter := func(v interface{}) {
				val[k] = v
			}
			return val.walkValue(walkContext, v, setter, callback)
		}()
		if err != nil {
			return err
		}
	}
	return nil
}

func (val MapValue) walkSlice(walkContext *WalkContext, slice []interface{}, callback MapValueCallback) error {
	for i, item := range slice {
		err := func() error {
			path := fmt.Sprintf("[%d]", i)
			walkContext.pushPath(path)
			defer walkContext.reset()

			setter := func(v interface{}) {
				slice[i] = v
			}
			return val.walkValue(walkContext, item, setter, callback)
		}()
		if err != nil {
			return err
		}
	}
	return nil
}

func (val MapValue) walkValue(walkContext *WalkContext, v interface{}, setter Setter, callback MapValueCallback) error {
	// 遍历  map
	if m, ok := v.(map[string]interface{}); ok {
		err := MapValue(m).walk(walkContext, callback)
		if err != nil {
			return err
		}
		return nil
	}
	// 遍历 slice
	if s, ok := v.([]interface{}); ok {
		err := val.walkSlice(walkContext, s, callback)
		if err != nil {
			return err
		}
		return nil
	}
	// 遍历 primitive.A
	if s, ok := v.(primitive.A); ok {
		err := val.walkSlice(walkContext, s, callback)
		if err != nil {
			return err
		}
		return nil
	}
	walkContext.Setter = setter
	return callback(walkContext, v)
}

type StringSetter func(s string)

type MapValueStringCallback func(walkContext *WalkContext, v string) error

func (val MapValue) WalkString(callback MapValueStringCallback) error {
	return val.walk(NewWalkContext(), func(walkContext *WalkContext, v interface{}) error {
		if s, ok := v.(string); ok {
			return callback(walkContext, s)
		}
		return nil
	})
}
