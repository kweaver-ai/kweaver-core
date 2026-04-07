package entity

import "github.com/kweaver-ai/adp/autoflow/flow-automation/common"

// InBox 消息队列
type InBox struct {
	BaseInfo `yaml:",inline" json:",inline" bson:"inline"`
	Msg      common.DocMsg `yaml:"msg,omitempty" json:"msg,omitempty" bson:"msg,omitempty"`
	Topic    string        `yaml:"topic,omitempty" json:"topic,omitempty" bson:"topic,omitempty"`
	DocID    string        `yaml:"docid,omitempty" json:"docid,omitempty" bson:"docid,omitempty"`
	Dags     []string      `yaml:"dag,omitempty" json:"dag,omitempty" bson:"dag,omitempty"`
}
