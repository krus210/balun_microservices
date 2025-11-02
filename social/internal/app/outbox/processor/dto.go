package processor

import (
	"time"

	"github.com/google/uuid"
)

// ----- Search options DTO -----

type searchEventsOptions struct {
	NotBefore *time.Time
	NotAfter  *time.Time

	AggregateType *AggregateType
	EventType     *EventType

	OnlyUnpublished bool
	MaxRetryCount   int
	DueAt           *time.Time
	Limit           int
	WithLock        bool
	// ...
}

type SearchEventsOption func(o *searchEventsOptions)

// ----- Option builders -----

func WithLimit(n int) SearchEventsOption {
	return func(o *searchEventsOptions) { o.Limit = n }
}

func WithOnlyUnpublished() SearchEventsOption {
	return func(o *searchEventsOptions) { o.OnlyUnpublished = true }
}

func WithMaxRetryCount(n int) SearchEventsOption {
	return func(o *searchEventsOptions) { o.MaxRetryCount = n }
}

func WithLock() SearchEventsOption {
	return func(o *searchEventsOptions) { o.WithLock = true }
}

func WithAggregateType(t AggregateType) SearchEventsOption {
	return func(o *searchEventsOptions) { o.AggregateType = &t }
}

func WithEventType(t EventType) SearchEventsOption {
	return func(o *searchEventsOptions) { o.EventType = &t }
}

func WithNotBefore(t time.Time) SearchEventsOption {
	return func(o *searchEventsOptions) { o.NotBefore = &t }
}

func WithNotAfter(t time.Time) SearchEventsOption {
	return func(o *searchEventsOptions) { o.NotAfter = &t }
}

func WithDueAt(t time.Time) SearchEventsOption {
	return func(o *searchEventsOptions) { o.DueAt = &t }
}

func CollectSearchEventsOptions(opts ...SearchEventsOption) searchEventsOptions {
	res := searchEventsOptions{
		Limit:         10,
		MaxRetryCount: 3,
	}

	for _, opt := range opts {
		opt(&res)
	}
	return res
}

// ----- Update options DTO -----

type updateEventsOptions struct {
	// window по времени для partition pruning (created_at)
	NotBefore *time.Time
	NotAfter  *time.Time

	AggregateType *AggregateType
	EventType     *EventType

	IDs []uuid.UUID

	// что обновляем
	SetPublishedAt   *time.Time
	IncRetryBy       int
	SetNextAttemptAt *time.Time

	// фильтры статуса
	OnlyUnpublished bool // по умолчанию true
}

type UpdateEventsOption func(*updateEventsOptions)

func CollectUpdateEventsOptions(opts ...UpdateEventsOption) updateEventsOptions {
	o := updateEventsOptions{
		OnlyUnpublished: true,
	}
	for _, opt := range opts {
		opt(&o)
	}
	return o
}

// ----- Option builders -----

func WithUpdateIDs(ids ...uuid.UUID) UpdateEventsOption {
	return func(o *updateEventsOptions) { o.IDs = append(o.IDs, ids...) }
}

func WithUpdateAggregateType(at AggregateType) UpdateEventsOption {
	return func(o *updateEventsOptions) { o.AggregateType = &at }
}

func WithUpdateEventType(et EventType) UpdateEventsOption {
	return func(o *updateEventsOptions) { o.EventType = &et }
}

func WithUpdateNotBefore(t time.Time) UpdateEventsOption {
	return func(o *updateEventsOptions) { o.NotBefore = &t }
}

func WithUpdateNotAfter(t time.Time) UpdateEventsOption {
	return func(o *updateEventsOptions) { o.NotAfter = &t }
}

func SetPublishedAt(ts time.Time) UpdateEventsOption {
	return func(o *updateEventsOptions) { o.SetPublishedAt = &ts }
}

func IncRetry(by int) UpdateEventsOption {
	return func(o *updateEventsOptions) { o.IncRetryBy = by }
}

func IncludePublished() UpdateEventsOption {
	return func(o *updateEventsOptions) { o.OnlyUnpublished = false }
}

func SetNextAttemptAt(ts time.Time) UpdateEventsOption {
	return func(o *updateEventsOptions) { o.SetNextAttemptAt = &ts }
}
