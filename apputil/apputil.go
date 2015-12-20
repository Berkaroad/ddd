package apputil

import (
	"fmt"
	"github.com/berkaroad/ddd/domain"
	"github.com/berkaroad/ddd/es"
)

// 创建或更新聚合，并持久化到事件存储中
// isCreateAction：是否为创建操作
// store：事件存储
// id：聚合标识
// state：聚合状态
// agg：聚合
// action：创建或更新聚合的操作
func CreateOrUpdateAggregate(isCreateAction bool, store es.EventStore, id domain.AggregateIdentity, state domain.AggregateState, agg domain.Aggregate, action func(agg domain.Aggregate) error) {
	stream := store.LoadEventStream(id)
	state.Init(stream.Events)
	if isCreateAction {
		if stream.StreamVersion > 0 {
			panic(fmt.Sprintf("Aggregate '%s' is already exists!", domain.AggIdToStr(id)))
		}
	} else {
		if stream.StreamVersion <= 0 {
			panic(fmt.Sprintf("Aggregate '%s' is not exists!", domain.AggIdToStr(id)))
		} else if state.Version() < 0 {
			panic(fmt.Sprintf("Aggregate '%s' is deleted, couldn't use it!", domain.AggIdToStr(id)))
		}
	}
	agg.Init(state)

	if err := action(agg); err != nil {
		panic(err.Error())
	}
	streamExpectedVersion := stream.StreamVersion
	if streamExpectedVersion < 0 {
		streamExpectedVersion = 0
	}
	if err := store.AppendEventsToStream(id, streamExpectedVersion, agg.Changes()); err != nil {
		panic(err.Error())
	}
}
