package icmp

// Logger 服务日志服务，可适配其他日志组件
//
//go:generate mockgen -source=./logger.go -destination ./cmpmock/logger_mock.go -package cmpmock
type Logger interface {
	Infof(format string, args ...interface{})
	Infoln(args ...interface{})
	Debugf(format string, args ...interface{})
	Debugln(args ...interface{})
	Errorf(format string, args ...interface{})
	Errorln(args ...interface{})
	Warnf(format string, args ...interface{})
	Warnln(args ...interface{})
	// Tracef(format string, args ...interface{})
	// Traceln(args ...interface{})
	Panicf(format string, args ...interface{})
	Panicln(args ...interface{})
	Fatalf(format string, args ...interface{})
	Fatalln(args ...interface{})
}
