package cdaenum

import "testing"

func TestAgentMode_EnumCheck(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		mode    AgentMode
		wantErr bool
	}{
		{name: "default", mode: AgentModeDefault, wantErr: false},
		{name: "dolphin", mode: AgentModeDolphin, wantErr: false},
		{name: "react", mode: AgentModeReact, wantErr: false},
		{name: "invalid", mode: AgentMode("invalid"), wantErr: true},
		{name: "empty", mode: AgentMode(""), wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.mode.EnumCheck()
			if (err != nil) != tt.wantErr {
				t.Fatalf("EnumCheck() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
