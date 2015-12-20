package domain

import (
	"github.com/berkaroad/uuid"
)

// 领域事件基类
type BaseEvent struct {
	ProcId uuid.UUID
}

// 新建普通领域事件
func NewNormalEvent() BaseEvent {
	return BaseEvent{ProcId: uuid.UUID{}}
}

// 新建支持Saga的领域事件
func NewSagaSupportedEvent() BaseEvent {
	return BaseEvent{ProcId: uuid.New()}
}

// 加载支持Saga的领域事件
func LoadSagaSupportedEvent(previousEvent Event) BaseEvent {
	return BaseEvent{ProcId: previousEvent.ProcessId()}
}

func (self BaseEvent) ProcessId() uuid.UUID {
	return self.ProcId
}
