package utils

import (
	"runtime"
	"time"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/log"
)

// StringsContain
func StringsContain(strs []string, str string) bool {
	for i := range strs {
		if strs[i] == str {
			return true
		}
	}
	return false
}

type KeyValueGetter func(key string) (interface{}, bool)
type KeyValueIterator func(KeyValueIterateFunc)
type KeyValueIterateFunc func(key, val string) (stop bool)

type KeyValueInterfaceGetter func(key string) (interface{}, bool)
type KeyValueInterfaceIterator func(KeyValueIterateInterfaceFunc)
type KeyValueIterateInterfaceFunc func(key, val interface{}) (stop bool)

const (
	LogKeyDagInsID = "dagInsId"
)

func TimeCost() func() {
	startTime := time.Now()
	funcName := getFunctionName(1)

	return func() {
		elapsedTime := time.Since(startTime)
		log.Infof("Function %s took %s\n", &funcName, elapsedTime.Milliseconds())
	}
}

func getFunctionName(skip int) string {
	pc, _, _, _ := runtime.Caller(skip)
	fullName := runtime.FuncForPC(pc).Name()
	lastSlash := 0

	for i := 0; i < len(fullName); i++ {
		if fullName[i] == '/' {
			lastSlash = i
		}
	}

	return fullName[lastSlash+1:]
}
