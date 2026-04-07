package rds

import (
	"context"
	"sync"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/db"
	traceLog "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/log"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/trace"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/rds"
	"go.opentelemetry.io/otel/attribute"
	"gorm.io/gorm"
)

type AgentDaoImpl struct {
	inner *gorm.DB
}

var (
	agent     rds.AgentDao
	agentOnce sync.Once
)

func NewAgent() rds.AgentDao {
	agentOnce.Do(func() {
		agent = &AgentDaoImpl{
			inner: db.NewDB(),
		}
	})

	return agent
}

func (d *AgentDaoImpl) GetAgents(ctx context.Context) (agents []*rds.AgentModel, err error) {
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(newCtx)
	sql := "select * from t_automation_agent"
	trace.SetAttributes(newCtx, attribute.String(trace.DB_SQL, sql))
	err = d.inner.Raw(sql).Scan(&agents).Error
	if err != nil {
		log.Warnf("[AgentDaoImpl.GetAgents] get failed: %s", err.Error())
	}
	return
}

func (d *AgentDaoImpl) GetAgentByName(ctx context.Context, name string) (agent *rds.AgentModel, err error) {
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(newCtx)

	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, rds.AGENT_TABLENAME))
	sql := "select * from t_automation_agent where f_name = ?"
	trace.SetAttributes(newCtx, attribute.String(trace.DB_SQL, sql))

	agent = &rds.AgentModel{}
	err = d.inner.Raw(sql, name).Scan(agent).Error

	if err != nil {
		log.Warnf("[AgentDaoImpl.GetAgentByName] get failed: %s", err.Error())
		return nil, err
	}

	return agent, nil
}

func (d *AgentDaoImpl) DeleteAgentByName(ctx context.Context, name string) (err error) {
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(newCtx)

	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, rds.AGENT_TABLENAME))
	sql := "delete from t_automation_agent where f_name = ?"
	trace.SetAttributes(newCtx, attribute.String(trace.DB_SQL, sql))

	result := d.inner.Exec(sql, name)
	if result.Error != nil {
		log.Warnf("[AgentDaoImpl.DeleteAgentByName] delete failed: %s", result.Error.Error())
		return result.Error
	}

	return nil
}

func (d *AgentDaoImpl) CreateAgent(ctx context.Context, agent *rds.AgentModel) (err error) {
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(newCtx)

	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, rds.AGENT_TABLENAME))
	sql := "insert into t_automation_agent (f_id, f_name, f_agent_id, f_version) values (?, ?, ?, ?)"
	trace.SetAttributes(newCtx, attribute.String(trace.DB_SQL, sql))

	result := d.inner.Exec(sql, agent.ID, agent.Name, agent.AgentID, agent.Version)
	if result.Error != nil {
		log.Warnf("[AgentDaoImpl.CreateAgent] insert failed: %s", result.Error.Error())
		return result.Error
	}

	return nil
}

func (d *AgentDaoImpl) UpdateAgent(ctx context.Context, agent *rds.AgentModel) (err error) {
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(newCtx)

	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, rds.AGENT_TABLENAME))
	sql := "update t_automation_agent set f_agent_id = ?, f_version = ? where f_name = ?"
	trace.SetAttributes(newCtx, attribute.String(trace.DB_SQL, sql))

	result := d.inner.Exec(sql, agent.AgentID, agent.Version, agent.Name)
	if result.Error != nil {
		log.Warnf("[AgentDaoImpl.UpdateAgent] update failed: %s", result.Error.Error())
		return result.Error
	}

	return nil
}
