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
	// 执行技能
	engine.POST("/skills/:skill_id/execute", r.SkillHandler.ExecuteSkill)
	/*管理态读接口*/
	engine.GET("/skills/:skill_id/management/content", r.SkillHandler.GetManagementContent)
	engine.POST("/skills/:skill_id/management/files/read", r.SkillHandler.ReadManagementFile)
	engine.GET("/skills/:skill_id/management/download", r.SkillHandler.DownloadManagementSkill)
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
	// 更新元数据
	engine.PUT("/skills/:skill_id/metadata", r.SkillHandler.UpdateSkillMetadata)
	// 更新技能包
	engine.PUT("/skills/:skill_id/package", r.SkillHandler.UpdateSkillPackage)
	// 将历史版本回灌到草稿态
	engine.POST("/skills/:skill_id/history/republish", r.SkillHandler.RepublishSkillHistory)
	// 直接发布历史版本
	engine.POST("/skills/:skill_id/history/publish", r.SkillHandler.PublishSkillHistory)
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
	// 执行技能
	engine.POST("/skills/:skill_id/execute", r.SkillHandler.ExecuteSkill)
	// 查询技能发布历史
	engine.GET("/skills/:skill_id/history", r.SkillHandler.GetSkillReleaseHistory)
	/*管理态读接口*/
	engine.GET("/skills/:skill_id/management/content", r.SkillHandler.GetManagementContent)
	engine.POST("/skills/:skill_id/management/files/read", r.SkillHandler.ReadManagementFile)
	engine.GET("/skills/:skill_id/management/download", r.SkillHandler.DownloadManagementSkill)
	/*构建接口*/
	engine.POST("/skills/index/build", r.SkillHandler.CreateSkillIndexBuildTask)
	engine.GET("/skills/index/build", r.SkillHandler.QuerySkillIndexBuildTaskList)
	engine.GET("/skills/index/build/:task_id", r.SkillHandler.GetSkillIndexBuildTask)
	engine.POST("/skills/index/build/:task_id/cancel", r.SkillHandler.CancelSkillIndexBuildTask)
	engine.POST("/skills/index/build/:task_id/retry", r.SkillHandler.RetrySkillIndexBuildTask)
}
