package releaseeo

import (
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/enum/daenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/valueobject/daconfvalobj"
	"github.com/stretchr/testify/assert"
)

func TestReleaseEO_IsPmsCtrlBool(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		eo   *ReleaseEO
		want bool
	}{
		{
			name: "IsPmsCtrl is 1",
			eo: &ReleaseEO{
				IsPmsCtrl: 1,
			},
			want: true,
		},
		{
			name: "IsPmsCtrl is 0",
			eo: &ReleaseEO{
				IsPmsCtrl: 0,
			},
			want: false,
		},
		{
			name: "IsPmsCtrl is 2",
			eo: &ReleaseEO{
				IsPmsCtrl: 2,
			},
			want: false,
		},
		{
			name: "IsPmsCtrl is -1",
			eo: &ReleaseEO{
				IsPmsCtrl: -1,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.eo.IsPmsCtrlBool()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestReleaseEO_SetIsPmsCtrl(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		eo        *ReleaseEO
		isPmsCtrl bool
		want      int
	}{
		{
			name:      "set to true",
			eo:        &ReleaseEO{IsPmsCtrl: 0},
			isPmsCtrl: true,
			want:      1,
		},
		{
			name:      "set to false",
			eo:        &ReleaseEO{IsPmsCtrl: 1},
			isPmsCtrl: false,
			want:      0,
		},
		{
			name:      "set to true when already true",
			eo:        &ReleaseEO{IsPmsCtrl: 1},
			isPmsCtrl: true,
			want:      1,
		},
		{
			name:      "set to false when already false",
			eo:        &ReleaseEO{IsPmsCtrl: 0},
			isPmsCtrl: false,
			want:      0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.eo.SetIsPmsCtrl(tt.isPmsCtrl)
			assert.Equal(t, tt.want, tt.eo.IsPmsCtrl)
		})
	}
}

func TestReleaseDAConfWrapperEO(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		eo    *ReleaseDAConfWrapperEO
		check func(t *testing.T, eo *ReleaseDAConfWrapperEO)
	}{
		{
			name: "valid wrapper with config",
			eo: &ReleaseDAConfWrapperEO{
				ReleaseEO: ReleaseEO{
					ID:           "release-1",
					AgentID:      "agent-1",
					AgentVersion: "v1.0.0",
				},
				Config: &daconfvalobj.Config{
					Input:  &daconfvalobj.Input{},
					Output: &daconfvalobj.Output{},
				},
			},
			check: func(t *testing.T, eo *ReleaseDAConfWrapperEO) {
				assert.Equal(t, "release-1", eo.ID)
				assert.Equal(t, "agent-1", eo.AgentID)
				assert.NotNil(t, eo.Config)
			},
		},
		{
			name: "wrapper without config",
			eo: &ReleaseDAConfWrapperEO{
				ReleaseEO: ReleaseEO{
					ID:           "release-2",
					AgentVersion: "v2.0.0",
				},
				Config: nil,
			},
			check: func(t *testing.T, eo *ReleaseDAConfWrapperEO) {
				assert.Equal(t, "release-2", eo.ID)
				assert.Nil(t, eo.Config)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.check != nil {
				tt.check(t, tt.eo)
			}
		})
	}
}

func TestReleaseEO_PublishToBes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		eo    *ReleaseEO
		check func(t *testing.T, eo *ReleaseEO)
	}{
		{
			name: "with publish targets",
			eo: &ReleaseEO{
				ID: "release-1",
				PublishToBes: []cdaenum.PublishToBe{
					cdaenum.PublishToBeAPIAgent,
					cdaenum.PublishToBeWebSDKAgent,
				},
			},
			check: func(t *testing.T, eo *ReleaseEO) {
				assert.Len(t, eo.PublishToBes, 2)
				assert.Equal(t, cdaenum.PublishToBeAPIAgent, eo.PublishToBes[0])
				assert.Equal(t, cdaenum.PublishToBeWebSDKAgent, eo.PublishToBes[1])
			},
		},
		{
			name: "with empty publish targets",
			eo: &ReleaseEO{
				ID:           "release-2",
				PublishToBes: []cdaenum.PublishToBe{},
			},
			check: func(t *testing.T, eo *ReleaseEO) {
				assert.Len(t, eo.PublishToBes, 0)
			},
		},
		{
			name: "with nil publish targets",
			eo: &ReleaseEO{
				ID:           "release-3",
				PublishToBes: nil,
			},
			check: func(t *testing.T, eo *ReleaseEO) {
				assert.Nil(t, eo.PublishToBes)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.check != nil {
				tt.check(t, tt.eo)
			}
		})
	}
}

func TestReleaseEO_PublishToWhere(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		eo    *ReleaseEO
		check func(t *testing.T, eo *ReleaseEO)
	}{
		{
			name: "with publish where targets",
			eo: &ReleaseEO{
				ID: "release-1",
				PublishToWhere: []daenum.PublishToWhere{
					daenum.PublishToWhereCustomSpace,
					daenum.PublishToWhereSquare,
				},
			},
			check: func(t *testing.T, eo *ReleaseEO) {
				assert.Len(t, eo.PublishToWhere, 2)
				assert.Equal(t, daenum.PublishToWhereCustomSpace, eo.PublishToWhere[0])
				assert.Equal(t, daenum.PublishToWhereSquare, eo.PublishToWhere[1])
			},
		},
		{
			name: "with empty publish where",
			eo: &ReleaseEO{
				ID:             "release-2",
				PublishToWhere: []daenum.PublishToWhere{},
			},
			check: func(t *testing.T, eo *ReleaseEO) {
				assert.Len(t, eo.PublishToWhere, 0)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.check != nil {
				tt.check(t, tt.eo)
			}
		})
	}
}
