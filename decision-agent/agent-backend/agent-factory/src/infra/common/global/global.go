package global

import (
	"sync"

	"github.com/kweaver-ai/decision-agent/agent-factory/conf"

	"github.com/kweaver-ai/proton-rds-sdk-go/sqlx"
)

var (
	GConfig *conf.Config // 全局配置
	GDB     *sqlx.DB     // 全局 DB

	loggerOnce sync.Once
)


