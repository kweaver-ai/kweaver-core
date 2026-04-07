package keeper

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	redis_v8 "github.com/go-redis/redis/v8"
	lock "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/lock"
	commonLog "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/log"
	rds "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/store"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/event"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/store"
	"github.com/shiningrush/goevent"
)

const (
	// LeaderKey leader key
	RdsLeaderKey             = "flow-automation:leader"
	RdsWorkerKey             = "flow-automation:worker"
	defaultHeartbeatInterval = 5 * time.Second
	defaultNodeTTL           = 30 * time.Second
	defaultLeaderLockTTL     = 30 * time.Second
	defaultCheckInterval     = 10 * time.Second
)

type Keeper struct {
	opt        *KeeperOption
	leaderFlag atomic.Value
	keyNumber  int
	logger     commonLog.Logger
	rds        rds.RDB
	lockClient lock.DistributeLock

	wg sync.WaitGroup

	ctx    context.Context
	cancel context.CancelFunc

	// leader 专属协程控制
	leaderMu     sync.Mutex
	leaderCtx    context.Context
	leaderCancel context.CancelFunc
}

type KeeperOption struct {
	Key string
}

// NewKeeper new keeper instance
func NewKeeper(opt *KeeperOption) *Keeper {
	k := &Keeper{
		opt:    opt,
		logger: commonLog.NewLogger(),
		rds:    rds.NewRedis(),
		lockClient: lock.NewDistributeLock(
			rds.NewRedis(),
			RdsLeaderKey,
			opt.Key),
	}
	k.leaderFlag.Store(false)
	return k
}

// WorkerNumber get the the key number of Worker key, if here is a WorkKey like `worker-1`, then it will return "1"
func (k *Keeper) WorkerNumber() int {
	return k.keyNumber
}

// WorkerKey must match `xxxx-1` format
func (k *Keeper) WorkerKey() string {
	return k.opt.Key
}

// Init 初始化
func (k *Keeper) Init() error {
	if err := k.readOpt(); err != nil {
		return err
	}

	store.InitFlakeGenerator(uint16(k.WorkerNumber()))

	k.ctx, k.cancel = context.WithCancel(context.Background())
	k.logger.Infof("[Keeper] worker: %s, init keeper start...", k.WorkerKey())

	err := k.register()
	if err != nil {
		k.cancel()
		return err
	}

	k.wg.Add(2)
	go k.heartBeatLoop()
	go k.goElect()

	return nil
}

func (k *Keeper) setLeaderFlag(isLeader bool) {
	k.leaderFlag.Store(isLeader)
	goevent.Publish(&event.LeaderChanged{
		IsLeader:  isLeader,
		WorkerKey: k.WorkerKey(),
	})
}

func (k *Keeper) readOpt() error {
	if k.opt.Key == "" {
		return fmt.Errorf("worker key string can not be empty")
	}

	number, err := CheckWorkerKey(k.opt.Key)
	if err != nil {
		return err
	}
	k.keyNumber = number

	return nil
}

func (k *Keeper) register() error {
	return k.rds.GetClient().ZAdd(k.ctx, RdsWorkerKey, &redis_v8.Z{
		Score:  float64(time.Now().UnixMilli()),
		Member: k.WorkerKey(),
	}).Err()
}

func (k *Keeper) unregister() error {
	return k.rds.GetClient().ZRem(context.Background(), RdsWorkerKey, k.WorkerKey()).Err()
}

func (k *Keeper) goElect() {
	defer k.wg.Done()

	// 启动后立即执行一次选举
	k.tryElection()

	ticker := time.NewTicker(defaultHeartbeatInterval)

	defer ticker.Stop()
	for {
		select {
		case <-k.ctx.Done():
			return
		case <-ticker.C:
			k.tryElection()
		}
	}
}

func (k *Keeper) tryElection() {
	k.leaderMu.Lock()
	defer k.leaderMu.Unlock()

	isLeader := k.leaderFlag.Load().(bool)

	if isLeader {
		return
	}

	k.lockClient.ClearErrChannel()

	err := k.lockClient.TryLock(k.ctx, defaultLeaderLockTTL, true)
	if err != nil {
		if !errors.Is(err, rds.ErrLockNotAcquired) {
			k.logger.Warnf("[Keeper] worker: %s, try election failed, detail: %s", k.WorkerKey(), err.Error())
		}
		return
	}

	k.logger.Infof("[Keeper] elected success, current leader: %s", k.WorkerKey())

	// 添加主节点配置信息
	k.setLeaderFlag(true)

	// 开启主节点任务
	k.startLeaderTasks()
}

// startLeaderTasks 当前节点被选举为主节点后，需要开启的后续任务
func (k *Keeper) startLeaderTasks() {
	k.leaderCtx, k.leaderCancel = context.WithCancel(k.ctx)

	k.wg.Add(3)
	go k.aliveCheckLoop()
	go k.watchLockError()
	go k.IsLockHeldCheck()

}

// aliveCheckLoop 是一个goroutine，用于定期检查活动节点
// 该goroutine会在k.ctx或者k.leaderCtx关闭时退出
// 在每个defaultCheckInterval时间间隔内，aliveCheckLoop会检查活动节点
func (k *Keeper) aliveCheckLoop() {
	defer k.wg.Done()

	// 主节点选举成功，立即进行一次活动节点检查
	k.aliveCheck()

	ticker := time.NewTicker(defaultCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-k.ctx.Done():
			return
		case <-k.leaderCtx.Done():
			return
		case <-ticker.C:
			k.aliveCheck()
		}
	}
}

// aliveCheck 用于定期检查活动节点。
// 该函数会检查在defaultNodeTTL时间内没有心跳的活动节点，并将其从redis中删除。
func (k *Keeper) aliveCheck() {
	deadline := float64(time.Now().Add(-defaultNodeTTL).UnixMilli())
	removed, err := k.rds.GetClient().ZRemRangeByScore(k.leaderCtx, RdsWorkerKey, "-inf", fmt.Sprintf("%f", deadline)).Result()
	if err != nil {
		k.logger.Warnf("[Keeper] alive check error: %s", err)
		return
	}

	if removed > 0 {
		k.logger.Infof("[Keeper] alive check removed %d unalive nodes", removed)
	}
}

func (k *Keeper) watchLockError() {
	defer k.wg.Done()

	errCh := k.lockClient.GetErrChannel()
	for {
		select {
		case <-k.ctx.Done():
			return
		case <-k.leaderCtx.Done():
			return
		case err, ok := <-errCh:
			if !ok {
				k.logger.Warnf("[Keeper] watch lock error, channel closed")
				k.stopLeaderTasks()
				return
			}
			if err != nil {
				k.logger.Warnf("[Keeper] worker: %s, watchlock error: %s", k.WorkerKey(), err.Error())
				k.stopLeaderTasks()
				return
			}
		}
	}
}

// IsLockHeldCheck 定时检测当前锁持有者，非锁持有者退出
func (k *Keeper) IsLockHeldCheck() {
	defer k.wg.Done()

	ticker := time.NewTicker(defaultHeartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-k.ctx.Done():
			return
		case <-k.leaderCtx.Done():
			return
		case <-ticker.C:
			if !k.IsLockHeld() {
				k.logger.Warnf("[Keeper] lock held check, lock ownership lost, worker: %s", k.WorkerKey())
				k.stopLeaderTasks()
				return
			}
		}
	}
}

func (k *Keeper) stopLeaderTasks() {
	k.leaderMu.Lock()
	defer k.leaderMu.Unlock()

	isLeader := k.leaderFlag.Load().(bool)
	if !isLeader {
		return
	}

	k.setLeaderFlag(false)
	k.lockClient.Release()

	if k.leaderCancel != nil {
		k.leaderCancel()
		k.leaderCancel = nil
	}

	k.logger.Infof("[Keeper] worker %s stopped leader tasks", k.WorkerKey())
}

func (k *Keeper) heartBeatLoop() {
	defer k.wg.Done()

	ticker := time.NewTicker(defaultHeartbeatInterval)
	defer ticker.Stop()
	for {
		select {
		case <-k.ctx.Done():
			return
		case <-ticker.C:
			err := k.rds.GetClient().ZAdd(k.ctx, RdsWorkerKey, &redis_v8.Z{
				Score:  float64(time.Now().UnixMilli()),
				Member: k.WorkerKey(),
			}).Err()
			if err != nil {
				k.logger.Errorf("[Keeper] worker %s, heartbeat failed, detail: %s", k.WorkerKey(), err.Error())
			}
		}
	}
}

// IsLeader indicate the component if is leader node
func (k *Keeper) IsLeader() bool {
	return k.leaderFlag.Load().(bool)
}

func (k *Keeper) AliveNodes() ([]string, error) {
	deadline := float64(time.Now().Add(-defaultNodeTTL).UnixMilli())
	return k.rds.GetClient().ZRangeByScore(k.ctx, RdsWorkerKey, &redis_v8.ZRangeBy{
		Min: fmt.Sprintf("%f", deadline),
		Max: "+inf",
	}).Result()
}

func (k *Keeper) IsAlive(workerKey string) (bool, error) {
	// 获取节点的心跳时间戳
	score, err := k.rds.GetClient().ZScore(k.ctx, RdsWorkerKey, workerKey).Result()
	if err != nil {
		if err == redis_v8.Nil {
			return false, nil
		}
		return false, err
	}

	// 判断是否在有效期内
	lastHeartbeat := int64(score)
	deadline := time.Now().Add(-defaultNodeTTL).UnixMilli()

	return lastHeartbeat > deadline, nil
}

// isLockHeld returns true if the current node holds the lock, false otherwise
func (k *Keeper) IsLockHeld() bool {
	val, err := k.rds.Get(k.lockClient.GetLockKey())
	if err != nil {
		return false
	}
	return val == k.opt.Key
}

func (k *Keeper) Close() {
	k.stopLeaderTasks()

	k.cancel()
	k.wg.Wait()

	err := k.unregister()
	if err != nil {
		k.logger.Warnf("[Keeper] worker %s, unregister worker failed: %s", k.WorkerKey(), err.Error())
	}
}
