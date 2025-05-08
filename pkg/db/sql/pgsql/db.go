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
	"strings"
	"time"
)

// DB is the goradd driver for postgresql databases.
type DB struct {
	sql2.DbHelper
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
//
// contextTimeout is the timeout that will be set in the context such that all individual database
// calls will need to complete within that time or the context will be canceled with an error.
// The PGX driver monitors this cancellation and will timeout the database call.
// PGX has no other mechanism of assuring a database query does not hang. The ConnectionTimeout
// setting in config only monitors the time it takes to establish a connection.
// contextTimeout will not be applied to transactions.
func NewDB(dbKey string,
	connectionString string,
	config *pgx.ConnConfig,
	contextTimeout time.Duration) (*DB, error) {
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
	m.DbHelper = sql2.NewSqlHelper(dbKey, db3, m, contextTimeout)
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
			config.Host = v.(string) // Typically, tcp or unix (for unix sockets).
		case "port":
			config.Port = uint16(v.(float64))
		case "runtimeParams":
			config.RuntimeParams = anyutil.StringMap(v.(map[string]interface{}))
		case "kerberosServerName":
			config.KerberosSrvName = v.(string)
		case "kerberosSPN":
			config.KerberosSpn = v.(string)
		case "connectionTimeout":
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
	case OpContains:
		// handle enum fields
		if o := operands[0]; o.NodeType_() == ColumnNodeType {
			cn := o.(*ColumnNode)
			if cn.SchemaType == schema.ColTypeEnumArray {
				s := operandStrings[0]
				s2 := operandStrings[1]
				// stored as a json array in the field
				return fmt.Sprintf(`%s @> '%s'`, s, s2)
			}
		}

		// TBD Json fields

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
// It returns the record id of the new record, and possibly an error if an error occurred.
func (m *DB) Insert(ctx context.Context, table string, pkName string, fields map[string]interface{}) (string, error) {
	sql, args := sql2.GenerateInsert(m, table, fields)
	if pkName != "" {
		sql += " RETURNING "
		sql += pkName
	}
	if rows, err := m.SqlQuery(ctx, sql, args...); err != nil {
		if pgErr, ok := anyutil.As[*pgconn.PgError](err); ok {
			if pgErr.Code == "23505" {
				return "", db.NewUniqueValueError(table, nil, err)
			}
		}
		return "", db.NewQueryError("SqlQuery", sql, args, err)
	} else {
		var id string
		defer sql2.RowClose(rows)
		for rows.Next() {
			err = rows.Scan(&id)
			return "", db.NewQueryError("Scan", sql, args, err)
		}
		if err = rows.Err(); err != nil {
			return "", db.NewQueryError("rows.Err", sql, args, err)
		} else {
			return id, nil
		}
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
