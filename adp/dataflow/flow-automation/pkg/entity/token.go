package entity

// Token token结构体
type Token struct {
	BaseInfo     `yaml:",inline" json:",inline" bson:"inline"`
	UserID       string `yaml:"userid,omitempty" json:"userid,omitempty" bson:"userid,omitempty"`
	UserName     string `yaml:"username,omitempty" json:"username,omitempty" bson:"username,omitempty"`
	RefreshToken string `yaml:"refresh_token,omitempty" json:"refresh_token,omitempty" bson:"refresh_token,omitempty"`
	Token        string `yaml:"token,omitempty" json:"token,omitempty" bson:"token,omitempty"`
	ExpiresIn    int    `yaml:"expires_in,omitempty" json:"expires_in,omitempty" bson:"expires_in,omitempty"`
	LoginIP      string `yaml:"login_ip,omitempty" json:"login_ip,omitempty" bson:"login_ip,omitempty"`
	IsApp        bool   `yaml:"isapp,omitempty" json:"isapp,omitempty" bson:"isapp,omitempty"`
}
