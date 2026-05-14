package skill

import (
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/config"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	logicsskill "github.com/kweaver-ai/adp/execution-factory/operator-integration/server/logics/skill"
)

type SkillHandler interface {
	RegisterSkill(c *gin.Context)
	CreateSkillIndexBuildTask(c *gin.Context)
	QuerySkillIndexBuildTaskList(c *gin.Context)
	GetSkillIndexBuildTask(c *gin.Context)
	CancelSkillIndexBuildTask(c *gin.Context)
	RetrySkillIndexBuildTask(c *gin.Context)
	UpdateSkillMetadata(c *gin.Context)
	UpdateSkillPackage(c *gin.Context)
	RepublishSkillHistory(c *gin.Context)
	PublishSkillHistory(c *gin.Context)
	DeleteSkill(c *gin.Context)
	UpdateSkillStatus(c *gin.Context)
	DownloadSkill(c *gin.Context)
	QuerySkillList(c *gin.Context)
	QuerySkillMarketList(c *gin.Context)
	GetSkillMarketDetail(c *gin.Context)
	GetSkillDetail(c *gin.Context)
	GetSkillContent(c *gin.Context)
	GetSkillReleaseHistory(c *gin.Context)
	ReadSkillFile(c *gin.Context)
	ExecuteSkill(c *gin.Context)
	// 管理态读接口
	GetManagementContent(c *gin.Context)
	ReadManagementFile(c *gin.Context)
	DownloadManagementSkill(c *gin.Context)
}

type skillHandler struct {
	Logger            interfaces.Logger
	Registry          interfaces.SkillRegistry
	Market            interfaces.SkillMarket
	Reader            interfaces.SkillReader
	MgmtReader        interfaces.SkillManagementReader
	IndexBuildService interfaces.SkillIndexBuildService
}

var (
	once sync.Once
	h    SkillHandler
)

func NewSkillHandler() SkillHandler {
	once.Do(func() {
		conf := config.NewConfigLoader()
		registry := logicsskill.NewSkillRegistry()
		market, _ := registry.(interfaces.SkillMarket)
		h = &skillHandler{
			Logger:            conf.GetLogger(),
			Registry:          registry,
			Market:            market,
			Reader:            logicsskill.NewSkillReader(),
			MgmtReader:        logicsskill.NewSkillManagementReader(),
			IndexBuildService: logicsskill.NewSkillIndexBuildService(),
		}
	})
	return h
}
