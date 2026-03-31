package agentsvc

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/bytedance/sonic"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/otel/oteltrace"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
)

func formatSSEMessage(data string) []byte {
	return []byte(fmt.Sprintf("data: %s\n\n", string(data)))
}

// type StreamDiffer struct {
// }

type Change struct {
	SeqID   int           `json:"seq_id"`
	Key     []interface{} `json:"key"`
	Content interface{}   `json:"content"`
	Action  string        `json:"action"`
}

// StreamDiff 对比 oldJSON/newJSON，并将增量 diff 以 SSE 格式写入 w。
func formatChange(ch Change) string {
	var sb strings.Builder

	sb.WriteString("{")
	sb.WriteString(fmt.Sprintf("\"seq_id\": %d, ", ch.SeqID))
	keyBytes, _ := sonic.Marshal(ch.Key)
	sb.WriteString(fmt.Sprintf("\"key\": %s, ", keyBytes))

	contentBytes, _ := sonic.Marshal(ch.Content)
	sb.WriteString(fmt.Sprintf("\"content\": %s, ", contentBytes))
	sb.WriteString(fmt.Sprintf("\"action\": \"%s\"", ch.Action))
	sb.WriteString("}")

	return sb.String()
}

func emitJSON(seq *int, out chan []byte, keyPath []interface{}, content interface{}, action string) {
	ch := Change{
		SeqID:   *seq,
		Key:     append([]interface{}{}, keyPath...),
		Content: content,
		Action:  action,
	}
	line := formatChange(ch)
	out <- []byte("data: " + line + "\n\n")

	*seq++
}

// StreamDiff 对比 oldJSON/newJSON，并把每条增量以 "data: {...}" 推送到 out chan
// chunkIndex: 流式周期索引，仅首(0)周期创建独立 span
func StreamDiff(ctx context.Context, lastSeq *int, oldJSON, newJSON []byte, out chan []byte, chunkIndex int) error {
	var err error

	if chunkIndex == 0 {
		ctx, _ = oteltrace.StartInternalSpan(ctx)
		defer oteltrace.EndSpan(ctx, err)
	}

	var oldVal, newVal interface{}
	if err := sonic.Unmarshal(oldJSON, &oldVal); err != nil {
		return err
	}

	if err := sonic.Unmarshal(newJSON, &newVal); err != nil {
		return err
	}

	// emit 会把一条格式化后的 SSE 消息发到 chan
	// 跟踪 emitJSON 调用次数
	emitCount := 0

	// 递归 diff
	var diff func(oldV, newV interface{}, path []interface{})
	diff = func(oldV, newV interface{}, path []interface{}) {
		// 类型不同或非复杂类型：直接 upsert 整个 newV
		if reflect.TypeOf(oldV) != reflect.TypeOf(newV) {
			emitJSON(lastSeq, out, path, newV, "upsert")

			emitCount++

			return
		}

		switch ov := oldV.(type) {
		case map[string]interface{}:
			nv := newV.(map[string]interface{})
			// upsert 或 递归
			for k, newChild := range nv {
				childPath := append(path, k)
				oldChild, ok := ov[k]

				if !ok {
					emitJSON(lastSeq, out, childPath, newChild, "upsert")

					emitCount++
				} else {
					diff(oldChild, newChild, childPath)
				}
			}
			// remove
			for k := range ov {
				if _, ok := nv[k]; !ok {
					emitJSON(lastSeq, out, append(path, k), nil, "remove")

					emitCount++
				}
			}

		case []interface{}:
			nv := newV.([]interface{})
			// 1. 处理重叠区间内的元素变更
			minLen := len(ov)
			if len(nv) < minLen {
				minLen = len(nv)
			}

			for i := 0; i < minLen; i++ {
				oElem, nElem := ov[i], nv[i]
				pathIdx := append(path, i)

				// 1.0 类型变化 -> 直接 upsert 整个元素
				if reflect.TypeOf(oElem) != reflect.TypeOf(nElem) {
					emitJSON(lastSeq, out, pathIdx, nElem, "upsert")

					emitCount++

					continue
				}

				// 1.1 对象类型，递归 diff
				if oMap, ok1 := oElem.(map[string]interface{}); ok1 {
					if nMap, ok2 := nElem.(map[string]interface{}); ok2 {
						diff(oMap, nMap, pathIdx)
						continue
					}
				}
				// 1.2 嵌套数组，递归 diff
				if oArr, ok1 := oElem.([]interface{}); ok1 {
					if nArr, ok2 := nElem.([]interface{}); ok2 {
						// 嵌套数组，调用自身
						diff(oArr, nArr, pathIdx)
						continue
					}
				}
				// 1.3 字符串追加场景
				if os, ok1 := oElem.(string); ok1 {
					if ns, ok2 := nElem.(string); ok2 && strings.HasPrefix(ns, os) {
						delta := ns[len(os):]
						if delta != "" {
							emitJSON(lastSeq, out, pathIdx, delta, "append")

							emitCount++
						}

						continue
					}
				}
				// 1.4 基本类型或"字符串替换"或其他不匹配追加的情况，用 upsert
				if !reflect.DeepEqual(oElem, nElem) {
					emitJSON(lastSeq, out, pathIdx, nElem, "upsert")

					emitCount++
				}
			}
			// 2. 处理新增元素（append）
			for i := len(ov); i < len(nv); i++ {
				emitJSON(lastSeq, out, append(path, i), nv[i], "append")

				emitCount++
			}
			// 3. 处理删除元素（remove）
			for i := len(nv); i < len(ov); i++ {
				emitJSON(lastSeq, out, append(path, i), nil, "remove")

				emitCount++
			}

		case string:
			ns := newV.(string)
			os := ov
			// 如果 new 以 old 为前缀，只发一次 append 整段 delta
			if strings.HasPrefix(ns, os) {
				if delta := ns[len(os):]; delta != "" {
					emitJSON(lastSeq, out, path, delta, "append")

					emitCount++
				}
			} else {
				// 否则整串替换
				emitJSON(lastSeq, out, path, ns, "upsert")

				emitCount++
			}

		default:
			if !reflect.DeepEqual(oldV, newV) {
				emitJSON(lastSeq, out, path, newV, "upsert")

				emitCount++
			}
		}
	}

	// 执行 diff
	diff(oldVal, newVal, []interface{}{})

	// 如果没有调用过 emitJSON，记录 warn 日志
	if emitCount == 0 {
		logger.Warnf("StreamDiff: no differences found between oldJSON and newJSON, new json :%s", string(newJSON))
	}

	// 结束标记
	// emit([]interface{}{}, nil, "end")
	return nil
}
