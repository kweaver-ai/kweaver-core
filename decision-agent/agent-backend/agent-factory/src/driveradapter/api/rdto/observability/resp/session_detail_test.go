package observabilityresp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSessionDetailResp_StructFields(t *testing.T) {
	t.Parallel()

	resp := SessionDetailResp{
		SessionID:             "session-123",
		StartTime:             1700000000000,
		EndTime:               1700003600000,
		SessionRunCount:       5,
		SessionDuration:       3600000,
		AvgRunExecuteDuration: 5000,
		AvgRunTTFTDuration:    150,
		RunErrorCount:         1,
		ToolFailCount:         2,
	}

	assert.Equal(t, "session-123", resp.SessionID)
	assert.Equal(t, int64(1700000000000), resp.StartTime)
	assert.Equal(t, int64(1700003600000), resp.EndTime)
	assert.Equal(t, 5, resp.SessionRunCount)
	assert.Equal(t, 3600000, resp.SessionDuration)
	assert.Equal(t, 5000, resp.AvgRunExecuteDuration)
	assert.Equal(t, 150, resp.AvgRunTTFTDuration)
	assert.Equal(t, 1, resp.RunErrorCount)
	assert.Equal(t, 2, resp.ToolFailCount)
}

func TestSessionDetailResp_Empty(t *testing.T) {
	t.Parallel()

	resp := SessionDetailResp{}

	assert.Empty(t, resp.SessionID)
	assert.Zero(t, resp.StartTime)
	assert.Zero(t, resp.EndTime)
	assert.Zero(t, resp.SessionRunCount)
	assert.Zero(t, resp.SessionDuration)
	assert.Zero(t, resp.AvgRunExecuteDuration)
	assert.Zero(t, resp.AvgRunTTFTDuration)
	assert.Zero(t, resp.RunErrorCount)
	assert.Zero(t, resp.ToolFailCount)
}

func TestSessionDetailResp_WithDurationMetrics(t *testing.T) {
	t.Parallel()

	resp := SessionDetailResp{
		SessionDuration:       10000,
		AvgRunExecuteDuration: 2000,
		AvgRunTTFTDuration:    100,
	}

	assert.Equal(t, 10000, resp.SessionDuration)
	assert.Equal(t, 2000, resp.AvgRunExecuteDuration)
	assert.Equal(t, 100, resp.AvgRunTTFTDuration)
}

func TestSessionDetailResp_WithErrorCounts(t *testing.T) {
	t.Parallel()

	resp := SessionDetailResp{
		RunErrorCount: 5,
		ToolFailCount: 3,
	}

	assert.Equal(t, 5, resp.RunErrorCount)
	assert.Equal(t, 3, resp.ToolFailCount)
}

func TestSessionDetailResp_WithZeroValues(t *testing.T) {
	t.Parallel()

	resp := SessionDetailResp{
		SessionID:       "session-zero",
		SessionRunCount: 0,
		RunErrorCount:   0,
		ToolFailCount:   0,
	}

	assert.Equal(t, "session-zero", resp.SessionID)
	assert.Zero(t, resp.SessionRunCount)
	assert.Zero(t, resp.RunErrorCount)
	assert.Zero(t, resp.ToolFailCount)
}

func TestSessionDetailResp_WithLargeValues(t *testing.T) {
	t.Parallel()

	resp := SessionDetailResp{
		SessionRunCount:       1000,
		SessionDuration:       3600000,
		AvgRunExecuteDuration: 10000,
		AvgRunTTFTDuration:    500,
		RunErrorCount:         100,
		ToolFailCount:         50,
	}

	assert.Equal(t, 1000, resp.SessionRunCount)
	assert.Equal(t, 3600000, resp.SessionDuration)
	assert.Equal(t, 10000, resp.AvgRunExecuteDuration)
	assert.Equal(t, 500, resp.AvgRunTTFTDuration)
	assert.Equal(t, 100, resp.RunErrorCount)
	assert.Equal(t, 50, resp.ToolFailCount)
}

func TestSessionDetailResp_WithAllFields(t *testing.T) {
	t.Parallel()

	resp := SessionDetailResp{
		SessionID:             "complete-session",
		StartTime:             1700000000000,
		EndTime:               1700010000000,
		SessionRunCount:       10,
		SessionDuration:       1000000,
		AvgRunExecuteDuration: 3000,
		AvgRunTTFTDuration:    200,
		RunErrorCount:         0,
		ToolFailCount:         0,
	}

	assert.Equal(t, "complete-session", resp.SessionID)
	assert.Equal(t, int64(1700000000000), resp.StartTime)
	assert.Equal(t, int64(1700010000000), resp.EndTime)
	assert.Equal(t, 10, resp.SessionRunCount)
	assert.Equal(t, 1000000, resp.SessionDuration)
	assert.Equal(t, 3000, resp.AvgRunExecuteDuration)
	assert.Equal(t, 200, resp.AvgRunTTFTDuration)
	assert.Equal(t, 0, resp.RunErrorCount)
	assert.Equal(t, 0, resp.ToolFailCount)
}

func TestSessionDetailResp_WithTimeRange(t *testing.T) {
	t.Parallel()

	resp := SessionDetailResp{
		SessionID: "time-range-session",
		StartTime: 1609459200000, // 2021-01-01
		EndTime:   1609545600000, // 2021-01-02
	}

	assert.Equal(t, "time-range-session", resp.SessionID)
	assert.Equal(t, int64(1609459200000), resp.StartTime)
	assert.Equal(t, int64(1609545600000), resp.EndTime)
	assert.Greater(t, resp.EndTime, resp.StartTime)
}
