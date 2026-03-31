package daconfvalobj

import (
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/daconfvalobj/datasourcevalobj"
	"github.com/stretchr/testify/assert"
)

func TestConfig_GetDolphinTplLength(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		config         *Config
		expectedLength int
	}{
		{
			name:           "empty config",
			config:         &Config{},
			expectedLength: 0,
		},
		{
			name: "only pre dolphin",
			config: &Config{
				PreDolphin: []*DolphinTpl{
					{Key: cdaenum.DolphinTplKeyDocRetrieve, Value: "value1"},
					{Key: cdaenum.DolphinTplKeyGraphRetrieve, Value: "value2"},
				},
			},
			expectedLength: 2,
		},
		{
			name: "only post dolphin",
			config: &Config{
				PostDolphin: []*DolphinTpl{
					{Key: cdaenum.DolphinTplKeyRelatedQuestions, Value: "value1"},
				},
			},
			expectedLength: 1,
		},
		{
			name: "both pre and post dolphin",
			config: &Config{
				PreDolphin: []*DolphinTpl{
					{Key: cdaenum.DolphinTplKeyDocRetrieve, Value: "value1"},
					{Key: cdaenum.DolphinTplKeyGraphRetrieve, Value: "value2"},
					{Key: cdaenum.DolphinTplKeyContextOrganize, Value: "value3"},
				},
				PostDolphin: []*DolphinTpl{
					{Key: cdaenum.DolphinTplKeyRelatedQuestions, Value: "value4"},
				},
			},
			expectedLength: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			length := tt.config.GetDolphinTplLength()
			assert.Equal(t, tt.expectedLength, length)
		})
	}
}

func TestConfig_IsOneDolphinTplDisabled(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		config   *Config
		key      cdaenum.DolphinTplKey
		expected bool
	}{
		{
			name: "template is enabled in pre dolphin",
			config: &Config{
				PreDolphin: []*DolphinTpl{
					{Key: cdaenum.DolphinTplKeyDocRetrieve, Value: "value", Enabled: true},
				},
			},
			key:      cdaenum.DolphinTplKeyDocRetrieve,
			expected: false,
		},
		{
			name: "template is disabled in pre dolphin",
			config: &Config{
				PreDolphin: []*DolphinTpl{
					{Key: cdaenum.DolphinTplKeyDocRetrieve, Value: "value", Enabled: false},
				},
			},
			key:      cdaenum.DolphinTplKeyDocRetrieve,
			expected: true,
		},
		{
			name: "template is disabled in post dolphin",
			config: &Config{
				PostDolphin: []*DolphinTpl{
					{Key: cdaenum.DolphinTplKeyRelatedQuestions, Value: "value", Enabled: false},
				},
			},
			key:      cdaenum.DolphinTplKeyRelatedQuestions,
			expected: true,
		},
		{
			name:     "template not found",
			config:   &Config{},
			key:      cdaenum.DolphinTplKeyDocRetrieve,
			expected: false,
		},
		{
			name: "same key in both pre and post, first is enabled but second is disabled",
			config: &Config{
				PreDolphin: []*DolphinTpl{
					{Key: cdaenum.DolphinTplKeyDocRetrieve, Value: "value1", Enabled: true},
				},
				PostDolphin: []*DolphinTpl{
					{Key: cdaenum.DolphinTplKeyDocRetrieve, Value: "value2", Enabled: false},
				},
			},
			key:      cdaenum.DolphinTplKeyDocRetrieve,
			expected: true, // Any disabled template returns true
		},
		{
			name: "same key in both pre and post, first is disabled",
			config: &Config{
				PreDolphin: []*DolphinTpl{
					{Key: cdaenum.DolphinTplKeyDocRetrieve, Value: "value1", Enabled: false},
				},
				PostDolphin: []*DolphinTpl{
					{Key: cdaenum.DolphinTplKeyDocRetrieve, Value: "value2", Enabled: true},
				},
			},
			key:      cdaenum.DolphinTplKeyDocRetrieve,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			disabled := tt.config.IsOneDolphinTplDisabled(tt.key)
			assert.Equal(t, tt.expected, disabled)
		})
	}
}

func TestConfig_IsOneDolphinTplEdited(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		config   *Config
		key      cdaenum.DolphinTplKey
		expected bool
	}{
		{
			name: "template is edited in pre dolphin",
			config: &Config{
				PreDolphin: []*DolphinTpl{
					{Key: cdaenum.DolphinTplKeyDocRetrieve, Value: "value", Edited: true},
				},
			},
			key:      cdaenum.DolphinTplKeyDocRetrieve,
			expected: true,
		},
		{
			name: "template is not edited in pre dolphin",
			config: &Config{
				PreDolphin: []*DolphinTpl{
					{Key: cdaenum.DolphinTplKeyDocRetrieve, Value: "value", Edited: false},
				},
			},
			key:      cdaenum.DolphinTplKeyDocRetrieve,
			expected: false,
		},
		{
			name: "template is edited in post dolphin",
			config: &Config{
				PostDolphin: []*DolphinTpl{
					{Key: cdaenum.DolphinTplKeyRelatedQuestions, Value: "value", Edited: true},
				},
			},
			key:      cdaenum.DolphinTplKeyRelatedQuestions,
			expected: true,
		},
		{
			name:     "template not found",
			config:   &Config{},
			key:      cdaenum.DolphinTplKeyDocRetrieve,
			expected: false,
		},
		{
			name: "same key in both pre and post, first is edited",
			config: &Config{
				PreDolphin: []*DolphinTpl{
					{Key: cdaenum.DolphinTplKeyDocRetrieve, Value: "value1", Edited: true},
				},
				PostDolphin: []*DolphinTpl{
					{Key: cdaenum.DolphinTplKeyDocRetrieve, Value: "value2", Edited: false},
				},
			},
			key:      cdaenum.DolphinTplKeyDocRetrieve,
			expected: true,
		},
		{
			name: "same key in both pre and post, first is not edited",
			config: &Config{
				PreDolphin: []*DolphinTpl{
					{Key: cdaenum.DolphinTplKeyDocRetrieve, Value: "value1", Edited: false},
				},
				PostDolphin: []*DolphinTpl{
					{Key: cdaenum.DolphinTplKeyDocRetrieve, Value: "value2", Edited: true},
				},
			},
			key:      cdaenum.DolphinTplKeyDocRetrieve,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			edited := tt.config.IsOneDolphinTplEdited(tt.key)
			assert.Equal(t, tt.expected, edited)
		})
	}
}

func TestConfig_RemoveDataSourceFromPreDolphin(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                    string
		config                  *Config
		contextOrganizeValue    string
		expectedPreDolphinCount int
		expectedContextValue    string
		panicCheck              func(*Config)
	}{
		{
			name: "remove unedited doc retrieve and graph retrieve",
			config: &Config{
				DataSource: nil,
				PreDolphin: []*DolphinTpl{
					{Key: cdaenum.DolphinTplKeyDocRetrieve, Value: "doc value", Edited: false},
					{Key: cdaenum.DolphinTplKeyGraphRetrieve, Value: "graph value", Edited: false},
					{Key: cdaenum.DolphinTplKeyContextOrganize, Value: "old context", Edited: false},
					{Key: cdaenum.DolphinTplKeyRelatedQuestions, Value: "agent instruction", Edited: false},
				},
			},
			contextOrganizeValue:    "new context",
			expectedPreDolphinCount: 2,
			expectedContextValue:    "new context",
			panicCheck: func(c *Config) {
				// No panic expected
			},
		},
		{
			name: "keep edited doc retrieve and graph retrieve",
			config: &Config{
				DataSource: nil,
				PreDolphin: []*DolphinTpl{
					{Key: cdaenum.DolphinTplKeyDocRetrieve, Value: "doc value", Edited: true},
					{Key: cdaenum.DolphinTplKeyGraphRetrieve, Value: "graph value", Edited: true},
					{Key: cdaenum.DolphinTplKeyContextOrganize, Value: "old context", Edited: false},
					{Key: cdaenum.DolphinTplKeyRelatedQuestions, Value: "agent instruction", Edited: false},
				},
			},
			contextOrganizeValue:    "new context",
			expectedPreDolphinCount: 4,
			expectedContextValue:    "new context",
			panicCheck: func(c *Config) {
				// No panic expected
			},
		},
		{
			name: "keep edited context organize with original value",
			config: &Config{
				DataSource: nil,
				PreDolphin: []*DolphinTpl{
					{Key: cdaenum.DolphinTplKeyDocRetrieve, Value: "doc value", Edited: false},
					{Key: cdaenum.DolphinTplKeyGraphRetrieve, Value: "graph value", Edited: false},
					{Key: cdaenum.DolphinTplKeyContextOrganize, Value: "original context", Edited: true},
					{Key: cdaenum.DolphinTplKeyRelatedQuestions, Value: "agent instruction", Edited: false},
				},
			},
			contextOrganizeValue:    "new context",
			expectedPreDolphinCount: 2,
			expectedContextValue:    "original context",
			panicCheck: func(c *Config) {
				// No panic expected
			},
		},
		{
			name: "mixed edit status",
			config: &Config{
				DataSource: nil,
				PreDolphin: []*DolphinTpl{
					{Key: cdaenum.DolphinTplKeyDocRetrieve, Value: "doc value", Edited: false},
					{Key: cdaenum.DolphinTplKeyGraphRetrieve, Value: "graph value", Edited: true},
					{Key: cdaenum.DolphinTplKeyContextOrganize, Value: "old context", Edited: false},
					{Key: cdaenum.DolphinTplKeyRelatedQuestions, Value: "agent instruction", Edited: true},
				},
			},
			contextOrganizeValue:    "new context",
			expectedPreDolphinCount: 3,
			expectedContextValue:    "new context",
			panicCheck: func(c *Config) {
				// No panic expected
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// First verify no panic happens by checking DataSource is nil or not set
			if tt.config.DataSource != nil && !tt.config.DataSource.IsNotSet() {
				assert.Panics(t, func() {
					tt.config.RemoveDataSourceFromPreDolphin(tt.contextOrganizeValue) //nolint:errcheck
				})
			} else {
				err := tt.config.RemoveDataSourceFromPreDolphin(tt.contextOrganizeValue)
				assert.NoError(t, err)
				assert.Len(t, tt.config.PreDolphin, tt.expectedPreDolphinCount)

				// Check context organize value
				for _, tpl := range tt.config.PreDolphin {
					if tpl.Key == cdaenum.DolphinTplKeyContextOrganize {
						assert.Equal(t, tt.expectedContextValue, tpl.Value)
					}
				}
			}
		})
	}
}

func TestConfig_RemoveDataSourceFromPreDolphin_Panics(t *testing.T) {
	t.Parallel()

	config := &Config{
		DataSource: &datasourcevalobj.RetrieverDataSource{
			Doc: []*datasourcevalobj.DocSource{
				{DsID: "test"},
			},
		},
		PreDolphin: []*DolphinTpl{
			{Key: cdaenum.DolphinTplKeyDocRetrieve, Value: "value", Edited: false},
		},
	}

	assert.Panics(t, func() {
		config.RemoveDataSourceFromPreDolphin("new context") //nolint:errcheck
	})
}
