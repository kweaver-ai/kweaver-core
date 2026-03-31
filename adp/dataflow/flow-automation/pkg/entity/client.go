// Package entity 实体信息
package entity

// Client 客户端信息
type Client struct {
	BaseInfo     `yaml:",inline" json:",inline" bson:"inline"`
	ClientName   string `yaml:"client_name,omitempty" json:"client_name,omitempty" bson:"client_name,omitempty"`
	ClientID     string `yaml:"client_id,omitempty" json:"client_id,omitempty" bson:"client_id,omitempty"`
	ClientSecret string `yaml:"client_secret,omitempty" json:"client_secret,omitempty" bson:"client_secret,omitempty"`
}
