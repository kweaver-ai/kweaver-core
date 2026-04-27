package rest

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/go-playground/assert"
	. "github.com/smartystreets/goconvey/convey"
)

func check(jsonStr string) (*JSONValueDesc, error) {
	var jsonV interface{}
	err := json.Unmarshal([]byte(jsonStr), &jsonV)
	if err != nil {
		return nil, err
	}

	objDesc := make(map[string]*JSONValueDesc)
	objDesc["client_id"] = &JSONValueDesc{Kind: reflect.String, Required: true}
	objDesc["redirect_uri"] = &JSONValueDesc{Kind: reflect.String, Required: true}
	objDesc["response_type"] = &JSONValueDesc{Kind: reflect.Float64, Required: true}
	objDesc["scope"] = &JSONValueDesc{Kind: reflect.Bool, Required: true}

	udidsElementObjDesc := make(map[string]*JSONValueDesc)
	udidsElementObjDesc["id"] = &JSONValueDesc{Kind: reflect.String, Required: true}
	udidsElementObjDesc["params"] = &JSONValueDesc{Kind: reflect.Map, Required: true}
	udidsElementDesc := &JSONValueDesc{Kind: reflect.Map, Required: true, ValueDesc: udidsElementObjDesc}
	objDesc["udids"] = &JSONValueDesc{Kind: reflect.Slice, Required: false, ValueDesc: map[string]*JSONValueDesc{"element": udidsElementDesc}}

	credentialObjDesc := make(map[string]*JSONValueDesc)
	credentialObjDesc["id"] = &JSONValueDesc{Kind: reflect.String, Required: true}
	credentialObjDesc["params"] = &JSONValueDesc{Kind: reflect.Map, Required: true}
	objDesc["credential"] = &JSONValueDesc{Kind: reflect.Map, Required: true, ValueDesc: credentialObjDesc}

	reqParamsDesc := &JSONValueDesc{Kind: reflect.Map, Required: true, ValueDesc: objDesc}
	err = CheckJSONValue("body", jsonV, reqParamsDesc)
	return reqParamsDesc, err
}

func TestCheckJSONValue(t *testing.T) {
	Convey("CheckJSONValue", t, func() {
		Convey("正确", func() {
			jsonStr := `{
				"client_id": "43ea6470-0117-4ba6-a33a-76baa27dc8d7",
				"redirect_uri": "https:127.0.0.1:9010/callback",
				"response_type": 6666,
				"scope": true,
				"udids": [{
					"id": "555",
					"params": {
					  "ticket": "f1fd967472f847438ed9408f431c0bee@!@sso_as"
					}
				  }],
				"credential": {
				  "id": "bqkj",
				  "params": {
					"ticket": "f1fd967472f847438ed9408f431c0bee@!@sso_as"
				  }
				}
			  }`
			r, err := check(jsonStr)
			assert.Equal(t, err, nil)
			assert.Equal(t, r.ValueDesc["udids"].Exist, true)
		})

		Convey("类型错误", func() {
			jsonStr := `{
				"client_id": "43ea6470-0117-4ba6-a33a-76baa27dc8d7",
				"redirect_uri": "https:127.0.0.1:9010/callback",
				"response_type": 6666,
				"scope": true,
				"udids": [{
					"id": 555,
					"params": {
					  "ticket": "f1fd967472f847438ed9408f431c0bee@!@sso_as"
					}
				  }],
				"credential": {
				  "id": "bqkj",
				  "params": {
					"ticket": "f1fd967472f847438ed9408f431c0bee@!@sso_as"
				  }
				}
			  }`
			_, err := check(jsonStr)
			assert.NotEqual(t, err, nil)
		})

		Convey("缺少必须参数", func() {
			jsonStr := `{
				"redirect_uri": "https:127.0.0.1:9010/callback",
				"response_type": 6666,
				"scope": true,
				"udids": [{
					"id": "555",
					"params": {
					  "ticket": "f1fd967472f847438ed9408f431c0bee@!@sso_as"
					}
				  }],
				"credential": {
				  "id": "bqkj",
				  "params": {
					"ticket": "f1fd967472f847438ed9408f431c0bee@!@sso_as"
				  }
				}
			  }`
			_, err := check(jsonStr)
			assert.NotEqual(t, err, nil)
		})

		Convey("未传非必须参数", func() {
			jsonStr := `{
				"client_id": "43ea6470-0117-4ba6-a33a-76baa27dc8d7",
				"redirect_uri": "https:127.0.0.1:9010/callback",
				"response_type": 6666,
				"scope": true,
				"credential": {
				  "id": "bqkj",
				  "params": {
					"ticket": "f1fd967472f847438ed9408f431c0bee@!@sso_as"
				  }
				}
			  }`
			r, err := check(jsonStr)
			assert.Equal(t, err, nil)
			assert.Equal(t, r.ValueDesc["udids"].Exist, false)
		})
	})
}
