package daconfvalobj

import "testing"

func TestBuiltInCanEditFields_ValObjCheck(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		p    *BuiltInCanEditFields
	}{
		{
			name: "全为true",
			p: &BuiltInCanEditFields{
				Name:                 true,
				Avatar:               true,
				Profile:              true,
				InputConfig:          true,
				SystemPrompt:         true,
				DataSourceKG:         true,
				DataSourceDoc:        true,
				Model:                true,
				Skills:               true,
				OpeningRemarkConfig:  true,
				PresetQuestions:      true,
				SkillsToolsToolInput: true,
				Memory:               true,
				RelatedQuestion:      true,
			},
		},
		{
			name: "全为false",
			p: &BuiltInCanEditFields{
				Name:                 false,
				Avatar:               false,
				Profile:              false,
				InputConfig:          false,
				SystemPrompt:         false,
				DataSourceKG:         false,
				DataSourceDoc:        false,
				Model:                false,
				Skills:               false,
				OpeningRemarkConfig:  false,
				PresetQuestions:      false,
				SkillsToolsToolInput: false,
				Memory:               false,
				RelatedQuestion:      false,
			},
		},
		{
			name: "混合值",
			p: &BuiltInCanEditFields{
				Name:         true,
				Avatar:       false,
				Profile:      true,
				InputConfig:  false,
				SystemPrompt: true,
			},
		},
		{
			name: "空结构体",
			p:    &BuiltInCanEditFields{},
		},
		{
			name: "nil指针",
			p:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.p == nil {
				t.Skip("nil指针不需要测试")
			}

			err := tt.p.ValObjCheck()
			if err != nil {
				t.Errorf("ValObjCheck() error = %v, want nil", err)
			}
		})
	}
}
