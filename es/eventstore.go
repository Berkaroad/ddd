package es

import (
	"fmt"
	"github.com/berkaroad/ddd/domain"
	"github.com/berkaroad/util"
)

// 事件存储
type EventStore interface {
	// 构造函数
	util.Initializer
	// 加载事件流
	LoadEventStream(id domain.AggregateIdentity) domain.EventStream
	// 追加事件流
	AppendEventsToStream(id domain.AggregateIdentity, expectedVersion int, events []domain.Event) error
}

// 乐观并发错误
type OptimisticConcurrencyError struct {
	// 实际版本
	ActualVersion int
	// 预期版本
	ExpectedVersion int
	// 聚合标识
	Id domain.AggregateIdentity
	// 实际一组事件
	ActualEvents []domain.Event
}

func (self *OptimisticConcurrencyError) Error() string {
	message := fmt.Sprintf("OptimisticConcurrencyError: Expected v%d but found v%d in stream '%s'", self.ExpectedVersion, self.ActualVersion, domain.AggIdToStr(self.Id))
	return message
}
