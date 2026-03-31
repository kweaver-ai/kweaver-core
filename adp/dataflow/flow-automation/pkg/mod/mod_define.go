package mod

import (
	"context"
	"time"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/entity"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

//go:generate mockgen -package mod -source ./mod_define.go -destination ./mod_define_mock.go

var (
	ActionMap = map[string]entity.Action{}

	defExc       Executor
	defStore     Store
	defKeeper    Keeper
	defParser    Parser
	defCommander Commander
)

// Commander used to execute command
type Commander interface {
	RunDag(ctx context.Context, dagId string, specVar map[string]string) (*entity.DagInstance, error)
	RetryDagIns(ctx context.Context, dagInsId string, ops ...CommandOptSetter) error
	RetryTask(ctx context.Context, taskInsIds []string, ops ...CommandOptSetter) error
	CancelTask(ctx context.Context, taskInsIds []string, ops ...CommandOptSetter) error
}

// CommandOption
type CommandOption struct {
	// isSync means commander will watch dag instance's cmd executing situation until it's command is executed
	// usually command executing time is very short, so async mode is enough,
	// but if you want a sync call, you set it to true
	isSync bool
	// syncTimeout is just work at sync mode, it is the timeout of watch dag instance
	// default is 5s
	syncTimeout time.Duration
	// syncInterval is just work at sync mode, it is the interval of watch dag instance
	// default is 500ms
	syncInterval time.Duration
}
type CommandOptSetter func(opt *CommandOption)

var (
	// CommSync means commander will watch dag instance's cmd executing situation until it's command is executed
	// usually command executing time is very short, so async mode is enough,
	// but if you want a sync call, you set it to true
	CommSync = func() CommandOptSetter {
		return func(opt *CommandOption) {
			opt.isSync = true
		}
	}
	// CommSync means commander will watch dag instance's cmd executing situation until it's command is executed
	// usually command executing time is very short, so async mode is enough,
	// but if you want a sync call, you set it to true
	CommSyncTimeout = func(duration time.Duration) CommandOptSetter {
		return func(opt *CommandOption) {
			if duration > 0 {
				opt.syncTimeout = duration
			}
		}
	}
	// CommSyncInterval is just work at sync mode, it is the interval of watch dag instance
	// default is 500ms
	CommSyncInterval = func(duration time.Duration) CommandOptSetter {
		return func(opt *CommandOption) {
			if duration > 0 {
				opt.syncInterval = duration
			}
		}
	}
)

// SetCommander
func SetCommander(c Commander) {
	defCommander = c
}

// GetCommander
func GetCommander() Commander {
	return defCommander
}

// Executor is used to execute task
type Executor interface {
	Push(dagIns *entity.DagInstance, taskIns *entity.TaskInstance)
	CancelTaskIns(taskInsIds []string) error
}

// SetExecutor
func SetExecutor(e Executor) {
	defExc = e
}

// GetExecutor
func GetExecutor() Executor {
	return defExc
}

// Closer means the component need be closeFunc
type Closer interface {
	Close()
}

// Store used to persist obj
type Store interface {
	Closer
	WithTransaction(ctx context.Context, fn func(context.Context, Store) error) error
	// WithTransaction(ctx context.Context, fn func(sessCtx mongo.SessionContext) error) error
	CreateToken(token *entity.Token) error
	UpdateToken(token *entity.Token) error
	DeleteToken(id string) error
	GetTokenByUserID(userID string) (*entity.Token, error)
	CreateClient(clientName, clientID, clientSecret string) error
	GetClient(clientName string) (client *entity.Client, err error)
	RemoveClient(clientName string) (err error)
	CreateDag(ctx context.Context, dag *entity.Dag) (string, error)
	BatchCreateDag(ctx context.Context, dags []*entity.Dag) ([]*entity.Dag, error)
	CreateDagIns(ctx context.Context, dagIns *entity.DagInstance) (string, error)
	BatchCreateDagIns(ctx context.Context, dagIns []*entity.DagInstance) ([]*entity.DagInstance, error)
	BatchDeleteDagIns(ctx context.Context, ids []string) error
	CreateTaskIns(ctx context.Context, taskIns *entity.TaskInstance) error
	BatchCreateTaskIns(ctx context.Context, taskIns []*entity.TaskInstance) ([]*entity.TaskInstance, error)
	PatchTaskIns(ctx context.Context, taskIns *entity.TaskInstance) error
	PatchDagIns(ctx context.Context, dagIns *entity.DagInstance, mustsPatchFields ...string) error
	UpdateDag(ctx context.Context, dagIns *entity.Dag) error
	UpdateDagIncValue(ctx context.Context, dagId string, incKey string, incValue any) error
	UpdateDagIns(ctx context.Context, dagIns *entity.DagInstance) error
	UpdateTaskIns(ctx context.Context, taskIns *entity.TaskInstance) error
	BatchUpdateDagIns(ctx context.Context, dagIns []*entity.DagInstance) error
	BatchUpdateTaskIns(taskIns []*entity.TaskInstance) error
	BatchDeleteTaskIns(ctx context.Context, ids []string) error
	GetTaskIns(ctx context.Context, taskIns string) (*entity.TaskInstance, error)
	GetDag(ctx context.Context, dagId string) (*entity.Dag, error)
	GetDagByFields(ctx context.Context, params map[string]interface{}) (*entity.Dag, error)
	GetDagWithOptionalVersion(ctx context.Context, dagID, versionID string) (*entity.Dag, error)
	GetDagInstance(ctx context.Context, dagInsId string) (*entity.DagInstance, error)
	GetDagInstanceByFields(ctx context.Context, params map[string]interface{}) (*entity.DagInstance, error)
	ListDag(ctx context.Context, input *ListDagInput) ([]*entity.Dag, error)
	ListDagByFields(ctx context.Context, filter bson.M, opt options.FindOptions) ([]*entity.Dag, error)
	ListDagInstance(ctx context.Context, input *ListDagInstanceInput) ([]*entity.DagInstance, error)
	DisdinctDagInstance(input *ListDagInstanceInput) ([]interface{}, error)
	ListTaskInstance(ctx context.Context, input *ListTaskInstanceInput) ([]*entity.TaskInstance, error)
	Marshal(obj interface{}) ([]byte, error)
	Unmarshal(bytes []byte, ptr interface{}) error
	BatchDeleteDagWithTransaction(ctx context.Context, ids []string) error
	GetDagCount(ctx context.Context, params map[string]interface{}) (int64, error)
	ListDagCount(ctx context.Context, input *ListDagInput) (int64, error)
	ListDagCountByFields(ctx context.Context, filter bson.M) (int64, error)
	GetDagInstanceCount(ctx context.Context, params map[string]interface{}) (int64, error)
	CreateInbox(ctx context.Context, msg *entity.InBox) error
	DeleteInbox(ctx context.Context, ids []string) error
	GetInbox(ctx context.Context, id string) (*entity.InBox, error)
	ListInbox(ctx context.Context, input *ListInboxInput) ([]*entity.InBox, error)
	GetSwitchStatus() (bool, error)
	SetSwitchStatus(status bool) error
	CreateLogs(ctx context.Context, ossLogs []*entity.Log) error
	ListHistoryDagIns(ctx context.Context, params map[string]interface{}, dataChannel chan []bson.M) error
	ListHistoryTaskIns(ctx context.Context, params map[string]interface{}, dataChannel chan []bson.M) error
	DeleteDagInsByID(ctx context.Context, params map[string]interface{}) error
	DeleteTaskInsByID(ctx context.Context, params map[string]interface{}) error
	DeleteTaskInsByDagInsID(ctx context.Context, dagInsID string) error
	GetTaskInstanceCount(ctx context.Context, params map[string]interface{}) (int64, error)
	CreatOutBoxMessage(ctx context.Context, outBox *entity.OutBox) error
	BatchCreatOutBoxMessage(ctx context.Context, outBox []*entity.OutBox) error
	DeleteOutBoxMessage(ctx context.Context, ids []string) error
	ListOutBoxMessage(ctx context.Context, input *entity.OutBoxInput) ([]*entity.OutBox, error)
	ListDagInstanceInRangeTime(ctx context.Context, status string, begin, end int64) ([]*entity.DagInstance, error)
	ListExistDagInsID(ctx context.Context, dagInsIDs []string) ([]string, error)
	ListExistDagID(ctx context.Context, dagIDs []string) ([]string, error)
	GroupDagInstance(ctx context.Context, input *GroupInput) ([]*entity.DagInstanceGroup, error)
	RetryDagIns(ctx context.Context, dagInsID string, taskInsIDs []string) error

	// DeleteDag 删除Dag配置,仅在组合算子注册失败时，删除dag配置时使用
	DeleteDag(ctx context.Context, id ...string) error
	CreateDagVersion(ctx context.Context, dagVersion *entity.DagVersion) (string, error)
	ListDagVersions(ctx context.Context, input *ListDagVersionInput) ([]entity.DagVersion, error)
	GetHistoryDagByVersionID(ctx context.Context, dagID, versionID string) (*entity.DagVersion, error)
}

// ListDagInput
type ListDagInput struct {
	UserID         string
	DagIDs         []string
	Trigger        []string
	Sources        []string
	KeyWord        string
	Limit          int64
	Offset         int64
	Order          int64
	SortBy         string
	Accessors      []string
	TriggerExclude []string
	TriggerType    string
	TriggerTypes   []string
	Type           string
	SelectField    []string
	Status         []entity.DagStatus
	BizDomainID    string
	Scope          string
}

// ListDagInstanceInput
type ListDagInstanceInput struct {
	Worker        string
	DagIDs        []string
	UpdatedEnd    int64
	Status        []entity.DagInstanceStatus
	HasCmd        bool
	Limit         int64
	Offset        int64
	Order         int64
	SortBy        string
	DistinctField string
	UserIDs       []string
	Priority      []interface{}
	TimeRange     *TimeRangeSearch
	ExcludeModeVM bool
	MatchQuery    *MatchQuery
	SelectField   []string
}

type TimeRangeSearch struct {
	Begin int64
	End   int64
	Field string
}

// MatchQuery 模糊搜索
type MatchQuery struct {
	Field string
	Value interface{}
}

// ListTaskInstanceInput
type ListTaskInstanceInput struct {
	IDs      []string
	DagInsID string
	Status   []entity.TaskInstanceStatus
	// query expired tasks(it will calculate task's timeout)
	Expired         bool
	SelectField     []string
	Order           int64
	SortBy          string
	DagInsIDs       []string
	ActionName      []string
	ActionNameRegex string
	Limit           int64
	Offset          int64
	Hash            string
}

// ListInboxInput
type ListInboxInput struct {
	DocID  string
	Topics []string
	Now    int64
	Limit  int64
	Offset int64
	Order  int64
	SortBy string
}

// ListOutboxInput list outbox参数
type ListOutboxInput struct {
	ID     string
	Now    int64
	Limit  int64
	Offset int64
	Order  int64
	SortBy string
}

type GroupInput struct {
	SearchOptions []*SearchOption
	TimeRange     *TimeRangeSearch
	Order         int64
	SortBy        string
	GroupBy       string
	GroupBys      []string
	// 是否按分组返回分组数量
	IsSum bool
	// 是否返回分组内最新记录
	IsFirst bool
	// 限制数据返回总数
	Limit int64
	// ProjectFields 分组返回指定字段
	ProjectFields []string
}

type SearchOption struct {
	Field     string
	Value     interface{}
	Condition string
}

func (g *GroupInput) BuildQuery() mongo.Pipeline {
	var pipeline mongo.Pipeline
	if len(g.SearchOptions) > 0 {
		query := bson.D{}
		for _, val := range g.SearchOptions {
			query = append(query, bson.E{Key: val.Field, Value: bson.M{val.Condition: val.Value}})
		}
		if g.TimeRange != nil {
			query = append(query, bson.E{Key: g.TimeRange.Field, Value: bson.M{"$gte": g.TimeRange.Begin, "$lte": g.TimeRange.End}})
		}
		pipeline = append(pipeline, bson.D{{Key: "$match", Value: query}})
	}

	if g.SortBy != "" {
		pipeline = append(pipeline, bson.D{{Key: "$sort", Value: bson.D{{Key: g.SortBy, Value: g.Order}}}})
	}

	if g.GroupBy != "" {
		groupQuery := bson.D{{Key: "_id", Value: "$" + g.GroupBy}}
		if g.IsSum {
			groupQuery = append(groupQuery, bson.E{Key: "total", Value: bson.D{{Key: "$sum", Value: 1}}})
		}
		if g.IsFirst {
			groupQuery = append(groupQuery, bson.E{Key: "latestRecord", Value: bson.D{{Key: "$first", Value: "$$ROOT"}}})
		}
		pipeline = append(pipeline, bson.D{{Key: "$group", Value: groupQuery}})
	}

	if len(g.GroupBys) > 0 {
		groupBys := bson.D{}
		for _, v := range g.GroupBys {
			groupBys = append(groupBys, bson.E{Key: v, Value: "$" + v})
		}
		groupQuery := bson.D{{Key: "_id", Value: groupBys}}
		if g.IsSum {
			groupQuery = append(groupQuery, bson.E{Key: "total", Value: bson.D{{Key: "$sum", Value: 1}}})
		}
		if g.IsFirst {
			groupQuery = append(groupQuery, bson.E{Key: "latestRecord", Value: bson.D{{Key: "$first", Value: "$$ROOT"}}})
		}
		pipeline = append(pipeline, bson.D{{Key: "$group", Value: groupQuery}})
	}

	if g.Limit > 0 {
		pipeline = append(pipeline, bson.D{{Key: "$limit", Value: g.Limit}})
	}

	projectQuery := bson.D{{Key: "_id", Value: 0}}
	if g.IsSum {
		projectQuery = append(projectQuery, bson.E{Key: "total", Value: 1})
	}
	if g.IsFirst {
		// 仅返回 latestRecord 指定子字段
		if len(g.ProjectFields) > 0 {
			for _, f := range g.ProjectFields {
				projectQuery = append(projectQuery, bson.E{Key: "latestRecord." + f, Value: 1})
			}
		} else {
			// 不指定则默认全量 latestRecord
			projectQuery = append(projectQuery, bson.E{Key: "latestRecord", Value: 1})
		}
	}
	pipeline = append(pipeline, bson.D{{Key: "$project", Value: projectQuery}})

	return pipeline
}

// ListDagVersionInput 列举流程历史版本输入参数
type ListDagVersionInput struct {
	DagID       string
	Limit       int64
	Offset      int64
	Order       int64
	SortBy      string
	SelectField []string
}

// SetStore
func SetStore(e Store) {
	defStore = e
}

// GetStore
func GetStore() Store {
	return defStore
}

// Keeper
type Keeper interface {
	Closer
	IsLeader() bool
	IsAlive(workerKey string) (bool, error)
	AliveNodes() ([]string, error)
	WorkerKey() string
	WorkerNumber() int
}

// SetKeeper
func SetKeeper(e Keeper) {
	defKeeper = e
}

// GetKeeper
func GetKeeper() Keeper {
	return defKeeper
}

// Parser used to execute command, init dag instance and push task instance
type Parser interface {
	InitialDagIns(ctx context.Context, dagIns *entity.DagInstance)
	EntryTaskIns(taskIns *entity.TaskInstance)
	RunDagIns(dagIns *entity.DagInstance) error
}

// SetParser
func SetParser(e Parser) {
	defParser = e
}

// GetParser
func GetParser() Parser {
	return defParser
}

// SecurityPolicyProcResultMsg 安全策略结果消息
type SecurityPolicyProcResultMsg struct {
	PID        string `json:"pid"`
	Result     string `json:"result"`
	PolicyType string `json:"policy_type"`
}
