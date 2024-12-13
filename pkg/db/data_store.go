package db

import (
	"context"
	"github.com/goradd/maps"
	. "github.com/goradd/orm/pkg/query"
	"github.com/goradd/orm/pkg/schema"
	"iter"
)

type DatabaseMap = maps.SliceMap[string, DatabaseI]

// The dataStore is the central database collection used in code generation and the orm.
var datastore *DatabaseMap

type TransactionID int

type SchemaExtractor interface {
	ExtractSchema(options map[string]any) schema.Database
}

// DatabaseI is the interface that describes the behaviors required for a database implementation.
type DatabaseI interface {
	// NewBuilder returns a newly created query builder
	NewBuilder(ctx context.Context) QueryBuilderI

	// Update will put the given values into a record that already exists in the database. The "fields" value
	// should include only fields that have changed.
	Update(ctx context.Context, table string, fields map[string]interface{}, pkName string, pkValue interface{})
	// Insert will insert a new record into the database with the given values, and return the new record's primary key value.
	// The fields value should include all the required values in the database.
	Insert(ctx context.Context, table string, fields map[string]interface{}) string
	// Delete will delete the given record from the database
	Delete(ctx context.Context, table string, pkName string, pkValue interface{})
	// Associate sets a many-many relationship to the given values.
	// The values are taken from the ORM, and are treated differently depending on whether this is a SQL or NoSQL database.
	Associate(ctx context.Context,
		table string,
		column string,
		pk interface{},
		relatedTable string,
		relatedColumn string,
		relatedPks interface{})

	// Begin will begin a transaction in the database and return the transaction id
	Begin(ctx context.Context) TransactionID
	// Commit will commit the given transaction
	Commit(ctx context.Context, txid TransactionID)
	// Rollback will roll back the given transaction PROVIDED it has not been committed. If it has been
	// committed, it will do nothing. Rollback can therefore be used in a defer statement as a safeguard in case
	// a transaction fails.
	Rollback(ctx context.Context, txid TransactionID)
	// NewContext is called early in the processing of a response to insert an empty context that the database can use if needed.
	NewContext(ctx context.Context) context.Context
}

// AddDatabase adds a database to the global database store. Only call this during app startup.
func AddDatabase(d DatabaseI, key string) {
	if datastore == nil {
		datastore = new(DatabaseMap)
	}

	datastore.Set(key, d)
}

// GetDatabase returns the database given the database's key.
func GetDatabase(key string) DatabaseI {
	return datastore.Get(key)
}

// DatabaseIter returns an iterator over the databases in key order.
func DatabaseIter() iter.Seq2[string, DatabaseI] {
	return datastore.All()
}

// NewContext returns a new context with the database contexts inserted into the given
// context. Pass nil to return a BackgroundContext with the database contexts.
//
// A database context is required by the various database calls to track results
// and transactions.
func NewContext(ctx context.Context) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	for _, d := range DatabaseIter() {
		ctx = d.NewContext(ctx)
	}
	return ctx
}

// ExecuteTransaction wraps the function in a database transaction
func ExecuteTransaction(ctx context.Context, d DatabaseI, f func()) {
	txid := d.Begin(ctx)
	defer d.Rollback(ctx, txid)
	f()
	d.Commit(ctx, txid)
}
