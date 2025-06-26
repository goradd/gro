package pgsql

import (
	"context"
	sqldb "database/sql"
	"fmt"
	"github.com/goradd/anyutil"
	"github.com/goradd/orm/pkg/db"
	sql2 "github.com/goradd/orm/pkg/db/sql"
	. "github.com/goradd/orm/pkg/query"
	"github.com/goradd/orm/pkg/schema"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/stdlib"
	"log/slog"
	"slices"
	"strings"
	"time"
)

// DB is the goradd driver for postgresql databases.
type DB struct {
	sql2.Base
	contextTimeout time.Duration
}

// NewDB returns a new Postgresql DB database object based on the pgx driver
// that you can add to the datastore.
// If connectionString is set, it will be used to create the configuration. Otherwise,
// use a config setting. Using a configSetting can potentially give you access to the
// underlying pgx database for advanced operations.
//
// The postgres driver specifies that you must use ParseConfig
// to create the initial configuration, although that can be sent a blank string to
// gather initial values from environment variables. You can then change items in
// the configuration structure. For example:
//
//	config,_ := pgx.ParseConfig(connectionString)
//	config.Password = "mysecret"
//	db := pgsql.NewDB(key, "", config)
func NewDB(dbKey string,
	connectionString string,
	config *pgx.ConnConfig) (*DB, error) {
	if connectionString == "" && config == nil {
		return nil, fmt.Errorf("must specify how to connect to the database")
	}

	if connectionString == "" {
		connectionString = stdlib.RegisterConnConfig(config)
	}

	db3, err := sqldb.Open("pgx", connectionString)
	if err != nil {
		return nil, fmt.Errorf("could not open database: %w", err)
	}
	err = db3.Ping()
	if err != nil {
		return nil, fmt.Errorf("could not ping database: %w", err)
	}

	m := new(DB)
	m.Base = sql2.NewBase(dbKey, db3, m)
	return m, nil
}

// OverrideConfigSettings will use a map read in from a json file to modify
// the given config settings
func OverrideConfigSettings(config *pgx.ConnConfig, jsonContent map[string]interface{}) {
	for k, v := range jsonContent {
		switch k {
		case "database":
			config.Database = v.(string)
		case "user":
			config.User = v.(string)
		case "password":
			config.Password = v.(string)
		case "host":
			config.Host = v.(string)
		case "port":
			config.Port = uint16(v.(float64))
		case "runtime_params":
			config.RuntimeParams = anyutil.StringMap(v.(map[string]interface{}))
		case "kerberos_server_name":
			config.KerberosSrvName = v.(string)
		case "kerberos_spn":
			config.KerberosSpn = v.(string)
		case "connection_timeout":
			d, err := time.ParseDuration(v.(string))
			if err != nil {
				config.ConnectTimeout = d
			}
		}
	}
}

// QuoteIdentifier surrounds the given identifier with quote characters
// appropriate for Postgres
func (m *DB) QuoteIdentifier(v string) string {
	var b strings.Builder
	b.WriteRune('"')

	if i := strings.Index(v, "."); i != -1 {
		b.WriteString(v[:i])
		b.WriteString(`"."`)
		b.WriteString(v[i+1:])
	} else {
		b.WriteString(v)
	}

	b.WriteRune('"')
	return b.String()
}

// FormatArgument formats the given argument number for embedding in a SQL statement.
func (m *DB) FormatArgument(n int) string {
	return fmt.Sprintf(`$%d`, n)
}

// OperationSql provides Postgres specific SQL for certain operators.
func (m *DB) OperationSql(op Operator, operands []Node, operandStrings []string) (sql string) {
	switch op {
	case OpDateAddSeconds:
		// Modifying a datetime in the query
		// Only works on date, datetime and timestamps. Not times.
		s := operandStrings[0]
		s2 := operandStrings[1]
		return fmt.Sprintf(`(%s + MAKE_INTERVAL(SECONDS => %s))`, s, s2)
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
			if pgErr, ok := anyutil.As[*pgconn.PgError](err); ok {
				if pgErr.Code == "23505" {
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

	if _, ok := fields[pkName]; ok {
		// If we are manually inserting a primary key, we need to potentially sync the next value.
		err = m.syncIdentity(ctx, table, pkName)
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
		if pgErr, ok := anyutil.As[*pgconn.PgError](err); ok {
			if pgErr.Code == "23505" {
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
		if pgErr, ok := anyutil.As[*pgconn.PgError](err); ok {
			if pgErr.Code == "23505" {
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

// syncIdentity should be called whenever a record is inserted that is manually setting a pk value
// that is normally generated by the database.
// Note that table can be of the form "schema.table".
func (m *DB) syncIdentity(ctx context.Context, table string, pk string) error {
	var schemaWithDot string
	parts := strings.Split(table, ".")
	if len(parts) > 1 {
		schemaWithDot = m.QuoteIdentifier(parts[0]) + "."
	}

	// Note: the sgoradd_sync_identity_sequence requires the first parameter to be quoted, but not the 2nd.
	s := fmt.Sprintf(`SELECT %sgoradd_sync_identity_sequence('%s'::TEXT, '%s'::TEXT)`, schemaWithDot, m.QuoteIdentifier(table), pk)

	_, err := m.SqlExec(ctx, s)
	if err != nil { // Should not normally happen
		slog.Warn("Error during syncIdentity",
			slog.String(db.LogTable, table),
			slog.String(db.LogColumn, pk),
			slog.Any(db.LogError, err))
	}
	return err
}

// SetConstraints turns constraints on and off.
// For postgres, this must be done inside a transaction.
func (m *DB) SetConstraints(on bool) error {
	var sqlStr string

	if !on {
		sqlStr = "SET CONSTRAINTS ALL DEFERRED"
	} else {
		sqlStr = "SET CONSTRAINTS ALL IMMEDIATE"
	}
	_, err := m.SqlDb().Exec(sqlStr)
	return err
}

type contextKey string

func (m *DB) constraintKey() contextKey {
	return contextKey("PostgresConstraint-" + m.DbKey())
}

func (m *DB) getConstraintsOff(ctx context.Context) bool {
	i := ctx.Value(m.constraintKey())
	return i != nil
}

// WithConstraintsOff makes sure operations in f occur with foreign key constraints turned off.
// Postgres (and many other drivers), require that this happens within a transaction.
func (m *DB) WithConstraintsOff(ctx context.Context, f func(ctx context.Context) error) (err error) {
	off := m.getConstraintsOff(ctx)
	if off {
		// constraints are already off, so just pass through
		err = f(ctx)
		return
	}

	err = m.WithTransaction(ctx, func(ctx context.Context) (err error) {
		ctx = context.WithValue(ctx, m.constraintKey(), true)
		_, err = m.SqlExec(ctx, "SET CONSTRAINTS ALL DEFERRED")
		if err != nil {
			return
		}
		defer func() {
			// Although a transaction will automatically turn constraints back on,
			// we do this in case the call is embedded in another transaction and the developer
			// is trying to carefully control transactions
			_, err = m.SqlExec(ctx, "SET CONSTRAINTS ALL IMMEDIATE")
		}()
		return f(ctx)
	})

	return
}

// DestroySchema removes all tables and data from the tables found in the given schema s.
func (m *DB) DestroySchema(ctx context.Context, s schema.Database) {
	// gather table names to delete
	var tables []string

	for _, table := range s.AssociationTables {
		tables = append(tables, m.QuoteIdentifier(table.QualifiedTableName()))
	}
	for _, table := range slices.Backward(s.Tables) {
		tables = append(tables, m.QuoteIdentifier(table.QualifiedName()))
	}
	for _, table := range s.EnumTables {
		tables = append(tables, m.QuoteIdentifier(table.QualifiedTableName()))
	}

	cmd := fmt.Sprintf(`DROP TABLE IF EXISTS %s CASCADE`, strings.Join(tables, ","))
	_, err := m.SqlExec(ctx, cmd)
	if err != nil {
		slog.Error("failed to drop tables",
			slog.Any(db.LogError, err),
		)
	}
}
