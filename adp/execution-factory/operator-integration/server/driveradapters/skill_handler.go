package driveradapters

import (
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/driveradapters/skill"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/logics/business_domain"
)

type SkillRestHandler interface {
	// RegisterPrivate 注册内部API
	RegisterPrivate(engine *gin.RouterGroup)
	// RegisterPublic 注册公开API
	RegisterPublic(engine *gin.RouterGroup)
}

type skillRestHandler struct {
	SkillHandler          skill.SkillHandler
	businessDomainService interfaces.IBusinessDomainService
}

var (
	sOnce    sync.Once
	sHandler SkillRestHandler
)

func NewSkillRestHandler() SkillRestHandler {
	sOnce.Do(func() {
		sHandler = &skillRestHandler{
			SkillHandler:          skill.NewSkillHandler(),
			businessDomainService: business_domain.NewBusinessDomainService(),
		}
	})
	return sHandler
}
func (r *skillRestHandler) RegisterPrivate(engine *gin.RouterGroup) {
	engine.Use(middlewareBusinessDomain(false, false, r.businessDomainService))
	/*市场接口*/
	// 查询技能市场列表
	engine.GET("/skills/market", r.SkillHandler.QuerySkillMarketList)
	// 查询技能市场详情
	engine.GET("/skills/market/:skill_id", r.SkillHandler.GetSkillMarketDetail)
	/*读取接口*/
	// 查询技能内容
	engine.GET("/skills/:skill_id/content", r.SkillHandler.GetSkillContent)
	// 读取技能文件
	engine.POST("/skills/:skill_id/files/read", r.SkillHandler.ReadSkillFile)
}

func (r *skillRestHandler) RegisterPublic(engine *gin.RouterGroup) {
	engine.Use(middlewareBusinessDomain(true, false, r.businessDomainService))
	/*管理接口*/
	// 注册技能
	engine.POST("/skills", r.SkillHandler.RegisterSkill)
	// 查询技能列表
	engine.GET("/skills", r.SkillHandler.QuerySkillList)
	// 查询技能详情
	engine.GET("/skills/:skill_id", r.SkillHandler.GetSkillDetail)
	// 下载技能
	engine.GET("/skills/:skill_id/download", r.SkillHandler.DownloadSkill)
	// 删除技能
	engine.DELETE("/skills/:skill_id", r.SkillHandler.DeleteSkill)
	// 更新状态
	engine.PUT("/skills/:skill_id/status", r.SkillHandler.UpdateSkillStatus)
	/*市场接口*/
	// 查询技能市场列表
	engine.GET("/skills/market", r.SkillHandler.QuerySkillMarketList)
	// 查询技能市场详情
	engine.GET("/skills/market/:skill_id", r.SkillHandler.GetSkillMarketDetail)
	/*读取接口*/
	// 查询技能内容
	engine.GET("/skills/:skill_id/content", r.SkillHandler.GetSkillContent)
	// 读取技能文件
	engine.POST("/skills/:skill_id/files/read", r.SkillHandler.ReadSkillFile)
}
