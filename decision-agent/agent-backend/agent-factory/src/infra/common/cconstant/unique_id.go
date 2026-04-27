package cconstant

type UniqueIDFlag int

// todo 不太懂这东西的用处，先放着
const (
	UniqueIDFlagDB       UniqueIDFlag = 1 // UniqueIDFlagDB is the flag for unique id in database
	UniqueIDFlagRedisDlm UniqueIDFlag = 2 // UniqueIDFlagRedisDlm is the flag for unique value in redis dlm
)
