// Package initial 初始化服务
package initial

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/bytedance/gopkg/util/logger"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/common"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/drivenadapters"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/driveradapters/middleware"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/keeper"
	cdb "github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/db"
	ierr "github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/errors"
	i18n "github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/i18n"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/lock"
	commonLog "github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/log"
	cmq "github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/mq"
	store "github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/store"
	telemetryvar "github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/telemetry/common"
	traceLog "github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/telemetry/log"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/actions"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/entity"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/event"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/mod"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/utils"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/utils/data"
	mongoStore "github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/store/mongo"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/store/rds"
	dagmodel "github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/store/rds/dag"
	msqclient "github.com/kweaver-ai/proton-mq-sdk-go"
	"github.com/shiningrush/goevent"
	"gopkg.in/yaml.v3"
)

// DocLibPermMsg 文档库权限分配消息体
type DocLibPermMsg struct {
	AppID      string              `json:"app_id"`
	DocLibType string              `json:"doc_lib_type"`
	Expires    string              `json:"expires_at"`
	Perm       map[string][]string `json:"perm"`
}

// DocPermMsg 文档权限分配结构体
type DocPermMsg struct {
	AppID   string              `json:"app_id"`
	DocID   string              `json:"doc_id"`
	Expires string              `json:"expires_at"`
	Perm    map[string][]string `json:"perm"`
}

const (
	letters              string = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ~-_."
	applyDocLibPermTopic string = "core.doc_share.doc_lib_perm.app.set"
	applyDocPermTopic    string = "core.doc_share.doc_perm.app.set"
)

var (
	closers         []mod.Closer
	applyDocLibPerm = []string{"read", "create", "modify"}
	docLibPermValue = 1
	docPermValue    = 61
	applyDocPerm    = []string{"display", "download", "modify", "create", "delete"}
)

// RegisterAction you need register all used action to it
func RegisterAction(acts []entity.Action) {
	for i := range acts {
		mod.ActionMap[acts[i].Name()] = acts[i]
	}
}

// GetAction 获取action
func GetAction(name string) (entity.Action, bool) {
	act, ok := mod.ActionMap[name]
	return act, ok
}

// InitialOption 初始化信息结构体
type InitialOption struct { //nolint
	Keeper mod.Keeper
	Store  mod.Store

	// ParserWorkersCnt default 20
	ParserWorkersCnt int
	// LowestExecutorWorkerCnt default 200
	LowestExecutorWorkerCnt int
	// LowExecutorWorkerCnt default 200
	LowExecutorWorkerCnt int
	// MediumExecutorWorkerCnt default 200
	MediumExecutorWorkerCnt int
	// HighExecutorWorkerCnt default 200
	HighExecutorWorkerCnt int
	// HighestExecutorWorkerCnt default 400
	HighestExecutorWorkerCnt int
	// ExecutorTimeout default 60s
	ExecutorTimeout time.Duration
	// DagScheduleTimeout default 15s
	DagScheduleTimeout time.Duration
	// DagRunningTimeout default 1d
	DagRunningTimeout time.Duration
	// ListInsCount default 500
	ListInsCount int

	// Read dag define from directory
	// each file will be pared to a dag, so you CAN'T define all dag in one file
	ReadDagFromDir string
}

// Start will block until accept system signal, if you don't want block, plz check "Init"
func Start(opt *InitialOption, afterInit ...func() error) error {
	if err := Init(opt); err != nil {
		return err
	}
	for i := range afterInit {
		if err := afterInit[i](); err != nil {
			return err
		}
	}

	log.Println("automation start success")
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT) //nolint
	sig := <-c
	log.Printf("get sig: %s, ready to close component", sig)
	Close()
	log.Println("close completed")
	return nil
}

// Init will not block, but you need to close automation after application closing
func Init(opt *InitialOption) error {
	config := common.NewConfig()

	initRds(config)
	if err := checkOption(opt); err != nil {
		return err
	}

	logout := os.Getenv("LOGOUT")
	logDir := "/var/log/contentAutoMation/"
	logName := "contentAutoMation.log"
	commonLog.InitLogger(logout, logDir, logName)
	// 替换可观测性日志上报索引库
	// traceLog.InitLogger(common.ServiceName)
	traceLog.InitARLog(&config.Telemetry)
	i18n.InitI18nTranslator(common.MultiResourcePath)
	ierr.InitServiceName(common.ErrCodeServiceName)
	middleware.SetMiddleware()
	store.InitRedis(&store.RedisConfiguration{
		Ctx:              context.Background(),
		Host:             config.Redis.Host,
		Port:             config.Redis.Port,
		SlaveHost:        config.Redis.SlaveHost,
		SlavePort:        config.Redis.SlavePort,
		UserName:         config.Redis.UserName,
		Password:         config.Redis.Password,
		SentinelPassword: config.Redis.SentinelPassword,
		SentinelUsername: config.Redis.SentinelUsername,
		MasterGroupName:  config.Redis.MasterGroupName,
		ClusterMode:      config.Redis.ClusterMode,
	})

	initKeeper(opt)
	initCluster(config)
	rds.InitSingleton()
	initMQClient(config)
	initFlowO11yLogger()

	initCommonComponent(opt)
	initLeaderChangedHandler(opt)

	initInternalAccount(config)
	// initAnyData(context.Background(), config)

	RegisterAction([]entity.Action{
		&actions.ManualTrigger{},
		&actions.FormTrigger{},
		&actions.CronTrigger{},
		&actions.SandboxExecute{},
		&actions.CronWeekTrigger{},
		&actions.CronCustomTrigger{},
		&actions.CronMonthTrigger{},
		&actions.AnyShareDirCopy{},
		&actions.AnyShareDirRemove{},
		&actions.AnyShareDirMove{},
		&actions.LogicBranch{},
		&actions.ParallelBranch{},
		&actions.AnyShareDirCreate{},
		&actions.AnyShareDirPath{},
		&actions.AnyShareDirRename{},
		&actions.AnyShareDirTag{},
		&actions.AnyShareFileCopy{},
		&actions.AnyShareFileMove{},
		&actions.AnyShareFilePath{},
		&actions.AnyShareFileRemove{},
		&actions.AnyShareFileRename{},
		&actions.AnyShareFileTag{},
		&actions.AnyshareFileCopyTrigger{},
		&actions.AnyshareFileMoveTrigger{},
		&actions.AnyshareFileRemoveTrigger{},
		&actions.AnyshareFileUploadTrigger{},
		&actions.AnyshareFileCopyTrigger{},
		&actions.AnyshareFolderCreateTrigger{},
		&actions.AnyshareFolderCopyTrigger{},
		&actions.AnyshareFolderCreateTrigger{},
		&actions.AnyshareFolderMoveTrigger{},
		&actions.AnyshareFolderRemoveTrigger{},
		&actions.AnyshareFileReversionTrigger{},
		&actions.AnyshareFileRenameTrigger{},
		&actions.AnyshareFileRestoreTrigger{},
		&actions.AnyshareUserCreateTrigger{},
		&actions.AnyshareUserDeleteTrigger{},
		&actions.AnyshareUserFreezeTrigger{},
		&actions.AnyshareOrgNameModifyTrigger{},
		&actions.AnyshareUserMovedTrigger{},
		&actions.AnyshareUserAddDeptTrigger{},
		&actions.AnyshareUserRemoveDeptTrigger{},
		&actions.AnyshareDeptCreateTrigger{},
		&actions.AnyshareDeptDeleteTrigger{},
		&actions.AnyshareDeptMovedTrigger{},
		&actions.AnyshareUserChangeTrigger{},
		&actions.AnyshareTagTreeCreateTrigger{},
		&actions.AnyshareTagTreeAddedTrigger{},
		&actions.AnyshareTagTreeEditedTrigger{},
		&actions.AnyshareTagTreeDeletedTrigger{},
		&actions.TextJoin{},
		&actions.TextSplit{},
		&actions.TextMatch{},
		&actions.AnyShareFileMatchContent{},
		&actions.AnyShareFileSetCsfLevel{},
		&actions.AnyShareDirSetTemplate{},
		&actions.AnyShareFileSetTemplate{},
		&actions.AnyShareFileGetPage{},
		&actions.PyExe{},
		&actions.OCR{},
		&actions.WorkflowAsyncTask{},
		&actions.AnyshareFileSetPerm{},
		&actions.IntelliinfoTranfer{},
		&actions.SecurityPolicyTrigger{},
		&actions.Return{},
		&actions.AnyShareDocRemove{},
		&actions.AnyShareDocRename{},
		&actions.AnyShareDocSetCsfLevel{},
		&actions.AnyShareDocSetPerm{},
		&actions.AnyShareDocSetTemplate{},
		&actions.AnyShareDocTag{},
		&actions.AnyShareDocGetPath{},
		&actions.TimeNow{},
		&actions.TimeRelative{},
		&actions.AnyshareCreateFile{},
		&actions.AnyshareExcelFileUpdate{},
		&actions.AnyshareDocxFileUpdate{},
		&actions.EleInvoice{},
		&actions.IDCard{},
		&actions.AudioTransfer{},
		&actions.DocSummarize{},
		&actions.MeetSummarize{},
		&actions.AnyShareSelectedFileTrigger{},
		&actions.AnyShareSelectedFolderTrigger{},
		&actions.AnyShareFileRelevance{},
		&actions.AnyShareFolderRelevance{},
		&actions.AnyShareFileGetByName{},
		&actions.AnyShareDocSetAllowSuffixDoc{},
		&actions.AnyShareDocSetSpaceQuota{},
		&actions.CustomExtract{},
		&actions.AnyshareFolderSetPerm{},
		&actions.AnyshareDocLibQuotaScale{},
		&actions.AnyShareFileStat{},
		&actions.AnyShareDirStat{},
		&actions.JsonGet{},
		&actions.JsonSet{},
		&actions.JsonTemplate{},
		&actions.JsonParse{},
		&actions.ContentAbstract{},
		&actions.ContentFullText{},
		&actions.ContentEntity{},
		&actions.LLMChatCompletion{},
		&actions.DataFlowDocTrigger{},
		&actions.DataFlowUserTrigger{},
		&actions.DataFlowDeptTrigger{},
		&actions.DataFlowTagTrigger{},
		&actions.CallAgent{},
		&actions.EcoconfigReindex{},
		&actions.Assign{},
		&actions.Define{},
		&actions.Loop{},
		&actions.MDLDataViewTrigger{},
		&actions.OpenSearchBulkUpsert{},
		&actions.ArrayFilter{},
		&actions.DatabaseWrite{},
		&actions.LLMEmbedding{},
		&actions.LLMReranker{},
		&actions.ContentPipelineFullText{},
		&actions.ContentPipelineDocFormatConvert{},
		&actions.ContentFileParse{},
		&actions.OCRNew{},
		&actions.AnyshareFileUpdate{},
		&actions.SubFlowCallWorkflow{},
		&actions.SubFlowCallDataflow{},
		&actions.DatasetWriteDocs{},
	})

	if opt.ReadDagFromDir != "" {
		return readDagFromDir(opt.ReadDagFromDir)
	}
	return nil
}

func initInternalAccount(config *common.Config) { //nolint
	retryCount := 0
	retryMaxCount := 9
	retryWaitTime := 10 * time.Second
	userMgntAdapters := drivenadapters.NewUserManagement()
	authAdapters := drivenadapters.NewAuthentication()
	logger := commonLog.NewLogger()
	mqAdapters := cmq.NewMQClient()
	rdsClientLock := "flow-automation:client"
	lockClient := lock.NewDistributeLock(store.NewRedis(), rdsClientLock, config.OAuth.ClientName)
	for {
		if retryCount > retryMaxCount {
			return
		}
		store := mod.GetStore()
		clientInfo, err := store.GetClient(config.OAuth.ClientName)
		if err != nil {
			commonLog.NewLogger().Warnln("get client failed: %s", err.Error())
			time.Sleep(retryWaitTime)
			retryCount++
			continue
		}

		if clientInfo.ClientID != "" && clientInfo.ClientSecret != "" {
			name, uerr := userMgntAdapters.QueryInternalAccount(clientInfo.ClientID)
			if uerr != nil {
				commonLog.NewLogger().Warnln("query internal app failed: %s", uerr.Error())
				time.Sleep(retryWaitTime)
				retryCount++
				continue
			}

			if name == clientInfo.ClientName {
				// 向内部账号配置获取任意用户访问令牌的权限
				err = authAdapters.ConfigAuthPerm(clientInfo.ClientID)
				if err != nil {
					commonLog.NewLogger().Errorf("config app perm failed: %s", err.Error())
					time.Sleep(retryWaitTime)
					retryCount++
					continue
				}
				// 内部账户配置权限
				docLibMessage, _ := json.Marshal(DocLibPermMsg{
					AppID:      clientInfo.ClientID,
					DocLibType: "all_doc_lib",
					Expires:    "1970-01-01T08:00:00+08:00",
					Perm: map[string][]string{
						"allow": applyDocLibPerm,
					},
				})

				docMessage, _ := json.Marshal(DocPermMsg{
					DocID:   "gns://",
					AppID:   clientInfo.ClientID,
					Expires: "1970-01-01T08:00:00+08:00",
					Perm: map[string][]string{
						"deny":  {},
						"allow": applyDocPerm,
					},
				})
				for {
					err = mqAdapters.Publish(applyDocLibPermTopic, docLibMessage)
					if err != nil {
						logger.Errorf("Failed to configure document library permissions; Detail: {%s}", err)
						continue
					}

					err = mqAdapters.Publish(applyDocPermTopic, docMessage)
					if err != nil {
						logger.Errorf("Failed to configure document permissions; Detail: {%s}", err)
						continue
					}

					break
				}
				config.OAuth.ClientID = clientInfo.ClientID
				config.OAuth.ClientSecret = clientInfo.ClientSecret
				return
			}

			terr := store.RemoveClient(config.OAuth.ClientName)
			if terr != nil {
				logger.Warnln("remove client failed: %s", terr.Error())
			}

			time.Sleep(retryWaitTime)
			retryCount++
			continue
		}

		// 创建锁
		err = lockClient.Lock(context.Background(), 60*time.Second)

		if err == nil {
			name := config.OAuth.ClientName
			pwd := generatePassword()
			id, err := userMgntAdapters.RegisterInternalAccount(name, pwd)

			if err != nil {
				logger.Errorf("register internal app failed: %s", err.Error())
				time.Sleep(retryWaitTime)
				retryCount++
				continue
			}

			config.OAuth.ClientID = id
			config.OAuth.ClientSecret = pwd

			err = store.CreateClient(name, id, pwd)
			if err != nil {
				logger.Errorf("create client failed: %s", err.Error())
				time.Sleep(retryWaitTime)
				retryCount++
				continue
			}

			_, rErr := lockClient.Release()

			if rErr != nil {
				logger.Errorf("release lock failed, detail: %s", rErr.Error())
			}
			continue
		}
	}
}

func initCluster(config *common.Config) {
	// 如果配置了access address，直接使用配置的地址
	if config.AccessAddress.Host != "" {
		return
	}

	var addr drivenadapters.ClusterAccess

	addr, err := drivenadapters.NewAnyshare().ClusterAccess()
	if err != nil {
		// panic(err)
		logger.Errorf("get cluster access failed: %s", err.Error())
	}
	go func() {
		// 每60s 获取anyshare最新地址
		ticker := time.NewTicker(60 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			addr, derr := drivenadapters.NewAnyshare().ClusterAccess()
			if derr == nil {
				config.AccessAddress.Host = addr.Host
				config.AccessAddress.Port = addr.Port
				config.AccessAddress.Schema = "https"
				config.AccessAddress.Path = ""
			}
		}
	}()

	config.AccessAddress.Host = addr.Host
	config.AccessAddress.Port = addr.Port
	config.AccessAddress.Schema = "https"
	config.AccessAddress.Path = ""
}

func initLeaderChangedHandler(opt *InitialOption) {
	h := &LeaderChangedHandler{
		opt: opt,
	}
	if err := goevent.Subscribe(h); err != nil {
		log.Fatalln(err)
	}
	closers = append(closers, h)

	// when application init, leader election should completed, so we need trigger it
	if opt.Keeper.IsLeader() {
		h.Handle(context.Background(), &event.LeaderChanged{
			IsLeader:  true,
			WorkerKey: opt.Keeper.WorkerKey(),
		})
	}
}

// SetDagInstanceLifecycleHook set hook handler for automation
// IMPORTANT: you MUST set hook before you call Init or Start to avoid lost changes.
// (because component will work immediately after you call Init or Start)
func SetDagInstanceLifecycleHook(hook entity.DagInstanceLifecycleHook) {
	entity.HookDagInstance = hook
}

// LeaderChangedHandler used to handle leader chaged event
type LeaderChangedHandler struct {
	opt *InitialOption

	leaderCloser []mod.Closer
}

// Topic topic
func (l *LeaderChangedHandler) Topic() []string {
	return []string{event.KeyLeaderChanged}
}

// Handle handle
func (l *LeaderChangedHandler) Handle(cxt context.Context, e goevent.Event) {
	lcEvent := e.(*event.LeaderChanged)

	// changed to leader
	if lcEvent.IsLeader {
		for {
			if len(l.leaderCloser) == 0 {
				break
			}
			time.Sleep(time.Second)
		}
		wg := mod.NewDefWatchDog(l.opt.DagScheduleTimeout, l.opt.DagRunningTimeout)
		wg.Init()
		l.leaderCloser = append(l.leaderCloser, wg)

		dis := mod.NewDefDispatcher(l.opt.ListInsCount)
		dis.Init()
		l.leaderCloser = append(l.leaderCloser, dis)
		log.Println("leader initial")
	}
	// continue leader failed
	if !lcEvent.IsLeader && len(l.leaderCloser) != 0 {
		l.Close()
	}
}

// Close leader component
func (l *LeaderChangedHandler) Close() {
	for i := range l.leaderCloser {
		l.leaderCloser[i].Close()
	}
	l.leaderCloser = []mod.Closer{}
}

// Close all closer
func Close() {
	for i := range closers {
		closers[i].Close()
	}
	goevent.Close()
}

func checkOption(opt *InitialOption) error {
	config := common.NewConfig()
	if opt.Store == nil {
		switch config.Server.DBType {
		case "mysql":
			opt.Store = dagmodel.NewDagRepository()
		default:
			st := mongoStore.NewStore(&mongoStore.StoreOption{
				// if your mongo does not set user/pwd, you should remove it
				ConnStr:  config.MongoDB.DSN(),
				Database: config.MongoDB.DBName(),
				Prefix:   "flow",
				MaxPool:  config.MongoDB.MaxPool(),
				MinPool:  config.MongoDB.MinPool(),
				Timeout:  100 * time.Second,
			})
			if err := st.Init(); err != nil {
				log.Fatal(fmt.Errorf("init store failed: %w", err))
			}
			opt.Store = st
		}
	}

	if opt.ExecutorTimeout == 0 {
		opt.ExecutorTimeout = 10 * 60 * time.Second
	}
	if opt.DagScheduleTimeout == 0 {
		opt.DagScheduleTimeout = 120 * time.Second
	}
	if opt.DagRunningTimeout == 0 {
		opt.DagRunningTimeout = 8 * 60 * 60 * time.Second
	}
	if opt.LowestExecutorWorkerCnt == 0 {
		opt.LowestExecutorWorkerCnt = 200
	}
	if opt.LowExecutorWorkerCnt == 0 {
		opt.LowExecutorWorkerCnt = 200
	}
	if opt.MediumExecutorWorkerCnt == 0 {
		opt.MediumExecutorWorkerCnt = 200
	}
	if opt.HighExecutorWorkerCnt == 0 {
		opt.HighExecutorWorkerCnt = 200
	}
	if opt.HighestExecutorWorkerCnt == 0 {
		opt.HighestExecutorWorkerCnt = 400
	}
	if opt.ListInsCount == 0 {
		opt.ListInsCount = 500
	}
	if opt.ParserWorkersCnt == 0 {
		opt.ParserWorkersCnt = 200
	}
	return nil
}

func initKeeper(opt *InitialOption) {
	if opt.Keeper != nil {
		return
	}
	machineID, _ := newMachineID()()
	keeper := keeper.NewKeeper(&keeper.KeeperOption{
		Key: fmt.Sprintf("worker-%d", machineID),
	})
	if err := keeper.Init(); err != nil {
		log.Fatal(fmt.Errorf("init keeper failed: %w", err))
	}

	opt.Keeper = keeper
}

func initCommonComponent(opt *InitialOption) {
	mod.SetKeeper(opt.Keeper)
	mod.SetStore(opt.Store)
	entity.StoreMarshal = opt.Store.Marshal
	entity.StoreUnmarshal = opt.Store.Unmarshal

	// Executor must init before parse otherwise will cause a error
	exe := mod.NewDefExecutor(opt.ExecutorTimeout, opt.LowestExecutorWorkerCnt, opt.LowExecutorWorkerCnt, opt.MediumExecutorWorkerCnt, opt.HighExecutorWorkerCnt, opt.HighExecutorWorkerCnt)
	mod.SetExecutor(exe)
	p := mod.NewDefParser(opt.ParserWorkersCnt, opt.ListInsCount, opt.ExecutorTimeout)
	mod.SetParser(p)

	go func() {
		exe.Init()
		closers = append(closers, exe)
		p.Init()
		closers = append(closers, p)

		comm := &mod.DefCommander{}
		mod.SetCommander(comm)

		// keeper and store must close latest
		closers = append(closers, opt.Store, opt.Keeper)
	}()
}

func readDagFromDir(dir string) error {
	paths, err := utils.DefaultReader.ReadPathsFromDir(dir)
	if err != nil {
		return err
	}

	for _, path := range paths {
		bs, err := utils.DefaultReader.ReadDag(path)
		if err != nil {
			return fmt.Errorf("read %s failed: %w", path, err)
		}

		dag := entity.Dag{
			Status: entity.DagStatusNormal,
		}
		err = yaml.Unmarshal(bs, &dag)
		if err != nil {
			return fmt.Errorf("unmarshal %s failed: %w", path, err)
		}

		if dag.ID == "" {
			dag.ID = strings.TrimSuffix(strings.TrimSuffix(filepath.Base(path), ".yaml"), ".yml")
		}

		if err := ensureDagLatest(&dag); err != nil {
			return err
		}
	}
	return nil
}

func ensureDagLatest(dag *entity.Dag) error {
	oDag, err := mod.GetStore().GetDag(context.Background(), dag.ID)
	if err != nil && !errors.Is(err, data.ErrDataNotFound) {
		return err
	}
	if oDag != nil {
		return mod.GetStore().UpdateDag(context.Background(), dag)
	}

	dag.SetCheckRootNode(func(t []entity.Task) error {
		_, bErr := mod.BuildRootNode(mod.MapTasksToGetter(dag.Tasks))
		if bErr != nil {
			return bErr
		}
		return nil
	})

	_, err = mod.GetStore().CreateDag(context.Background(), dag)
	return err
}

// https://github.com/tinrab/makaroni/tree/master/utilities/unique-id
// 根据ip获取唯一id
func newMachineID() func() (uint16, error) {
	return func() (uint16, error) {
		ipStr := os.Getenv("POD_IP")
		if ipStr == "" {
			ipStr = "127.0.0.1"
		}
		ip := net.ParseIP(ipStr)
		ip = ip.To16()
		if ip == nil || len(ip) < 4 {
			return 0, errors.New("invalid IP")
		}
		return uint16(ip[14])<<8 + uint16(ip[15]), nil
	}
}

func generatePassword() string {
	// 生成[6, 100]以内的随机数
	letter := []rune(letters)
	b := make([]rune, 12)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := range b {
		b[i] = letter[r.Intn(len(letter))]
	}
	return string(b)
}

func initRds(config *common.Config) {
	err := cdb.InitGormDB(&cdb.Config{
		Host:            config.DB.HOST,
		Port:            config.DB.PORT,
		Driver:          config.DB.TYPE,
		Name:            config.DB.NAME,
		User:            config.DB.USER,
		Password:        config.DB.PASSWORD,
		Timezone:        "UTC",
		MaxIdleConns:    "2",
		MaxOpenConns:    "0",
		ConnMaxLifetime: "0m",
	})
	if err != nil {
		panic(err)
	}
}

func initMQClient(config *common.Config) {
	err := cmq.InitMQClient(&cmq.MQConfig{
		ConfigPath:               common.CMS_CONFIG_SERVICE_ACCESS,
		PollIntervalMilliseconds: 100,
		MaxInFlight:              200,
		ProtonMQInfo: msqclient.ProtonMQInfo{
			Host:        config.MQ.Host,
			Port:        config.MQ.Port,
			LookupdHost: config.MQ.LookupdHost,
			LookupdPort: config.MQ.LookupdPort,
			MQType:      config.MQ.ConnectorType,
			Auth: &msqclient.AuthOpts{
				Username:  config.MQ.Auth.UserName,
				Password:  config.MQ.Auth.PassWord,
				Mechanism: config.MQ.Auth.Mechanism,
			},
		},
	})
	if err != nil {
		panic(err)
	}
	closers = append(closers, cmq.NewMQClient())
}

func initFlowO11yLogger() {
	config := common.NewConfig()
	conf := &telemetryvar.TelemetryConf{
		ServerName:    config.Telemetry.ServerName,
		ServerVersion: config.Telemetry.ServerVersion,
	}

	f, err := os.Open("/conf/flow_o11y_data.yaml")
	if err != nil {
		commonLog.NewLogger().Errorf("open flow_o11y_data.yaml failed: %s", err.Error())
		return
	}
	defer f.Close()

	if err := yaml.NewDecoder(f).Decode(&conf); err != nil {
		commonLog.NewLogger().Errorf("decode flow_o11y_data.yaml failed: %s", err.Error())
		return
	}

	traceLog.InitFlowO11yLogger(conf)
}
