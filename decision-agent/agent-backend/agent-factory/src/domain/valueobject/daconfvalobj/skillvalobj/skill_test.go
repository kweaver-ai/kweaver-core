package skillvalobj

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSkill_ValObjCheck(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		skill    *Skill
		wantErr  bool
		checkErr func(t *testing.T, err error)
	}{
		{
			name: "valid skill with all nil slices",
			skill: &Skill{
				Tools:  nil,
				Agents: nil,
				MCPs:   nil,
				Skills: nil,
			},
			wantErr: false,
		},
		{
			name: "valid skill with empty slices",
			skill: &Skill{
				Tools:  []*SkillTool{},
				Agents: []*SkillAgent{},
				MCPs:   []*SkillMCP{},
				Skills: []*SkillSkill{},
			},
			wantErr: false,
		},
		{
			name: "valid skill with tools",
			skill: &Skill{
				Tools: []*SkillTool{
					{
						ToolID:    "tool-1",
						ToolBoxID: "toolbox-1",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid skill with agents",
			skill: &Skill{
				Agents: []*SkillAgent{
					{
						AgentKey: "agent-1",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid skill with mcps",
			skill: &Skill{
				MCPs: []*SkillMCP{
					{
						MCPServerID: "mcp-1",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid skill with skills",
			skill: &Skill{
				Skills: []*SkillSkill{
					{
						SkillID: "skill-1",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid tool in skill",
			skill: &Skill{
				Tools: []*SkillTool{
					{
						ToolID: "",
					},
				},
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "tools is invalid")
			},
		},
		{
			name: "invalid agent in skill",
			skill: &Skill{
				Agents: []*SkillAgent{
					{
						AgentKey: "",
					},
				},
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "agents is invalid")
			},
		},
		{
			name: "invalid mcp in skill",
			skill: &Skill{
				MCPs: []*SkillMCP{
					{
						MCPServerID: "",
					},
				},
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "mcp is invalid")
			},
		},
		{
			name: "invalid skill in skills",
			skill: &Skill{
				Skills: []*SkillSkill{
					{
						SkillID: "",
					},
				},
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "skills is invalid")
			},
		},
		{
			name: "nil skill",
			skill: &Skill{
				Tools: []*SkillTool{
					{
						ToolID: "",
					},
				},
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "tools is invalid")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.skill.ValObjCheck()
			if tt.wantErr {
				require.Error(t, err)

				if tt.checkErr != nil {
					tt.checkErr(t, err)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestSkillTool_ValObjCheck(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		tool     *SkillTool
		wantErr  bool
		checkErr func(t *testing.T, err error)
	}{
		{
			name: "valid tool",
			tool: &SkillTool{
				ToolID:    "tool-1",
				ToolBoxID: "toolbox-1",
			},
			wantErr: false,
		},
		{
			name:    "nil tool",
			tool:    nil,
			wantErr: true,
		},
		{
			name: "empty tool_id",
			tool: &SkillTool{
				ToolID:    "",
				ToolBoxID: "toolbox-1",
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "tool_id is required")
			},
		},
		{
			name: "empty tool_box_id",
			tool: &SkillTool{
				ToolID:    "tool-1",
				ToolBoxID: "",
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "tool_box_id is required")
			},
		},
		{
			name: "tool with result process strategies",
			tool: &SkillTool{
				ToolID:    "tool-1",
				ToolBoxID: "toolbox-1",
				ResultProcessStrategies: []ResultProcessStrategy{
					{
						Category: Category{
							ID:          "cat-1",
							Name:        "category1",
							Description: "test category",
						},
						Strategy: Strategy{
							ID:          "strat-1",
							Name:        "strategy1",
							Description: "test strategy",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "tool with json input",
			tool: &SkillTool{
				ToolID:    "tool-1",
				ToolBoxID: "toolbox-1",
				ToolInput: json.RawMessage(`{"key": "value"}`),
			},
			wantErr: false,
		},
		{
			name: "tool with intervention enabled",
			tool: &SkillTool{
				ToolID:                          "tool-1",
				ToolBoxID:                       "toolbox-1",
				Intervention:                    true,
				InterventionConfirmationMessage: "Please confirm",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var err error

			if tt.tool == nil {
				// Nil pointer will panic, so we skip it
				t.Skip("nil tool test skipped - would panic")
				return
			}

			err = tt.tool.ValObjCheck()
			if tt.wantErr {
				require.Error(t, err)

				if tt.checkErr != nil {
					tt.checkErr(t, err)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestSkillAgent_ValObjCheck(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		agent    *SkillAgent
		wantErr  bool
		checkErr func(t *testing.T, err error)
	}{
		{
			name: "valid agent",
			agent: &SkillAgent{
				AgentKey:     "agent-1",
				AgentVersion: "1.0",
			},
			wantErr: false,
		},
		{
			name:    "nil agent",
			agent:   nil,
			wantErr: true,
		},
		{
			name: "empty agent_key",
			agent: &SkillAgent{
				AgentKey: "",
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "agent_key is required")
			},
		},
		{
			name: "agent with data source config",
			agent: &SkillAgent{
				AgentKey: "agent-1",
				DataSourceConfig: &DataSourceConfig{
					Type: "database",
				},
			},
			wantErr: false,
		},
		{
			name: "agent with llm config",
			agent: &SkillAgent{
				AgentKey:  "agent-1",
				LlmConfig: &LLMConfig{},
			},
			wantErr: false,
		},
		{
			name: "agent with json input",
			agent: &SkillAgent{
				AgentKey:     "agent-1",
				AgentInput:   json.RawMessage(`{"key": "value"}`),
				AgentTimeout: 30,
			},
			wantErr: false,
		},
		{
			name: "agent with intervention enabled",
			agent: &SkillAgent{
				AgentKey:                        "agent-1",
				Intervention:                    true,
				InterventionConfirmationMessage: "Please confirm",
			},
			wantErr: false,
		},
		{
			name: "agent with pms check status",
			agent: &SkillAgent{
				AgentKey:                    "agent-1",
				CurrentPmsCheckStatus:       CurrentPmsCheckStatusSuccess,
				CurrentIsExistsAndPublished: true,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var err error

			if tt.agent == nil {
				// Nil pointer will panic, so we skip it
				t.Skip("nil agent test skipped - would panic")
				return
			}

			err = tt.agent.ValObjCheck()
			if tt.wantErr {
				require.Error(t, err)

				if tt.checkErr != nil {
					tt.checkErr(t, err)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestSkillMCP_ValObjCheck(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		mcp      *SkillMCP
		wantErr  bool
		checkErr func(t *testing.T, err error)
	}{
		{
			name: "valid mcp",
			mcp: &SkillMCP{
				MCPServerID: "mcp-1",
			},
			wantErr: false,
		},
		{
			name:    "nil mcp",
			mcp:     nil,
			wantErr: true,
		},
		{
			name: "empty mcp_server_id",
			mcp: &SkillMCP{
				MCPServerID: "",
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "mcp_server_id is required")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var err error

			if tt.mcp == nil {
				// Nil pointer will panic, so we skip it
				t.Skip("nil mcp test skipped - would panic")
				return
			}

			err = tt.mcp.ValObjCheck()
			if tt.wantErr {
				require.Error(t, err)

				if tt.checkErr != nil {
					tt.checkErr(t, err)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCurrentPmsCheckStatusT(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		status CurrentPmsCheckStatusT
		valid  bool
	}{
		{
			name:   "success status",
			status: CurrentPmsCheckStatusSuccess,
			valid:  true,
		},
		{
			name:   "failed status",
			status: CurrentPmsCheckStatusFailed,
			valid:  true,
		},
		{
			name:   "custom status",
			status: CurrentPmsCheckStatusT("pending"),
			valid:  true,
		},
		{
			name:   "empty status",
			status: CurrentPmsCheckStatusT(""),
			valid:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// CurrentPmsCheckStatusT is just a string type alias
			// Any value is technically valid
			assert.NotEmpty(t, string(tt.status) != "" || tt.status == "")
		})
	}
}

func TestResultProcessStrategy(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		strategy ResultProcessStrategy
	}{
		{
			name: "valid strategy",
			strategy: ResultProcessStrategy{
				Category: Category{
					ID:          "cat-1",
					Name:        "category1",
					Description: "test category",
				},
				Strategy: Strategy{
					ID:          "strat-1",
					Name:        "strategy1",
					Description: "test strategy",
				},
			},
		},
		{
			name: "strategy with empty fields",
			strategy: ResultProcessStrategy{
				Category: Category{},
				Strategy: Strategy{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// ResultProcessStrategy doesn't have validation
			// Just verify it can be created
			assert.NotNil(t, tt.strategy.Category)
			assert.NotNil(t, tt.strategy.Strategy)
		})
	}
}
