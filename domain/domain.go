package domain

import (
	"github.com/berkaroad/util"
	"github.com/berkaroad/uuid"
)

func init() {
	util.TypeContainer.RegisterTypeOf((*Event)(nil))
}

// 领域事件
type Event interface {
	// Saga 处理ID
	ProcessId() uuid.UUID
}

// 事件流
type EventStream struct {
	// 事件流版本
	StreamVersion int
	// 关联的一组事件
	Events []Event
}

// 聚合标识
type AggregateIdentity interface {
	// 单个聚合类型的唯一标识
	Id() string
	// 聚合类型标签
	Tag() string
}

// 聚合状态
type AggregateState interface {
	// 初始化聚合状态
	Init(events []Event)
	// 聚合状态版本
	Version() int
	// 通过事件改变聚合状态
	Mutate(e Event)
}

// 聚合
type Aggregate interface {
	// 初始化聚合
	Init(state AggregateState)
	// 当前聚合操作产生的变化，以一组事件来表示
	Changes() []Event
	// 将聚合操作创建的事件进行应用：更改聚合状态、记录到Changes中
	Apply(e Event)
}

// 聚合标识字符串形式
func AggIdToStr(id AggregateIdentity) string {
	return id.Tag() + "|" + id.Id()
}

// 应用服务
type ApplicationService interface {
	// 构造函数
	util.Initializer
}
