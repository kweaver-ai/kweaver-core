package entity

// Log 服务日志转储信息
type Log struct {
	BaseInfo `yaml:",inline" json:",inline" bson:"inline"`
	OssID    string `yaml:"ossid,omitempty" json:"ossid,omitempty" bson:"ossid,omitempty"`
	Key      string `yaml:"key,omitempty" json:"key,omitempty" bson:"key,omitempty"`
	FileName string `yaml:"filename,omitempty" json:"filename,omitempty" bson:"filename,omitempty"`
}
