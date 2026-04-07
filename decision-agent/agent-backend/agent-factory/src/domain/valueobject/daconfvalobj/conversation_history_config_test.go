package daconfvalobj

import (
	"strconv"
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/constant"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/stretchr/testify/assert"
)

func TestConversationHistoryConfig_ValObjCheck(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		config  *ConversationHistoryConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid count strategy with count_limit",
			config: &ConversationHistoryConfig{
				Strategy: cdaenum.HistoryStrategyCount,
				CountParams: &CountParams{
					CountLimit: 10,
				},
			},
			wantErr: false,
		},
		{
			name: "valid count strategy with count_limit=1 (min)",
			config: &ConversationHistoryConfig{
				Strategy: cdaenum.HistoryStrategyCount,
				CountParams: &CountParams{
					CountLimit: 1,
				},
			},
			wantErr: false,
		},
		{
			name: "valid count strategy with count_limit=1000 (max)",
			config: &ConversationHistoryConfig{
				Strategy: cdaenum.HistoryStrategyCount,
				CountParams: &CountParams{
					CountLimit: 1000,
				},
			},
			wantErr: false,
		},
		{
			name: "count strategy with count_limit=0 is invalid",
			config: &ConversationHistoryConfig{
				Strategy: cdaenum.HistoryStrategyCount,
				CountParams: &CountParams{
					CountLimit: 0,
				},
			},
			wantErr: true,
			errMsg:  "count_limit must be between 1 and " + strconv.Itoa(constant.MaxHistoryLimit),
		},
		{
			name: "count strategy with count_limit > 1000 is invalid",
			config: &ConversationHistoryConfig{
				Strategy: cdaenum.HistoryStrategyCount,
				CountParams: &CountParams{
					CountLimit: 1001,
				},
			},
			wantErr: true,
			errMsg:  "count_limit must be between 1 and " + strconv.Itoa(constant.MaxHistoryLimit),
		},
		{
			name: "count strategy with negative count_limit is invalid",
			config: &ConversationHistoryConfig{
				Strategy: cdaenum.HistoryStrategyCount,
				CountParams: &CountParams{
					CountLimit: -1,
				},
			},
			wantErr: true,
			errMsg:  "count_limit must be between 1 and " + strconv.Itoa(constant.MaxHistoryLimit),
		},
		{
			name: "count strategy with nil CountParams is valid",
			config: &ConversationHistoryConfig{
				Strategy:    cdaenum.HistoryStrategyCount,
				CountParams: nil,
			},
			wantErr: false,
		},
		{
			name: "none strategy is valid",
			config: &ConversationHistoryConfig{
				Strategy: cdaenum.HistoryStrategyNone,
			},
			wantErr: false,
		},
		{
			name: "time_window strategy with nil TimeWindowParams is invalid",
			config: &ConversationHistoryConfig{
				Strategy:         cdaenum.HistoryStrategyTimeWindow,
				TimeWindowParams: nil,
			},
			wantErr: true,
			errMsg:  "time_window_params is required",
		},
		{
			name: "time_window strategy with empty TimeWindowParams is valid",
			config: &ConversationHistoryConfig{
				Strategy:         cdaenum.HistoryStrategyTimeWindow,
				TimeWindowParams: &TimeWindowParams{},
			},
			wantErr: false,
		},
		{
			name: "token strategy with nil TokenLimitParams is invalid",
			config: &ConversationHistoryConfig{
				Strategy:         cdaenum.HistoryStrategyToken,
				TokenLimitParams: nil,
			},
			wantErr: true,
			errMsg:  "token_limit_params is required",
		},
		{
			name: "token strategy with empty TokenLimitParams is valid",
			config: &ConversationHistoryConfig{
				Strategy:         cdaenum.HistoryStrategyToken,
				TokenLimitParams: &TokenLimitParams{},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.config.ValObjCheck()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCountParams_DefaultValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		countLimit  int
		expected    int
		shouldPanic bool
	}{
		{
			name:        "count_limit within valid range",
			countLimit:  10,
			expected:    10,
			shouldPanic: false,
		},
		{
			name:        "count_limit at minimum",
			countLimit:  1,
			expected:    1,
			shouldPanic: false,
		},
		{
			name:        "count_limit at maximum",
			countLimit:  constant.MaxHistoryLimit,
			expected:    constant.MaxHistoryLimit,
			shouldPanic: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			config := &ConversationHistoryConfig{
				Strategy: cdaenum.HistoryStrategyCount,
				CountParams: &CountParams{
					CountLimit: tt.countLimit,
				},
			}

			err := config.ValObjCheck()
			if tt.shouldPanic {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, config.CountParams.CountLimit)
			}
		})
	}
}
