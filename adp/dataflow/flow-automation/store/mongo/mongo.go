// Package mongo 数据库操作
package mongo

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"slices"

	jsoniter "github.com/json-iterator/go"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/common"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/telemetry/trace"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/entity"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/event"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/log"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/mod"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/utils"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/utils/data"
	"github.com/shiningrush/goevent"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.opentelemetry.io/otel/attribute"
)

// IndexExpiredTime 索引自动过期时间
const IndexExpiredTime = 60

var nonContext = context.TODO()

// StoreOption Store 配置项
type StoreOption struct {
	// mongo connection string
	ConnStr  string
	Database string
	// Timeout access mongo timeout.default 5s
	Timeout time.Duration
	// the prefix will append to the database
	Prefix  string
	MaxPool *uint64
	MinPool *uint64
}

// Store 结构体
type Store struct {
	config            *common.Config
	opt               *StoreOption
	dagClsName        string
	dagInsClsName     string
	taskInsClsName    string
	tokenClsName      string
	inboxClsName      string
	clientClsName     string
	switchClsName     string
	logClsName        string
	outBoxClsName     string
	dagVersionClsName string

	mongoClient *mongo.Client
	mongoDB     *mongo.Database
}

// NewStore 创建Store实例
func NewStore(option *StoreOption) *Store {
	return &Store{
		config: common.NewConfig(),
		opt:    option,
	}
}

var _ mod.Store = (*Store)(nil)

// Init store 初始化
func (s *Store) Init() error {
	if err := s.readOpt(); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), s.opt.Timeout)
	defer cancel()
	c := options.Client()
	c.ApplyURI(s.opt.ConnStr)
	c.MaxPoolSize = s.opt.MaxPool
	c.MinPoolSize = s.opt.MinPool
	client, err := mongo.Connect(ctx, c.ApplyURI(s.opt.ConnStr))
	if err != nil {
		return fmt.Errorf("connect client failed: %w", err)
	}
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		return fmt.Errorf("ping client failed: %w", err)
	}
	s.mongoClient = client
	s.mongoDB = s.mongoClient.Database(s.opt.Database)

	err = s.dropCollectionIfExists(ctx, "flow_election", "flow_heartbeat", "flow_lock", "flow_mutex")
	if err != nil {
		return err
	}

	for _, clsName := range []string{
		s.dagClsName,
		s.dagInsClsName,
		s.taskInsClsName,
		s.tokenClsName,
		s.inboxClsName,
		s.clientClsName,
		s.switchClsName,
		s.logClsName,
		s.outBoxClsName,
		s.dagVersionClsName,
	} {
		err = s.createCollection(ctx, clsName)
		if err != nil {
			return fmt.Errorf("create collection failed: %w", err)
		}
	}

	err = s.dropIndex(ctx, s.dagInsClsName, []string{"dagId_1", "dagId_1_status_1"})
	if err != nil {
		return fmt.Errorf("drop index failed: %w", err)
	}

	_, err = s.mongoDB.Collection(s.dagInsClsName).Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "createdAt", Value: 1},
			},
		},
		{
			Keys: bson.D{
				{Key: "endedAt", Value: 1},
			},
		},
		{
			Keys: bson.D{
				{Key: "status", Value: 1},
				{Key: "userid", Value: 1},
				{Key: "priority", Value: 1},
				{Key: "mode", Value: 1},
			},
		},
		{
			Keys: bson.D{
				{Key: "dagId", Value: 1},
				{Key: "status", Value: 1},
				{Key: "createdAt", Value: 1},
				{Key: "mode", Value: 1},
			},
		},
		{
			Keys: bson.D{
				{Key: "updatedAt", Value: 1},
			},
		},
		{
			Keys: bson.D{
				{Key: "status", Value: 1},
				{Key: "updatedAt", Value: 1},
				{Key: "mode", Value: 1},
			},
		},
	})

	if err != nil {
		return fmt.Errorf("create index failed: %w", err)
	}

	err = s.dropIndex(ctx, s.taskInsClsName, []string{"idx_dagInsId_taskId_unique"})
	if err != nil {
		return fmt.Errorf("drop task index failed: %w", err)
	}

	_, err = s.mongoDB.Collection(s.taskInsClsName).Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "dagInsId", Value: 1},
			},
		},
		{
			Keys: bson.D{
				{Key: "status", Value: 1},
			},
		},
		{
			Keys: bson.D{
				{Key: "updatedAt", Value: 1},
			},
		},
		{
			Keys: bson.D{
				{Key: "lastModifiedAt", Value: 1},
			},
		},
		{
			Keys: bson.D{
				{Key: "status", Value: 1},
				{Key: "updatedAt", Value: 1},
			},
		},
		{
			Keys: bson.D{
				{Key: "status", Value: 1},
				{Key: "actionName", Value: 1},
			},
		},
		{
			Keys: bson.D{
				{Key: "hash", Value: 1},
			},
		},
		// 索引：dagInsId和taskId (移除唯一限制以支持VM模式下的循环任务)
		{
			Keys: bson.D{
				{Key: "dagInsId", Value: 1},
				{Key: "taskId", Value: 1},
			},
			Options: options.Index().SetName("idx_dagInsId_taskId"),
		},
	})

	if err != nil {
		return fmt.Errorf("create index failed: %w", err)
	}

	_, err = s.mongoDB.Collection(s.inboxClsName).Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "docid", Value: 1},
				{Key: "topic", Value: 1},
			},
		},
		{
			Keys: bson.D{
				{Key: "createdAt", Value: 1},
			},
		},
	})

	if err != nil {
		return fmt.Errorf("create index failed: %w", err)
	}

	_, err = s.mongoDB.Collection(s.outBoxClsName).Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "createdAt", Value: 1},
		},
	})

	if err != nil {
		return fmt.Errorf("create index failed: %w", err)
	}

	_, err = s.mongoDB.Collection(s.dagVersionClsName).Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "dagId", Value: 1},
				{Key: "versionId", Value: 1},
			},
		},
		{
			Keys: bson.D{
				{Key: "dagId", Value: 1},
				{Key: "sortTime", Value: 1},
			},
		},
	})

	if err != nil {
		return fmt.Errorf("create index failed: %w", err)
	}

	return nil
}

func (s *Store) dropIndex(ctx context.Context, clsName string, dropIndexNames []string) error {
	var dropIndexNameMap = make(map[string]struct{})
	for _, v := range dropIndexNames {
		dropIndexNameMap[v] = struct{}{}
	}
	cursor, err := s.mongoDB.Collection(clsName).Indexes().List(ctx)
	if err != nil {
		return err
	}

	for cursor.Next(context.Background()) {
		var index bson.M
		if err := cursor.Decode(&index); err != nil {
			return err
		}
		indexName := index["name"].(string)

		if _, ok := dropIndexNameMap[indexName]; ok {
			_, err = s.mongoDB.Collection(clsName).Indexes().DropOne(ctx, indexName)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Store) createCollection(ctx context.Context, clsName string) error {
	collections, err := s.mongoDB.ListCollectionNames(ctx, map[string]interface{}{})
	if err != nil {
		return err
	}

	exists := slices.Contains(collections, clsName)

	// 如果不存在就创建
	if !exists {
		err = s.mongoDB.CreateCollection(ctx, clsName)
		if err != nil {
			// 并发情况下可能重复创建，忽略此错误
			if mongo.IsDuplicateKeyError(err) {
				return nil
			}
			return err
		}
	}

	return nil
}

func (s *Store) dropCollectionIfExists(ctx context.Context, clsNames ...string) error {
	// 列出所有 collection 名称
	collections, err := s.mongoDB.ListCollectionNames(ctx, bson.M{})
	if err != nil {
		return err
	}

	collectionMap := map[string]struct{}{}
	for _, v := range collections {
		collectionMap[v] = struct{}{}
	}

	for _, clsName := range clsNames {
		if _, ok := collectionMap[clsName]; !ok {
			continue
		}
		err := s.mongoDB.Collection(clsName).Drop(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Store) readOpt() error {
	if s.opt.ConnStr == "" {
		return fmt.Errorf("connect string cannot be empty")
	}
	if s.opt.Database == "" {
		s.opt.Database = "flow"
	}
	if s.opt.Timeout == 0 {
		s.opt.Timeout = 5 * time.Second
	}
	s.dagClsName = "dag"
	s.dagInsClsName = "dag_instance"
	s.taskInsClsName = "task_instance"
	s.tokenClsName = "token"
	s.inboxClsName = "inbox"
	s.clientClsName = "client"
	s.switchClsName = "switch"
	s.logClsName = "log"
	s.outBoxClsName = "outbox"
	s.dagVersionClsName = "dag_version"
	if s.opt.Prefix != "" {
		s.dagClsName = fmt.Sprintf("%s_%s", s.opt.Prefix, s.dagClsName)
		s.dagInsClsName = fmt.Sprintf("%s_%s", s.opt.Prefix, s.dagInsClsName)
		s.taskInsClsName = fmt.Sprintf("%s_%s", s.opt.Prefix, s.taskInsClsName)
		s.tokenClsName = fmt.Sprintf("%s_%s", s.opt.Prefix, s.tokenClsName)
		s.inboxClsName = fmt.Sprintf("%s_%s", s.opt.Prefix, s.inboxClsName)
		s.clientClsName = fmt.Sprintf("%s_%s", s.opt.Prefix, s.clientClsName)
		s.switchClsName = fmt.Sprintf("%s_%s", s.opt.Prefix, s.switchClsName)
		s.logClsName = fmt.Sprintf("%s_%s", s.opt.Prefix, s.logClsName)
		s.outBoxClsName = fmt.Sprintf("%s_%s", s.opt.Prefix, s.outBoxClsName)
		s.dagVersionClsName = fmt.Sprintf("%s_%s", s.opt.Prefix, s.dagVersionClsName)
	}

	return nil
}

// Close component when we not use it anymore
func (s *Store) Close() {
	ctx, cancel := context.WithTimeout(context.TODO(), s.opt.Timeout)
	defer cancel()

	if err := s.mongoClient.Disconnect(ctx); err != nil {
		log.Errorf("close store client failed: %s", err)
	}
}

// WithTransaction 通用事务封装，自动管理会话与超时，并接入链路追踪
func (s *Store) WithTransaction(ctx context.Context, fn func(ctx context.Context, _ mod.Store) error) error {
	var err error
	if ctx != nonContext {
		newCtx, span := trace.StartInternalSpan(ctx)
		defer func() { trace.TelemetrySpanEnd(span, err) }()
		ctx = newCtx
	}

	// 设置超时
	ctx, cancel := context.WithTimeout(ctx, s.opt.Timeout)
	defer cancel()

	// 检查是否配置了副本集，如果没有配置则直接执行函数而不使用事务
	// 这是为了支持本地开发环境中的单节点 MongoDB
	if s.config.MongoDB.ReplicaSet == "" {
		// 非副本集模式：直接执行函数，不使用事务
		log.Debugf("[Store.WithTransaction] Running without transaction (non-replica-set mode)")
		// 创建一个假的 SessionContext 来满足函数签名
		// 在非事务模式下，直接使用 context 即可
		return fn(ctx, s)
	}

	// 副本集模式：使用事务
	// 启动会话
	session, err := s.mongoClient.StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)

	// 执行事务
	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		if err := fn(sessCtx, s); err != nil {
			return nil, err
		}
		return nil, nil
	})

	return err
}

// GetSwitchStatus 获取开关状态
func (s *Store) GetSwitchStatus() (bool, error) {
	ret := new(entity.Switch)
	ctx, cancel := context.WithTimeout(context.TODO(), s.opt.Timeout)
	defer cancel()

	if err := s.mongoDB.Collection(s.switchClsName).FindOne(ctx, bson.M{"name": entity.SwitchName}).Decode(ret); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return true, nil
		}
		return false, fmt.Errorf("get switch status failed: %w", err)
	}

	return ret.Status, nil
}

// SetSwitchStatus 设置开关状态
func (s *Store) SetSwitchStatus(status bool) error {
	update := bson.M{
		"status": status,
	}
	update = bson.M{
		"$set": update,
	}
	opts := options.Update().SetUpsert(true)

	ctx, cancel := context.WithTimeout(context.TODO(), s.opt.Timeout)
	defer cancel()

	if _, err := s.mongoDB.Collection(s.switchClsName).UpdateOne(ctx, bson.M{"name": entity.SwitchName}, update, opts); err != nil {
		return fmt.Errorf("set switch status failed: %w", err)
	}

	return nil
}

// CreateToken 创建token记录
func (s *Store) CreateToken(tokenInfo *entity.Token) error {
	baseInfo := tokenInfo.GetBaseInfo()
	baseInfo.Initial()

	ctx, cancel := context.WithTimeout(context.TODO(), s.opt.Timeout)
	defer cancel()

	if _, err := s.mongoDB.Collection(s.tokenClsName).InsertOne(ctx, tokenInfo); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return fmt.Errorf("%s key[ %s ] already existed: %w", s.tokenClsName, baseInfo.ID, data.ErrDataConflicted)
		}

		return fmt.Errorf("insert instance failed: %w", err)
	}

	return nil
}

// UpdateToken 更新token记录
func (s *Store) UpdateToken(tokenInfo *entity.Token) error {
	baseInfo := tokenInfo.GetBaseInfo()
	baseInfo.Update()

	update := bson.M{
		"updatedAt":  time.Now().Unix(),
		"token":      tokenInfo.Token,
		"expires_in": tokenInfo.ExpiresIn,
	}
	update = bson.M{
		"$set": update,
	}

	ctx, cancel := context.WithTimeout(context.TODO(), s.opt.Timeout)
	defer cancel()
	ret, err := s.mongoDB.Collection(s.tokenClsName).UpdateOne(ctx, bson.M{"userid": tokenInfo.UserID}, update)
	if err != nil {
		return fmt.Errorf("update token failed: %w", err)
	}
	if ret.MatchedCount == 0 {
		return fmt.Errorf("%s has no key[ %s ] to update: %w", s.tokenClsName, baseInfo.ID, data.ErrDataNotFound)
	}
	return nil
}

// DeleteToken 删除token记录
func (s *Store) DeleteToken(id string) error {
	ctx, cancel := context.WithTimeout(context.TODO(), s.opt.Timeout)
	defer cancel()
	_, err := s.mongoDB.Collection(s.tokenClsName).DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("delete token failed: %w", err)
	}
	return nil
}

// GetTokenByUserID 获取token记录
func (s *Store) GetTokenByUserID(userID string) (*entity.Token, error) {
	ret := new(entity.Token)
	ctx, cancel := context.WithTimeout(context.TODO(), s.opt.Timeout)
	defer cancel()

	if err := s.mongoDB.Collection(s.tokenClsName).FindOne(ctx, bson.M{"userid": userID}).Decode(ret); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return ret, nil
		}
		return ret, fmt.Errorf("get token failed: %w", err)
	}

	return ret, nil
}

// CreateInbox 创建inbox记录
func (s *Store) CreateInbox(ctx context.Context, msg *entity.InBox) error {
	var err error
	if ctx != nonContext {
		newCtx, span := trace.StartInternalSpan(ctx)
		defer func() { trace.TelemetrySpanEnd(span, err) }()
		ctx = newCtx
	}

	_, err = s.genericCreate(ctx, msg, s.inboxClsName)
	return err
}

// DeleteInbox 删除inbox记录
func (s *Store) DeleteInbox(ctx context.Context, ids []string) error {
	var err error
	if ctx != nonContext {
		newCtx, span := trace.StartInternalSpan(ctx)
		defer func() { trace.TelemetrySpanEnd(span, err) }()
		ctx = newCtx
	}

	err = s.genericBatchDelete(ctx, ids, s.inboxClsName)
	return err
}

// GetInbox 获取inbox记录
func (s *Store) GetInbox(ctx context.Context, id string) (*entity.InBox, error) {
	var err error
	if ctx != nonContext {
		newCtx, span := trace.StartInternalSpan(ctx)
		defer func() { trace.TelemetrySpanEnd(span, err) }()
		ctx = newCtx
	}

	ret := new(entity.InBox)
	if err := s.genericGet(ctx, s.inboxClsName, id, ret); err != nil {
		return nil, err
	}

	return ret, nil
}

// ListInbox 列举inbox记录
func (s *Store) ListInbox(ctx context.Context, input *mod.ListInboxInput) ([]*entity.InBox, error) {
	var err error
	if ctx != nonContext {
		newCtx, span := trace.StartInternalSpan(ctx)
		defer func() { trace.TelemetrySpanEnd(span, err) }()
		ctx = newCtx
	}

	query := bson.M{}
	if input.DocID != "" {
		query["docid"] = input.DocID
	}
	if len(input.Topics) != 0 {
		query["topic"] = bson.M{
			"$in": input.Topics,
		}
	}
	if input.Now != 0 {
		query["createdAt"] = bson.M{
			"$lte": input.Now - 2*60,
		}
	}
	opt := &options.FindOptions{}
	if input.Limit > 0 {
		opt.Limit = &input.Limit
		offset := input.Limit * input.Offset
		opt.Skip = &offset
	}
	if input.SortBy != "" {
		opt.Sort = map[string]interface{}{input.SortBy: input.Order}
	}

	var ret []*entity.InBox
	err = s.genericList(ctx, &ret, s.inboxClsName, query, opt)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

// CreateDag 创建dag记录
func (s *Store) CreateDag(ctx context.Context, dag *entity.Dag) (string, error) {
	var dagID string
	var err error
	if ctx != nonContext {
		newCtx, span := trace.StartInternalSpan(ctx)
		defer func() { trace.TelemetrySpanEnd(span, err) }()
		ctx = newCtx
	}

	// check task's connection
	_, err = mod.BuildRootNode(mod.MapTasksToGetter(dag.Tasks))
	if err != nil {
		return dagID, err
	}
	dagID, err = s.genericCreate(ctx, dag, s.dagClsName)
	return dagID, err
}

// BatchCreateDag 批量创建dag记录
func (s *Store) BatchCreateDag(ctx context.Context, dags []*entity.Dag) ([]*entity.Dag, error) {
	var err error
	if ctx != nonContext {
		newCtx, span := trace.StartInternalSpan(ctx)
		defer func() { trace.TelemetrySpanEnd(span, err) }()
		ctx = newCtx
	}

	if len(dags) == 0 {
		return dags, nil
	}

	ctx, cancel := context.WithTimeout(ctx, s.opt.Timeout)
	defer cancel()

	datas := []any{}
	for _, dag := range dags {
		dag.Initial()
		// check task's connection
		_, err = mod.BuildRootNode(mod.MapTasksToGetter(dag.Tasks))
		if err != nil {
			return nil, err
		}
		datas = append(datas, dag)
	}

	var order = false
	if _, err = s.mongoDB.Collection(s.dagClsName).InsertMany(ctx, datas, &options.InsertManyOptions{Ordered: &order}); err != nil {
		return nil, fmt.Errorf("batch insert dags failed: %w", err)
	}
	return dags, nil
}

// CreateDagIns 创建dag instance记录
func (s *Store) CreateDagIns(ctx context.Context, dagIns *entity.DagInstance) (string, error) {

	var err error
	if ctx != nonContext {
		newCtx, span := trace.StartInternalSpan(ctx)
		defer func() { trace.TelemetrySpanEnd(span, err) }()
		ctx = newCtx
	}

	dagInsID, err := s.genericCreate(ctx, dagIns, s.dagInsClsName)
	return dagInsID, err
}

// BatchCreateDagIns 批量创建dag instance
func (s *Store) BatchCreateDagIns(ctx context.Context, dagIns []*entity.DagInstance) ([]*entity.DagInstance, error) {

	var err error
	if ctx != nonContext {
		_, span := trace.StartInternalSpan(ctx)
		defer func() { trace.TelemetrySpanEnd(span, err) }()
	}

	ctx, cancel := context.WithTimeout(context.TODO(), s.opt.Timeout)
	defer cancel()

	session, err := s.mongoClient.StartSession()
	if err != nil {
		return dagIns, err
	}
	defer session.EndSession(ctx)

	var dagArr = make([][]interface{}, len(dagIns)/1000+1)

	for i := range dagIns {
		index := i / 1000
		dagIns[i].Initial()
		dagArr[index] = append(dagArr[index], dagIns[i])
	}

	var order = false

	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		for i := range dagArr {
			if len(dagArr[i]) == 0 {
				log.Warnf("dagArr %d is empty", i)
				continue
			}

			if _, err := s.mongoDB.Collection(s.dagInsClsName).InsertMany(sessCtx, dagArr[i], &options.InsertManyOptions{Ordered: &order}); err != nil {
				return nil, fmt.Errorf("insert dag instance failed: %w", err)
			}
		}

		return nil, nil
	})

	if err != nil {
		return dagIns, err
	}
	return dagIns, nil
}

// CreateTaskIns 创建task instance
func (s *Store) CreateTaskIns(ctx context.Context, taskIns *entity.TaskInstance) error {
	var err error
	if ctx != nonContext {
		newCtx, span := trace.StartInternalSpan(ctx)
		defer func() { trace.TelemetrySpanEnd(span, err) }()
		ctx = newCtx
	}

	_, err = s.genericCreate(ctx, taskIns, s.taskInsClsName)
	return err
}

func (s *Store) genericCreate(ctx context.Context, input entity.BaseInfoGetter, clsName string) (string, error) {
	var id string
	var err error
	if ctx != nonContext {
		newCtx, span := trace.StartInternalSpan(ctx)
		msgStr, _ := jsoniter.MarshalToString(input)
		trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, clsName), attribute.String(trace.DB_QUERY, msgStr))
		defer func() { trace.TelemetrySpanEnd(span, err) }()
		ctx = newCtx
	}

	baseInfo := input.GetBaseInfo()
	baseInfo.Initial()
	id = baseInfo.ID

	ctx, cancel := context.WithTimeout(ctx, s.opt.Timeout)
	defer cancel()

	if _, err := s.mongoDB.Collection(clsName).InsertOne(ctx, input); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return id, fmt.Errorf("%s key[ %s ] already existed: %w", clsName, baseInfo.ID, data.ErrDataConflicted)
		}

		return id, fmt.Errorf("insert instance failed: %w", err)
	}
	return id, nil
}

// BatchCreateTaskIns 批量创建task instance
func (s *Store) BatchCreateTaskIns(ctx context.Context, taskIns []*entity.TaskInstance) ([]*entity.TaskInstance, error) {
	// ctx, cancel := context.WithTimeout(context.TODO(), s.opt.Timeout)
	// defer cancel()

	// for i := range taskIns {
	// 	taskIns[i].Initial()
	// 	if _, err := s.mongoDB.Collection(s.taskInsClsName).InsertOne(ctx, taskIns[i]); err != nil {
	// 		return fmt.Errorf("insert task instance failed: %w", err)
	// 	}
	// }
	// return nil
	var err error
	if ctx != nonContext {
		_, span := trace.StartInternalSpan(ctx)
		defer func() { trace.TelemetrySpanEnd(span, err) }()
	}

	ctx, cancel := context.WithTimeout(context.TODO(), s.opt.Timeout)
	defer cancel()

	session, err := s.mongoClient.StartSession()
	if err != nil {
		return taskIns, err
	}
	defer session.EndSession(ctx)

	var taskinsArr = make([][]interface{}, len(taskIns)/1000+1)

	for i := range taskIns {
		index := i / 1000
		taskIns[i].Initial()
		taskinsArr[index] = append(taskinsArr[index], taskIns[i])
	}

	var order = false

	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		for i := range taskinsArr {
			if _, err := s.mongoDB.Collection(s.taskInsClsName).InsertMany(sessCtx, taskinsArr[i], &options.InsertManyOptions{Ordered: &order}); err != nil {
				return nil, fmt.Errorf("insert task instance failed: %w", err)
			}
		}

		return nil, nil
	})

	if err != nil {
		return taskIns, err
	}
	return taskIns, nil
}

// PatchTaskIns 修改task instance
func (s *Store) PatchTaskIns(ctx context.Context, taskIns *entity.TaskInstance) error {
	var err error
	if ctx != nonContext {
		_, span := trace.StartInternalSpan(ctx)
		defer func() { trace.TelemetrySpanEnd(span, err) }()
	}

	if taskIns.ID == "" {
		return fmt.Errorf("id cannot be empty")
	}
	update := bson.M{
		"updatedAt": time.Now().Unix(),
	}
	if taskIns.Status != "" {
		update["status"] = taskIns.Status
	}
	if taskIns.Reason != "" {
		update["reason"] = taskIns.Reason
	}
	if len(taskIns.Traces) > 0 {
		update["traces"] = taskIns.Traces
	}
	if taskIns.Results != nil {
		update["results"] = taskIns.Results
	}
	if taskIns.LastModifiedAt != 0 {
		update["lastModifiedAt"] = taskIns.LastModifiedAt
	}
	if taskIns.RenderedParams != nil {
		update["renderedParams"] = taskIns.RenderedParams
	}

	if taskIns.DependOn != nil {
		update["dependOn"] = taskIns.DependOn
	}

	if taskIns.Hash != "" {
		update["hash"] = taskIns.Hash
	}

	if taskIns.MetaData != nil {
		update["metadata"] = taskIns.MetaData
	}

	update = bson.M{
		"$set": update,
	}

	ctx, cancel := context.WithTimeout(context.TODO(), s.opt.Timeout)
	defer cancel()
	if _, err := s.mongoDB.Collection(s.taskInsClsName).UpdateOne(ctx, bson.M{"_id": taskIns.ID}, update); err != nil {
		return fmt.Errorf("patch task instance failed: %w", err)
	}
	return nil
}

// PatchDagIns 修改dag instance
func (s *Store) PatchDagIns(ctx context.Context, dagIns *entity.DagInstance, mustsPatchFields ...string) error {
	var err error

	if ctx != nonContext {
		_, span := trace.StartInternalSpan(ctx)
		defer func() { trace.TelemetrySpanEnd(span, err) }()
	}

	update := bson.M{
		"updatedAt": time.Now().Unix(),
	}

	if dagIns.EndedAt != 0 {
		update["endedAt"] = dagIns.EndedAt
	}

	if dagIns.EventPersistence == 0 {
		if dagIns.ShareData != nil {
			if dagIns.ShareDataExt != nil {
				update["shareDataExt"] = dagIns.ShareDataExt
				update["shareData"] = nil
			} else {
				update["shareDataExt"] = nil
				update["shareData"] = dagIns.ShareData
			}
		}

		if dagIns.Dump != "" {
			if dagIns.DumpExt != nil {
				update["dumpExt"] = dagIns.DumpExt
				update["dump"] = ""
			} else {
				update["dumpExt"] = nil
				update["dump"] = dagIns.Dump
			}
		}
	}

	if dagIns.EventPersistence != 0 {
		update["eventPersistence"] = dagIns.EventPersistence
	}

	if dagIns.EventOssPath != "" {
		update["eventOssPath"] = dagIns.EventOssPath
	}

	if dagIns.Status != "" {
		update["status"] = dagIns.Status
	}
	if utils.StringsContain(mustsPatchFields, "Cmd") || dagIns.Cmd != nil {
		update["cmd"] = dagIns.Cmd
	}
	if dagIns.Worker != "" {
		update["worker"] = dagIns.Worker
	}
	if utils.StringsContain(mustsPatchFields, "Reason") || dagIns.Reason != "" {
		update["reason"] = dagIns.Reason
	}

	if dagIns.ResumeData != "" {
		update["resume_data"] = dagIns.ResumeData
	}

	if dagIns.ResumeStatus != "" {
		update["resume_status"] = dagIns.ResumeStatus
	}

	if dagIns.Source != "" {
		update["source"] = dagIns.Source
	}

	update = bson.M{
		"$set": update,
	}

	if len(dagIns.Keywords) > 0 {
		update["$addToSet"] = bson.M{
			"keywords": bson.M{"$each": dagIns.Keywords},
		}
	}

	ctx, cancel := context.WithTimeout(context.TODO(), s.opt.Timeout)
	defer cancel()

	session, err := s.mongoClient.StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)

	if _, err := s.mongoDB.Collection(s.dagInsClsName).UpdateOne(ctx, bson.M{"_id": dagIns.ID}, update); err != nil {
		return fmt.Errorf("patch dag instance failed: %w", err)
	}

	goevent.Publish(&event.DagInstancePatched{
		Payload:         dagIns,
		MustPatchFields: mustsPatchFields,
	})

	return nil
}

// UpdateDag 更新dag
func (s *Store) UpdateDag(ctx context.Context, dag *entity.Dag) error {
	var err error
	if ctx != nonContext {
		newCtx, span := trace.StartInternalSpan(ctx)
		defer func() { trace.TelemetrySpanEnd(span, err) }()
		ctx = newCtx
	}

	// check task's connection
	_, err = mod.BuildRootNode(mod.MapTasksToGetter(dag.Tasks))
	if err != nil {
		return err
	}
	err = s.genericUpdate(ctx, dag, s.dagClsName)
	return err
}

func (s *Store) UpdateDagIncValue(ctx context.Context, dagId string, incKey string, incValue any) error {
	var err error
	if ctx != nonContext {
		newCtx, span := trace.StartInternalSpan(ctx)
		defer func() { trace.TelemetrySpanEnd(span, err) }()
		ctx = newCtx
	}

	session, sessErr := s.mongoClient.StartSession()
	if sessErr != nil {
		return sessErr
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(ctx, func(sc mongo.SessionContext) (any, error) {
		_, uerr := s.mongoDB.Collection(s.dagClsName).UpdateOne(sc,
			bson.M{
				"_id": dagId,
				"$or": bson.A{
					bson.M{"inc_values": bson.M{"$exists": false}},
					bson.M{"inc_values": nil},
					bson.M{"$expr": bson.M{"$ne": bson.A{bson.M{"$type": "$inc_values"}, "object"}}},
				},
			},
			bson.M{
				"$set": bson.M{
					"inc_values": bson.M{},
				},
			},
		)

		if uerr != nil {
			return nil, uerr
		}

		update := bson.M{
			"$set": bson.M{
				"inc_values." + incKey: incValue,
			},
		}
		if _, uerr := s.mongoDB.Collection(s.dagClsName).UpdateOne(sc, bson.M{"_id": dagId}, update, options.Update().SetUpsert(true)); uerr != nil {
			return nil, uerr
		}
		return nil, nil
	})

	if err != nil {
		return fmt.Errorf("update dag inc value failed: %w", err)
	}
	return nil
}

// UpdateDagIns 更新dag instance
func (s *Store) UpdateDagIns(ctx context.Context, dagIns *entity.DagInstance) error {
	var err error
	if ctx != nonContext {
		newCtx, span := trace.StartInternalSpan(ctx)
		defer func() { trace.TelemetrySpanEnd(span, err) }()
		ctx = newCtx
	}

	if err := s.genericUpdate(ctx, dagIns, s.dagInsClsName); err != nil {
		return err
	}

	goevent.Publish(&event.DagInstanceUpdated{Payload: dagIns})
	return nil
}

// UpdateTaskIns // 更新task instance
func (s *Store) UpdateTaskIns(ctx context.Context, taskIns *entity.TaskInstance) error {
	var err error
	if ctx != nonContext {
		newCtx, span := trace.StartInternalSpan(ctx)
		defer func() { trace.TelemetrySpanEnd(span, err) }()
		ctx = newCtx
	}

	err = s.genericUpdate(ctx, taskIns, s.taskInsClsName)
	return err
}

// genericUpdate
func (s *Store) genericUpdate(ctx context.Context, input entity.BaseInfoGetter, clsName string) error {
	var err error
	if ctx != nonContext {
		newCtx, span := trace.StartInternalSpan(ctx)
		msgStr, _ := jsoniter.MarshalToString(input)
		trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, clsName), attribute.String(trace.DB_QUERY, msgStr))
		defer func() { trace.TelemetrySpanEnd(span, err) }()
	}

	baseInfo := input.GetBaseInfo()
	baseInfo.Update()

	ctx, cancel := context.WithTimeout(ctx, s.opt.Timeout)
	defer cancel()
	ret, err := s.mongoDB.Collection(clsName).ReplaceOne(ctx, bson.M{"_id": baseInfo.ID}, input)
	if err != nil {
		return fmt.Errorf("update dag instance failed: %w", err)
	}
	if ret.MatchedCount == 0 {
		return fmt.Errorf("%s has no key[ %s ] to update: %w", clsName, baseInfo.ID, data.ErrDataNotFound)
	}
	return nil
}

// BatchUpdateDagIns 批量更新dag instance
func (s *Store) BatchUpdateDagIns(ctx context.Context, dagIns []*entity.DagInstance) error {
	var err error
	if ctx != nonContext {
		_, span := trace.StartInternalSpan(ctx)
		defer func() { trace.TelemetrySpanEnd(span, err) }()
	}

	ctx, cancel := context.WithTimeout(context.TODO(), s.opt.Timeout)
	defer cancel()

	if err != nil {
		return err
	}

	var dagInsArr = make([][]*entity.DagInstance, len(dagIns)/1000+1)

	for i := range dagIns {
		index := i / 1000
		dagIns[i].Update()
		dagInsArr[index] = append(dagInsArr[index], dagIns[i])
	}

	for i := range dagInsArr {
		var bulkOps []mongo.WriteModel
		for j := range dagInsArr[i] {
			_dafIns := dagInsArr[i][j]
			bulkOps = append(bulkOps, mongo.NewReplaceOneModel().SetFilter(bson.M{"_id": _dafIns.ID}).SetReplacement(*_dafIns))
		}
		if len(bulkOps) == 0 {
			continue
		}
		if _, err := s.mongoDB.Collection(s.dagInsClsName).BulkWrite(ctx, bulkOps); err != nil {
			return fmt.Errorf("update dag instances failed: %w", err)
		}
	}

	if err != nil {
		return err
	}
	return nil

	// errChan := make(chan error)
	// defer close(errChan)

	// errs := &data.Errors{}
	// go func() {
	// 	for err := range errChan {
	// 		errs.Append(err)
	// 	}
	// }()

	// wg := sync.WaitGroup{}
	// for i := range dagIns {
	// 	wg.Add(1)
	// 	go func(dag *entity.DagInstance, ch chan error) {
	// 		dag.Update()
	// 		if _, err := s.mongoDB.Collection(s.dagInsClsName).ReplaceOne(
	// 			ctx,
	// 			bson.M{"_id": dag.ID}, dag); err != nil {
	// 			ch <- fmt.Errorf("batch update dag instance failed: %w", err)
	// 		}
	// 		wg.Done()
	// 	}(dagIns[i], errChan)
	// }
	// wg.Wait()
	// return nil
}

// BatchUpdateTaskIns 批量更新task instance
func (s *Store) BatchUpdateTaskIns(taskIns []*entity.TaskInstance) error {
	ctx, cancel := context.WithTimeout(context.TODO(), s.opt.Timeout)
	defer cancel()
	for i := range taskIns {
		taskIns[i].Update()
		if _, err := s.mongoDB.Collection(s.taskInsClsName).ReplaceOne(
			ctx,
			bson.M{"_id": taskIns[i].ID}, taskIns[i]); err != nil {
			return fmt.Errorf("batch update task instance failed: %w", err)
		}
	}
	return nil
}

// GetTaskIns 获取task instance
func (s *Store) GetTaskIns(ctx context.Context, taskInsID string) (*entity.TaskInstance, error) {
	var err error
	if ctx != nonContext {
		newCtx, span := trace.StartInternalSpan(ctx)
		defer func() { trace.TelemetrySpanEnd(span, err) }()
		ctx = newCtx
	}

	ret := new(entity.TaskInstance)
	if err := s.genericGet(ctx, s.taskInsClsName, taskInsID, ret); err != nil {
		return nil, err
	}

	return ret, nil
}

// GetDag 获取dag
func (s *Store) GetDag(ctx context.Context, dagID string) (*entity.Dag, error) {
	var err error
	if ctx != nonContext {
		newCtx, span := trace.StartInternalSpan(ctx)
		defer func() { trace.TelemetrySpanEnd(span, err) }()
		ctx = newCtx
	}

	ret := new(entity.Dag)
	if err := s.genericGet(ctx, s.dagClsName, dagID, ret); err != nil {
		return nil, err
	}

	return ret, nil
}

// GetDagByFields 根据fields获取dag
func (s *Store) GetDagByFields(ctx context.Context, params map[string]interface{}) (*entity.Dag, error) {
	var err error
	if ctx != nonContext {
		newCtx, span := trace.StartInternalSpan(ctx)
		defer func() { trace.TelemetrySpanEnd(span, err) }()
		ctx = newCtx
	}

	ret := new(entity.Dag)
	if err := s.genericGetByFields(ctx, s.dagClsName, params, ret); err != nil {
		return nil, err
	}

	return ret, nil
}

// GetDagWithOptionalVersion 根据 dagID 和 versionID 获取 dag 配置, versionID 为空时获取最新版本的 dag 配置
func (s *Store) GetDagWithOptionalVersion(ctx context.Context, dagID, versionID string) (*entity.Dag, error) {
	var err error
	if ctx != nonContext {
		newCtx, span := trace.StartInternalSpan(ctx)
		defer func() { trace.TelemetrySpanEnd(span, err) }()
		ctx = newCtx
	}

	ret := &entity.Dag{}
	if versionID != "" {
		query := bson.M{
			"dagId":     dagID,
			"versionId": versionID,
		}

		res := &entity.DagVersion{}
		err = s.genericGetByFields(ctx, s.dagVersionClsName, query, res)
		if err != nil {
			return ret, err
		}

		ret, err = res.Config.ParseToDag()
	} else {
		err = s.genericGet(ctx, s.dagClsName, dagID, ret)
	}

	return ret, err
}

// GetDagInstance 获取dag instance
func (s *Store) GetDagInstance(ctx context.Context, dagInsID string) (*entity.DagInstance, error) {
	var err error
	if ctx != nonContext {
		newCtx, span := trace.StartInternalSpan(ctx)
		defer func() { trace.TelemetrySpanEnd(span, err) }()
		ctx = newCtx
	}

	ret := new(entity.DagInstance)
	if err := s.genericGet(ctx, s.dagInsClsName, dagInsID, ret); err != nil {
		return nil, err
	}

	return ret, nil
}

// GetDagInstanceByFields 依据fields获取dag instance
func (s *Store) GetDagInstanceByFields(ctx context.Context, params map[string]interface{}) (*entity.DagInstance, error) {
	var err error
	if ctx != nonContext {
		newCtx, span := trace.StartInternalSpan(ctx)
		defer func() { trace.TelemetrySpanEnd(span, err) }()
		ctx = newCtx
	}

	ret := new(entity.DagInstance)
	if err := s.genericGetByFields(ctx, s.dagInsClsName, params, ret); err != nil {
		return nil, err
	}

	return ret, nil
}

func (s *Store) genericGet(ctx context.Context, clsName, id string, ret interface{}) error {
	var err error
	if ctx != nonContext {
		newCtx, span := trace.StartInternalSpan(ctx)
		trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, clsName), attribute.String(trace.DB_QUERY, id))
		defer func() { trace.TelemetrySpanEnd(span, err) }()
	}

	ctx, cancel := context.WithTimeout(context.TODO(), s.opt.Timeout)
	defer cancel()

	query := bson.M{"_id": id}

	_, ok := query["removed"]
	if clsName == s.dagClsName && !ok {
		query["removed"] = bson.M{
			"$in": []interface{}{false, nil}, // matches both false and non-existent field
		}
	}

	if err := s.mongoDB.Collection(clsName).FindOne(ctx, query).Decode(ret); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return err
			// return fmt.Errorf("%s key[ %s ] not found: %w", clsName, id, data.ErrDataNotFound)
		}
		return fmt.Errorf("get dag instance failed: %w", err)
	}

	return nil
}

// genericGetByFields query by custom fields
func (s *Store) genericGetByFields(ctx context.Context, clsName string, params map[string]interface{}, ret interface{}) error {
	var err error
	if ctx != nonContext {
		_, span := trace.StartInternalSpan(ctx)
		msgStr, _ := jsoniter.MarshalToString(params)
		trace.SetAttributes(ctx, attribute.String(trace.TABLE_NAME, clsName), attribute.String(trace.DB_QUERY, msgStr))
		defer func() { trace.TelemetrySpanEnd(span, err) }()
	}

	ctx, cancel := context.WithTimeout(context.TODO(), s.opt.Timeout)
	defer cancel()

	filter := bson.M{}
	for k, v := range params {
		filter[k] = v
	}

	_, ok := params["removed"]
	if clsName == s.dagClsName && !ok {
		filter["removed"] = bson.M{
			"$in": []interface{}{false, nil}, // matches both false and non-existent field
		}
	}

	if err := s.mongoDB.Collection(clsName).FindOne(ctx, filter).Decode(ret); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return err
		}
		return fmt.Errorf("get dag instance failed: %w", err)
	}

	return nil
}

// ListDag 列举dag
func (s *Store) ListDag(ctx context.Context, input *mod.ListDagInput) ([]*entity.Dag, error) {
	var err error
	if ctx != nonContext {
		newCtx, span := trace.StartInternalSpan(ctx)
		defer func() { trace.TelemetrySpanEnd(span, err) }()
		ctx = newCtx
	}

	// 排除软删除和安全策略的工作流
	query := bson.M{
		"type": bson.M{
			"$nin": []string{
				common.DagTypeSecurityPolicy,
				common.DagTypeDataFlow,
				common.DagTypeDataFlowForBot,
				common.DagTypeComboOperator,
			},
		},
	}
	if input.Type != "" {
		if input.Type == "all" {
			delete(query, "type")
		} else {
			query["type"] = input.Type
		}
	}

	if input.TriggerType != "" {
		query["trigger"] = input.TriggerType
	}

	if len(input.TriggerTypes) > 0 {
		query["trigger"] = bson.M{
			"$in": input.TriggerTypes,
		}
	}

	if input.UserID != "" {
		query["userid"] = input.UserID
	}
	if input.KeyWord != "" {
		query["name"] = bson.M{
			"$regex": input.KeyWord,
		}
	}

	if len(input.DagIDs) > 0 {
		query["_id"] = bson.M{
			"$in": input.DagIDs,
		}
	}

	if len(input.Status) > 0 {
		query["status"] = bson.M{
			"$in": input.Status,
		}
	}

	if len(input.Trigger) > 0 {
		if len(input.Sources) != 0 {
			query["$or"] = []bson.M{
				{
					"steps": bson.M{
						"$elemMatch": bson.M{
							"operator": bson.M{
								"$in": input.Trigger,
							},
							"$or": bson.A{
								bson.D{{Key: "parameters.docid", Value: bson.M{
									"$in": input.Sources,
								}}},
								bson.D{{Key: "parameters.docids", Value: bson.M{
									"$in": input.Sources,
								}}},
							},
						},
					},
				},
				{
					"trigger_config.parameters.docids": bson.M{
						"$in": input.Sources,
					},
					"trigger_config.operator": bson.M{
						"$in": input.Trigger,
					},
				},
			}
		} else {
			query["$or"] = []bson.M{
				{
					"steps": bson.M{
						"$elemMatch": bson.M{
							"operator": bson.M{
								"$in": input.Trigger,
							},
						},
					},
				},
				{
					"trigger_config.operator": bson.M{
						"$in": input.Trigger,
					},
				},
			}
		}
	} else if len(input.TriggerExclude) > 0 && input.Scope != "all" {
		query["steps"] = bson.M{
			"$not": bson.M{
				"$elemMatch": bson.M{
					"operator": bson.M{
						"$in": input.TriggerExclude,
					},
				},
			},
		}
	}

	if input.Accessors != nil && input.UserID == "" {
		query["accessors.id"] = bson.M{
			"$in": input.Accessors,
		}
	}

	// scope 查询条件仅在子流程调用接口使用
	if input.Scope == "all" {
		query["$or"] = []bson.M{
			{
				"steps": bson.M{
					"$not": bson.M{
						"$elemMatch": bson.M{
							"operator": bson.M{
								"$in": input.TriggerExclude,
							},
						},
					},
				},
				"accessors.id": bson.M{
					"$in": input.Accessors,
				},
			},
			{
				"userid": input.UserID,
			},
		}
		// 子流程调用，过滤批量触发的工作流
		query["steps"] = bson.M{
			"$not": bson.M{
				"$elemMatch": bson.M{
					"datasource": bson.M{"$ne": nil, "$exists": true},
				},
			},
		}
		delete(query, "userid")
		delete(query, "accessors.id")
	}

	opt := &options.FindOptions{}
	if input.Limit > 0 {
		opt.Limit = &input.Limit
		offset := input.Limit * input.Offset
		opt.Skip = &offset
	}
	if input.SortBy != "" {
		opt.Sort = map[string]interface{}{input.SortBy: input.Order}
	}

	if len(input.SelectField) > 0 {
		fields := bson.M{}
		for _, f := range input.SelectField {
			fields[f] = 1
		}
		opt.Projection = fields
	}

	var ret []*entity.Dag
	err = s.genericList(ctx, &ret, s.dagClsName, query, opt)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (s *Store) ListDagByFields(ctx context.Context, filter bson.M, opt options.FindOptions) ([]*entity.Dag, error) {
	var err error
	if ctx != nonContext {
		newCtx, span := trace.StartInternalSpan(ctx)
		defer func() { trace.TelemetrySpanEnd(span, err) }()
		ctx = newCtx
	}

	var ret []*entity.Dag

	err = s.genericList(ctx, &ret, s.dagClsName, filter, &opt)

	if err != nil {
		return nil, err
	}

	return ret, nil
}

// ListDagCount 获取dag数量
func (s *Store) ListDagCount(ctx context.Context, input *mod.ListDagInput) (int64, error) {
	var err error
	if ctx != nonContext {
		newCtx, span := trace.StartInternalSpan(ctx)
		defer func() { trace.TelemetrySpanEnd(span, err) }()
		ctx = newCtx
	}

	query := bson.M{
		"type": bson.M{
			"$nin": []string{
				common.DagTypeSecurityPolicy,
				common.DagTypeDataFlow,
				common.DagTypeDataFlowForBot,
				common.DagTypeComboOperator,
			},
		},
	}
	if input.Type != "" {
		if input.Type == "all" {
			delete(query, "type")
		} else {
			query["type"] = input.Type
		}
	}

	if input.TriggerType != "" {
		query["trigger"] = input.TriggerType
	}

	if input.UserID != "" {
		query["userid"] = input.UserID
	}
	if input.KeyWord != "" {
		query["name"] = bson.M{
			"$regex": input.KeyWord,
		}
	}

	if len(input.DagIDs) > 0 {
		query["_id"] = bson.M{
			"$in": input.DagIDs,
		}
	}

	if len(input.Status) > 0 {
		query["status"] = bson.M{
			"$in": input.Status,
		}
	}

	if len(input.Sources) != 0 && len(input.Trigger) > 0 {
		query["steps"] = bson.M{
			"$elemMatch": bson.M{
				"operator": bson.M{
					"$in": input.Trigger,
				},
				"$or": bson.A{
					bson.D{{Key: "parameters.docid", Value: bson.M{
						"$in": input.Sources,
					}}},
					bson.D{{Key: "parameters.docids", Value: bson.M{
						"$in": input.Sources,
					}}},
				},
			},
		}
	}

	if input.Accessors != nil && input.UserID == "" {
		query["accessors.id"] = bson.M{
			"$in": input.Accessors,
		}
	}

	if input.BizDomainID != "" {
		if input.BizDomainID == common.BizDomainDefaultID {
			query["$or"] = []bson.M{
				{"biz_domain_id": ""},
				{"biz_domain_id": common.BizDomainDefaultID},
				{"biz_domain_id": bson.M{"$exists": false}},
			}
		} else {
			query["biz_domain_id"] = input.BizDomainID
		}
	}

	count, err := s.genericGetCount(ctx, s.dagClsName, query)
	if err != nil {
		return count, err
	}
	return count, nil
}

func (s *Store) ListDagCountByFields(ctx context.Context, filter bson.M) (int64, error) {
	var err error
	if ctx != nonContext {
		newCtx, span := trace.StartInternalSpan(ctx)
		defer func() { trace.TelemetrySpanEnd(span, err) }()
		ctx = newCtx
	}

	count, err := s.genericGetCount(ctx, s.dagClsName, filter)
	if err != nil {
		return count, err
	}
	return count, nil
}

// ListDagInstance 列举dag instance
func (s *Store) ListDagInstance(ctx context.Context, input *mod.ListDagInstanceInput) ([]*entity.DagInstance, error) {
	var err error
	if ctx != nonContext {
		newCtx, span := trace.StartInternalSpan(ctx)
		defer func() { trace.TelemetrySpanEnd(span, err) }()
		ctx = newCtx
	}

	query := bson.M{}
	if len(input.DagIDs) > 0 {
		query["dagId"] = bson.M{
			"$in": input.DagIDs,
		}
	}
	if len(input.Status) > 0 {
		query["status"] = bson.M{
			"$in": input.Status,
		}
	}
	if len(input.UserIDs) > 0 {
		query["userid"] = bson.M{
			"$in": input.UserIDs,
		}
	}
	if len(input.Priority) > 0 {
		query["priority"] = bson.M{
			"$in": input.Priority,
		}
	}
	if input.Worker != "" {
		query["worker"] = input.Worker
	}
	if input.UpdatedEnd > 0 {
		query["updatedAt"] = bson.M{
			"$lte": input.UpdatedEnd,
		}
	}
	if input.HasCmd {
		query["cmd"] = bson.M{
			"$ne": nil,
		}
	}

	if input.ExcludeModeVM {
		query["mode"] = bson.M{
			"$ne": entity.DagInstanceModeVM,
		}
	}

	if input.TimeRange != nil {
		query[input.TimeRange.Field] = bson.M{
			"$gte": input.TimeRange.Begin,
			"$lte": input.TimeRange.End,
		}
	}

	if input.MatchQuery != nil {
		query[input.MatchQuery.Field] = input.MatchQuery.Value
	}

	opt := &options.FindOptions{}
	if input.Limit > 0 {
		opt.Limit = &input.Limit
		offset := input.Limit * input.Offset
		opt.Skip = &offset
	}
	if input.SortBy != "" {
		opt.Sort = map[string]interface{}{input.SortBy: input.Order}
	}

	if len(input.SelectField) > 0 {
		fields := bson.M{}
		for _, f := range input.SelectField {
			fields[f] = 1
		}
		opt.Projection = fields
	}

	var ret []*entity.DagInstance
	err = s.genericList(ctx, &ret, s.dagInsClsName, query, opt)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

// DisdinctDagInstance 按列给dag instance去重
func (s *Store) DisdinctDagInstance(input *mod.ListDagInstanceInput) ([]interface{}, error) {
	query := bson.M{}
	if len(input.Status) > 0 {
		query["status"] = bson.M{
			"$in": input.Status,
		}
	}
	if len(input.DagIDs) > 0 {
		query["dagId"] = bson.M{
			"$in": input.DagIDs,
		}
	}
	if input.Worker != "" {
		query["worker"] = input.Worker
	}
	if input.UpdatedEnd > 0 {
		query["updatedAt"] = bson.M{
			"$lte": input.UpdatedEnd,
		}
	}

	if input.ExcludeModeVM {
		query["mode"] = bson.M{
			"$ne": entity.DagInstanceModeVM,
		}
	}

	opt := &options.FindOptions{}
	if input.Limit > 0 {
		opt.Limit = &input.Limit
		offset := input.Limit * input.Offset
		opt.Skip = &offset
	}
	if input.SortBy != "" {
		opt.Sort = map[string]interface{}{input.SortBy: input.Order}
	}

	ctx, cancel := context.WithTimeout(context.TODO(), s.opt.Timeout)
	defer cancel()

	res, err := s.mongoDB.Collection(s.dagInsClsName).Distinct(ctx, input.DistinctField, query)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// ListTaskInstance 列举task instance
func (s *Store) ListTaskInstance(ctx context.Context, input *mod.ListTaskInstanceInput) ([]*entity.TaskInstance, error) {
	var err error
	if ctx != nonContext {
		newCtx, span := trace.StartInternalSpan(ctx)
		defer func() { trace.TelemetrySpanEnd(span, err) }()
		ctx = newCtx
	}

	query := bson.M{}
	actionNameQuery := bson.A{}

	if len(input.IDs) > 0 {
		query["_id"] = bson.M{
			"$in": input.IDs,
		}
	}

	if len(input.ActionName) > 0 {
		actionNameQuery = append(actionNameQuery, bson.M{
			"actionName": bson.M{
				"$in": input.ActionName,
			},
		})
	}

	if input.ActionNameRegex != "" {
		actionNameQuery = append(actionNameQuery, bson.M{
			"actionName": bson.M{
				"$regex": input.ActionNameRegex,
			},
		})
	}

	if len(input.Status) > 0 {
		query["status"] = bson.M{
			"$in": input.Status,
		}
	}
	if input.Expired {
		query["$expr"] = bson.M{
			"$lte": bson.A{
				"$updatedAt",
				bson.M{
					"$subtract": bson.A{
						// delay is prevent watch dog conflicted with task's context timeout
						time.Now().Unix() - 5,
						"$timeoutSecs",
					},
				},
			},
		}
	}

	if input.DagInsID != "" {
		query["dagInsId"] = input.DagInsID
	} else if len(input.DagInsIDs) > 0 {
		query["dagInsId"] = bson.M{
			"$in": input.DagInsIDs,
		}
	}

	if input.Hash != "" {
		query["hash"] = input.Hash
	}

	if len(actionNameQuery) > 0 {
		query = bson.M{
			"$and": bson.A{
				bson.M{
					"$or": actionNameQuery,
				},
				query,
			},
		}
	}

	opt := &options.FindOptions{}
	if len(input.SelectField) > 0 {
		fields := bson.M{}
		for _, f := range input.SelectField {
			fields[f] = 1
		}
		opt.Projection = fields
	}

	if input.SortBy != "" {
		opt.Sort = map[string]interface{}{input.SortBy: input.Order}
	}

	if input.Limit > 0 {
		opt.Limit = &input.Limit
		offset := input.Limit * input.Offset
		opt.Skip = &offset
	}

	var ret []*entity.TaskInstance
	err = s.genericList(ctx, &ret, s.taskInsClsName, query, opt)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (s *Store) genericList(ctx context.Context, ret interface{}, clsName string, query bson.M, opts ...*options.FindOptions) error {
	var err error
	if ctx != nonContext {
		newCtx, span := trace.StartInternalSpan(ctx)
		msgStr, _ := jsoniter.MarshalToString(query)
		trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, clsName), attribute.String(trace.DB_QUERY, msgStr))
		defer func() { trace.TelemetrySpanEnd(span, err) }()
	}

	ctx, cancel := context.WithTimeout(context.TODO(), s.opt.Timeout)
	defer cancel()

	_, ok := query["removed"]
	if clsName == s.dagClsName && !ok {
		query["removed"] = bson.M{
			"$ne": true,
		}
	}

	_, ok = query["is_debug"]
	if clsName == s.dagClsName && !ok {
		query["is_debug"] = bson.M{
			"$ne": true,
		}
	}

	cur, err := s.mongoDB.Collection(clsName).Find(ctx, query, opts...)
	if err != nil {
		return fmt.Errorf("find %s failed: %w", clsName, err)
	}
	if err := cur.All(ctx, ret); err != nil {
		return fmt.Errorf("decode failed: %w", err)
	}
	return nil
}

// BatchDeleteDag 批量删除dag
func (s *Store) BatchDeleteDag(ctx context.Context, ids []string) error {
	var err error
	if ctx != nonContext {
		newCtx, span := trace.StartInternalSpan(ctx)
		defer func() { trace.TelemetrySpanEnd(span, err) }()
		ctx = newCtx
	}

	err = s.genericBatchDelete(ctx, ids, s.dagClsName)
	return err
}

// BatchDeleteDagIns 批量删除dag instance
func (s *Store) BatchDeleteDagIns(ctx context.Context, ids []string) error {
	var err error
	if ctx != nonContext {
		newCtx, span := trace.StartInternalSpan(ctx)
		defer func() { trace.TelemetrySpanEnd(span, err) }()
		ctx = newCtx
	}

	err = s.genericBatchDelete(ctx, ids, s.dagInsClsName)
	return err
}

// BatchDeleteTaskIns 批量删除task instance
func (s *Store) BatchDeleteTaskIns(ctx context.Context, ids []string) error {
	var err error
	if ctx != nonContext {
		newCtx, span := trace.StartInternalSpan(ctx)
		defer func() { trace.TelemetrySpanEnd(span, err) }()
		ctx = newCtx
	}

	err = s.genericBatchDelete(ctx, ids, s.taskInsClsName)
	return err
}

func (s *Store) genericBatchDelete(ctx context.Context, ids []string, clsName string) error {
	var err error
	if ctx != nonContext {
		newCtx, span := trace.StartInternalSpan(ctx)
		msgStr, _ := jsoniter.MarshalToString(ids)
		trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, clsName), attribute.String(trace.DB_QUERY, msgStr))
		defer func() { trace.TelemetrySpanEnd(span, err) }()
	}

	ctx, cancel := context.WithTimeout(ctx, s.opt.Timeout)
	defer cancel()

	_, err = s.mongoDB.Collection(clsName).DeleteMany(ctx, bson.M{
		"_id": bson.M{
			"$in": ids,
		},
	})
	if err != nil {
		return fmt.Errorf("delete failed: %w", err)
	}

	return nil
}

func (s *Store) genericBatchDeleteWithTransaction(ctx context.Context, query bson.M, clsName string) error {
	var err error
	if ctx != nonContext {
		newCtx, span := trace.StartInternalSpan(ctx)
		msgStr, _ := jsoniter.MarshalToString(query)
		trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, clsName), attribute.String(trace.DB_QUERY, msgStr))
		defer func() { trace.TelemetrySpanEnd(span, err) }()
	}

	_, err = s.mongoDB.Collection(clsName).DeleteMany(ctx, query)
	if err != nil {
		return fmt.Errorf("delete failed: %w", err)
	}

	return nil
}

// Marshal marshal
func (s *Store) Marshal(obj interface{}) ([]byte, error) {
	return bson.Marshal(obj)
}

// Unmarshal unmashal
func (s *Store) Unmarshal(bytes []byte, ptr interface{}) error {
	return bson.Unmarshal(bytes, ptr)
}

// BatchDeleteDagWithTransaction batch delete dag if exisit
func (s *Store) BatchDeleteDagWithTransaction(ctx context.Context, ids []string) error {
	var err error
	if ctx != nonContext {
		newCtx, span := trace.StartInternalSpan(ctx)
		defer func() { trace.TelemetrySpanEnd(span, err) }()
		ctx = newCtx
	}

	//  delete dag
	// 软删除工作流
	updates := bson.M{
		"$set": bson.M{
			"removed": true,
		},
	}
	if _, err := s.mongoDB.Collection(s.dagClsName).UpdateOne(ctx, bson.M{
		"_id":  bson.M{"$in": ids},
		"type": bson.M{"$ne": common.DagTypeSecurityPolicy},
	}, updates); err != nil {
		return err
	}

	query := bson.M{
		"dagId":    bson.M{"$in": ids},
		"dag_type": bson.M{"$ne": common.DagTypeSecurityPolicy},
	}

	for {
		cursor, err := s.mongoDB.Collection(s.dagInsClsName).Find(ctx, query, options.Find().SetSort(bson.M{"updatedAt": 1}).SetLimit(1000))
		if err != nil {
			return err
		}

		// 存储记录的_id
		var daginsToDelete []bson.M
		for cursor.Next(ctx) {
			var result bson.M
			if err := cursor.Decode(&result); err != nil {
				return err
			}
			daginsToDelete = append(daginsToDelete, bson.M{"_id": result["_id"]})
		}

		if len(daginsToDelete) == 0 {
			break // 没有更多记录可删除
		}

		// Define the callback that specifies the sequence of operations to perform inside the transaction.
		callback := func(sessCtx mongo.SessionContext) (interface{}, error) {
			// Important: You must pass sessCtx as the Context parameter to the operations for them to be executed in the
			// transaction.
			// 删除这些记录
			if err := s.genericBatchDeleteWithTransaction(sessCtx, bson.M{"$or": daginsToDelete}, s.dagInsClsName); err != nil {
				return nil, err
			}

			if err := s.genericBatchDeleteWithTransaction(sessCtx, bson.M{"dagInsId": bson.M{"$in": daginsToDelete}}, s.taskInsClsName); err != nil {
				return nil, err
			}
			return nil, nil
		}

		//  Start a session and run the callback using WithTransaction.
		session, err := s.mongoClient.StartSession()
		if err != nil {
			return err
		}

		_, err = session.WithTransaction(ctx, callback)
		if err != nil {
			session.EndSession(ctx)
			continue
		}
		session.EndSession(ctx)

		// 检查是否处理了所有匹配的文档
		if len(daginsToDelete) < 1000 {
			break
		}
	}

	return nil
}

// GetDagCount 获取dag数量
func (s *Store) GetDagCount(ctx context.Context, params map[string]interface{}) (int64, error) {
	var err error
	if ctx != nonContext {
		newCtx, span := trace.StartInternalSpan(ctx)
		defer func() { trace.TelemetrySpanEnd(span, err) }()
		ctx = newCtx
	}

	// 排除软删除和安全策略的工作流
	query := bson.M{
		"type": bson.M{
			"$ne": common.DagTypeSecurityPolicy,
		},
	}

	for k, v := range params {
		query[k] = v
	}
	count, err := s.genericGetCount(ctx, s.dagClsName, query)
	return count, err
}

// GetDagInstanceCount 获取dag instance数量
func (s *Store) GetDagInstanceCount(ctx context.Context, params map[string]interface{}) (int64, error) {
	var err error
	if ctx != nonContext {
		newCtx, span := trace.StartInternalSpan(ctx)
		defer func() { trace.TelemetrySpanEnd(span, err) }()
		ctx = newCtx
	}

	query := bson.M{}
	for k, v := range params {
		query[k] = v
	}
	count, err := s.genericGetCount(ctx, s.dagInsClsName, query)
	return count, err
}

func (s *Store) genericGetCount(ctx context.Context, clsName string, query bson.M, opts ...*options.CountOptions) (int64, error) {
	var err error
	if ctx != nonContext {
		newCtx, span := trace.StartInternalSpan(ctx)
		msgStr, _ := jsoniter.MarshalToString(query)
		trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, clsName), attribute.String(trace.DB_QUERY, msgStr))
		defer func() { trace.TelemetrySpanEnd(span, err) }()
	}

	// 适配日志清理线程查询超时时间
	ttl := s.opt.Timeout
	if _, ok := query["log_clean_ttl"]; ok {
		ttl = query["log_clean_ttl"].(time.Duration)
		delete(query, "log_clean_ttl")
	}

	ctx, cancel := context.WithTimeout(context.TODO(), ttl)
	defer cancel()

	_, ok := query["removed"]
	if clsName == s.dagClsName && !ok {
		query["removed"] = bson.M{
			"$ne": true,
		}
	}

	_, ok = query["is_debug"]
	if clsName == s.dagClsName && !ok {
		query["is_debug"] = bson.M{
			"$ne": true,
		}
	}

	var count int64
	count, err = s.mongoDB.Collection(clsName).CountDocuments(ctx, query, opts...)
	if err != nil {
		return count, err
	}

	return count, nil
}

// CreateClient 创建客户端
func (s *Store) CreateClient(clientName, clientID, clientSecret string) error {
	ctx, cancel := context.WithTimeout(context.TODO(), s.opt.Timeout)
	defer cancel()
	_, err := s.mongoDB.Collection(s.clientClsName).InsertOne(ctx, bson.M{
		"client_id":     clientID,
		"client_secret": clientSecret,
		"client_name":   clientName,
	})
	if err != nil {
		return fmt.Errorf("create client failed: %w", err)
	}
	return nil
}

// GetClient 获取客户端信息
func (s *Store) GetClient(clientName string) (client *entity.Client, err error) {
	ret := new(entity.Client)
	ctx, cancel := context.WithTimeout(context.TODO(), s.opt.Timeout)
	defer cancel()

	if err = s.mongoDB.Collection(s.clientClsName).FindOne(ctx, bson.M{"client_name": clientName}).Decode(ret); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return ret, nil
		}
		return
	}

	return ret, nil
}

// RemoveClient 移除客户端
func (s *Store) RemoveClient(name string) error {
	ctx, cancel := context.WithTimeout(context.TODO(), s.opt.Timeout)
	defer cancel()
	_, err := s.mongoDB.Collection(s.clientClsName).DeleteOne(ctx, bson.M{
		"client_name": name,
	})
	if err != nil {
		return fmt.Errorf("remove client failed: %w", err)
	}
	return nil
}

// CreateLog 创建日志信息
func (s *Store) CreateLogs(ctx context.Context, ossLogs []*entity.Log) error {
	var err error
	if ctx != nonContext {
		newCtx, span := trace.StartInternalSpan(ctx)
		defer func() { trace.TelemetrySpanEnd(span, err) }()
		ctx = newCtx
	}

	ctx, cancel := context.WithTimeout(ctx, s.opt.Timeout)
	defer cancel()

	session, err := s.mongoClient.StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)

	var ossLogArr = make([][]interface{}, len(ossLogs)/1000+1)

	for i := range ossLogs {
		index := i / 1000
		ossLogs[i].Initial()
		ossLogArr[index] = append(ossLogArr[index], ossLogs[i])
	}

	var order = false

	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		for i := range ossLogArr {
			if _, err := s.mongoDB.Collection(s.logClsName).InsertMany(sessCtx, ossLogArr[i], &options.InsertManyOptions{Ordered: &order}); err != nil {
				return nil, fmt.Errorf("insert oss log failed: %w", err)
			}
		}

		return nil, nil
	})

	return err
}

// ListHistoryDagIns 列举历史dag实例数据
func (s *Store) ListHistoryDagIns(ctx context.Context, params map[string]interface{}, dataChannel chan []bson.M) error {
	var err error
	if ctx != nonContext {
		newCtx, span := trace.StartInternalSpan(ctx)
		defer func() { trace.TelemetrySpanEnd(span, err) }()
		ctx = newCtx
	}

	query := bson.M{
		"$and": []bson.M{
			{"status": bson.M{"$in": params["status"]}},
			{"updatedAt": bson.M{"$lte": params["updatedAt"]}},
		},
	}
	opt := &options.FindOptions{}
	opt.Sort = map[string]interface{}{"_id": 1}

	return s.generateListWithCursor(ctx, s.dagInsClsName, query, dataChannel, opt)
}

// ListHistoryTaskIns 列举历史task实例数据
func (s *Store) ListHistoryTaskIns(ctx context.Context, params map[string]interface{}, dataChannel chan []bson.M) error {
	var err error
	if ctx != nonContext {
		newCtx, span := trace.StartInternalSpan(ctx)
		defer func() { trace.TelemetrySpanEnd(span, err) }()
		ctx = newCtx
	}

	query := bson.M{
		"$and": []bson.M{
			{"status": bson.M{"$in": params["status"]}},
			{"updatedAt": bson.M{"$lte": params["updatedAt"]}},
		},
	}
	opt := &options.FindOptions{}
	opt.Sort = map[string]interface{}{"_id": 1}

	return s.generateListWithCursor(ctx, s.taskInsClsName, query, dataChannel, opt)
}

// generateListWithCursor 使用游标列举数据
func (s *Store) generateListWithCursor(ctx context.Context, clsName string, query bson.M, dataChannel chan []bson.M, opts ...*options.FindOptions) error {
	var (
		err   error
		exist bool
	)
	if ctx != nonContext {
		newCtx, span := trace.StartInternalSpan(ctx)
		defer func() { trace.TelemetrySpanEnd(span, err) }()
		ctx = newCtx
	}

	var cursor *mongo.Cursor
	// 固定超时时间30分钟，当前一批删除数据如果超过执行时间则在下一批数据中再进行删除
	ctx, cancel := context.WithTimeout(ctx, 30*time.Minute)
	defer cancel()

	// 守护线程，收到子线程通知立即退出
	go func() {
		<-ctx.Done()
		exist = true
	}()
	// 执行查询
	cursor, err = s.mongoDB.Collection(clsName).Find(ctx, query, opts...)
	if err != nil {
		return fmt.Errorf("list history data failed: %w", err)
	}
	defer cursor.Close(ctx) //nolint

	// 创建一个切片来存储查询结果
	var results []bson.M

	// 创建一个批次计数器
	batchSize := common.DefaultQuerySize

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			// 遍历游标
			for cursor.Next(ctx) {
				if exist {
					close(dataChannel)
					return nil
				}
				var result bson.M
				if err := cursor.Decode(&result); err != nil {
					return fmt.Errorf("list history data failed: %w", err)
				}

				results = append(results, result)

				// 当达到批次大小时，将数据发送到通道，并清空切片
				if len(results) == batchSize && !exist {
					dataChannel <- results
					results = nil
				}
			}
			if len(results) > 0 && !exist {
				dataChannel <- results
			}
			close(dataChannel)
			return nil
		}
	}
}

// ListExistDagInsID 返回已存在的dagid
func (s *Store) ListExistDagInsID(ctx context.Context, dagInsIDs []string) ([]string, error) {
	var err error
	if ctx != nonContext {
		newCtx, span := trace.StartInternalSpan(ctx)
		defer func() { trace.TelemetrySpanEnd(span, err) }()
		ctx = newCtx
	}

	query := bson.M{}
	query["_id"] = bson.M{"$in": dagInsIDs}

	opt := &options.FindOptions{}
	opt.SetProjection(bson.M{"_id": 1})

	var ret []*entity.DagInstance
	err = s.genericList(ctx, &ret, s.dagInsClsName, query, opt)
	if err != nil {
		return nil, err
	}

	var result []string
	for _, val := range ret {
		result = append(result, val.ID)
	}
	return result, nil
}

// ListExistDagID 返回已存在的dagid
func (s *Store) ListExistDagID(ctx context.Context, dagIDs []string) ([]string, error) {
	var err error
	if ctx != nonContext {
		newCtx, span := trace.StartInternalSpan(ctx)
		defer func() { trace.TelemetrySpanEnd(span, err) }()
		ctx = newCtx
	}

	query := bson.M{}
	query["_id"] = bson.M{"$in": dagIDs}

	opt := &options.FindOptions{}
	opt.SetProjection(bson.M{"_id": 1})

	var ret []*entity.DagInstance
	err = s.genericList(ctx, &ret, s.dagClsName, query, opt)
	if err != nil {
		return nil, err
	}

	var result []string
	for _, val := range ret {
		result = append(result, val.ID)
	}
	return result, nil
}

// DeleteDagInsByID 根据id删除dag实例
func (s *Store) DeleteDagInsByID(ctx context.Context, params map[string]interface{}) error {
	// 构建查询条件
	query := bson.M{
		"$and": []bson.M{
			{"_id": bson.M{"$lte": params["_id"]}},
			{"status": bson.M{"$in": params["status"]}},
			{"updatedAt": bson.M{"$lte": params["updatedAt"]}},
		},
	}
	return s.genericBatchDeleteWithTransaction(ctx, query, s.dagInsClsName)
}

// DeleteTaskInsByID 根据id删除task实例
func (s *Store) DeleteTaskInsByID(ctx context.Context, params map[string]interface{}) error {
	// 构建查询条件
	query := bson.M{
		"$and": []bson.M{
			{"_id": bson.M{"$lte": params["_id"]}},
			{"dagInsId": bson.M{"$in": params["dagInsIDs"]}},
			{"status": bson.M{"$in": params["status"]}},
			{"updatedAt": bson.M{"$lte": params["updatedAt"]}},
		},
	}
	return s.genericBatchDeleteWithTransaction(ctx, query, s.taskInsClsName)
}

// DeleteTaskInsByDagInsID 根据dag实例id删除task实例
func (s *Store) DeleteTaskInsByDagInsID(ctx context.Context, dagInsID string) error {
	var err error
	if ctx != nonContext {
		_ctx, span := trace.StartInternalSpan(ctx)
		defer func() { trace.TelemetrySpanEnd(span, err) }()
		ctx = _ctx
	}

	ctx, cancel := context.WithTimeout(ctx, s.opt.Timeout)
	defer cancel()
	query := bson.M{"dagInsId": dagInsID}

	queryBytes, _ := bson.Marshal(query)
	trace.SetAttributes(ctx, attribute.String(trace.TABLE_NAME, s.taskInsClsName), attribute.String(trace.DB_QUERY, string(queryBytes)))

	return s.genericBatchDeleteWithTransaction(ctx, query, s.taskInsClsName)
}

// GetTaskInstanceCount task instance数量
func (s *Store) GetTaskInstanceCount(ctx context.Context, params map[string]interface{}) (int64, error) {
	var err error
	if ctx != nonContext {
		newCtx, span := trace.StartInternalSpan(ctx)
		defer func() { trace.TelemetrySpanEnd(span, err) }()
		ctx = newCtx
	}

	query := bson.M{}
	for k, v := range params {
		query[k] = v
	}
	count, err := s.genericGetCount(ctx, s.taskInsClsName, query)
	return count, err
}

// CreatOutBoxMessage 创建OutBox消息记录
func (s *Store) CreatOutBoxMessage(ctx context.Context, outBox *entity.OutBox) error {
	var err error
	if ctx != nonContext {
		newCtx, span := trace.StartInternalSpan(ctx)
		defer func() { trace.TelemetrySpanEnd(span, err) }()
		ctx = newCtx
	}

	baseInfo := outBox.GetBaseInfo()
	baseInfo.Initial()

	ctx, cancel := context.WithTimeout(ctx, s.opt.Timeout)
	defer cancel()

	if _, err = s.mongoDB.Collection(s.outBoxClsName).InsertOne(ctx, outBox); err != nil {
		return fmt.Errorf("insert out box message failed: %w", err)
	}
	return nil
}

// BatchCreatOutBoxMessage 批量创建OutBox消息记录
func (s *Store) BatchCreatOutBoxMessage(ctx context.Context, outBox []*entity.OutBox) error {
	var err error
	if ctx != nonContext {
		newCtx, span := trace.StartInternalSpan(ctx)
		defer func() { trace.TelemetrySpanEnd(span, err) }()
		ctx = newCtx
	}

	ctx, cancel := context.WithTimeout(ctx, s.opt.Timeout)
	defer cancel()

	var data []interface{}
	for _, val := range outBox {
		data = append(data, val)
	}

	var order = false
	if _, err = s.mongoDB.Collection(s.outBoxClsName).InsertMany(ctx, data, &options.InsertManyOptions{Ordered: &order}); err != nil {
		return fmt.Errorf("batch insert out box message failed: %w", err)
	}
	return nil
}

// ListOutBoxMessage 列举outbox消息列表
func (s *Store) ListOutBoxMessage(ctx context.Context, input *entity.OutBoxInput) ([]*entity.OutBox, error) {
	var err error
	if ctx != nonContext {
		newCtx, span := trace.StartInternalSpan(ctx)
		defer func() { trace.TelemetrySpanEnd(span, err) }()
		ctx = newCtx
	}

	query := bson.M{"createdAt": bson.M{"$lte": input.CreateTime}}
	opt := &options.FindOptions{}
	opt.Limit = &input.Limit

	var ret []*entity.OutBox
	err = s.genericList(ctx, &ret, s.outBoxClsName, query, opt)
	if err != nil {
		return nil, err
	}

	return ret, nil
}

// DeleteOutBoxMessage 根据id删除outbox消息
func (s *Store) DeleteOutBoxMessage(ctx context.Context, ids []string) error {
	var err error
	if ctx != nonContext {
		newCtx, span := trace.StartInternalSpan(ctx)
		defer func() { trace.TelemetrySpanEnd(span, err) }()
		ctx = newCtx
	}

	err = s.genericBatchDelete(ctx, ids, s.outBoxClsName)
	return err
}

// ListDagInstanceInRangeTime 列举时间范围内的DagIns记录
func (s *Store) ListDagInstanceInRangeTime(ctx context.Context, status string, begin, end int64) ([]*entity.DagInstance, error) {
	var err error
	if ctx != nonContext {
		newCtx, span := trace.StartInternalSpan(ctx)
		defer func() { trace.TelemetrySpanEnd(span, err) }()
		ctx = newCtx
	}

	query := bson.M{
		"$and": []bson.M{
			{"status": status},
			{"updatedAt": bson.M{"$gte": begin, "$lte": end}},
		},
	}

	var ret []*entity.DagInstance
	err = s.genericList(ctx, &ret, s.dagInsClsName, query, &options.FindOptions{})
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (s *Store) GroupDagInstance(ctx context.Context, input *mod.GroupInput) ([]*entity.DagInstanceGroup, error) {
	var err error
	if ctx != nonContext {
		newCtx, span := trace.StartInternalSpan(ctx)
		defer func() { trace.TelemetrySpanEnd(span, err) }()
		ctx = newCtx
	}

	var ret []*entity.DagInstanceGroup
	err = s.genericGroup(ctx, &ret, s.dagInsClsName, input.BuildQuery())
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (s *Store) genericGroup(ctx context.Context, ret interface{}, clsName string, pipeline mongo.Pipeline) error {
	var err error
	if ctx != nonContext {
		newCtx, span := trace.StartInternalSpan(ctx)
		msgStr, _ := jsoniter.MarshalToString(pipeline)
		trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, clsName), attribute.String(trace.DB_QUERY, msgStr))
		defer func() { trace.TelemetrySpanEnd(span, err) }()
	}

	ctx, cancel := context.WithTimeout(context.TODO(), s.opt.Timeout)
	defer cancel()

	cur, err := s.mongoDB.Collection(clsName).Aggregate(ctx, pipeline)
	if err != nil {
		return err
	}

	if err := cur.All(ctx, ret); err != nil {
		return fmt.Errorf("decode failed: %w", err)
	}

	return nil
}

func (s *Store) RetryDagIns(ctx context.Context, dagInsID string, taskInsIDs []string) error {
	var err error
	if ctx != nonContext {
		newCtx, span := trace.StartInternalSpan(ctx)
		defer func() { trace.TelemetrySpanEnd(span, err) }()
		trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, s.dagInsClsName), attribute.String(trace.DB_QUERY, dagInsID))
		ctx = newCtx
	}

	ctx, cancel := context.WithTimeout(ctx, s.opt.Timeout)
	defer cancel()

	session, err := s.mongoClient.StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		now := time.Now().Unix()
		fileds := bson.M{
			"updatedAt": now,
			"status":    entity.TaskInstanceStatusInit,
		}

		update := bson.M{
			"$set": fileds,
		}

		if _, err := s.mongoDB.Collection(s.taskInsClsName).UpdateMany(sessCtx, bson.M{"_id": bson.M{"$in": taskInsIDs}}, update); err != nil {
			return nil, fmt.Errorf("update task instance failed: %w", err)
		}

		fileds["status"] = entity.DagInstanceStatusInit
		fileds["endedAt"] = now

		update = bson.M{
			"$set": fileds,
		}

		if _, err := s.mongoDB.Collection(s.dagInsClsName).UpdateOne(sessCtx, bson.M{"_id": dagInsID}, update); err != nil {
			return nil, fmt.Errorf("update dag instance failed: %w", err)
		}

		return nil, nil
	})

	return err
}

// DeleteDag 删除Dag配置,仅在组合算子注册或配置权限失败时，删除dag配置时使用
func (s *Store) DeleteDag(ctx context.Context, ids ...string) error {
	var err error
	if ctx != nonContext {
		newCtx, span := trace.StartInternalSpan(ctx)
		defer func() { trace.TelemetrySpanEnd(span, err) }()
		trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, s.dagClsName), attribute.String(trace.DB_QUERY, strings.Join(ids, ",")))
		ctx = newCtx
	}

	if len(ids) == 0 {
		return nil
	}

	err = s.genericBatchDelete(ctx, ids, s.dagClsName)
	return err
}

// CreateDagVersion 创建DAG版本记录
func (s *Store) CreateDagVersion(ctx context.Context, dagVersion *entity.DagVersion) (string, error) {
	var err error
	if ctx != nonContext {
		newCtx, span := trace.StartInternalSpan(ctx)
		defer func() { trace.TelemetrySpanEnd(span, err) }()
		ctx = newCtx
	}

	id, err := s.genericCreate(ctx, dagVersion, s.dagVersionClsName)
	return id, err
}

// ListDagVersions 列举数据流历史版本
func (s *Store) ListDagVersions(ctx context.Context, input *mod.ListDagVersionInput) ([]entity.DagVersion, error) {
	var err error
	if ctx != nonContext {
		newCtx, span := trace.StartInternalSpan(ctx)
		defer func() { trace.TelemetrySpanEnd(span, err) }()
		ctx = newCtx
	}

	query := bson.M{}
	if input.DagID != "" {
		query["dagId"] = input.DagID
	}

	opt := &options.FindOptions{}
	if input.Limit > 0 {
		opt.Limit = &input.Limit
		offset := input.Limit * input.Offset
		opt.Skip = &offset
	}
	if input.SortBy != "" {
		opt.Sort = map[string]interface{}{input.SortBy: input.Order}
	}

	if len(input.SelectField) > 0 {
		fields := bson.M{}
		for _, f := range input.SelectField {
			fields[f] = 1
		}
		opt.Projection = fields
	}

	var ret []entity.DagVersion
	err = s.genericList(ctx, &ret, s.dagVersionClsName, query, opt)
	return ret, err
}

// GetHistoryDagByVersionID 根据版本ID获取历史DAG
func (s *Store) GetHistoryDagByVersionID(ctx context.Context, dagID, versionID string) (*entity.DagVersion, error) {
	var err error
	if ctx != nonContext {
		newCtx, span := trace.StartInternalSpan(ctx)
		defer func() { trace.TelemetrySpanEnd(span, err) }()
		ctx = newCtx
	}

	query := bson.M{
		"dagId":     dagID,
		"versionId": versionID,
	}

	res := &entity.DagVersion{}
	err = s.genericGetByFields(ctx, s.dagVersionClsName, query, res)
	return res, err
}
