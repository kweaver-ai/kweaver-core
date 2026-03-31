package datasourcevalobj

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMetricSource_ValObjCheck_Valid(t *testing.T) {
	t.Parallel()

	metric := &MetricSource{
		MetricModelID: "metric-123",
	}

	err := metric.ValObjCheck()
	assert.NoError(t, err)
}

func TestMetricSource_ValObjCheck_Error(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		m    *MetricSource
	}{
		{
			name: "empty metric model id",
			m: &MetricSource{
				MetricModelID: "",
			},
		},
		{
			name: "nil struct",
			m:    &MetricSource{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.m.ValObjCheck()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "metric_model_id is required")
		})
	}
}

func TestMetricSource_GetErrMsgMap(t *testing.T) {
	t.Parallel()

	metric := &MetricSource{}
	errMap := metric.GetErrMsgMap()

	assert.NotNil(t, errMap)
	assert.Len(t, errMap, 1)
	assert.Equal(t, `"metric_model_id"不能为空`, errMap["MetricModelID.required"])
}

func TestMetricSource_Fields(t *testing.T) {
	t.Parallel()

	metric := &MetricSource{
		MetricModelID: "metric-456",
	}

	assert.Equal(t, "metric-456", metric.MetricModelID)
}
