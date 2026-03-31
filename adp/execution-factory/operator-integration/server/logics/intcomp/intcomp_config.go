// Package intcomp internal component config
// @file intcomp_config.go
// @description: 内置组件配置操作
package intcomp

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/dbaccess"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/config"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/errors"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/validator"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces/model"
	validator10 "github.com/go-playground/validator/v10"
)

// IntCompConfigImpl internal component config impl
type intCompConfigImpl struct {
	IntCompDB model.IInternalComponentConfigDB
	Logger    interfaces.Logger
	Validator interfaces.Validator
}

var (
	iOnce sync.Once
	ic    interfaces.IIntCompConfigService
)

// NewIntCompConfigService new internal component config service
func NewIntCompConfigService() interfaces.IIntCompConfigService {
	iOnce.Do(func() {
		confLoader := config.NewConfigLoader()
		ic = &intCompConfigImpl{
			IntCompDB: dbaccess.NewInternalComponentConfigDBSingleton(),
			Logger:    confLoader.GetLogger(),
			Validator: validator.NewValidator(),
		}
	})
	return ic
}

// DeleteConfig 删除配置
func (i *intCompConfigImpl) DeleteConfig(ctx context.Context, tx *sql.Tx, configType, configID string) (err error) {
	// 检查是否存在
	exist, _, err := i.IntCompDB.SelectConfig(ctx, configType, configID)
	if err != nil {
		i.Logger.WithContext(ctx).Errorf("select config failed, err: %v", err)
		err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	if !exist {
		return
	}
	err = i.IntCompDB.DeleteConfig(ctx, tx, configType, configID)
	if err != nil {
		i.Logger.WithContext(ctx).Errorf("delete internal component config error: %v", err)
		err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
	}
	return
}

// UpdateConfig 更新或添加配置
func (i *intCompConfigImpl) UpdateConfig(ctx context.Context, tx *sql.Tx, config *interfaces.IntCompConfig) (err error) {
	// 检查配置是否存在
	exist, dbConfig, err := i.IntCompDB.SelectConfig(ctx, config.ComponentType.String(), config.ComponentID)
	if err != nil {
		i.Logger.WithContext(ctx).Errorf("select config failed, err: %v", err)
		err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	if exist {
		// 更新配置
		dbConfig.ConfigSource = config.ConfigSource.String()
		dbConfig.ConfigVersion = config.ConfigVersion
		dbConfig.ProtectedFlag = config.ProtectedFlag
		err = i.IntCompDB.UpdateConfig(ctx, tx, dbConfig)
	} else {
		// 添加配置
		err = i.IntCompDB.InsertConfig(ctx, tx, &model.InternalComponentConfigDB{
			ComponentType: config.ComponentType.String(),
			ComponentID:   config.ComponentID,
			ConfigVersion: config.ConfigVersion,
			ConfigSource:  config.ConfigSource.String(),
			ProtectedFlag: config.ProtectedFlag,
		})
	}
	if err != nil {
		i.Logger.WithContext(ctx).Errorf("update internal component config failed, err: %v", err)
		err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
	}
	return
}

// CompareConfig 比较当前配置和待检查的配置，返回结果
func (i *intCompConfigImpl) CompareConfig(ctx context.Context, check *interfaces.IntCompConfig) (action interfaces.IntCompConfigAction, err error) {
	// 检查请求参数是否合法
	err = validator10.New().Struct(check)
	if err != nil {
		return
	}
	err = i.Validator.ValidatorIntCompVersion(ctx, check.ConfigVersion)
	if err != nil {
		return
	}
	// 检查组件配置是否存在
	exist, current, err := i.IntCompDB.SelectConfig(ctx, check.ComponentType.String(), check.ComponentID)
	if err != nil {
		i.Logger.WithContext(ctx).Errorf("select config failed, err: %v", err)
		err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	if !exist {
		action = interfaces.IntCompConfigActionTypeUpdate
		return
	}
	// 检查版本是否一致
	if current.ConfigVersion != check.ConfigVersion {
		// 版本比较
		var result int
		result, err = compareVersions(check.ConfigVersion, current.ConfigVersion)
		if err != nil {
			i.Logger.WithContext(ctx).Errorf("compare versions failed, err: %v", err)
			err = errors.DefaultHTTPError(ctx, http.StatusBadRequest, err.Error())
			return
		}
		if result > 0 { // check.ConfigVersion > current.ConfigVersion
			action = interfaces.IntCompConfigActionTypeUpdate
		} else { // check.ConfigVersion < current.ConfigVersion
			action = interfaces.IntCompConfigActionTypeSkip
		}
		return
	}
	// 当版本一致时
	switch check.ConfigSource {
	case interfaces.ConfigSourceTypeManual: // 手动配置，直接更新
		action = interfaces.IntCompConfigActionTypeUpdate
	case interfaces.ConfigSourceTypeAuto: // 自动配置，检查是否需要更新
		if current.ConfigSource == interfaces.ConfigSourceTypeManual.String() && current.ProtectedFlag { // 保护锁，不更新
			action = interfaces.IntCompConfigActionTypeSkip
			return
		}
		// 不保护，需要更新
		action = interfaces.IntCompConfigActionTypeUpdate
	default: // 未知配置，不更新
		action = interfaces.IntCompConfigActionTypeSkip
	}
	return
}

// compareVersions 比较两个语义化版本号
// 返回：-1表示v1 < v2，0表示相等，1表示v1 > v2
func compareVersions(v1, v2 string) (result int, err error) {
	// 将版本号拆分为数字组件
	parts1, err := splitVersion(v1)
	if err != nil {
		return 0, err
	}
	parts2, err := splitVersion(v2)
	if err != nil {
		return 0, err
	}
	// 比较每个数字组件
	maxLen := max(len(parts1), len(parts2))
	for i := 0; i < maxLen; i++ {
		num1 := getPart(parts1, i)
		num2 := getPart(parts2, i)

		if num1 < num2 {
			return -1, nil
		}
		if num1 > num2 {
			return 1, nil
		}
	}
	return 0, nil
}

// splitVersion 将版本字符串拆分为数字数组
func splitVersion(v string) ([]int, error) {
	parts := strings.Split(v, ".")
	result := make([]int, len(parts))

	for i, part := range parts {
		num, err := strconv.Atoi(part)
		if err != nil {
			return nil, fmt.Errorf("invalid version format: %s", v)
		}
		result[i] = num
	}
	return result, nil
}

// getPart 安全获取版本号组件
func getPart(parts []int, index int) int {
	if index < len(parts) {
		return parts[index]
	}
	return 0
}
