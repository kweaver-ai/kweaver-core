package main

import (

	// _ "net/http/pprof"

	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/spf13/cobra"
	_ "go.uber.org/automaxprocs"

	"flow-stream-data-pipeline/common"
	access "flow-stream-data-pipeline/pipeline-worker/drivenadapters"
	"flow-stream-data-pipeline/pipeline-worker/interfaces"
	"flow-stream-data-pipeline/pipeline-worker/logics"
)

var (
	server *logics.WorkerService

	//  pipeline-mgmt pass worker id and user id
	workerID    string
	accountID   string
	accountType string

	rootCmd = &cobra.Command{
		Use:   "start",
		Short: "start the worker",
		Long:  "start the worker",
		Run:   startFunc,
	}
)

func init() {
	rootCmd.Flags().StringVar(&workerID, "worker_id", "", "worker id")
	rootCmd.Flags().StringVar(&accountID, "account_id", "", "account id")
	rootCmd.Flags().StringVar(&accountType, "account_type", "", "account type")

	_ = rootCmd.MarkFlagRequired("worker_id")
	_ = rootCmd.MarkFlagRequired("account_id")
	_ = rootCmd.MarkFlagRequired("account_type")
}

func main() {
	_ = rootCmd.Execute()
}

// start application
func startFunc(cmd *cobra.Command, args []string) {

	logger.Info("Worker Server Starting")

	// 初始化服务配置
	appSetting := common.NewSetting()
	logger.Info("Worker Server Init Setting Success")

	// 设置错误码语言
	rest.SetLang(appSetting.ServerSetting.Language)
	logger.Info("Worker Server Set Language Success")

	logics.SetIndexMgmtAccess(access.NewIndexBaseAccess(appSetting))
	logics.SetMQAccess(access.NewMQAccess(appSetting))
	logics.SetOpensearchAccess(access.NewOpenSearchAccess(appSetting))
	logics.SetPipelineMgmtAccess(access.NewPipelineMgmt(appSetting))

	ctx := context.WithValue(context.Background(), rest.XLangKey, rest.DefaultLanguage)
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, interfaces.AccountInfo{
		ID:   accountID,
		Type: accountType,
	})

	// 监听中断信号（SIGINT、SIGTERM）
	go monitorSignal(ctx)

	// 开启 worker server
	server = logics.NewWorkerService(ctx, appSetting, workerID)
	server.Start(ctx)
}

// monitor the system signal for restful existing
func monitorSignal(ctx context.Context) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	for s := range c {
		switch s {
		case syscall.SIGINT, syscall.SIGTERM:
			logger.Infof("Worker Server detect exit signal: %v, exiting...", s)
			// 暂停 worker server
			server.Stop(ctx)
		default:
			logger.Infof("Worker Server detect unknown signal: %v", s)
		}
	}
}
