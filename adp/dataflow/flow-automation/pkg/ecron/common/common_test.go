package common

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
)

func TestCommonGetString(t *testing.T) {
	Convey("GetString", t, func() {
		c := LangLoader{
			Lang: map[string]string{
				"hello": "world",
			},
		}
		value := c.GetString("hello")
		assert.Equal(t, value, "world")
	})
}

func TestCommonFormatAddress(t *testing.T) {
	Convey("FormatAddress", t, func() {
		assert.Equal(t, FormatAddress("127.0.0.1", 12345), "127.0.0.1:12345")
		assert.Equal(t, FormatAddress("localhost", 12345), "localhost:12345")
		assert.Equal(t, FormatAddress("eisoo.com", 0), "eisoo.com")
	})
}

func TestCommonTimeStampToString(t *testing.T) {
	Convey("TimeStampToString", t, func() {
		assert.NotEqual(t, TimeStampToString(456), "")
		assert.Equal(t, TimeStampToString(0), "")
		// 以下两个判断跟测试机器所在的时区有关
		// assert.Equal(t, TimeStampToString(1577151370), "2019-12-24T09:36:10+08:00")
		// assert.NotEqual(t, TimeStampToString(1577151370), "2019-12-24T09:36:10")
	})
}

func TestCommonStringToTimeStamp(t *testing.T) {
	Convey("StringToTimeStamp", t, func() {
		ts, err := StringToTimeStamp("")
		assert.Equal(t, ts, int64(0))
		assert.Equal(t, err, nil)

		ts, err = StringToTimeStamp("123")
		assert.Equal(t, ts, int64(0))
		assert.NotEqual(t, err, nil)

		ts, err = StringToTimeStamp("2019-12-24T09:36:10")
		assert.Equal(t, ts, int64(0))
		assert.NotEqual(t, err, nil)

		ts, err = StringToTimeStamp("2019-12-24T09:36:10 08:00")
		assert.Equal(t, ts, int64(0))
		assert.NotEqual(t, err, nil)

		ts, err = StringToTimeStamp("2019-12-24T09:36:10Z08:00")
		assert.Equal(t, ts, int64(0))
		assert.NotEqual(t, err, nil)

		ts, err = StringToTimeStamp("2019-12-24T09:36:10.555555+08:00")
		assert.Equal(t, ts, int64(1577151370))
		assert.Equal(t, err, nil)

		ts, err = StringToTimeStamp("2019-12-24T09:36:10.555+08:00")
		assert.Equal(t, ts, int64(1577151370))
		assert.Equal(t, err, nil)

		ts, err = StringToTimeStamp("2019-12-24T09:36:10+08:00")
		assert.Equal(t, ts, int64(1577151370))
		assert.Equal(t, err, nil)
	})
}

func TestCommonECronError(t *testing.T) {
	Convey("Error", t, func() {
		err := ECronError{
			Cause:   "for test",
			Code:    123,
			Message: "",
			Detail:  nil,
		}
		json1 := "{\"cause\":\"for test\",\"code\":123,\"message\":\"\",\"detail\":null}"
		assert.Equal(t, err.Error(), json1)

		err = ECronError{
			Cause:   "for test",
			Code:    123,
			Message: "",
			Detail: map[string]interface{}{
				DetailConflicts: map[string]string{
					"job_id": "456",
				},
			},
		}
		json1 = "{\"cause\":\"for test\",\"code\":123,\"message\":\"\",\"detail\":{\"conflicts\":{\"job_id\":\"456\"}}}"
		assert.Equal(t, err.Error(), json1)
	})
}

func TestCommonNewDataDict(t *testing.T) {
	tests := []struct {
		name       string
		dataInt    []int
		dataString []string
	}{
		0: {
			"job type",
			[]int{1, 2, 3},
			[]string{TIMING, PERIODICITY, IMMEDIATE},
		},
		1: {
			"job status",
			[]int{1, 2, 3, 4, 5},
			[]string{SUCCESS, EXECUTING, FAILURE, INTERRUPT, ABANDON},
		},
		2: {
			"job operation",
			[]int{1, 2, 3, 4, 5},
			[]string{CREATE, UPDATE, DELETE, ENABLE, NOTIFY},
		},
		3: {
			"job execution",
			[]int{1, 2},
			[]string{HTTP, EXE, HTTPS},
		},
	}

	Convey("NewDataDict", t, func() {
		dd := NewDataDict()
		assert.NotEqual(t, dd, nil)

		tt := tests[0]
		Convey(tt.name, func() {
			for i := 0; i < len(tt.dataInt); i++ {
				v, ok1 := dd.DJobType.StringToInt(tt.dataString[i])
				s, ok2 := dd.DJobType.IntToString(tt.dataInt[i])
				assert.Equal(t, tt.dataInt[i], v)
				assert.Equal(t, true, ok1)
				assert.Equal(t, tt.dataString[i], s)
				assert.Equal(t, true, ok2)
			}
			v, ok1 := dd.DJobType.StringToInt("")
			s, ok2 := dd.DJobType.IntToString(0)
			assert.Equal(t, 0, v)
			assert.Equal(t, false, ok1)
			assert.Equal(t, "", s)
			assert.Equal(t, false, ok2)
		})

		tt = tests[1]
		Convey(tt.name, func() {
			for i := 0; i < len(tt.dataInt); i++ {
				v, ok1 := dd.DJobStatus.StringToInt(tt.dataString[i])
				s, ok2 := dd.DJobStatus.IntToString(tt.dataInt[i])
				assert.Equal(t, tt.dataInt[i], v)
				assert.Equal(t, true, ok1)
				assert.Equal(t, tt.dataString[i], s)
				assert.Equal(t, true, ok2)
			}
			v, ok1 := dd.DJobStatus.StringToInt("")
			s, ok2 := dd.DJobStatus.IntToString(0)
			assert.Equal(t, 0, v)
			assert.Equal(t, false, ok1)
			assert.Equal(t, "", s)
			assert.Equal(t, false, ok2)
		})

		tt = tests[2]
		Convey(tt.name, func() {
			for i := 0; i < len(tt.dataInt); i++ {
				v, ok1 := dd.DJobOperation.StringToInt(tt.dataString[i])
				s, ok2 := dd.DJobOperation.IntToString(tt.dataInt[i])
				assert.Equal(t, tt.dataInt[i], v)
				assert.Equal(t, true, ok1)
				assert.Equal(t, tt.dataString[i], s)
				assert.Equal(t, true, ok2)
			}
			v, ok1 := dd.DJobOperation.StringToInt("")
			s, ok2 := dd.DJobOperation.IntToString(0)
			assert.Equal(t, 0, v)
			assert.Equal(t, false, ok1)
			assert.Equal(t, "", s)
			assert.Equal(t, false, ok2)
		})

		tt = tests[3]
		Convey(tt.name, func() {
			for i := 0; i < len(tt.dataInt); i++ {
				v, ok1 := dd.DJobExecution.StringToInt(tt.dataString[i])
				s, ok2 := dd.DJobExecution.IntToString(tt.dataInt[i])
				assert.Equal(t, tt.dataInt[i], v)
				assert.Equal(t, true, ok1)
				assert.Equal(t, tt.dataString[i], s)
				assert.Equal(t, true, ok2)
			}
			v, ok1 := dd.DJobExecution.StringToInt("")
			s, ok2 := dd.DJobExecution.IntToString(0)
			assert.Equal(t, 0, v)
			assert.Equal(t, false, ok1)
			assert.Equal(t, "", s)
			assert.Equal(t, false, ok2)
		})
	})
}

func TestCommonGetHTTPAccess(t *testing.T) {
	Convey("GetHTTPAccess()", t, func() {
		Convey("ssl on", func() {
			access := GetHTTPAccess("127.0.0.1", 12345, true)
			assert.Equal(t, access, "https://127.0.0.1:12345")
		})
		Convey("ssl off", func() {
			access := GetHTTPAccess("127.0.0.1", 12345, false)
			assert.Equal(t, access, "http://127.0.0.1:12345")
		})
	})
}

func TestCommonGetSleepDuration(t *testing.T) {
	Convey("GetSleepDuration()", t, func() {
		Convey("less than 10, return 10", func() {
			d := GetSleepDuration(5)
			assert.Equal(t, d, time.Duration(1e9*10))
		})
		Convey("more than 10, return itself", func() {
			d := GetSleepDuration(30)
			assert.Equal(t, d, time.Duration(1e9*30))
		})
	})
}

func TestCommonGetIntMoreThanLowerLimit(t *testing.T) {
	Convey("GetIntMoreThanLowerLimit()", t, func() {
		Convey("less than lower, return lower", func() {
			i := GetIntMoreThanLowerLimit(0, 300)
			assert.Equal(t, i, 300)
		})
		Convey("more than lower, return itself", func() {
			i := GetIntMoreThanLowerLimit(50, 30)
			assert.Equal(t, i, 50)
		})
	})
}
