package productreq

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateReq_StructFields(t *testing.T) {
	t.Parallel()

	req := UpdateReq{
		Name:    "Updated Product Name",
		Profile: "Updated product profile",
	}

	assert.Equal(t, "Updated Product Name", req.Name)
	assert.Equal(t, "Updated product profile", req.Profile)
}

func TestUpdateReq_Empty(t *testing.T) {
	t.Parallel()

	req := UpdateReq{}

	assert.Empty(t, req.Name)
	assert.Empty(t, req.Profile)
}

func TestUpdateReq_GetErrMsgMap(t *testing.T) {
	t.Parallel()

	req := UpdateReq{}

	errMsgMap := req.GetErrMsgMap()

	assert.NotNil(t, errMsgMap)
	assert.Equal(t, `"name"不能为空`, errMsgMap["Name.required"])
	assert.Equal(t, `"name"长度不能超过50`, errMsgMap["Name.max"])
	assert.Equal(t, `"profile"长度不能超过100`, errMsgMap["Profile.max"])
}

func TestUpdateReq_CustomCheck(t *testing.T) {
	t.Parallel()

	req := UpdateReq{
		Name: "Test Product",
	}

	err := req.CustomCheck()

	assert.NoError(t, err)
}

func TestUpdateReq_WithName(t *testing.T) {
	t.Parallel()

	names := []string{
		"Updated Product Name",
		"更新的产品名称",
		"Product with numbers 123",
	}

	for _, name := range names {
		req := UpdateReq{
			Name: name,
		}
		assert.Equal(t, name, req.Name)
	}
}

func TestUpdateReq_WithProfile(t *testing.T) {
	t.Parallel()

	profiles := []string{
		"Updated product profile",
		"更新的产品简介",
		"Profile with numbers 123",
		"",
	}

	for _, profile := range profiles {
		req := UpdateReq{
			Profile: profile,
		}
		assert.Equal(t, profile, req.Profile)
	}
}

func TestUpdateReq_WithMaxLengths(t *testing.T) {
	t.Parallel()

	req := UpdateReq{
		Name:    strings.Repeat("a", 50),
		Profile: strings.Repeat("b", 100),
	}

	assert.Len(t, req.Name, 50)
	assert.Len(t, req.Profile, 100)
}

func TestUpdateReq_WithAllFields(t *testing.T) {
	t.Parallel()

	req := UpdateReq{
		Name:    "Complete Update Name",
		Profile: "Complete update profile with description",
	}

	assert.Equal(t, "Complete Update Name", req.Name)
	assert.Equal(t, "Complete update profile with description", req.Profile)

	err := req.CustomCheck()
	assert.NoError(t, err)
}

func TestUpdateReq_WithChineseName(t *testing.T) {
	t.Parallel()

	req := UpdateReq{
		Name: "更新的产品名称",
	}

	assert.Equal(t, "更新的产品名称", req.Name)
}

func TestUpdateReq_WithMixedName(t *testing.T) {
	t.Parallel()

	req := UpdateReq{
		Name: "Updated更新的产品Name",
	}

	assert.Equal(t, "Updated更新的产品Name", req.Name)
}

func TestUpdateReq_WithEmptyName(t *testing.T) {
	t.Parallel()

	req := UpdateReq{
		Name: "",
	}

	assert.Empty(t, req.Name)
}

func TestUpdateReq_WithLongProfile(t *testing.T) {
	t.Parallel()

	req := UpdateReq{
		Profile: strings.Repeat("a", 100),
	}

	assert.Len(t, req.Profile, 100)
}

func TestUpdateReq_D2e(t *testing.T) {
	t.Parallel()

	t.Run("with all fields", func(t *testing.T) {
		t.Parallel()

		req := UpdateReq{
			Name:    "Updated Product",
			Profile: "Updated profile",
		}

		eo, err := req.D2e()

		require.NoError(t, err)
		require.NotNil(t, eo)
		assert.Equal(t, req.Name, eo.Name)
		assert.Equal(t, req.Profile, eo.Profile)
	})

	t.Run("with only name", func(t *testing.T) {
		t.Parallel()

		req := UpdateReq{
			Name: "Minimal Update",
		}

		eo, err := req.D2e()

		require.NoError(t, err)
		require.NotNil(t, eo)
		assert.Equal(t, req.Name, eo.Name)
		assert.Empty(t, eo.Profile)
	})

	t.Run("with empty profile", func(t *testing.T) {
		t.Parallel()

		req := UpdateReq{
			Name:    "Product Empty Profile",
			Profile: "",
		}

		eo, err := req.D2e()

		require.NoError(t, err)
		require.NotNil(t, eo)
		assert.Equal(t, req.Name, eo.Name)
		assert.Empty(t, eo.Profile)
	})

	t.Run("with chinese characters", func(t *testing.T) {
		t.Parallel()

		req := UpdateReq{
			Name:    "更新的产品名称",
			Profile: "更新的产品简介",
		}

		eo, err := req.D2e()

		require.NoError(t, err)
		require.NotNil(t, eo)
		assert.Equal(t, req.Name, eo.Name)
		assert.Equal(t, req.Profile, eo.Profile)
	})

	t.Run("with long values", func(t *testing.T) {
		t.Parallel()

		req := UpdateReq{
			Name:    strings.Repeat("a", 50),
			Profile: strings.Repeat("b", 100),
		}

		eo, err := req.D2e()

		require.NoError(t, err)
		require.NotNil(t, eo)
		assert.Len(t, eo.Name, 50)
		assert.Len(t, eo.Profile, 100)
	})

	t.Run("with empty name", func(t *testing.T) {
		t.Parallel()

		req := UpdateReq{
			Name:    "",
			Profile: "test profile",
		}

		eo, err := req.D2e()

		require.NoError(t, err)
		require.NotNil(t, eo)
		assert.Empty(t, eo.Name)
		assert.Equal(t, req.Profile, eo.Profile)
	})

	t.Run("with mixed content", func(t *testing.T) {
		t.Parallel()

		req := UpdateReq{
			Name:    "Updated更新的产品Name",
			Profile: "Updated profile简介",
		}

		eo, err := req.D2e()

		require.NoError(t, err)
		require.NotNil(t, eo)
		assert.Equal(t, req.Name, eo.Name)
		assert.Equal(t, req.Profile, eo.Profile)
	})
}
