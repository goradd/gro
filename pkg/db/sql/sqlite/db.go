package sqlite

import (
	"context"
	sqldb "database/sql"
	"fmt"
	"github.com/goradd/orm/pkg/db"
	sql2 "github.com/goradd/orm/pkg/db/sql"
	. "github.com/goradd/orm/pkg/query"
	"github.com/goradd/orm/pkg/schema"
	"log/slog"
	"modernc.org/sqlite"
	"strings"
	"time"
)

// DB is the goradd driver for a modernc sqlite database.
type DB struct {
	sql2.DbHelper
	contextTimeout time.Duration
}

func init() {
	// this is required to enable foreign key checks on every connection, even ones that are created blindly
	// by the driver's connection pool mechanism.
	sqlite.RegisterConnectionHook(func(conn sqlite.ExecQuerierContext, dsn string) error {
		// enable FK enforcement (or any other PRAGMAs you like)
		_, err := conn.ExecContext(context.Background(), "PRAGMA foreign_keys = ON", nil)
		return err
	})
}

// NewDB returns a new Sqlite database object based on the modernc driver.
// See https://sqlite.org/uri.html for the format of the connection string.
// An empty connection string will create a memory only database that is shared across
// all commections from within the same process.
func NewDB(dbKey string,
	connectionString string) (*DB, error) {
	if connectionString == "" {
		connectionString = "file::memory:?cache=shared"
	}

	db3, err := sqldb.Open("sqlite", connectionString)
	if err != nil {
		return nil, fmt.Errorf("could not open database: %w", err)
	}
	err = db3.Ping()
	if err != nil {
		return nil, fmt.Errorf("could not ping database: %w", err)
	}

	var version string
	db3.QueryRow("SELECT sqlite_version();").Scan(&version)
	slog.Info("SQLite version:" + version)

	m := new(DB)
	m.DbHelper = sql2.NewSqlHelper(dbKey, db3, m)
	return m, nil
}

// QuoteIdentifier surrounds the given identifier with quote characters
// appropriate for Postgres
func (m *DB) QuoteIdentifier(v string) string {
	var b strings.Builder
	b.WriteRune('"')
	b.WriteString(v)
	b.WriteRune('"')
	return b.String()
}

// FormatArgument formats the given argument number for embedding in a SQL statement.
func (m *DB) FormatArgument(n int) string {
	return "?"
}

// OperationSql provides SQLite specific SQL for certain operators.
// operandStrings will already be escaped.
func (m *DB) OperationSql(op Operator, operands []Node, operandStrings []string) (sql string) {
	switch op {
	case OpContains:
		// handle enum fields
		if o := operands[0]; o.NodeType_() == ColumnNodeType {
			cn := o.(*ColumnNode)
			if cn.SchemaType == schema.ColTypeEnumArray {
				s := operandStrings[0]
				s2 := operandStrings[1]
				return fmt.Sprintf(`EXISTS (
  					SELECT 1
  					FROM   json_each(%s)
  					WHERE  json_each.value = %s)`, s, s2)
			}
		}

	case OpDateAddSeconds:
		// Modifying a datetime in the query
		// Only works on date, datetime and timestamps. Not times.
		s := operandStrings[0]
		s2 := operandStrings[1]
		var modifier string
		if len(s2) > 0 && s2[0] != '-' && s2[0] != '+' {
			modifier = "+"
		}
		return fmt.Sprintf(`datetime(%s, '%s%s seconds')`, s, modifier, s2)
	}
	return
}

// Insert inserts the given data as a new record in the database.
// If the table has an auto-generated primary key, pass the name of that field to pkName.
// Insert will then return the new auto-generated primary key.
// If fields contains the auto-generated primary key, Insert will also synchronize postgres to make
// sure it will not auto generate another key that matches the manually set primary key.
// Set pkName to empty if the table has a manually set primary key.
// Table can include a schema name separated with a period.
func (m *DB) Insert(ctx context.Context, table string, pkName string, fields map[string]interface{}) (string, error) {
	sql, args := sql2.GenerateInsert(m, table, fields)
	if pkName == "" {
		if _, err := m.SqlExec(ctx, sql, args...); err != nil {
			if sqliteErr, ok := err.(interface{ Code() int }); ok {
				if sqliteErr.Code() == 2067 {
					return "", db.NewUniqueValueError(table, nil, err)
				}
			}
			return "", db.NewQueryError("SqlQuery", sql, args, err)
		}
		return "", nil // success
	}

	id, err := m.insertWithReturning(ctx, table, pkName, sql, args)

	if err != nil {
		return "", err
	}

	return id, err
}

func (m *DB) insertWithReturning(ctx context.Context, table string, pkName string, sql string, args []interface{}) (string, error) {
	sql += fmt.Sprintf(" RETURNING %s", m.QuoteIdentifier(pkName))
	rows, err := m.SqlQuery(ctx, sql, args...)

	if rows != nil {
		defer sql2.RowClose(rows)
	}

	if err != nil {
		if sqliteErr, ok := err.(interface{ Code() int }); ok {
			if sqliteErr.Code() == 2067 {
				return "", db.NewUniqueValueError(table, nil, err)
			}
		}
		return "", db.NewQueryError("SqlQuery", sql, args, err)
	} else {
		var id string
		// get id
		if rows == nil || !rows.Next() {
			// Theoretically this should not happen.
			return "", fmt.Errorf("primary key column not found")
		}
		err = rows.Scan(&id)
		if err != nil {
			return "", db.NewQueryError("Scan", sql, args, err)
		}

		if err = rows.Err(); err != nil {
			return "", db.NewQueryError("rows.Err", sql, args, err)
		}

		return id, err
	}
}

// Update sets specific fields of a single record that exists in the database.
// optLockFieldName is the name of a version field that will implement an optimistic locking check while doing the update.
// If optLockFieldName is provided:
//   - That field will be used to limit the update,
//   - That field will be updated with a new version
//   - If the record was deleted, or if the record was previously updated, an OptimisticLockError will be returned.
//     You will need to query further to determine if the record still exists.
//
// Otherwise, if optLockFieldName is blank, and the record we are attempting to change does not exist, the database
// will not be altered, and no error will be returned.
func (m *DB) Update(ctx context.Context,
	table string,
	pkName string,
	pkValue any,
	fields map[string]any,
	optLockFieldName string,
	optLockFieldValue int64,
) (newLock int64, err error) {
	if len(fields) == 0 {
		panic("fields must not be empty")
	}
	where := map[string]any{pkName: pkValue}
	if optLockFieldName != "" {
		newLock = db.RecordVersion(optLockFieldValue)
		where[optLockFieldName] = optLockFieldValue
		fields[optLockFieldName] = newLock
	}
	s, args := sql2.GenerateUpdate(m, table, fields, where, false)
	var result sqldb.Result
	result, err = m.SqlExec(ctx, s, args...)
	if err != nil {
		if sqliteErr, ok := err.(interface{ Code() int }); ok {
			if sqliteErr.Code() == 2067 {
				return 0, db.NewUniqueValueError(table, nil, err)
			}
		}
		return 0, db.NewQueryError("SqlExec", s, args, err)
	}
	if rows, _ := result.RowsAffected(); rows == 0 {
		if optLockFieldName != "" {
			return 0, db.NewOptimisticLockError(table, pkValue, nil)
		} /*else {
			Note: We cannot determine that a record was not found, because another possibility is simply that the record
				  did not change. The above works because the optimistic lock forces a change.
			return 0, db.NewRecordNotFoundError(table, pkValue)
		} */
	}
	return
}

func (m *DB) SupportsForUpdate() bool {
	return true
}
