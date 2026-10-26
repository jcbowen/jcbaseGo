// Package snowflake 提供基于雪花算法（Snowflake）的分布式唯一 ID 生成能力。
//
// 该实现为纯算法封装，不依赖任何业务类型或全局变量，可直接用于任意项目。
// 默认起始时间戳为 2024-01-01 00:00:00 UTC，节点位宽 10（0~1023）、序列位宽 12。
package snowflake

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

const (
	// defaultEpochMillis 默认起始时间戳（2024-01-01 00:00:00 UTC）
	defaultEpochMillis int64 = 1704067200000

	// nodeBits 节点 ID 占用位数
	nodeBits uint8 = 10
	// sequenceBits 序列号占用位数
	sequenceBits uint8 = 12

	// nodeMax 最大节点 ID
	nodeMax int64 = -1 ^ (-1 << nodeBits)
	// sequenceMask 序列号掩码
	sequenceMask int64 = -1 ^ (-1 << sequenceBits)

	// nodeShift 节点 ID 左移位数
	nodeShift uint8 = sequenceBits
	// timestampShift 时间戳左移位数
	timestampShift uint8 = nodeBits + sequenceBits

	// maxBackwardMillis 允许的最大时钟回拨毫秒数，超过则拒绝生成 ID
	maxBackwardMillis int64 = 5000
)

// Snowflake 雪花 ID 生成器
type Snowflake struct {
	mu       sync.Mutex
	node     int64
	epoch    int64
	sequence int64
	lastTime int64

	// now 可注入的毫秒时钟源，仅用于测试；nil 时使用 time.Now().UnixMilli()
	now func() int64
}

// NewSnowflake 创建雪花 ID 生成器（使用默认起始时间戳）。
//
// 参数:
//   - node: 节点 ID，范围为 [0, 1023]
//
// 返回值:
//   - *Snowflake: 生成器实例
//   - error: 节点 ID 非法时返回错误
func NewSnowflake(node uint16) (*Snowflake, error) {
	return NewSnowflakeWithEpoch(node, defaultEpochMillis)
}

// NewSnowflakeWithEpoch 创建雪花 ID 生成器，可自定义起始时间戳（毫秒）。
//
// 参数:
//   - node: 节点 ID，范围为 [0, 1023]
//   - epochMillis: 起始时间戳（毫秒）
//
// 返回值:
//   - *Snowflake: 生成器实例
//   - error: 节点 ID 非法时返回错误
func NewSnowflakeWithEpoch(node uint16, epochMillis int64) (*Snowflake, error) {
	if int64(node) > nodeMax {
		return nil, errors.New("节点 ID 超出允许范围")
	}
	return &Snowflake{
		node:     int64(node),
		epoch:    epochMillis,
		sequence: 0,
		lastTime: -1,
	}, nil
}

// currentTime 返回当前毫秒时间戳（优先使用注入时钟）
func (s *Snowflake) currentTime() int64 {
	if s.now != nil {
		return s.now()
	}
	return time.Now().UnixMilli()
}

// NextID 生成下一个唯一 ID。
// 整个生成过程持有互斥锁：序列号在当前毫秒内用尽、或时钟回拨时，
// 在锁内等待时钟严格越过上次生成时间后再继续，确保任何情况下序列号不会被复用。
//
// 返回值:
//   - uint64: 生成的雪花 ID
//   - error: 时钟回拨超过阈值时返回错误
func (s *Snowflake) NextID() (uint64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.currentTime()
	// 时钟回拨：等待时钟追上 lastTime，回拨超过阈值直接报错
	for now < s.lastTime {
		if s.lastTime-now > maxBackwardMillis {
			return 0, fmt.Errorf("时钟回拨超过阈值 %d ms，lastTime=%d now=%d", maxBackwardMillis, s.lastTime, now)
		}
		time.Sleep(time.Millisecond)
		now = s.currentTime()
	}

	if now == s.lastTime {
		s.sequence = (s.sequence + 1) & sequenceMask
		if s.sequence == 0 {
			// 当前毫秒序列号已用尽：必须进入下一毫秒后从 0 重新计数，
			// 期间保持持锁，防止时钟抖动/回跳导致同一毫秒的序列号被复用
			for now = s.currentTime(); now <= s.lastTime; now = s.currentTime() {
				time.Sleep(time.Millisecond)
			}
			s.sequence = 0
		}
	} else {
		s.sequence = 0
	}

	s.lastTime = now
	id := ((now - s.epoch) << timestampShift) | (s.node << nodeShift) | s.sequence
	return uint64(id), nil
}
