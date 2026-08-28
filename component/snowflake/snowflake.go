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

// NextID 生成下一个唯一 ID。
// 发生时钟回拨时，小回拨在释放锁后短等，大回拨（超过 5 秒）直接返回错误。
//
// 返回值:
//   - uint64: 生成的雪花 ID
//   - error: 时钟回拨超过阈值或序列号溢出等待失败时返回错误
func (s *Snowflake) NextID() (uint64, error) {
	for {
		s.mu.Lock()
		now := time.Now().UnixMilli()

		if now < s.lastTime {
			backward := s.lastTime - now
			if backward > maxBackwardMillis {
				s.mu.Unlock()
				return 0, fmt.Errorf("时钟回拨超过阈值 %d ms，lastTime=%d now=%d", maxBackwardMillis, s.lastTime, now)
			}
			// 小回拨释放锁后短等，避免阻塞其他 ID 生成请求
			s.mu.Unlock()
			time.Sleep(time.Millisecond)
			continue
		}

		if now == s.lastTime {
			s.sequence = (s.sequence + 1) & sequenceMask
			if s.sequence == 0 {
				// 当前毫秒内序列号已用完，释放锁等待下一毫秒
				s.mu.Unlock()
				time.Sleep(time.Millisecond)
				continue
			}
		} else {
			s.sequence = 0
		}

		s.lastTime = now
		id := ((now - s.epoch) << timestampShift) | (s.node << nodeShift) | s.sequence
		s.mu.Unlock()
		return uint64(id), nil
	}
}
