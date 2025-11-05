package repository

import (
	"context"
	"fmt"
	"lib/postgres"

	appoutbox "social/internal/app/outbox/processor"

	"github.com/Masterminds/squirrel"
)

func (r *Repository) UpdateEvents(ctx context.Context, opts ...appoutbox.UpdateEventsOption) error {
	const api = "outbox.Repository.UpdateEvents"

	o := appoutbox.CollectUpdateEventsOptions(opts...)

	// защита от noop
	if o.SetPublishedAt == nil && o.IncRetryBy == 0 && o.SetNextAttemptAt == nil {
		return nil
	}

	qb := r.qb.
		Update(tableOutboxEvents)

	// setters
	if o.SetPublishedAt != nil {
		qb = qb.Set(columnOutboxPublishedAt, *o.SetPublishedAt)
	}
	if o.IncRetryBy > 0 {
		qb = qb.Set(columnOutboxRetryCount, squirrel.Expr(columnOutboxRetryCount+" + ?", o.IncRetryBy))
	}
	if o.SetNextAttemptAt != nil {
		qb = qb.Set(columnOutboxNextAttemptAt, *o.SetNextAttemptAt)
	}

	// filters (для partition pruning)
	if o.AggregateType != nil {
		qb = qb.Where(squirrel.Eq{columnOutboxAggType: string(*o.AggregateType)})
	}
	if len(o.IDs) > 0 {
		qb = qb.Where(squirrel.Eq{columnOutboxID: o.IDs}) // id IN (...)
	}
	if o.NotBefore != nil {
		qb = qb.Where(squirrel.GtOrEq{columnOutboxCreatedAt: *o.NotBefore})
	}
	if o.NotAfter != nil {
		qb = qb.Where(squirrel.LtOrEq{columnOutboxCreatedAt: *o.NotAfter})
	}
	if o.OnlyUnpublished {
		qb = qb.Where(squirrel.Eq{columnOutboxPublishedAt: nil})
	}

	conn := r.db.GetQueryEngine(ctx)
	if _, err := conn.Execx(ctx, qb); err != nil {
		return fmt.Errorf("%s: %w", api, postgres.ConvertPGError(err))
	}
	return nil
}
