package httphelper

import (
	"log"
	"os"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/cenvhelper"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
)

var logger = log.New(os.Stdout, "[chttpclient-DEBUG]", log.LstdFlags)

// ----------分割线-------------
type debugReqLogger struct {
	URL    string
	Method string
	Data   interface{}
}

func (l *debugReqLogger) Log() {
	logger.Printf("url:\t %s\n", l.URL)
	logger.Printf("method:\t %s\n", l.Method)

	dataJSONStr, _ := cutil.JSON().MarshalIndent(l.Data, "", "  ")
	logger.Printf("data:\n%v\n", string(dataJSONStr))
}

func debugReqLog(l debugReqLogger) {
	if cenvhelper.IsDebugMode() {
		l.Log()
	}
}

//----------分割线-------------

type debugResLogger struct {
	Err      error
	RespBody []byte
}

func (l *debugResLogger) Log() {
	res, _ := cutil.FormatJSONString(string(l.RespBody))

	logger.Printf("err:\t %s\n", l.Err)

	logger.Printf("resp body:\n%v\n", string(l.RespBody))

	logger.Printf("resp body json formated:\n%v\n", res)
}

func debugResLog(l debugResLogger) {
	if cenvhelper.IsDebugMode() {
		l.Log()
	}
}
