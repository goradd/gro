package db

import (
	"context"
	"errors"
	"github.com/goradd/maps"
	. "github.com/goradd/orm/pkg/query"
	"github.com/goradd/orm/pkg/schema"
	"iter"
)

// List of supported database drivers
const (
	DriverTypeMysql    = "mysql"
	DriverTypePostgres = "postgres"
	DriverTypeSQLite   = "sqlite"
)

// The dataStore is the central database collection used in code generation and the orm.
var datastore maps.SliceMap[string, DatabaseI]

type contextKey string

type TransactionID int

type SchemaExtractor interface {
	ExtractSchema(options map[string]any) schema.Database
}

type SchemaRebuilder interface {
	DestroySchema(ctx context.Context, s schema.Database) error
	CreateSchema(ctx context.Context, s schema.Database) error
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
	// Update will put the given values into a single record that already exists in the database.
	// The fields value should include only fields that have changed.
	// pkName is the query name of the primary key field and pkValue its value.
	// optLockFieldName and optLockFieldValue points to a version field in the record that helps implement optimistic locking. These can be empty if no optimistic locking is required.
	// returns a new value of the lock if the update is successful.
	Update(ctx context.Context, table string, pkName string, pkValue any, fields map[string]any, optLockFieldName string, optLockFieldValue int64) (int64, error)
	// Insert will insert a new record into the database with the given values.
	// If the primary key is auto generated, then the name of the primary key column should be passed in autoPkName, in which
	// case the newly generated primary key will be returned.
	// If the primary key is set in fields, it will be used instead of the generated value and the database will
	// be updated, if needed, to prevent the database from generating a future value that conflicts with this value.
	// Otherwise, if the primary key is manually set, autoPkName should be empty.
	// If fields does not include all the required values in the database, the database may return an error.
	Insert(ctx context.Context, table string, autoPkName string, fields map[string]any) (string, error)
	// Delete will delete records from the database that match the colName and colValue.
	// If optLockFieldName is provided, the optLockFieldValue will also constrain the delete, and if no
	// records are found, it will return an OptimisticLockError. If optLockFieldName is empty, and
	// no record is found, a NoRecordFound error will be returned.
	Delete(ctx context.Context, table string, colName string, colValue any, optLockFieldName string, optLockFieldValue int64) error
	// DeleteAll will efficiently delete all the records from a table.
	DeleteAll(ctx context.Context, table string) error
	// Query executes a simple query on a single table using fields, where the keys of fields are the names of database fields to select,
	// and the values are the types of data to return for each field.
	// If orderBy is not nil, it specifies field names to sort the data on, in ascending order.
	// If the database supports transactions and row locking, and a transaction is active, it will lock the rows read, and
	// depending on the setting in the transaction, it will be either a read or a write lock.
	Query(ctx context.Context, table string, fields map[string]ReceiverType, where map[string]any, orderBy []string) (CursorI, error)
	// BuilderQuery performs a complex query using a query builder.
	// The data returned will depend on the command inside the builder.
	BuilderQuery(ctx context.Context, builder *Builder) (any, error)
	// NewContext is called early in the processing of a response to insert an empty context that the database can use if needed.
	NewContext(ctx context.Context) context.Context
	// SetConstraints will turn foreign key constraints on or off.
	// Databases that do not support constraints can make this a no-op.
	// This is used during import in case there are circular or forward references in the data.
	// Applications should not call this directly, but rather use the WithConstraintsDisabled wrapper.
	// This MUST be done inside a transaction.
	SetConstraints(on bool) error
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

type transactioner interface {
	WithTransaction(ctx context.Context, f func(ctx context.Context) error) error
}

// WithTransaction wraps the function f in a database transaction if the driver supports transactions.
// Otherwise, just executes f with ctx.
//
// While the ORM by default will wrap individual database calls with a timeout,
// it will not apply this timeout to a transaction. It is up to you to pass a context that
// has a timeout to prevent the overall transaction from hanging.
func WithTransaction(ctx context.Context, d DatabaseI, f func(ctx context.Context) error) error {
	if t, ok := d.(transactioner); ok {
		return t.WithTransaction(ctx, f)
	}
	return f(ctx) // pass through without transaction
}

type constrainter interface {
	WithConstraintsOff(ctx context.Context, f func(ctx context.Context) error) error
}

// WithConstraintsOff turns off constraints for databases that support foreign key constraints.
// Otherwise, will just call f with ctx.
func WithConstraintsOff(ctx context.Context, d DatabaseI, f func(ctx context.Context) error) error {
	if c, ok := d.(constrainter); ok {
		return c.WithConstraintsOff(ctx, f)
	}
	return f(ctx) // pass through without constraint change
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
	relatedPks []K) error {
	err := WithTransaction(ctx, d, func(ctx context.Context) error {
		if err := d.Delete(ctx, assnTable, srcColumnName, pk, "", 0); err != nil {
			var rErr *RecordNotFoundError
			if !errors.As(err, &rErr) { // ignore record not found errors
				return err
			}
		}
		for _, relatedPk := range relatedPks {
			if err := Associate(ctx, d, assnTable, srcColumnName, pk, relatedColumnName, relatedPk); err != nil {
				return err
			}
		}
		return nil
	})
	return err
}

// Associate adds a record to the assnTable table.
func Associate[J, K any](ctx context.Context,
	d DatabaseI,
	assnTable string,
	srcColumnName string,
	pk J,
	relatedColumnName string,
	relatedPk K) error {
	_, err := d.Insert(ctx, assnTable, "", map[string]any{srcColumnName: pk, relatedColumnName: relatedPk})
	return err
}
