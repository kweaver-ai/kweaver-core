package utils

import (
	"testing"

	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/ecron/common"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
)

func TestErrorsNewECronError(t *testing.T) {
	err1 := &common.ECronError{}
	err2 := &common.ECronError{}

	Convey("function NewECronError() create and normal create", t, func() {
		err1 = NewECronError("for test", common.BadRequest, nil)
		err2 = &common.ECronError{
			Cause:   "for test",
			Code:    common.BadRequest,
			Detail:  nil,
			Message: NewConfiger().Lang().GetString("IDS_BAD_REQUEST"),
		}
		assert.Equal(t, err1, err2)
	})

	Convey("json marshal and unmarshal ECronError", t, func() {
		err1 = NewECronError("for test", common.BadRequest, map[string]interface{}{
			"conflicts": map[string]string{
				"job_id": "123",
			},
		})

		json1, err := jsoniter.Marshal(err1)
		assert.Equal(t, err, nil)

		json2, err := jsoniter.Marshal(gin.H{
			"cause":   "for test",
			"code":    common.BadRequest,
			"message": NewConfiger().Lang().GetString("IDS_BAD_REQUEST"),
			"detail": gin.H{
				"conflicts": gin.H{
					"job_id": "123",
				},
			},
		})
		assert.Equal(t, err, nil)

		err3 := common.ECronError{}
		err4 := common.ECronError{}
		err = jsoniter.Unmarshal(json1, &err3)
		assert.Equal(t, err, nil)

		err = jsoniter.Unmarshal(json2, &err4)
		assert.Equal(t, err, nil)
		assert.Equal(t, err3, err4)
	})
}
