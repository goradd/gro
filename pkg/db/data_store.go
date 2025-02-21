package db

import (
	"context"
	"github.com/goradd/maps"
	. "github.com/goradd/orm/pkg/query"
	"github.com/goradd/orm/pkg/schema"
	"iter"
)

var DriverType string

// List of supported database drivers
const (
	DriverTypeMysql    = "mysql"
	DriverTypePostgres = "postgres"
)

// The dataStore is the central database collection used in code generation and the orm.
var datastore maps.SliceMap[string, DatabaseI]

type TransactionID int

type SchemaExtractor interface {
	ExtractSchema(options map[string]any) schema.Database
}

// DatabaseI is the interface that describes the behaviors required for a database implementation.
//
// Time values are converted to whatever time format the database prefers.
//
// JSON values must already be encoded as strings or []byte values.
//
// If where is not nil, it specifies fields and values that will limit the search.
// Multiple field-value combinations will be Or'd together.
// If a value is a map[string]any type, its key is ignored, and the keys and values of the enclosed type will be
// And'd together. This Or-And pattern is recursive.
// If a value is a slice of int or strings, those values will be put in an "IN" test.
// For example, {"vals":[]int{1,2,3}} will result in SQL of "vals IN (1,2,3)".
type DatabaseI interface {
	// Update will put the given values into a record that already exists in the database. The fields value
	// should include only fields that have changed. All records that match the keys and values in where are changed.
	Update(ctx context.Context, table string, fields map[string]interface{}, where map[string]interface{})
	// Insert will insert a new record into the database with the given values, and return the new record's primary key value.
	// The field's value should include all the required values in the database.
	Insert(ctx context.Context, table string, fields map[string]interface{}) string
	// Delete will delete records from the database that match the key value pairs in where.
	// If where is nil, all the data will be deleted.
	Delete(ctx context.Context, table string, where map[string]interface{})
	// Query executes a simple query on a single table using fields, where the keys of fields are the names of database fields to select,
	// and the values are the types of data to return for each field.
	//
	// If orderBy is not nil, it specifies field names to sort the data on, in ascending order.
	Query(ctx context.Context, table string, fields map[string]ReceiverType, where map[string]any, orderBy []string) CursorI
	// BuilderQuery performs a complex query using a query builder.
	// The data returned will depend on the command inside the builder.
	BuilderQuery(builder *Builder) any
	// Begin will begin a transaction in the database and return the transaction id
	// Instead of calling Begin directly, use the ExecuteTransaction wrapper.
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
func ExecuteTransaction(ctx context.Context, d DatabaseI, f func() error) error {
	txid := d.Begin(ctx)
	defer d.Rollback(ctx, txid)
	err := f()
	d.Commit(ctx, txid)
	return err
}

// AssociateOnly resets a many-many relationship in the database.
// The assnTable is the name of the association table that contains the many-many relationships.
// The srcColumnName is the name of the column that points to the primary key in the source table.
// The value of that column is pk.
// The relatedColumnName is the name of the column in the association table that points to the destination table's primary key.
// with relatedPks having all the primary keys of objects that should be associated with the object with
// primary key pk.
// All previous associations with the source object are deleted.
func AssociateOnly[J, K any](ctx context.Context,
	d DatabaseI,
	assnTable string,
	srcColumnName string,
	pk J,
	relatedColumnName string,
	relatedPks []K) {
	_ = ExecuteTransaction(ctx, d, func() error {
		d.Delete(ctx, assnTable, map[string]interface{}{srcColumnName: pk})
		for _, relatedPk := range relatedPks {
			d.Insert(ctx, assnTable, map[string]any{srcColumnName: pk, relatedColumnName: relatedPk})
		}
		return nil
	})
}

// Associate adds a record to the assnTable table.
func Associate[J, K any](ctx context.Context,
	d DatabaseI,
	assnTable string,
	srcColumnName string,
	pk J,
	relatedColumnName string,
	relatedPk K) {
	d.Insert(ctx, assnTable, map[string]any{srcColumnName: pk, relatedColumnName: relatedPk})
}
