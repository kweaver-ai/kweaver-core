package sqlhelper2

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_struct2SqlValPairsMapByTag(t *testing.T) {
	t.Parallel()

	// 1. 基本类型
	type User struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	user := User{
		Name: "tom",
		Age:  19,
	}
	pairs, err := struct2SQLValPairsMapByTag(user, "json")

	assert.Nil(t, err)

	assert.Equal(t, "tom", pairs["name"])
	assert.Equal(t, 19, pairs["age"])

	// 2. 指针类型
	type User2 struct {
		Name   string  `json:"name"`
		Age    int     `json:"age"`
		School *string `json:"school"`
	}

	school := "school1"
	user2 := User2{
		Name:   "tom",
		Age:    19,
		School: &school,
	}

	pairs, err = struct2SQLValPairsMapByTag(user2, "json")

	assert.Nil(t, err)

	assert.Equal(t, 3, len(pairs))
	assert.Equal(t, "tom", pairs["name"])
	assert.Equal(t, 19, pairs["age"])
	assert.Equal(t, "school1", pairs["school"])

	// 3. 嵌套结构体
	type User3 struct {
		Name   string `json:"name"`
		Age    int    `json:"age"`
		School struct {
			Name    string `json:"name"`
			Address string `json:"address"`
		} `json:"school"`
	}

	user3 := User3{
		Name: "tom",
		Age:  19,
		School: struct {
			Name    string `json:"name"`
			Address string `json:"address"`
		}{
			Name:    "school1",
			Address: "address1",
		},
	}

	pairs, err = struct2SQLValPairsMapByTag(user3, "json") //nolint:staticcheck,ineffassign
	assert.Equal(t, err.Error(), "only support string number *string *number, but field School is struct", "ib.ToInsertSQL() failed")

	// 4. 零值
	type User4 struct {
		Name   string  `json:"name"`
		Age    int     `json:"age"`
		School *string `json:"school"`
	}

	user4 := User4{
		Name:   "",
		Age:    0,
		School: nil,
	}

	pairs, err = struct2SQLValPairsMapByTag(user4, "json")

	assert.Nil(t, err)

	assert.Equal(t, 0, len(pairs))

	// 5. 非基本类型
	type User5 struct {
		Name   string  `json:"name"`
		Age    bool    `json:"age"`
		School *string `json:"school"`
	}

	user5 := User5{
		Name:   "tom",
		Age:    true,
		School: &school,
	}

	pairs, err = struct2SQLValPairsMapByTag(user5, "json") //nolint:staticcheck,ineffassign

	assert.Equal(t, err.Error(), "only support string number *string *number, but field Age is bool", "ib.ToInsertSQL() failed")
}
