package daconfvalobj

import (
	"strings"
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/enum/cdaenum"
	"github.com/stretchr/testify/assert"
)

func TestValidate(t *testing.T) {
	t.Parallel()

	// 测试用例：成功验证
	t.Run("Valid Config", func(t *testing.T) {
		t.Parallel()

		maxFileCount := 10
		singleChatMaxSelectFileCount := 3
		config := &TempZoneConfig{
			Name:                         "临时区",
			TmpFileUseType:               cdaenum.TmpFileUseTypeUpload,
			MaxFileCount:                 &maxFileCount,
			SingleChatMaxSelectFileCount: &singleChatMaxSelectFileCount,
			SingleFileSizeLimit:          50,
			SingleFileSizeLimitUnit:      cdaenum.MB,
			SupportDataType:              cdaenum.SupportDataTypes{"file"},
			AllowedFileCategories:        cdaenum.AllowedFileCategories{"document"},
		}

		if err := config.Validate(); err != nil {
			t.Fatalf("预期验证通过，但得到错误: %v", err)
		}
	})

	// 测试用例：TmpFileUseType 缺失
	t.Run("Missing TmpFileUseType", func(t *testing.T) {
		t.Parallel()

		maxFileCount := 10
		singleChatMaxSelectFileCount := 3
		config := &TempZoneConfig{
			Name:                         "临时区",
			MaxFileCount:                 &maxFileCount,
			SingleChatMaxSelectFileCount: &singleChatMaxSelectFileCount,
			SingleFileSizeLimit:          50,
			SingleFileSizeLimitUnit:      cdaenum.MB,
			SupportDataType:              cdaenum.SupportDataTypes{"file"},
			AllowedFileCategories:        cdaenum.AllowedFileCategories{"document"},
		}

		err := config.Validate()
		if err == nil {
			t.Fatal("预期验证失败，但没有得到错误")
		}

		if !strings.Contains(err.Error(), "TmpFileUseType") {
			t.Fatalf("预期错误包含 'TmpFileUseType'，但得到: %v", err)
		}
	})

	// 测试用例：MaxFileCount 超过最大值
	// 注意：Validate()方法只检查required字段，不检查数值范围，所以这个测试应该通过
	t.Run("MaxFileCount Exceeds Maximum", func(t *testing.T) {
		t.Parallel()

		maxFileCount := 51 // 超过最大值50，但Validate()不会检查这个
		singleChatMaxSelectFileCount := 3
		config := &TempZoneConfig{
			Name:                         "临时区",
			TmpFileUseType:               cdaenum.TmpFileUseTypeUpload,
			MaxFileCount:                 &maxFileCount,
			SingleChatMaxSelectFileCount: &singleChatMaxSelectFileCount,
			SingleFileSizeLimit:          50,
			SingleFileSizeLimitUnit:      cdaenum.MB,
			SupportDataType:              cdaenum.SupportDataTypes{"file"},
			AllowedFileCategories:        cdaenum.AllowedFileCategories{"document"},
		}

		err := config.Validate()
		// Validate()只检查required字段，不检查数值范围，所以应该通过
		if err != nil {
			t.Fatalf("预期验证通过，但得到错误: %v", err)
		}
	})

	// 测试用例：MaxFileCount 低于最小值
	// 注意：Validate()方法只检查required字段，不检查数值范围
	t.Run("MaxFileCount Below Minimum", func(t *testing.T) {
		t.Parallel()

		maxFileCount := 0 // 低于最小值1，但Validate()不会检查这个
		singleChatMaxSelectFileCount := 3
		config := &TempZoneConfig{
			Name:                         "临时区",
			TmpFileUseType:               cdaenum.TmpFileUseTypeUpload,
			MaxFileCount:                 &maxFileCount,
			SingleChatMaxSelectFileCount: &singleChatMaxSelectFileCount,
			SingleFileSizeLimit:          50,
			SingleFileSizeLimitUnit:      cdaenum.MB,
			SupportDataType:              cdaenum.SupportDataTypes{"file"},
			AllowedFileCategories:        cdaenum.AllowedFileCategories{"document"},
		}

		err := config.Validate()
		// Validate()只检查required字段，不检查数值范围，所以应该通过
		if err != nil {
			t.Fatalf("预期验证通过，但得到错误: %v", err)
		}
	})

	// 测试用例：SingleChatMaxSelectFileCount 超过最大值
	// 注意：Validate()方法只检查required字段，不检查数值范围
	t.Run("SingleChatMaxSelectFileCount Exceeds Maximum", func(t *testing.T) {
		t.Parallel()

		maxFileCount := 10
		singleChatMaxSelectFileCount := 6 // 超过最大值5，但Validate()不会检查这个
		config := &TempZoneConfig{
			Name:                         "临时区",
			TmpFileUseType:               cdaenum.TmpFileUseTypeUpload,
			MaxFileCount:                 &maxFileCount,
			SingleChatMaxSelectFileCount: &singleChatMaxSelectFileCount,
			SingleFileSizeLimit:          50,
			SingleFileSizeLimitUnit:      cdaenum.MB,
			SupportDataType:              cdaenum.SupportDataTypes{"file"},
			AllowedFileCategories:        cdaenum.AllowedFileCategories{"document"},
		}

		err := config.Validate()
		// Validate()只检查required字段，不检查数值范围，所以应该通过
		if err != nil {
			t.Fatalf("预期验证通过，但得到错误: %v", err)
		}
	})

	// 测试用例：SingleChatMaxSelectFileCount 低于最小值
	// 注意：Validate()方法只检查required字段，不检查数值范围
	t.Run("SingleChatMaxSelectFileCount Below Minimum", func(t *testing.T) {
		t.Parallel()

		maxFileCount := 10
		singleChatMaxSelectFileCount := 0 // 低于最小值1，但Validate()不会检查这个
		config := &TempZoneConfig{
			Name:                         "临时区",
			TmpFileUseType:               cdaenum.TmpFileUseTypeUpload,
			MaxFileCount:                 &maxFileCount,
			SingleChatMaxSelectFileCount: &singleChatMaxSelectFileCount,
			SingleFileSizeLimit:          50,
			SingleFileSizeLimitUnit:      cdaenum.MB,
			SupportDataType:              cdaenum.SupportDataTypes{"file"},
			AllowedFileCategories:        cdaenum.AllowedFileCategories{"document"},
		}

		err := config.Validate()
		// Validate()只检查required字段，不检查数值范围，所以应该通过
		if err != nil {
			t.Fatalf("预期验证通过，但得到错误: %v", err)
		}
	})

	// 测试用例：SingleFileSizeLimit 缺失
	t.Run("Missing SingleFileSizeLimit", func(t *testing.T) {
		t.Parallel()

		maxFileCount := 10
		singleChatMaxSelectFileCount := 3
		config := &TempZoneConfig{
			Name:                         "临时区",
			TmpFileUseType:               cdaenum.TmpFileUseTypeUpload,
			MaxFileCount:                 &maxFileCount,
			SingleChatMaxSelectFileCount: &singleChatMaxSelectFileCount,
			SingleFileSizeLimitUnit:      cdaenum.MB,
			SupportDataType:              cdaenum.SupportDataTypes{"file"},
			AllowedFileCategories:        cdaenum.AllowedFileCategories{"document"},
		}

		err := config.Validate()
		if err == nil {
			t.Fatal("预期验证失败，但没有得到错误")
		}

		if !strings.Contains(err.Error(), "SingleFileSizeLimit") {
			t.Fatalf("预期错误包含 'SingleFileSizeLimit'，但得到: %v", err)
		}
	})

	// 测试用例：SingleFileSizeLimitUnit 缺失
	t.Run("Missing SingleFileSizeLimitUnit", func(t *testing.T) {
		t.Parallel()

		maxFileCount := 10
		singleChatMaxSelectFileCount := 3
		config := &TempZoneConfig{
			Name:                         "临时区",
			TmpFileUseType:               cdaenum.TmpFileUseTypeUpload,
			MaxFileCount:                 &maxFileCount,
			SingleChatMaxSelectFileCount: &singleChatMaxSelectFileCount,
			SingleFileSizeLimit:          50,
			SupportDataType:              cdaenum.SupportDataTypes{"file"},
			AllowedFileCategories:        cdaenum.AllowedFileCategories{"document"},
		}

		err := config.Validate()
		if err == nil {
			t.Fatal("预期验证失败，但没有得到错误")
		}

		if !strings.Contains(err.Error(), "SingleFileSizeLimitUnit") {
			t.Fatalf("预期错误包含 'SingleFileSizeLimitUnit'，但得到: %v", err)
		}
	})

	// 测试用例：SupportDataType 缺失
	t.Run("Missing SupportDataType", func(t *testing.T) {
		t.Parallel()

		maxFileCount := 10
		singleChatMaxSelectFileCount := 3
		config := &TempZoneConfig{
			Name:                         "临时区",
			TmpFileUseType:               cdaenum.TmpFileUseTypeUpload,
			MaxFileCount:                 &maxFileCount,
			SingleChatMaxSelectFileCount: &singleChatMaxSelectFileCount,
			SingleFileSizeLimit:          50,
			SingleFileSizeLimitUnit:      cdaenum.MB,
			AllowedFileCategories:        cdaenum.AllowedFileCategories{"document"},
		}

		err := config.Validate()
		if err == nil {
			t.Fatal("预期验证失败，但没有得到错误")
		}

		if !strings.Contains(err.Error(), "SupportDataType") {
			t.Fatalf("预期错误包含 'SupportDataType'，但得到: %v", err)
		}
	})

	// 测试用例：AllowedFileCategories 缺失
	t.Run("Missing AllowedFileCategories", func(t *testing.T) {
		t.Parallel()

		maxFileCount := 10
		singleChatMaxSelectFileCount := 3
		config := &TempZoneConfig{
			Name:                         "临时区",
			TmpFileUseType:               cdaenum.TmpFileUseTypeUpload,
			MaxFileCount:                 &maxFileCount,
			SingleChatMaxSelectFileCount: &singleChatMaxSelectFileCount,
			SingleFileSizeLimit:          50,
			SingleFileSizeLimitUnit:      cdaenum.MB,
			SupportDataType:              cdaenum.SupportDataTypes{"file"},
		}

		err := config.Validate()
		if err == nil {
			t.Fatal("预期验证失败，但没有得到错误")
		}

		if !strings.Contains(err.Error(), "AllowedFileCategories") {
			t.Fatalf("预期错误包含 'AllowedFileCategories'，但得到: %v", err)
		}
	})
}

func TestValObjCheck(t *testing.T) {
	t.Parallel()

	// 测试用例：成功验证
	t.Run("Valid Config", func(t *testing.T) {
		t.Parallel()

		maxFileCount := 10
		singleChatMaxSelectFileCount := 3
		config := &TempZoneConfig{
			Name:                         "临时区",
			TmpFileUseType:               cdaenum.TmpFileUseTypeUpload,
			MaxFileCount:                 &maxFileCount,
			SingleChatMaxSelectFileCount: &singleChatMaxSelectFileCount,
			SingleFileSizeLimit:          50,
			SingleFileSizeLimitUnit:      cdaenum.MB,
			SupportDataType:              cdaenum.SupportDataTypes{"file"},
			AllowedFileCategories:        cdaenum.AllowedFileCategories{"document"},
		}

		err := config.ValObjCheck()
		if err != nil {
			t.Fatalf("预期验证通过，但得到错误: %v", err)
		}
	})

	// 测试用例：基本参数校验失败
	t.Run("Basic Validation Fails", func(t *testing.T) {
		t.Parallel()

		maxFileCount := 51 // 超过最大值50
		singleChatMaxSelectFileCount := 3
		config := &TempZoneConfig{
			Name:                         "临时区",
			TmpFileUseType:               cdaenum.TmpFileUseTypeUpload,
			MaxFileCount:                 &maxFileCount,
			SingleChatMaxSelectFileCount: &singleChatMaxSelectFileCount,
			SingleFileSizeLimit:          50,
			SingleFileSizeLimitUnit:      cdaenum.MB,
			SupportDataType:              cdaenum.SupportDataTypes{"file"},
			AllowedFileCategories:        cdaenum.AllowedFileCategories{"document"},
		}

		err := config.ValObjCheck()
		if err == nil {
			t.Fatal("预期验证失败，但没有得到错误")
		}

		// 修正：检查实际的错误消息内容
		if !strings.Contains(err.Error(), "max_file_count must be between 1 and 50") {
			t.Fatalf("预期错误包含 'max_file_count must be between 1 and 50'，但得到: %v", err)
		}
	})

	// 测试用例：临时文件使用类型无效
	t.Run("Invalid TmpFileUseType", func(t *testing.T) {
		t.Parallel()

		maxFileCount := 10
		singleChatMaxSelectFileCount := 3
		config := &TempZoneConfig{
			Name:                         "临时区",
			TmpFileUseType:               cdaenum.TmpFileUseType("invalid_type"), // 无效的枚举值
			MaxFileCount:                 &maxFileCount,
			SingleChatMaxSelectFileCount: &singleChatMaxSelectFileCount,
			SingleFileSizeLimit:          50,
			SingleFileSizeLimitUnit:      cdaenum.MB,
			SupportDataType:              cdaenum.SupportDataTypes{"file"},
			AllowedFileCategories:        cdaenum.AllowedFileCategories{"document"},
		}

		err := config.ValObjCheck()
		if err == nil {
			t.Fatal("预期验证失败，但没有得到错误")
		}

		if !strings.Contains(err.Error(), "tmp_file_use_type") {
			t.Fatalf("预期错误包含 'tmp_file_use_type'，但得到: %v", err)
		}
	})

	// 测试用例：支持的数据类型无效
	t.Run("Invalid SupportDataType", func(t *testing.T) {
		t.Parallel()

		maxFileCount := 10
		singleChatMaxSelectFileCount := 3
		config := &TempZoneConfig{
			Name:                         "临时区",
			TmpFileUseType:               cdaenum.TmpFileUseTypeUpload,
			MaxFileCount:                 &maxFileCount,
			SingleChatMaxSelectFileCount: &singleChatMaxSelectFileCount,
			SingleFileSizeLimit:          50,
			SingleFileSizeLimitUnit:      cdaenum.MB,
			SupportDataType:              cdaenum.SupportDataTypes{"invalid_type"}, // 无效的枚举值
			AllowedFileCategories:        cdaenum.AllowedFileCategories{"document"},
		}

		err := config.ValObjCheck()
		if err == nil {
			t.Fatal("预期验证失败，但没有得到错误")
		}

		if !strings.Contains(err.Error(), "support_data_type") {
			t.Fatalf("预期错误包含 'support_data_type'，但得到: %v", err)
		}
	})

	// 测试用例：允许的文件类别无效
	t.Run("Invalid AllowedFileCategories", func(t *testing.T) {
		t.Parallel()

		maxFileCount := 10
		singleChatMaxSelectFileCount := 3
		config := &TempZoneConfig{
			Name:                         "临时区",
			TmpFileUseType:               cdaenum.TmpFileUseTypeUpload,
			MaxFileCount:                 &maxFileCount,
			SingleChatMaxSelectFileCount: &singleChatMaxSelectFileCount,
			SingleFileSizeLimit:          50,
			SingleFileSizeLimitUnit:      cdaenum.MB,
			SupportDataType:              cdaenum.SupportDataTypes{"file"},
			AllowedFileCategories:        cdaenum.AllowedFileCategories{"invalid_category"}, // 无效的枚举值
		}

		err := config.ValObjCheck()
		if err == nil {
			t.Fatal("预期验证失败，但没有得到错误")
		}

		if !strings.Contains(err.Error(), "allowed_file_categories") {
			t.Fatalf("预期错误包含 'allowed_file_categories'，但得到: %v", err)
		}
	})

	// 测试用例：单文件大小限制单位无效
	t.Run("Invalid SingleFileSizeLimitUnit", func(t *testing.T) {
		t.Parallel()

		maxFileCount := 10
		singleChatMaxSelectFileCount := 3
		config := &TempZoneConfig{
			Name:                         "临时区",
			TmpFileUseType:               cdaenum.TmpFileUseTypeUpload,
			MaxFileCount:                 &maxFileCount,
			SingleChatMaxSelectFileCount: &singleChatMaxSelectFileCount,
			SingleFileSizeLimit:          50,
			SingleFileSizeLimitUnit:      cdaenum.BitUnit("invalid_unit"), // 无效的枚举值
			SupportDataType:              cdaenum.SupportDataTypes{"file"},
			AllowedFileCategories:        cdaenum.AllowedFileCategories{"document"},
		}

		err := config.ValObjCheck()
		if err == nil {
			t.Fatal("预期验证失败，但没有得到错误")
		}

		if !strings.Contains(err.Error(), "single_file_size_limit_unit") {
			t.Fatalf("预期错误包含 'single_file_size_limit_unit'，但得到: %v", err)
		}
	})

	// 测试用例：单文件大小超出最大限制（100MB）
	t.Run("SingleFileSizeLimit Exceeds Maximum", func(t *testing.T) {
		t.Parallel()

		maxFileCount := 10
		singleChatMaxSelectFileCount := 3
		config := &TempZoneConfig{
			Name:                         "临时区",
			TmpFileUseType:               cdaenum.TmpFileUseTypeUpload,
			MaxFileCount:                 &maxFileCount,
			SingleChatMaxSelectFileCount: &singleChatMaxSelectFileCount,
			SingleFileSizeLimit:          200, // 超过最大限制100MB
			SingleFileSizeLimitUnit:      cdaenum.MB,
			SupportDataType:              cdaenum.SupportDataTypes{"file"},
			AllowedFileCategories:        cdaenum.AllowedFileCategories{"document"},
		}

		err := config.ValObjCheck()
		if err == nil {
			t.Fatal("预期验证失败，但没有得到错误")
		}

		if !strings.Contains(err.Error(), "exceeds maximum allowed") {
			t.Fatalf("预期错误包含 'exceeds maximum allowed'，但得到: %v", err)
		}
	})
}

func TestTempZoneConfig_GetErrMsgMap(t *testing.T) {
	t.Parallel()

	config := &TempZoneConfig{}
	msgMap := config.GetErrMsgMap()

	tests := []struct {
		name string
		key  string
		want string
	}{
		{
			name: "TmpFileUseType.required",
			key:  "TmpFileUseType.required",
			want: `"tmp_file_use_type"不能为空`,
		},
		{
			name: "SingleFileSizeLimit.required",
			key:  "SingleFileSizeLimit.required",
			want: `"single_file_size_limit"不能为空`,
		},
		{
			name: "SingleFileSizeLimitUnit.required",
			key:  "SingleFileSizeLimitUnit.required",
			want: `"single_file_size_limit_unit"不能为空`,
		},
		{
			name: "SupportDataType.required",
			key:  "SupportDataType.required",
			want: `"support_data_type"不能为空`,
		},
		{
			name: "AllowedFileCategories.required",
			key:  "AllowedFileCategories.required",
			want: `"allowed_file_categories"不能为空`,
		},
		{
			name: "不存在的key",
			key:  "Unknown.key",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := msgMap[tt.key]
			if got != tt.want {
				t.Errorf("GetErrMsgMap()[%q] = %q, want %q", tt.key, got, tt.want)
			}
		})
	}
}

func TestTempZoneConfig_GenAllowedFileTypes_Valid(t *testing.T) {
	t.Parallel()

	config := &TempZoneConfig{
		AllowedFileCategories: cdaenum.AllowedFileCategories{"document"},
	}

	err := config.GenAllowedFileTypes()
	if err != nil {
		t.Fatalf("GenAllowedFileTypes failed: %v", err)
	}

	if config.AllowedFileTypes == nil {
		t.Fatal("AllowedFileTypes should not be nil")
	}

	if len(config.AllowedFileTypes) == 0 {
		t.Fatal("AllowedFileTypes should not be empty")
	}
}

func TestTempZoneConfig_GenAllowedFileTypes_NilCategories(t *testing.T) {
	t.Parallel()

	config := &TempZoneConfig{
		AllowedFileCategories: nil,
	}

	err := config.GenAllowedFileTypes()
	if err != nil {
		t.Fatalf("GenAllowedFileTypes failed: %v", err)
	}

	if config.AllowedFileTypes != nil {
		t.Fatal("AllowedFileTypes should be nil when categories is nil")
	}
}

func TestTempZoneConfig_GenAllowedFileTypes_MultipleCategories(t *testing.T) {
	t.Parallel()

	config := &TempZoneConfig{
		AllowedFileCategories: cdaenum.AllowedFileCategories{"document", "pdf", "text"},
	}

	err := config.GenAllowedFileTypes()
	if err != nil {
		t.Fatalf("GenAllowedFileTypes failed: %v", err)
	}

	if config.AllowedFileTypes == nil {
		t.Fatal("AllowedFileTypes should not be nil")
	}

	if len(config.AllowedFileTypes) == 0 {
		t.Fatal("AllowedFileTypes should not be empty")
	}
}

func TestTempZoneConfig_GenAllowedFileTypes_InvalidCategory(t *testing.T) {
	t.Parallel()

	config := &TempZoneConfig{
		AllowedFileCategories: cdaenum.AllowedFileCategories{"invalid_category"},
	}

	err := config.GenAllowedFileTypes()
	// Should return error for invalid category
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "allowed_file_categories is invalid")
}

// Test ValObjCheck with SingleChatMaxSelectFileCount range errors
func TestValObjCheck_SingleChatMaxSelectFileCountRange(t *testing.T) {
	t.Parallel()

	maxFileCount := 10

	t.Run("SingleChatMaxSelectFileCount too low", func(t *testing.T) {
		t.Parallel()

		singleChatMaxSelectFileCount := 0 // Below minimum of 1
		config := &TempZoneConfig{
			Name:                         "临时区",
			TmpFileUseType:               cdaenum.TmpFileUseTypeUpload,
			MaxFileCount:                 &maxFileCount,
			SingleChatMaxSelectFileCount: &singleChatMaxSelectFileCount,
			SingleFileSizeLimit:          50,
			SingleFileSizeLimitUnit:      cdaenum.MB,
			SupportDataType:              cdaenum.SupportDataTypes{"file"},
			AllowedFileCategories:        cdaenum.AllowedFileCategories{"document"},
		}

		err := config.ValObjCheck()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "single_chat_max_select_file_count must be between 1 and 5")
	})

	t.Run("SingleChatMaxSelectFileCount too high", func(t *testing.T) {
		t.Parallel()

		singleChatMaxSelectFileCount := 6 // Above maximum of 5
		config := &TempZoneConfig{
			Name:                         "临时区",
			TmpFileUseType:               cdaenum.TmpFileUseTypeUpload,
			MaxFileCount:                 &maxFileCount,
			SingleChatMaxSelectFileCount: &singleChatMaxSelectFileCount,
			SingleFileSizeLimit:          50,
			SingleFileSizeLimitUnit:      cdaenum.MB,
			SupportDataType:              cdaenum.SupportDataTypes{"file"},
			AllowedFileCategories:        cdaenum.AllowedFileCategories{"document"},
		}

		err := config.ValObjCheck()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "single_chat_max_select_file_count must be between 1 and 5")
	})
}

// Test ValObjCheck with MaxFileCount range errors
func TestValObjCheck_MaxFileCountRange(t *testing.T) {
	t.Parallel()

	singleChatMaxSelectFileCount := 3

	t.Run("MaxFileCount too low", func(t *testing.T) {
		t.Parallel()

		maxFileCount := 0 // Below minimum of 1
		config := &TempZoneConfig{
			Name:                         "临时区",
			TmpFileUseType:               cdaenum.TmpFileUseTypeUpload,
			MaxFileCount:                 &maxFileCount,
			SingleChatMaxSelectFileCount: &singleChatMaxSelectFileCount,
			SingleFileSizeLimit:          50,
			SingleFileSizeLimitUnit:      cdaenum.MB,
			SupportDataType:              cdaenum.SupportDataTypes{"file"},
			AllowedFileCategories:        cdaenum.AllowedFileCategories{"document"},
		}

		err := config.ValObjCheck()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "max_file_count must be between 1 and 50")
	})

	t.Run("MaxFileCount too high", func(t *testing.T) {
		t.Parallel()

		maxFileCount := 51 // Above maximum of 50
		config := &TempZoneConfig{
			Name:                         "临时区",
			TmpFileUseType:               cdaenum.TmpFileUseTypeUpload,
			MaxFileCount:                 &maxFileCount,
			SingleChatMaxSelectFileCount: &singleChatMaxSelectFileCount,
			SingleFileSizeLimit:          50,
			SingleFileSizeLimitUnit:      cdaenum.MB,
			SupportDataType:              cdaenum.SupportDataTypes{"file"},
			AllowedFileCategories:        cdaenum.AllowedFileCategories{"document"},
		}

		err := config.ValObjCheck()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "max_file_count must be between 1 and 50")
	})
}

// Test ValObjCheck with nil MaxFileCount and SingleChatMaxSelectFileCount
func TestValObjCheck_NilOptionalFields(t *testing.T) {
	t.Parallel()

	config := &TempZoneConfig{
		Name:                         "临时区",
		TmpFileUseType:               cdaenum.TmpFileUseTypeUpload,
		MaxFileCount:                 nil, // nil should be allowed
		SingleChatMaxSelectFileCount: nil, // nil should be allowed
		SingleFileSizeLimit:          50,
		SingleFileSizeLimitUnit:      cdaenum.MB,
		SupportDataType:              cdaenum.SupportDataTypes{"file"},
		AllowedFileCategories:        cdaenum.AllowedFileCategories{"document"},
	}

	err := config.ValObjCheck()
	assert.NoError(t, err)
}
