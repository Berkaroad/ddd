package es

import (
	"github.com/berkaroad/ddd/domain"
	"github.com/berkaroad/storage"
	"github.com/berkaroad/util"
)

type AppendOnlyEventStore struct {
	isInitialized bool
	store         storage.AppendOnlyStore
}

func (self *AppendOnlyEventStore) InitFunc() interface{} {
	return func(store storage.AppendOnlyStore) {
		if !self.isInitialized {
			self.store = store
		}
		self.isInitialized = true
	}
}

func (self *AppendOnlyEventStore) LoadEventStream(id domain.AggregateIdentity) domain.EventStream {
	stream := domain.EventStream{StreamVersion: -1}
	for _, record := range self.store.ReadRecords(domain.AggIdToStr(id), 0, 10000) {
		if eventData, err := util.UnmarshalObj(record.Data); err == nil {
			if e, ok := eventData.(domain.Event); ok {
				stream.Events = append(stream.Events, e)
			}
		}
		stream.StreamVersion = record.StreamVersion
	}
	return stream
}

func (self *AppendOnlyEventStore) AppendEventsToStream(id domain.AggregateIdentity, expectedVersion int, events []domain.Event) error {
	if len(events) == 0 {
		return nil
	}
	for _, event := range events {
		if eventData, err := util.MarshalObj(event); err == nil {
			err = self.store.Append(domain.AggIdToStr(id), eventData, expectedVersion)
			if err != nil {
				if _, ok := err.(*storage.AppendOnlyStoreConcurrencyError); ok {
					server := self.LoadEventStream(id)
					err = &OptimisticConcurrencyError{ActualVersion: server.StreamVersion, ExpectedVersion: expectedVersion, Id: id, ActualEvents: server.Events}
				}
			}
			return err
		} else {
			return err
		}
	}
	return nil
}
