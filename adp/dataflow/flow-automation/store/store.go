// Package store 存储模块
package store

import (
	"strconv"

	"github.com/sony/sonyflake"
)

var generator *sonyflake.Sonyflake

// InitFlakeGenerator 初始化雪花id生成器
func InitFlakeGenerator(machineID uint16) {
	generator = sonyflake.NewSonyflake(sonyflake.Settings{
		MachineID: func() (uint16, error) {
			return machineID, nil
		},
	})
}

// NextID 递增id
func NextID() uint64 {
	id, err := generator.NextID()
	if err != nil {
		panic(err)
	}
	return id
}

// NextStringID uint转string
func NextStringID() string {
	return strconv.FormatUint(NextID(), 10)
}
