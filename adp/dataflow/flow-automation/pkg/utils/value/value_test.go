package value

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	val = map[string]interface{}{
		"string":  "b",
		"int":     1,
		"float64": 1.2,
		"map": map[string]interface{}{
			"internal": "internal",
		},
		"slice": []interface{}{
			map[string]interface{}{
				"sliceMap": map[string]interface{}{
					"sMapK": "sMapV",
					"mapSlice": []interface{}{
						"mapSliceI1",
						"mapSliceI2",
					},
				},
			},
			"sliceI1",
			"sliceI2",
		},
	}
)

type logger struct {
	logs []interface{}
}

// Log Log
func (l *logger) Log(v interface{}) {
	l.logs = append(l.logs, v)
}

// Logs Logs
func (l *logger) Logs() []interface{} {
	return l.logs
}

func TestValue_Walk(t *testing.T) {
	type TestCase struct {
		name            string
		val             MapValue
		logger          *logger
		callbackBuilder func(tc TestCase) MapValueCallback
		wantErr         assert.ErrorAssertionFunc
		assert          func(t *testing.T, tt TestCase)
	}

	tests := []TestCase{
		{
			name: "walk",
			val:  val,
			callbackBuilder: func(tc TestCase) MapValueCallback {
				return func(walkContext *WalkContext, v interface{}) error {
					tc.logger.Log(v)
					return nil
				}
			},
			wantErr: assert.NoError,

			assert: func(t *testing.T, tt TestCase) {
				// fix map order
				wantLogs := []interface{}{"b", 1,
					1.2, "internal", "sMapV", "mapSliceI1", "mapSliceI2", "sliceI1", "sliceI2"}
				logs := tt.logger.Logs()
				assert.ElementsMatch(t, wantLogs, logs)
			},
			logger: &logger{},
		},
		{
			name: "break",
			val:  val,
			callbackBuilder: func(tc TestCase) MapValueCallback {
				return func(walkContext *WalkContext, v interface{}) error {
					tc.logger.Log(v)
					if v == "internal" {
						return errors.New("break")
					}
					return nil
				}
			},
			wantErr: assert.Error,
			assert: func(t *testing.T, tt TestCase) {
				logs := tt.logger.Logs()
				assert.Equal(t, "internal", logs[len(logs)-1])
			},
			logger: &logger{},
		},
		{
			name: "path",
			val:  val,
			callbackBuilder: func(tc TestCase) MapValueCallback {
				return func(walkContext *WalkContext, v interface{}) error {
					tc.logger.Log(walkContext.Path())
					return nil
				}
			},
			wantErr: assert.NoError,
			assert: func(t *testing.T, tt TestCase) {
				expected := []interface{}{
					"string", "int", "float64", "map.internal",
					"slice[0].sliceMap.sMapK", "slice[0].sliceMap.mapSlice[0]",
					"slice[0].sliceMap.mapSlice[1]", "slice[1]", "slice[2]"}
				logs := tt.logger.Logs()
				assert.ElementsMatch(t, expected, logs)
			},
			logger: &logger{},
		},
		{
			name: "path2",
			val: map[string]interface{}{
				"a": map[string]interface{}{
					"b": "vb",
					"c": map[string]interface{}{
						"d": map[string]interface{}{
							"e": "f",
							"g": []interface{}{
								map[string]interface{}{
									"h": "i",
								},
							},
						},
					},
				},
			},
			callbackBuilder: func(tc TestCase) MapValueCallback {
				return func(walkContext *WalkContext, v interface{}) error {
					tc.logger.Log(walkContext.Path())
					return nil
				}
			},
			wantErr: assert.NoError,
			assert: func(t *testing.T, tt TestCase) {
				expected := []interface{}{"a.b", "a.c.d.e", "a.c.d.g[0].h"}
				logs := tt.logger.Logs()
				assert.ElementsMatch(t, expected, logs)
			},
			logger: &logger{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.val.Walk(tt.callbackBuilder(tt))
			tt.wantErr(t, err, fmt.Sprintf("Walk()"))
			tt.assert(t, tt)
		})
	}
}

func TestValue_WalkString(t *testing.T) {
	type TestCase struct {
		name            string
		val             MapValue
		logger          *logger
		callbackBuilder func(tt TestCase) MapValueStringCallback
		wantErr         assert.ErrorAssertionFunc
		assert          func(t *testing.T, tt TestCase)
	}
	tests := []TestCase{
		{
			name: "walk string",
			val:  val,
			callbackBuilder: func(tt TestCase) MapValueStringCallback {
				return func(walkContext *WalkContext, v string) error {
					tt.logger.Log(v)
					return nil
				}
			},
			wantErr: assert.NoError,
			assert: func(t *testing.T, tt TestCase) {
				wantLogs := []interface{}{"b", "internal", "sMapV", "mapSliceI1",
					"mapSliceI2", "sliceI1", "sliceI2"}
				logs := tt.logger.Logs()
				assert.ElementsMatch(t, wantLogs, logs)
			},
			logger: &logger{},
		},
		{
			name: "break",
			val:  val,
			callbackBuilder: func(tt TestCase) MapValueStringCallback {
				return func(walkContext *WalkContext, v string) error {
					tt.logger.Log(v)
					if v == "internal" {
						return errors.New("aa")
					}
					return nil
				}
			},
			wantErr: assert.Error,
			assert: func(t *testing.T, tt TestCase) {
				logs := tt.logger.Logs()
				assert.Equal(t, "internal", logs[len(logs)-1])
			},
			logger: &logger{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.wantErr(t, tt.val.WalkString(tt.callbackBuilder(tt)), fmt.Sprintf("WalkString()"))
			tt.assert(t, tt)
		})

	}
}
