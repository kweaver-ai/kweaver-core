package entity

// OutBox out box 结构体
type OutBox struct {
	BaseInfo `yaml:",inline" json:",inline" bson:"inline"`
	Topic    string `yaml:"topic,omitempty" json:"topic,omitempty" bson:"topic,omitempty"`
	Msg      string `yaml:"msg,omitempty" json:"msg,omitempty" bson:"msg,omitempty"`
}

// OutBoxInput outbox查询入参结构体
type OutBoxInput struct {
	CreateTime int64
	Limit      int64
}
