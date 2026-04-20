package skill

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
	infracommon "github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/common"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/config"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
)

const (
	skillIndexBuildTaskType  = "execution-factory:skill:index:build"
	skillIndexBuildQueueName = "skill_index_build"
)

type skillIndexBuildTaskPayload struct {
	TaskID           string                  `json:"task_id"`
	BusinessDomainID string                  `json:"business_domain_id"`
	AccountID        string                  `json:"account_id"`
	AccountType      interfaces.AccessorType `json:"account_type"`
}

type skillIndexBuildTaskEnqueuer interface {
	Enqueue(ctx context.Context, payload *skillIndexBuildTaskPayload) error
	Close() error
}

type skillIndexBuildInspector interface {
	GetTaskInfo(queue, id string) (*asynq.TaskInfo, error)
	DeleteTask(queue, id string) error
	CancelProcessing(id string) error
	Close() error
}

type skillIndexBuildAsynqClient struct {
	client *asynq.Client
}

type skillIndexBuildAsynqInspector struct {
	inspector *asynq.Inspector
}

func newSkillIndexBuildTaskEnqueuer() (skillIndexBuildTaskEnqueuer, error) {
	redisOpt, err := buildSkillIndexBuildRedisConnOpt()
	if err != nil {
		return nil, err
	}
	return &skillIndexBuildAsynqClient{
		client: asynq.NewClient(redisOpt),
	}, nil
}

func newSkillIndexBuildInspector() (skillIndexBuildInspector, error) {
	redisOpt, err := buildSkillIndexBuildRedisConnOpt()
	if err != nil {
		return nil, err
	}
	return &skillIndexBuildAsynqInspector{
		inspector: asynq.NewInspector(redisOpt),
	}, nil
}

func (c *skillIndexBuildAsynqClient) Enqueue(ctx context.Context, payload *skillIndexBuildTaskPayload) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	task := asynq.NewTask(
		skillIndexBuildTaskType,
		body,
		asynq.TaskID(payload.TaskID),
		asynq.MaxRetry(10),
		asynq.Queue(skillIndexBuildQueueName),
		asynq.Timeout(30*time.Minute),
	)
	_, err = c.client.EnqueueContext(ctx, task)
	return err
}

func (c *skillIndexBuildAsynqClient) Close() error {
	return c.client.Close()
}

func (i *skillIndexBuildAsynqInspector) GetTaskInfo(queue, id string) (*asynq.TaskInfo, error) {
	return i.inspector.GetTaskInfo(queue, id)
}

func (i *skillIndexBuildAsynqInspector) DeleteTask(queue, id string) error {
	return i.inspector.DeleteTask(queue, id)
}

func (i *skillIndexBuildAsynqInspector) CancelProcessing(id string) error {
	return i.inspector.CancelProcessing(id)
}

func (i *skillIndexBuildAsynqInspector) Close() error {
	return i.inspector.Close()
}

func buildSkillIndexBuildRedisConnOpt() (asynq.RedisConnOpt, error) {
	conf := config.NewConfigLoader()
	tlsConf, err := conf.RedisConfig.GetTLSConfig()
	if err != nil {
		return nil, err
	}
	switch conf.RedisConfig.ConnectType {
	case config.RedisTypeSentinel:
		return asynq.RedisFailoverClientOpt{
			MasterName:       conf.RedisConfig.ConnectInfo.MasterGroupName,
			SentinelAddrs:    []string{fmt.Sprintf("%s:%d", conf.RedisConfig.ConnectInfo.SentinelHost, conf.RedisConfig.ConnectInfo.SentinelPort)},
			Username:         conf.RedisConfig.ConnectInfo.Username,
			Password:         conf.RedisConfig.ConnectInfo.Password,
			SentinelUsername: conf.RedisConfig.ConnectInfo.SentinelUsername,
			SentinelPassword: conf.RedisConfig.ConnectInfo.SentinelPassword,
			TLSConfig:        tlsConf,
			PoolSize:         conf.RedisConfig.ConnectInfo.PoolSize,
		}, nil
	case config.RedisTypeMasterSlave:
		return asynq.RedisClientOpt{
			Addr:      fmt.Sprintf("%s:%d", conf.RedisConfig.ConnectInfo.MasterlHost, conf.RedisConfig.ConnectInfo.MasterlPort),
			Username:  conf.RedisConfig.ConnectInfo.Username,
			Password:  conf.RedisConfig.ConnectInfo.Password,
			TLSConfig: tlsConf,
			PoolSize:  conf.RedisConfig.ConnectInfo.PoolSize,
		}, nil
	case config.RedisTypeStandalone:
		return asynq.RedisClientOpt{
			Addr:      fmt.Sprintf("%s:%d", conf.RedisConfig.ConnectInfo.Host, conf.RedisConfig.ConnectInfo.Port),
			Username:  conf.RedisConfig.ConnectInfo.Username,
			Password:  conf.RedisConfig.ConnectInfo.Password,
			TLSConfig: tlsConf,
			PoolSize:  conf.RedisConfig.ConnectInfo.PoolSize,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported redis connect type: %s", conf.RedisConfig.ConnectType)
	}
}

func newSkillIndexBuildTaskPayload(ctx context.Context, taskID string) *skillIndexBuildTaskPayload {
	payload := &skillIndexBuildTaskPayload{
		TaskID: taskID,
	}
	if businessDomain, ok := infracommon.GetBusinessDomainFromCtx(ctx); ok {
		payload.BusinessDomainID = businessDomain
	}
	if authCtx, ok := infracommon.GetAccountAuthContextFromCtx(ctx); ok && authCtx != nil {
		payload.AccountID = authCtx.AccountID
		payload.AccountType = authCtx.AccountType
	}
	return payload
}
