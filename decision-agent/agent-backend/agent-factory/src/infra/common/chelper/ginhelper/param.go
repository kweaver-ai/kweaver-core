package ginhelper

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/capierr"
)

func GetParmID(c *gin.Context) (id string, err error) {
	id = c.Param("id")
	if id == "" {
		err = capierr.New400Err(c, "id is empty")
		return
	}

	return
}

type ParamIDInt64 struct {
	ID int64 `json:"id" binding:"required" uri:"id"`
}

func GetParmIDInt64(c *gin.Context) (id int64, err error) {
	var param ParamIDInt64
	if err = c.ShouldBindUri(&param); err != nil {
		err = capierr.New400Err(c, "id is empty")
		return
	}

	id = param.ID

	return
}

type ParamKey struct {
	Key string `json:"key" binding:"required" uri:"key"`
}

func GetParmKey(c *gin.Context) (key string, err error) {
	var param ParamKey
	if err = c.ShouldBindUri(&param); err != nil {
		err = capierr.New400Err(c, "key is empty")
		return
	}

	key = param.Key

	return
}

func GetParmInt64(c *gin.Context, key string) (val int64, err error) {
	valStr := c.Param(key)

	if valStr == "" {
		err = capierr.New400Err(c, fmt.Sprintf(`"%s" is empty`, key))
		return
	}

	val, err = strconv.ParseInt(valStr, 10, 64)
	if err != nil {
		err = capierr.New400Err(c, fmt.Sprintf(`"%s" is not integer`, key))
		return
	}

	return
}
