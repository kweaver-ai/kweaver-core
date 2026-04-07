package entity

// Switch 开关状态
type Switch struct {
	BaseInfo `yaml:",inline" json:",inline" bson:"inline"`
	Name     string `yaml:"name,omitempty" json:"name,omitempty" bson:"name,omitempty"`
	Status   bool   `yaml:"status,omitempty" json:"status,omitempty" bson:"status,omitempty"`
}

// SwitchName 开关名称
const SwitchName = "automation"
