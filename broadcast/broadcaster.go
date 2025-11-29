package broadcast

import (
	"context"
)

// Broadcaster is the injected broadcaster that the generated forms use to notify the application
// that the database has changed. The application will start with a default that does nothing.
var Broadcaster BroadcasterI

type BroadcasterI interface {
	Insert(ctx context.Context, dbId string, table string, pk interface{})
	Update(ctx context.Context, dbId string, table string, pk interface{}, fieldnames ...string)
	Delete(ctx context.Context, dbId string, table string, pk interface{})
	BulkChange(ctx context.Context, dbId string, table string)
}

// DefaultBroadcaster broadcasts database changes to the application
type DefaultBroadcaster struct {
}

func (b DefaultBroadcaster) Insert(ctx context.Context, dbId string, table string, pk interface{}) {
}

func (b DefaultBroadcaster) Update(ctx context.Context, dbId string, table string, pk interface{}, fieldnames ...string) {
}

func (b DefaultBroadcaster) Delete(ctx context.Context, dbId string, table string, pk interface{}) {
}

func (b DefaultBroadcaster) BulkChange(ctx context.Context, dbId string, table string) {
}

func Insert(ctx context.Context, dbId string, table string, pk interface{}) {
	if Broadcaster != nil {
		Broadcaster.Insert(ctx, dbId, table, pk)
	}
}

func Update(ctx context.Context, dbId string, table string, pk interface{}, fieldnames ...string) {
	if Broadcaster != nil {
		Broadcaster.Update(ctx, dbId, table, pk, fieldnames...)
	}
}

func Delete(ctx context.Context, dbId string, table string, pk interface{}) {
	if Broadcaster != nil {
		Broadcaster.Delete(ctx, dbId, table, pk)
	}
}

func BulkChange(ctx context.Context, dbId string, table string) {
	if Broadcaster != nil {
		Broadcaster.BulkChange(ctx, dbId, table)
	}
}
