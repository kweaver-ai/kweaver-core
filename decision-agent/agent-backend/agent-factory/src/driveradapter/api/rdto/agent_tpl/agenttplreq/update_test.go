package agenttplreq

import (
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/valueobject/daconfvalobj"
	"github.com/stretchr/testify/assert"
)

func TestUpdateReq_GetErrMsgMap(t *testing.T) {
	t.Parallel()

	req := &UpdateReq{}
	errMsgMap := req.GetErrMsgMap()

	assert.NotNil(t, errMsgMap)
	assert.NotEmpty(t, errMsgMap)

	expectedKeys := []string{
		"Name.required",
		"Name.checkAgentAndTplName",
		"Name.max",
		"Config.required",
		"Profile.max",
	}

	for _, key := range expectedKeys {
		assert.Contains(t, errMsgMap, key)
	}

	assert.Equal(t, `"name"不能为空`, errMsgMap["Name.required"])
	assert.Equal(t, `"config"不能为空`, errMsgMap["Config.required"])
	assert.Equal(t, `"profile"长度不能超过500`, errMsgMap["Profile.max"])
	assert.Equal(t, `"name"长度不能超过50`, errMsgMap["Name.max"])
}

func TestUpdateReq_GetErrMsgMapConsistency(t *testing.T) {
	t.Parallel()

	req1 := &UpdateReq{}
	req2 := &UpdateReq{Name: "Test"}
	req3 := &UpdateReq{Profile: strPtr("test profile")}

	map1 := req1.GetErrMsgMap()
	map2 := req2.GetErrMsgMap()
	map3 := req3.GetErrMsgMap()

	assert.Equal(t, map1, map2)
	assert.Equal(t, map2, map3)
}

func TestUpdateReq_New(t *testing.T) {
	t.Parallel()

	req := &UpdateReq{}

	assert.NotNil(t, req)
	assert.Empty(t, req.Name)
	assert.Nil(t, req.Profile)
	assert.Empty(t, req.Avatar)
	assert.Equal(t, 0, req.AvatarType)
	assert.Nil(t, req.Config)
}

func TestUpdateReq_WithName(t *testing.T) {
	t.Parallel()

	req := &UpdateReq{
		Name: "Test Agent Template",
	}

	assert.Equal(t, "Test Agent Template", req.Name)
}

func TestUpdateReq_WithProfile(t *testing.T) {
	t.Parallel()

	profile := "This is a test profile"
	req := &UpdateReq{
		Profile: &profile,
	}

	assert.NotNil(t, req.Profile)
	assert.Equal(t, profile, *req.Profile)
}

func TestUpdateReq_WithNilProfile(t *testing.T) {
	t.Parallel()

	req := &UpdateReq{
		Profile: nil,
	}

	assert.Nil(t, req.Profile)
}

func TestUpdateReq_WithAvatar(t *testing.T) {
	t.Parallel()

	req := &UpdateReq{
		Avatar: "avatar.png",
	}

	assert.Equal(t, "avatar.png", req.Avatar)
}

func TestUpdateReq_WithAvatarType(t *testing.T) {
	t.Parallel()

	avatarTypes := []int{0, 1, 2, 3}

	for _, avatarType := range avatarTypes {
		req := &UpdateReq{
			AvatarType: avatarType,
		}
		assert.Equal(t, avatarType, req.AvatarType)
	}
}

func TestUpdateReq_WithConfig(t *testing.T) {
	t.Parallel()

	config := &daconfvalobj.Config{
		Input:  &daconfvalobj.Input{},
		Output: &daconfvalobj.Output{},
	}
	req := &UpdateReq{
		Config: config,
	}

	assert.NotNil(t, req.Config)
	assert.Same(t, config, req.Config)
}

func TestUpdateReq_WithNilConfig(t *testing.T) {
	t.Parallel()

	req := &UpdateReq{
		Config: nil,
	}

	assert.Nil(t, req.Config)
}

func TestUpdateReq_AllFields(t *testing.T) {
	t.Parallel()

	profile := "Test profile"
	config := &daconfvalobj.Config{
		Input:  &daconfvalobj.Input{},
		Output: &daconfvalobj.Output{},
	}

	req := &UpdateReq{
		Name:       "Test Template",
		Profile:    &profile,
		Avatar:     "avatar.jpg",
		AvatarType: 1,
		Config:     config,
	}

	assert.Equal(t, "Test Template", req.Name)
	assert.Equal(t, profile, *req.Profile)
	assert.Equal(t, "avatar.jpg", req.Avatar)
	assert.Equal(t, 1, req.AvatarType)
	assert.Same(t, config, req.Config)
}

func TestUpdateReq_D2e(t *testing.T) {
	t.Parallel()

	t.Run("valid request", func(t *testing.T) {
		t.Parallel()

		profile := "Test profile"
		req := &UpdateReq{
			Name:       "Test Template",
			Profile:    &profile,
			Avatar:     "avatar.png",
			AvatarType: 1,
			Config: &daconfvalobj.Config{
				Input:  &daconfvalobj.Input{},
				Output: &daconfvalobj.Output{},
			},
		}

		eo, err := req.D2e()

		assert.NoError(t, err)
		assert.NotNil(t, eo)
		assert.Equal(t, req.Name, eo.Name)
		assert.Equal(t, req.Profile, eo.Profile)
		assert.Equal(t, req.Avatar, eo.Avatar)
		assert.Equal(t, cdaenum.AvatarType(req.AvatarType), eo.AvatarType)
	})

	t.Run("with nil profile", func(t *testing.T) {
		t.Parallel()

		req := &UpdateReq{
			Name:    "Test Template",
			Profile: nil,
			Config: &daconfvalobj.Config{
				Input:  &daconfvalobj.Input{},
				Output: &daconfvalobj.Output{},
			},
		}

		eo, err := req.D2e()

		assert.NoError(t, err)
		assert.NotNil(t, eo)
		assert.Nil(t, eo.Profile)
	})

	t.Run("with empty name", func(t *testing.T) {
		t.Parallel()

		req := &UpdateReq{
			Name: "",
			Config: &daconfvalobj.Config{
				Input:  &daconfvalobj.Input{},
				Output: &daconfvalobj.Output{},
			},
		}

		eo, err := req.D2e()

		assert.NoError(t, err)
		assert.NotNil(t, eo)
		assert.Equal(t, "", eo.Name)
	})

	t.Run("with zero avatar type", func(t *testing.T) {
		t.Parallel()

		req := &UpdateReq{
			Name:       "Test Template",
			AvatarType: 0,
			Config: &daconfvalobj.Config{
				Input:  &daconfvalobj.Input{},
				Output: &daconfvalobj.Output{},
			},
		}

		eo, err := req.D2e()

		assert.NoError(t, err)
		assert.NotNil(t, eo)
		assert.Equal(t, cdaenum.AvatarType(0), eo.AvatarType)
	})
}

func TestUpdateReq_ProfileMaxLength(t *testing.T) {
	t.Parallel()

	shortProfile := "Short profile"
	req := &UpdateReq{
		Profile: &shortProfile,
	}

	assert.Equal(t, "Short profile", *req.Profile)
}

func TestUpdateReq_NameMaxLength(t *testing.T) {
	t.Parallel()

	name50 := "12345678901234567890123456789012345678901234567890"
	req := &UpdateReq{
		Name: name50,
	}

	assert.Equal(t, 50, len(req.Name))
	assert.Equal(t, name50, req.Name)
}

func TestUpdateReq_StructFieldsIndependent(t *testing.T) {
	t.Parallel()

	t.Run("name independent", func(t *testing.T) {
		t.Parallel()

		req := &UpdateReq{Name: "Test"}
		assert.Equal(t, "Test", req.Name)
		assert.Nil(t, req.Profile)
		assert.Nil(t, req.Config)
	})

	t.Run("profile independent", func(t *testing.T) {
		t.Parallel()

		profile := "Profile"
		req := &UpdateReq{Profile: &profile}
		assert.Empty(t, req.Name)
		assert.NotNil(t, req.Profile)
		assert.Nil(t, req.Config)
	})

	t.Run("avatar independent", func(t *testing.T) {
		t.Parallel()

		req := &UpdateReq{Avatar: "test.png"}
		assert.Empty(t, req.Name)
		assert.Nil(t, req.Profile)
		assert.Equal(t, "test.png", req.Avatar)
		assert.Nil(t, req.Config)
	})

	t.Run("avatar type independent", func(t *testing.T) {
		t.Parallel()

		req := &UpdateReq{AvatarType: 2}
		assert.Empty(t, req.Name)
		assert.Nil(t, req.Profile)
		assert.Empty(t, req.Avatar)
		assert.Equal(t, 2, req.AvatarType)
		assert.Nil(t, req.Config)
	})

	t.Run("config independent", func(t *testing.T) {
		t.Parallel()

		config := &daconfvalobj.Config{}
		req := &UpdateReq{Config: config}
		assert.Empty(t, req.Name)
		assert.Nil(t, req.Profile)
		assert.Empty(t, req.Avatar)
		assert.Equal(t, 0, req.AvatarType)
		assert.NotNil(t, req.Config)
	})
}

// Helper function
func strPtr(s string) *string {
	return &s
}
