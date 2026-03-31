package productreq

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateReq_StructFields(t *testing.T) {
	t.Parallel()

	req := CreateReq{
		Name:    "Smart Customer Service",
		Profile: "This is a smart customer service product",
		Key:     "smart-customer-service",
	}

	assert.Equal(t, "Smart Customer Service", req.Name)
	assert.Equal(t, "This is a smart customer service product", req.Profile)
	assert.Equal(t, "smart-customer-service", req.Key)
}

func TestCreateReq_Empty(t *testing.T) {
	t.Parallel()

	req := CreateReq{}

	assert.Empty(t, req.Name)
	assert.Empty(t, req.Profile)
	assert.Empty(t, req.Key)
}

func TestCreateReq_GetErrMsgMap(t *testing.T) {
	t.Parallel()

	req := CreateReq{}

	errMsgMap := req.GetErrMsgMap()

	assert.NotNil(t, errMsgMap)
	assert.Equal(t, `"name"不能为空`, errMsgMap["Name.required"])
	assert.Equal(t, `"name"长度不能超过50`, errMsgMap["Name.max"])
	assert.Equal(t, `"profile"长度不能超过100`, errMsgMap["Profile.max"])
	assert.Equal(t, `"key"长度不能超过50`, errMsgMap["Key.max"])
}

func TestCreateReq_CustomCheck(t *testing.T) {
	t.Parallel()

	req := CreateReq{
		Name: "Test Product",
	}

	err := req.CustomCheck()

	assert.NoError(t, err)
}

func TestCreateReq_WithName(t *testing.T) {
	t.Parallel()

	names := []string{
		"Smart Customer Service",
		"智能客服产品",
		"Product with numbers 123",
		"Product with special chars !@#$%",
	}

	for _, name := range names {
		req := CreateReq{
			Name: name,
		}
		assert.Equal(t, name, req.Name)
	}
}

func TestCreateReq_WithProfile(t *testing.T) {
	t.Parallel()

	profiles := []string{
		"This is a smart customer service product",
		"这是一个智能客服产品",
		"Profile with numbers 123",
		"",
	}

	for _, profile := range profiles {
		req := CreateReq{
			Profile: profile,
		}
		assert.Equal(t, profile, req.Profile)
	}
}

func TestCreateReq_WithKey(t *testing.T) {
	t.Parallel()

	keys := []string{
		"smart-customer-service",
		"product-key-123",
		"产品-key-中文",
		"",
	}

	for _, key := range keys {
		req := CreateReq{
			Key: key,
		}
		assert.Equal(t, key, req.Key)
	}
}

func TestCreateReq_WithMaxLengths(t *testing.T) {
	t.Parallel()

	req := CreateReq{
		Name:    strings.Repeat("a", 50),
		Profile: strings.Repeat("b", 100),
		Key:     strings.Repeat("c", 50),
	}

	assert.Len(t, req.Name, 50)
	assert.Len(t, req.Profile, 100)
	assert.Len(t, req.Key, 50)
}

func TestCreateReq_WithAllFields(t *testing.T) {
	t.Parallel()

	req := CreateReq{
		Name:    "Complete Product Name",
		Profile: "Complete product profile with description",
		Key:     "complete-product-key",
	}

	assert.Equal(t, "Complete Product Name", req.Name)
	assert.Equal(t, "Complete product profile with description", req.Profile)
	assert.Equal(t, "complete-product-key", req.Key)

	err := req.CustomCheck()
	assert.NoError(t, err)
}

func TestCreateReq_WithChineseName(t *testing.T) {
	t.Parallel()

	req := CreateReq{
		Name: "智能客服产品",
	}

	assert.Equal(t, "智能客服产品", req.Name)
}

func TestCreateReq_WithMixedName(t *testing.T) {
	t.Parallel()

	req := CreateReq{
		Name: "Smart智能客服Product",
	}

	assert.Equal(t, "Smart智能客服Product", req.Name)
}

func TestCreateReq_WithEmptyName(t *testing.T) {
	t.Parallel()

	req := CreateReq{
		Name: "",
	}

	assert.Empty(t, req.Name)
}

func TestCreateReq_WithSpecialCharsInKey(t *testing.T) {
	t.Parallel()

	keys := []string{
		"smart-customer-service",
		"smart_customer_service",
		"smart.customer.service",
		"smartcustomerservice",
	}

	for _, key := range keys {
		req := CreateReq{
			Key: key,
		}
		assert.Equal(t, key, req.Key)
	}
}

func TestCreateReq_D2e(t *testing.T) {
	t.Parallel()

	t.Run("with all fields", func(t *testing.T) {
		t.Parallel()

		req := CreateReq{
			Name:    "Smart Customer Service",
			Profile: "This is a smart customer service product",
			Key:     "smart-customer-service",
		}

		eo, err := req.D2e()

		require.NoError(t, err)
		require.NotNil(t, eo)
		assert.Equal(t, req.Name, eo.Name)
		assert.Equal(t, req.Profile, eo.Profile)
		assert.Equal(t, req.Key, eo.Key)
	})

	t.Run("with empty key generates ulid", func(t *testing.T) {
		t.Parallel()

		req := CreateReq{
			Name:    "Test Product",
			Profile: "Test profile",
			Key:     "",
		}

		eo, err := req.D2e()

		require.NoError(t, err)
		require.NotNil(t, eo)
		assert.Equal(t, req.Name, eo.Name)
		assert.Equal(t, req.Profile, eo.Profile)
		assert.NotEmpty(t, eo.Key, "Key should be generated ULID when empty")
	})

	t.Run("with only name", func(t *testing.T) {
		t.Parallel()

		req := CreateReq{
			Name: "Minimal Product",
		}

		eo, err := req.D2e()

		require.NoError(t, err)
		require.NotNil(t, eo)
		assert.Equal(t, req.Name, eo.Name)
		assert.NotEmpty(t, eo.Key, "Key should be generated ULID when empty")
		assert.Empty(t, eo.Profile)
	})

	t.Run("with empty profile", func(t *testing.T) {
		t.Parallel()

		req := CreateReq{
			Name:    "Product No Profile",
			Key:     "product-no-profile",
			Profile: "",
		}

		eo, err := req.D2e()

		require.NoError(t, err)
		require.NotNil(t, eo)
		assert.Equal(t, req.Name, eo.Name)
		assert.Equal(t, req.Key, eo.Key)
		assert.Empty(t, eo.Profile)
	})

	t.Run("with chinese characters", func(t *testing.T) {
		t.Parallel()

		req := CreateReq{
			Name:    "智能客服产品",
			Profile: "这是一个智能客服产品",
			Key:     "zhineng-kefu",
		}

		eo, err := req.D2e()

		require.NoError(t, err)
		require.NotNil(t, eo)
		assert.Equal(t, req.Name, eo.Name)
		assert.Equal(t, req.Profile, eo.Profile)
		assert.Equal(t, req.Key, eo.Key)
	})

	t.Run("with long values", func(t *testing.T) {
		t.Parallel()

		req := CreateReq{
			Name:    strings.Repeat("a", 50),
			Profile: strings.Repeat("b", 100),
			Key:     strings.Repeat("c", 50),
		}

		eo, err := req.D2e()

		require.NoError(t, err)
		require.NotNil(t, eo)
		assert.Len(t, eo.Name, 50)
		assert.Len(t, eo.Profile, 100)
		assert.Len(t, eo.Key, 50)
	})

	t.Run("with empty name", func(t *testing.T) {
		t.Parallel()

		req := CreateReq{
			Name: "",
			Key:  "test-key",
		}

		eo, err := req.D2e()

		require.NoError(t, err)
		require.NotNil(t, eo)
		assert.Empty(t, eo.Name)
		assert.Equal(t, req.Key, eo.Key)
	})
}
