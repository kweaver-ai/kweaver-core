package mgnt

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/drivenadapters"
	ierrors "github.com/kweaver-ai/adp/autoflow/flow-automation/errors"
	traceLog "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/log"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/trace"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/entity"
)

const MAX_LIMIT = 10000
const DEFAULT_BATCH_SIZE = 1000

type SyncMode string

const (
	SyncModeFull        = "full"
	SyncModeIncremental = "incremental"
)

type DataViewParam struct {
	ID        string   `json:"id,omitempty"`
	SyncMode  SyncMode `json:"syncMode,omitempty"`
	BatchSize int      `json:"batchSize,omitempty"`

	// 全量模式
	Mode     string `json:"mode,omitempty"`
	Duration int64  `json:"duration,omitempty"`
	Start    int64  `json:"start,omitempty"`
	End      int64  `json:"end,omitempty"`

	//增量模式
	IncrementField string `json:"incrementField,omitempty"`
	IncrementValue any    `json:"incrementValue,omitempty"`
	Filter         string `json:"filter,omitempty"`
}

func encodeSQLLiteral(v any) string {
	if v == nil {
		return "NULL"
	}
	switch val := v.(type) {
	case string:
		escaped := strings.ReplaceAll(val, "'", "''")
		return fmt.Sprintf("'%s'", escaped)
	case []byte:
		escaped := strings.ReplaceAll(string(val), "'", "''")
		return fmt.Sprintf("'%s'", escaped)
	case fmt.Stringer:
		s := val.String()
		escaped := strings.ReplaceAll(s, "'", "''")
		return fmt.Sprintf("'%s'", escaped)
	default:
		return fmt.Sprintf("%v", v)
	}
}

func convertValue(v any, typ string) any {
	switch strings.ToLower(typ) {
	case "number", "float", "real", "double":
		switch val := v.(type) {
		case nil:
			return nil
		case float64:
			return val
		case int64:
			return float64(val)
		case int, int32, int16, int8:
			return float64(reflect.ValueOf(val).Int())
		case uint, uint64, uint32, uint16, uint8:
			return float64(reflect.ValueOf(val).Uint())
		case string:
			if val == "" {
				return nil
			}
			if f, err := strconv.ParseFloat(val, 64); err == nil {
				return f
			}
			var f64 float64
			if _, err := fmt.Sscanf(val, "%e", &f64); err == nil {
				return f64
			}
			return nil
		default:
			return nil
		}
	case "int", "integer":
		switch val := v.(type) {
		case nil:
			return nil
		case int64:
			return val
		case int, int32, int16, int8:
			return reflect.ValueOf(val).Int()
		case uint, uint64, uint32, uint16, uint8:
			return int64(reflect.ValueOf(val).Uint())
		case float64:
			return int64(val)
		case string:
			if val == "" {
				return nil
			}
			if i, err := strconv.ParseInt(val, 10, 64); err == nil {
				return i
			}
			if f, err := strconv.ParseFloat(val, 64); err == nil {
				return int64(f)
			}
			return nil
		default:
			return nil
		}
	}

	return v
}

func (p *DataViewParam) Init(params map[string]any) error {
	d, err := json.Marshal(params)

	if err != nil {
		return err
	}
	err = json.Unmarshal(d, &p)
	return err
}

func (m *mgnt) triggerFromMDLDataView(
	ctx context.Context,
	dag *entity.Dag,
	triggerType entity.Trigger,
	runVar map[string]string,
	userid, userType, ip string,
) error {
	var err error
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(ctx)

	if dag.Status != entity.DagStatusNormal {
		return ierrors.NewIError(ierrors.Forbidden, ierrors.DagStatusNotNormal, map[string]interface{}{"id:": dag.ID, "status": dag.Status})
	}

	triggerStep := dag.Steps[0]
	var param DataViewParam
	err = param.Init(triggerStep.Parameters)

	if err != nil {
		log.Warnf("[logic.triggerFromMDLDataView] Parse DataViewParam err %s", err.Error())
		return err
	}

	ch := make(chan error)

	go func() {
		if param.SyncMode == SyncModeIncremental {
			m.triggerFromMDLDataViewIncremental(context.Background(), ch, dag, &param, triggerType, runVar, userid, userType, ip)
		} else {
			m.triggerFromMDLDataViewFull(context.Background(), ch, dag, &param, triggerType, runVar, userid, userType, ip)
		}
	}()

	err = <-ch

	if err != nil {
		dagIns, dagErr := dag.Run(ctx, triggerType, runVar)

		if dagErr != nil {
			log.Warnf("[logic.triggerFromMDLDataView] dag.Run err: %s", err.Error())
			return err
		}

		dagIns.Initial()
		dagIns.Status = entity.DagInstanceStatusFailed
		taskIns := &entity.TaskInstance{
			TaskID:     triggerStep.ID,
			DagInsID:   dagIns.ID,
			Name:       triggerStep.Title,
			ActionName: triggerStep.Operator,
			Params:     triggerStep.Parameters,
			Status:     entity.TaskInstanceStatusFailed,
			Reason:     err,
		}
		dagIns.Source = `{"_type":"dataview"}`

		reason := map[string]any{
			"taskId":     taskIns.TaskID,
			"name":       taskIns.Name,
			"actionName": taskIns.ActionName,
			"detail":     taskIns.Reason,
		}

		b, _ := json.Marshal(reason)

		dagIns.Reason = string(b)

		_, dbErr := m.mongo.CreateDagIns(ctx, dagIns)

		if dbErr != nil {
			log.Warnf("[logic.triggerFromMDLDataView] CreateDagIns err: %s", dbErr.Error())
			return err
		}

		_, dbErr = m.mongo.BatchCreateTaskIns(ctx, []*entity.TaskInstance{taskIns})

		if dbErr != nil {
			log.Warnf("[logic.triggerFromMDLDataView] BatchCreateTaskIns err: %s", dbErr.Error())
			return err
		}
	}

	return err
}

func (m *mgnt) triggerFromMDLDataViewFull(
	ctx context.Context,
	ch chan error,
	dag *entity.Dag,
	param *DataViewParam,
	triggerType entity.Trigger,
	runVar map[string]string,
	userid, userType, ip string,
) {

	var err error
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(ctx)

	var (
		id        = param.ID
		mode      = param.Mode
		batchSize = param.BatchSize
		duration  = param.Duration
		start     = param.Start
		end       = param.End
	)

	if mode == "duration" {
		if duration <= 0 {
			ch <- ierrors.NewIError(ierrors.InvalidParameter, "", map[string]any{
				"duration": duration,
				"reason":   "Duration must be greater than 0.",
			})
			return
		}
		end = time.Now().UnixMilli()
		start = end - duration
	} else if mode == "range" {
		end = time.Now().UnixMilli()
	} else {
		start = 0
		end = 0
	}

	if batchSize == 0 {
		batchSize = DEFAULT_BATCH_SIZE
	}

	var (
		isFirstPage = true
		searchAfter []any
		entryBuffer = make([]any, 0, MAX_LIMIT)
		sentOnce    = false
	)

	runVar["id"] = id
	dag.SetPushMessage(m.executeMethods.Publish)

	for {
		data, err := m.uniquery.UniqueryDataView(ctx, id, &drivenadapters.UniqueryDataViewOptions{
			Start:          start,
			End:            end,
			Limit:          MAX_LIMIT,
			NeedTotal:      isFirstPage,
			UseSearchAfter: true,
			SearchAfter:    searchAfter,
			Format:         "original",
		}, userid, userType)

		if err != nil {
			log.Warnf("[logic.triggerFromMDLDataViewFull] UniqueryDataView err, start %d, end %d, search_after %v, detail: %s", start, end, searchAfter, err.Error())
			ch <- err
			return
		}

		isLastPage := len(data.SearchAfter) == 0

		if isFirstPage {
			isFirstPage = false
			if isLastPage && len(data.Entries) == 0 {
				err := ierrors.NewIError(ierrors.DataSourceIsEmpty, "", map[string]any{
					"data_view_id": param.ID,
					"start":        start,
					"end":          end,
					"search_after": searchAfter,
					"reason":       "data souce is empty",
				})
				ch <- err
				return
			}
		}
		searchAfter = data.SearchAfter
		entryBuffer = append(entryBuffer, data.Entries...)
		if len(entryBuffer) >= batchSize || isLastPage {
			entryBuffer, err = m.batchCreateDagInsFromDataView(ctx, dag, param, triggerType, runVar, entryBuffer, batchSize, isLastPage)
			if err != nil {
				ch <- err
				return
			}
		}

		if !sentOnce {
			ch <- nil
			sentOnce = true
		}

		if isLastPage {
			return
		}
	}
}

func (m *mgnt) triggerFromMDLDataViewIncremental(
	ctx context.Context,
	ch chan error,
	dag *entity.Dag,
	param *DataViewParam,
	triggerType entity.Trigger,
	runVar map[string]string,
	userid, userType, ip string,
) {
	var err error
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(ctx)

	batchSize := param.BatchSize
	incrementField := param.IncrementField

	dataview, err := m.uniquery.GetDataViewByID(ctx, param.ID, userid, userType)

	if err != nil {
		log.Warnf("[logic.triggerFromMDLDataViewIncremental] GetDataViewByID err %s", err.Error())
		ch <- err
		return
	}

	var field *drivenadapters.DataViewField

	for _, f := range dataview.Fields {
		if f.Name == incrementField {
			field = f
			break
		}
	}

	if field == nil {
		log.Warnf("[logic.triggerFromMDLDataViewIncremental] unknown field: %s", incrementField)
		err = ierrors.NewIError(ierrors.DataSourceIsEmpty, "", map[string]any{
			"data_view_id": param.ID,
			"reason":       fmt.Sprintf("unknown field: %s", incrementField),
		})
		ch <- err
		return
	}

	incrementValue, ok := dag.IncValues[incrementField]

	if !ok {
		incrementValue = param.IncrementValue
	}

	incrementValue = convertValue(incrementValue, field.Type)

	filter := param.Filter

	if batchSize == 0 {
		batchSize = DEFAULT_BATCH_SIZE
	}

	lockKey := entity.DataViewTriggerLock + dag.ID
	err = m.rdb.TryLock(lockKey, "", time.Second*300)

	if err != nil {
		log.Warnf("[logic.triggerFromMDLDataViewIncremental] task already in progress, dagId %s", dag.ID)
		ch <- ierrors.NewIError(ierrors.TaskAlreayInProgress, "", map[string]any{"id": dag.ID})
		return
	}

	lockCtx, lockCancel := context.WithCancel(ctx)
	defer func() {
		m.rdb.Unlock(lockKey, "")
		lockCancel()
	}()

	go func() {
		select {
		case <-lockCtx.Done():
			return
		default:
			time.Sleep(time.Second * 250)
			m.rdb.Renew(lockKey, "", time.Second*300)
		}
	}()

	sentOnce := false
	entryBuffer := make([]any, 0, MAX_LIMIT)

	for {
		sqlStr := fmt.Sprintf("SELECT * FROM %s", dataview.TableName)
		hasWhere := false
		if incrementValue != nil {
			op := ">"

			// 如果 incrementValue 是第一次运行，则使用 >=
			if incrementValue == param.IncrementValue {
				op = ">="
			}

			encodedVal := encodeSQLLiteral(incrementValue)
			sqlStr = fmt.Sprintf("%s WHERE \"%s\" %s %s", sqlStr, incrementField, op, encodedVal)
			hasWhere = true
		}

		if filter != "" {
			if !hasWhere {
				sqlStr = fmt.Sprintf("%s WHERE %s", sqlStr, filter)
			} else {
				sqlStr = fmt.Sprintf("%s AND (%s)", sqlStr, filter)
			}
		}

		sqlStr = fmt.Sprintf("%s ORDER BY \"%s\" ASC", sqlStr, incrementField)

		entries, err := m.uniquery.UniqueryDataView(ctx, param.ID, &drivenadapters.UniqueryDataViewOptions{
			QueryType: "SQL",
			Limit:     MAX_LIMIT,
			NeedTotal: true,
			Sql:       sqlStr,
		}, userid, userType)

		log.Warnf("[logic.triggerFromMDLDataViewIncremental] UniqueryDataView  data view id %s, sql %s", param.ID, sqlStr)

		if err != nil {
			log.Warnf("[logic.triggerFromMDLDataViewIncremental] UniqueryDataView err %s,  data view id %s, sql %s", err.Error(), param.ID, sqlStr)
			ch <- err
			return
		}

		isLastPage := len(entries.Entries) == 0

		if isLastPage && !sentOnce {
			log.Warnf("[logic.triggerFromMDLDataViewIncremental] data source is empty, data view id %s, sql %s", param.ID, sqlStr)
			err = ierrors.NewIError(ierrors.DataSourceIsEmpty, "", map[string]any{
				"data_view_id": param.ID,
				"sql":          sqlStr,
				"reason":       "data souce is empty",
			})
			ch <- err
			return
		}

		if !isLastPage {
			latestItem, ok := entries.Entries[len(entries.Entries)-1].(map[string]any)
			if !ok {
				log.Warnf("[logic.triggerFromMDLDataViewIncremental] item must be an object, data view id %s, sql %s", param.ID, sqlStr)
				err = ierrors.NewIError(ierrors.InternalError, "", map[string]any{
					"data_view_id": param.ID,
					"sql":          sqlStr,
					"item":         latestItem,
					"reason":       "item must be an object",
				})
				ch <- err
				return
			}
			latestIncrementValue, ok := latestItem[incrementField]
			if !ok {
				log.Warnf("[logic.triggerFromMDLDataViewIncremental] item.%s is required, data view id %s, sql %s", incrementField, param.ID, sqlStr)
				err = ierrors.NewIError(ierrors.InternalError, "", map[string]any{
					"data_view_id": param.ID,
					"sql":          sqlStr,
					"item":         latestItem,
					"reason":       fmt.Sprintf("item.%s is required", incrementField),
				})
				ch <- err
				return
			}

			incrementValue = latestIncrementValue
			dag.IncValues[incrementField] = latestIncrementValue
			err = m.mongo.UpdateDagIncValue(ctx, dag.ID, incrementField, incrementValue)

			if err != nil {
				log.Warnf("[logic.triggerFromMDLDataViewIncremental] UpdateDagIncValue err %s", err.Error())
				ch <- err
				return
			}
			entryBuffer = append(entryBuffer, entries.Entries...)
		}

		if l := len(entryBuffer); l >= batchSize || isLastPage && l > 0 {
			entryBuffer, err = m.batchCreateDagInsFromDataView(ctx, dag, param, triggerType, runVar, entryBuffer, batchSize, isLastPage)
			if err != nil {
				ch <- err
				return
			}
		}

		if !sentOnce {
			ch <- nil
			sentOnce = true
		}

		if isLastPage {
			return
		}
	}
}

func (m *mgnt) batchCreateDagInsFromDataView(ctx context.Context,
	dag *entity.Dag,
	param *DataViewParam,
	triggerType entity.Trigger,
	runVar map[string]string,
	entryBuffer []any,
	batchSize int,
	isLastPage bool,
) ([]any, error) {

	var err error
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(ctx)
	l := len(entryBuffer)
	dagInsTotal := l / batchSize
	if isLastPage && l%batchSize != 0 {
		dagInsTotal += 1
	}

	dagInstances := make([]*entity.DagInstance, 0, dagInsTotal)

	for i := range dagInsTotal {
		start, end := i*batchSize, (i+1)*batchSize

		if end > l {
			end = l
		}

		items := entryBuffer[start:end]
		dagIns, err := dag.Run(ctx, triggerType, runVar)

		if err != nil {
			log.Warnf("[logic.batchCreateDagInsFromDataView] dag.Run err, detail: %s", err.Error())
			return nil, err
		}

		dagIns.Initial()

		b, _ := json.Marshal(map[string]any{
			"_type": "dataview",
			"total": len(items),
		})
		dagIns.Source = string(b)

		werr := dagIns.WriteEventByVariableMap(ctx, map[string]any{
			"__" + dag.Steps[0].ID: map[string]any{
				"data": items,
			},
		}, time.Now().UnixMicro())

		if werr != nil {
			dagIns.Status = entity.DagInstanceStatusFailed
			b, _ := json.Marshal(map[string]any{
				"actionName": dag.Steps[0].Operator,
				"name":       dag.Steps[0].Title,
				"taskId":     dag.Steps[0].ID,
				"detail":     werr,
			})
			dagIns.Reason = string(b)
		}

		dagInstances = append(dagInstances, dagIns)
	}

	if len(dagInstances) > 0 {
		_, err := m.mongo.BatchCreateDagIns(ctx, dagInstances)
		if err != nil {
			log.Warnf("[logic.batchCreateDagInsFromDataView] BatchCreateDagIns err, detail: %s", err.Error())
			return nil, ierrors.NewIError(ierrors.InternalError, "", err.Error())
		}
	}

	newEntryBuffer := make([]any, 0, MAX_LIMIT)

	if itemCount := dagInsTotal * batchSize; itemCount < l {
		newEntryBuffer = append(newEntryBuffer, entryBuffer[itemCount:]...)
	}

	return newEntryBuffer, nil
}
