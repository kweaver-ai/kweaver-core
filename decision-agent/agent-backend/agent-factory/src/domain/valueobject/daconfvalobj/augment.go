package daconfvalobj

import "github.com/pkg/errors"

// Augment 表示query增强配置
type Augment struct {
	Enable     *bool              `json:"enable" binding:"required"`      // 是否启用
	DataSource *AugmentDataSource `json:"data_source" binding:"required"` // 图谱数据源配置
}

func (a *Augment) ValObjCheck() (err error) {
	// 检查Enable是否为空
	if a.Enable == nil {
		err = errors.New("enable is required")
		return
	}

	// 如果未启用增强功能，直接返回
	if !*a.Enable {
		return
	}

	// 如果启用了增强功能，检查DataSource是否为空
	if a.DataSource == nil {
		err = errors.New("data_source is required when enable is true")
		return
	}

	// 验证DataSource的有效性
	if err = a.DataSource.ValObjCheck(); err != nil {
		return
	}

	return
}

// AugmentDataSource 表示query增强的数据源配置
type AugmentDataSource struct {
	Kg []KgSource `json:"kg"` // 图谱类型数据源
}

func (ads *AugmentDataSource) ValObjCheck() error {
	// 检查Kg列表是否为空
	if len(ads.Kg) == 0 {
		return errors.New("kg is required")
	}

	// 验证每个图谱数据源的有效性
	for _, kg := range ads.Kg {
		if err := kg.ValObjCheck(); err != nil {
			return err
		}
	}

	return nil
}
