package cpmsresp

// AgentPermission Agent权限
type AgentPermission struct {
	Publish                  bool `json:"publish"`                       // 是否有发布权限
	Unpublish                bool `json:"unpublish"`                     // 是否有取消发布权限
	UnpublishOtherUserAgent  bool `json:"unpublish_other_user_agent"`    // 是否有取消其他用户发布的agent的权限
	PublishToBeSkillAgent    bool `json:"publish_to_be_skill_agent"`     // 是否有发布为技能agent的权限
	PublishToBeWebSdkAgent   bool `json:"publish_to_be_web_sdk_agent"`   // 是否有发布为web sdk agent的权限
	PublishToBeApiAgent      bool `json:"publish_to_be_api_agent"`       // 是否有发布为api agent的权限
	PublishToBeDataFlowAgent bool `json:"publish_to_be_data_flow_agent"` // 是否有发布为数据流agent的权限
	CreateSystemAgent        bool `json:"create_system_agent"`           // 是否有创建系统agent的权限
	MgntBuiltInAgent         bool `json:"mgnt_built_in_agent"`           // 是否有管理内置agent的权限
	SeeTrajectoryAnalysis    bool `json:"see_trajectory_analysis"`       // 是否有查看轨迹分析的权限
}

// AgentTplPermission Agent模板权限
type AgentTplPermission struct {
	Publish                    bool `json:"publish"`                        // 是否有发布权限
	Unpublish                  bool `json:"unpublish"`                      // 是否有取消发布权限
	UnpublishOtherUserAgentTpl bool `json:"unpublish_other_user_agent_tpl"` // 是否有取消其他用户发布的agent模板的权限
}

// UserStatusResp 用户权限状态响应
type UserStatusResp struct {
	// CustomSpace CustomSpacePermission `json:"custom_space"` // 自定义空间权限
	Agent    AgentPermission    `json:"agent"`     // Agent权限
	AgentTpl AgentTplPermission `json:"agent_tpl"` // Agent模板权限
}

func NewUserStatusResp() *UserStatusResp {
	return &UserStatusResp{
		// CustomSpace: CustomSpacePermission{},
		Agent:    AgentPermission{},
		AgentTpl: AgentTplPermission{},
	}
}

func NewUserStatusRespAllAllowed() *UserStatusResp {
	return &UserStatusResp{
		//CustomSpace: CustomSpacePermission{
		//    Create: true,
		//},
		Agent: AgentPermission{
			Publish:                  true,
			Unpublish:                true,
			UnpublishOtherUserAgent:  true,
			PublishToBeSkillAgent:    true,
			PublishToBeWebSdkAgent:   true,
			PublishToBeApiAgent:      true,
			PublishToBeDataFlowAgent: true,
			CreateSystemAgent:        true,
			MgntBuiltInAgent:         true,
			SeeTrajectoryAnalysis:    true,
		},
		AgentTpl: AgentTplPermission{
			Publish:                    true,
			Unpublish:                  true,
			UnpublishOtherUserAgentTpl: true,
		},
	}
}
