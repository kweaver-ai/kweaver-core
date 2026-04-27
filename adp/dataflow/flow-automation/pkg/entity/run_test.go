package entity

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefExecuteContext_Tracef(t *testing.T) {
	tests := []struct {
		name      string
		operation func(e *DefExecuteContext)
		wantLogs  []interface{}
	}{
		{
			name: "simple",
			operation: func(e *DefExecuteContext) {
				e.Tracef(context.TODO(), "aaa")
				e.Tracef(context.TODO(), "bbb")
				e.Trace(context.TODO(), "ccc")
			},
			wantLogs: []interface{}{"aaa", "bbb", "ccc"},
		},
		{
			name: "format",
			operation: func(e *DefExecuteContext) {
				e.Tracef(context.TODO(), "int:%d", 1)
				e.Tracef(context.TODO(), "str:%s", "aa")
				e.Tracef(context.TODO(), "+v:%+v", struct{ A int }{1})
			},
			wantLogs: []interface{}{"int:1", "str:aa", "+v:{A:1}"},
		},
		{
			name: "format with opt",
			operation: func(e *DefExecuteContext) {
				e.Tracef(context.TODO(), "str:%s", "aa", "ss", TraceOpPersistAfterAction)
			},
			wantLogs: []interface{}{"str:aa%!(EXTRA string=ss)", TraceOpPersistAfterAction},
		},
		{
			name: "multi opt",
			operation: func(e *DefExecuteContext) {
				e.Tracef(context.TODO(), "str:%s", TraceOpPersistAfterAction, TraceOpPersistAfterAction, TraceOpPersistAfterAction)
			},
			wantLogs: []interface{}{"str:%!s(MISSING)", TraceOpPersistAfterAction,
				TraceOpPersistAfterAction, TraceOpPersistAfterAction,
			},
		},
		{
			name: "not match format",
			operation: func(e *DefExecuteContext) {
				e.Tracef(context.TODO(), "int:%d", 1, TraceOpPersistAfterAction)
				e.Trace(context.TODO(), "cc", TraceOpPersistAfterAction)
				e.Tracef(context.TODO(), "str:%s", "aa")
				e.Tracef(context.TODO(), "+v:%+v", struct{ A int }{1})
				e.Tracef(context.TODO(), "str:%s", "aa", "ss", TraceOpPersistAfterAction)
			},
			wantLogs: []interface{}{"int:1", TraceOpPersistAfterAction,
				"cc", TraceOpPersistAfterAction,
				"str:aa",
				"+v:{A:1}",
				"str:aa%!(EXTRA string=ss)", TraceOpPersistAfterAction,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var logs []interface{}
			e := &DefExecuteContext{
				trace: func(ctx context.Context, msg string, opt ...TraceOp) {
					logs = append(logs, msg)
					for _, op := range opt {
						logs = append(logs, op)
					}
				},
			}
			tt.operation(e)

			want := fmt.Sprintf("%#v", tt.wantLogs)
			got := fmt.Sprintf("%#v", logs)
			assert.Equal(t, want, got)
		})
	}
}
