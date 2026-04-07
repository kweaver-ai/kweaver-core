package log

import (
	"log"
	"os"
	"sync"
	"time"

	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_log"
	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/config"
	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/public"
	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/resource"
	"github.com/kweaver-ai/TelemetrySDK-Go/span/v2/encoder"
	"github.com/kweaver-ai/TelemetrySDK-Go/span/v2/exporter"
	"github.com/kweaver-ai/TelemetrySDK-Go/span/v2/field"
	spanLog "github.com/kweaver-ai/TelemetrySDK-Go/span/v2/log"
	"github.com/kweaver-ai/TelemetrySDK-Go/span/v2/open_standard"
	"github.com/kweaver-ai/TelemetrySDK-Go/span/v2/runtime"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/common"
)

var (
	fo     spanLog.Logger
	foOnce sync.Once
)

// InitFlowO11yLogger 初始化flow_o11y日志记录器
func InitFlowO11yLogger(config *common.TelemetryConf) {
	foOnce.Do(func() {
		// ARLogger 程序日志记录器，使用异步发送模式，无返回值。
		//spanLog.AllLevel
		logLevel := config.GetLogLevel()
		var ARLogger = spanLog.NewSamplerLogger(spanLog.WithSample(1.0), spanLog.WithLevel(logLevel))
		// 初始化ar_log
		hostName, _ := os.Hostname()
		public.SetServiceInfo(config.ServerName, config.ServerVersion, hostName)
		exporters := initExporters(config.Exporters)

		systemLogWriter := open_standard.OpenTelemetryWriter(
			encoder.NewJsonEncoderWithExporters(exporters...),
			resource.LogResource())
		systemLogRunner := runtime.NewRuntime(systemLogWriter, field.NewSpanFromPool)
		systemLogRunner.SetUploadInternalAndMaxLog(3*time.Second, 10)
		// 运行SystemLogger日志器。
		go systemLogRunner.Run()

		ARLogger.SetLevel(logLevel)
		ARLogger.SetRuntime(systemLogRunner)
		fo = ARLogger
	})
}

// NewFlowO11yLogger 创建一个flow_o11y日志记录器
func NewFlowO11yLogger() spanLog.Logger {
	return fo
}

// CloseFlowO11yLogger 关闭flow_o11y日志记录器
func CloseFlowO11yLogger() {
	if fo != nil {
		fo.Close()
	}
}

// initExporters 拷贝自TelemetrySDK初始化函数
func initExporters(logConfig *config.ExportersTypConfig) []exporter.LogExporter {
	//初始化默认silent
	var systemLogExporters []exporter.LogExporter
	systemLogExporters = append(systemLogExporters, ar_log.NewExporter(public.NewSilentClient()))

	if logConfig.ConsoleExporter != nil && logConfig.ConsoleExporter.Enable {
		systemLogExporters = append(systemLogExporters, initConsoleExporter(logConfig.ConsoleExporter))
	}

	if logConfig.HttpExporters != nil && logConfig.HttpExporters.Enable {
		systemLogExporters = append(systemLogExporters, initHttpExporter(logConfig.HttpExporters))
	}
	if logConfig.ProtonMqExporters != nil && logConfig.ProtonMqExporters.Enable {
		systemLogExporters = append(systemLogExporters, initProtonMqExporter(logConfig.ProtonMqExporters))
	}

	return systemLogExporters
}

// initHttpExportersClient 2024-03-25 最新版本的配置
func initHttpExporter(config *config.HttpExporterTyp) exporter.LogExporter {
	// 设置日志通过HTTP上报
	var logEndpoint = config.Config.Endpoint
	systemLogClient := public.NewHTTPClient(public.WithAnyRobotURL(logEndpoint),
		public.WithCompression(1),
		public.WithTimeout(10*time.Second),
		public.WithRetry(true, 5*time.Second, 20*time.Second, 1*time.Minute))
	return ar_log.NewExporter(systemLogClient)
}

// initProtonMqExporter 初始化protonmq输出
func initProtonMqExporter(config *config.ProtonMqExporterTyp) exporter.LogExporter {
	systemLogClient, err := public.NewProtonMqClient(config)
	if err != nil {
		log.Fatalf("%+v", err)
	}
	return ar_log.NewExporter(systemLogClient)
}

// initConsoleExporter 初始化console输出
func initConsoleExporter(config *config.ConsoleExporterTyp) exporter.LogExporter {
	return exporter.GetRealTimeExporter()
}
