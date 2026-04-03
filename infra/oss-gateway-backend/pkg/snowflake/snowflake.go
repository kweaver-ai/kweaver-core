package snowflake

import (
	"errors"
	"sync"
	"time"
)

const (
	// 64位ID的划分
	workerIDBits     = 5  // 机器ID位数
	datacenterIDBits = 5  // 数据中心ID位数
	sequenceBits     = 12 // 序列号位数

	// 最大取值
	maxWorkerID     = -1 ^ (-1 << workerIDBits)     // 31 (0b11111)
	maxDatacenterID = -1 ^ (-1 << datacenterIDBits) // 31 (0b11111)
	maxSequence     = -1 ^ (-1 << sequenceBits)     // 4095 (0b111111111111)

	// 移位偏移
	workerIDShift      = sequenceBits                                   // 12
	datacenterIDShift  = sequenceBits + workerIDBits                    // 17
	timestampLeftShift = sequenceBits + workerIDBits + datacenterIDBits // 22

	// Twitter元年时间戳 (2010-11-04 09:42:54)
	twepoch int64 = 1288834974657
)

// IDWorker 雪花ID生成器
type IDWorker struct {
	mu            sync.Mutex
	datacenterID  int64
	workerID      int64
	sequence      int64
	lastTimestamp int64
}

// NewIDWorker 创建雪花ID生成器
// datacenterID: 数据中心（机器区域）ID，范围 0-31
// workerID: 机器ID，范围 0-31
func NewIDWorker(datacenterID, workerID int64) (*IDWorker, error) {
	if workerID > maxWorkerID || workerID < 0 {
		return nil, errors.New("worker_id值越界，有效范围: 0-31")
	}

	if datacenterID > maxDatacenterID || datacenterID < 0 {
		return nil, errors.New("datacenter_id值越界，有效范围: 0-31")
	}

	return &IDWorker{
		datacenterID:  datacenterID,
		workerID:      workerID,
		sequence:      0,
		lastTimestamp: -1,
	}, nil
}

// GetID 获取新的雪花ID（19位数字）
func (w *IDWorker) GetID() (int64, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	timestamp := w.genTimestamp()

	// 时钟回拨检测
	if timestamp < w.lastTimestamp {
		return 0, errors.New("clock is moving backwards, rejecting requests")
	}

	// 同一毫秒内
	if timestamp == w.lastTimestamp {
		w.sequence = (w.sequence + 1) & maxSequence
		if w.sequence == 0 {
			// 序列号用尽，等待下一毫秒
			timestamp = w.tilNextMillis(w.lastTimestamp)
		}
	} else {
		w.sequence = 0
	}

	w.lastTimestamp = timestamp

	// 生成ID
	// 时间戳部分 | 数据中心ID | 机器ID | 序列号
	newID := ((timestamp - twepoch) << timestampLeftShift) |
		(w.datacenterID << datacenterIDShift) |
		(w.workerID << workerIDShift) |
		w.sequence

	return newID, nil
}

// genTimestamp 生成当前时间戳（毫秒）
func (w *IDWorker) genTimestamp() int64 {
	return time.Now().UnixNano() / 1e6
}

// tilNextMillis 等待到下一毫秒
func (w *IDWorker) tilNextMillis(lastTimestamp int64) int64 {
	timestamp := w.genTimestamp()
	for timestamp <= lastTimestamp {
		timestamp = w.genTimestamp()
	}
	return timestamp
}

// 全局默认 worker 实例（datacenterID=1, workerID=1）
var defaultWorker *IDWorker

func init() {
	var err error
	defaultWorker, err = NewIDWorker(1, 1)
	if err != nil {
		panic("failed to initialize snowflake worker: " + err.Error())
	}
}

// GetDefaultWorker 获取默认的 worker 实例
func GetDefaultWorker() *IDWorker {
	return defaultWorker
}

// GenerateID 使用默认 worker 生成雪花ID
func GenerateID() (int64, error) {
	return defaultWorker.GetID()
}
