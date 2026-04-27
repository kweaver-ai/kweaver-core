package v3agentconfigsvc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOpeningRemarksSystemPrompt(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		language     string
		wantContains []string
	}{
		{
			name:     "English language",
			language: "en-US",
			wantContains: []string{
				"Generate an opening statement",
				"Language requirements",
			},
		},
		{
			name:     "Traditional Chinese",
			language: "zh-TW",
			wantContains: []string{
				"根據用戶輸入的內容生成開場白",
				"生成開場白的語言要求",
			},
		},
		{
			name:     "Simplified Chinese (default)",
			language: "zh-CN",
			wantContains: []string{
				"根据用户输入的内容生成开场白",
				"生成开场白的语言要求",
			},
		},
		{
			name:     "Unknown language defaults to zh-CN",
			language: "fr",
			wantContains: []string{
				"根据用户输入的内容生成开场白",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := openingRemarksSystemPrompt(tt.language)
			assert.NotEmpty(t, result)

			for _, substr := range tt.wantContains {
				assert.Contains(t, result, substr)
			}
		})
	}
}

func TestSystemPrompt(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		language     string
		wantContains []string
	}{
		{
			name:     "English language",
			language: "en-US",
			wantContains: []string{
				`Generate a "Personality and Response Logic"`,
				"Language requirements",
			},
		},
		{
			name:     "Traditional Chinese",
			language: "zh-TW",
			wantContains: []string{
				"根據用戶輸入的內容生成",
			},
		},
		{
			name:     "Simplified Chinese (default)",
			language: "zh-CN",
			wantContains: []string{
				"你是一个提示词工程师",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := systemPrompt(tt.language)
			assert.NotEmpty(t, result)

			for _, substr := range tt.wantContains {
				assert.Contains(t, result, substr)
			}
		})
	}
}

func TestUserPromptForOpenRemarks(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		language     string
		agentName    string
		agentProfile string
		agentSkills  []string
		agentSources []string
		wantContains []string
	}{
		{
			name:         "English language",
			language:     "en-US",
			agentName:    "Test Agent",
			agentProfile: "A test assistant",
			agentSkills:  []string{"skill1", "skill2"},
			agentSources: []string{"source1"},
			wantContains: []string{
				"Name: Test Agent",
				"Please generate an opening statement",
			},
		},
		{
			name:         "Simplified Chinese",
			language:     "zh-CN",
			agentName:    "测试助手",
			agentProfile: "测试简介",
			agentSkills:  []string{"技能1"},
			agentSources: []string{"来源1"},
			wantContains: []string{
				"名称：测试助手",
				"请根据上面提供的信息生成一个开场白",
			},
		},
		{
			name:         "Traditional Chinese",
			language:     "zh-TW",
			agentName:    "測試助手",
			agentProfile: "測試簡介",
			agentSkills:  []string{"技能1"},
			agentSources: []string{"來源1"},
			wantContains: []string{
				"名稱：測試助手",
				"請根據上面提供的信息生成一個開場白",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := userPromptForOpenRemarks(tt.language, tt.agentName, tt.agentProfile, tt.agentSkills, tt.agentSources)
			assert.NotEmpty(t, result)

			for _, substr := range tt.wantContains {
				assert.Contains(t, result, substr)
			}
		})
	}
}

func TestUserPromptForPresetQuestion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		language     string
		agentName    string
		agentProfile string
		wantContains []string
	}{
		{
			name:         "English language",
			language:     "en-US",
			agentName:    "Test Agent",
			agentProfile: "A test assistant",
			wantContains: []string{
				"Please generate 3 preset questions",
				"Each question should be a text of no more than 30 characters",
			},
		},
		{
			name:         "Simplified Chinese",
			language:     "zh-CN",
			agentName:    "测试助手",
			agentProfile: "测试简介",
			wantContains: []string{
				"请根据上面提供的信息生成3个预设问题",
				"每一个问题是一个不超过30字的文本",
			},
		},
		{
			name:         "Traditional Chinese",
			language:     "zh-TW",
			agentName:    "測試助手",
			agentProfile: "測試簡介",
			wantContains: []string{
				"請根據上面提供的信息生成3個預設問題",
				"每一個問題是一個不超過30個字的文本",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := userPromptForPresetQuestion(tt.language, tt.agentName, tt.agentProfile, []string{}, []string{})
			assert.NotEmpty(t, result)

			for _, substr := range tt.wantContains {
				assert.Contains(t, result, substr)
			}
		})
	}
}

func TestUserPromptForSystem(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		language     string
		agentName    string
		agentProfile string
		wantContains []string
	}{
		{
			name:         "English language",
			language:     "en-US",
			agentName:    "Test Agent",
			agentProfile: "A test assistant",
			wantContains: []string{
				"Please generate a personality and instruction",
				"Use Markdown syntax or plain text to output",
			},
		},
		{
			name:         "Simplified Chinese",
			language:     "zh-CN",
			agentName:    "测试助手",
			agentProfile: "测试简介",
			wantContains: []string{
				"请根据上面提供的信息生成一个人设和指令",
				"使用Markdown语法或普通文本来输出",
			},
		},
		{
			name:         "Traditional Chinese",
			language:     "zh-TW",
			agentName:    "測試助手",
			agentProfile: "測試簡介",
			wantContains: []string{
				"請根據上面提供的信息生成一個人設和指令",
				"使用Markdown語法或普通文本来輸出",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := userPromptForSystem(tt.language, tt.agentName, tt.agentProfile, []string{}, []string{})
			assert.NotEmpty(t, result)

			for _, substr := range tt.wantContains {
				assert.Contains(t, result, substr)
			}
		})
	}
}
