package cdaenum

// TempZoneStatus 是否启用临时区
type TempZoneStatus int

const (
	// TempZoneDisabled 禁用临时区
	TempZoneDisabled TempZoneStatus = 0

	// TempZoneEnabled 启用临时区
	TempZoneEnabled TempZoneStatus = 1
)
