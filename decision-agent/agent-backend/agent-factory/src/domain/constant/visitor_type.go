package constant

type VisitorType string

const (
	RealName  VisitorType = "realname"  // 实名用户
	Anonymous VisitorType = "anonymous" // 匿名用户
	Business  VisitorType = "business"  // 应用账户
)

func (v VisitorType) String() string {
	return string(v)
}
