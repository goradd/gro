// Package sql contains helper functions that connect a standard Go database/sql object
// to the GoRADD system.
//
// Most of the functionality in this package is used by database implementations. GoRADD users would
// not normally directly call functions in this package.
package sql

import (
	"context"
	"database/sql"
	"io"
	"log/slog"

	"github.com/goradd/gro/db"
	schema2 "github.com/goradd/gro/internal/schema"
)

// The DbI interface describes the interface that a sql database needs to implement so that
// it will work with the Builder object. If you know a DatabaseI object is a
// standard Go database/sql database, you can
// cast it to this type to gain access to the low level SQL driver and send it raw SQL commands.
type DbI interface {
	// SqlExec executes a query that does not expect to return values.
	// It will time out the query if contextTimeout is exceeded
	SqlExec(ctx context.Context, sql string, args ...interface{}) (r sql.Result, err error)
	// SqlQuery executes a SQL query that returns values.
	// It will time out the query if contextTimeout is exceeded
	SqlQuery(ctx context.Context, sql string, args ...interface{}) (r *sql.Rows, err error)
	// QuoteIdentifier will put quotes around an identifier in a database specific way.
	QuoteIdentifier(string) string
	// FormatArgument will return the placeholder string for the n'th argument
	// in a sql string. Some sqls just use "?" while others identify an argument
	// by location, like Postgres's $1, $2, etc.
	FormatArgument(n int) string
	// SupportsForUpdate will return true if it supports SELECT ... FOR UPDATE clauses for row level locking in a transaction
	SupportsForUpdate() bool
	// TableDefinitionSql returns the sql that will create table.
	TableDefinitionSql(d *schema2.Database, table *schema2.Table) (tableSql string, extraSql []string)
}

type contextKey string

func RowClose(c io.Closer) {
	if err := c.Close(); err != nil {
		slog.Warn("Error closing sql row cursor", slog.Any(db.LogError, err))
	}
}
