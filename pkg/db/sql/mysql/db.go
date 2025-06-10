package mysql

import (
	"context"
	sqldb "database/sql"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"github.com/goradd/anyutil"
	"github.com/goradd/orm/pkg/db"
	sql2 "github.com/goradd/orm/pkg/db/sql"
	. "github.com/goradd/orm/pkg/query"
	"github.com/goradd/orm/pkg/schema"
	"strings"
	"time"
)

// DB is the goradd driver for mysql databases. It works through the excellent go-sql-driver driver,
// to supply functionality above go's built in driver. To use it, call NewDB, but afterward,
// work through the DB parent interface so that the underlying database can be swapped out later if needed.
//
// # Primary Keys
//
// Historically, Mysql uses auto generated integers as primary keys. Newer versions of Mysql are able to generate
// UUIDs, but the implementation is v1 only, and is not the most secure, nor the best for balancing storage needs.
// To use UUIDs, instead override the getXXXinsertFields function and manually set your own UUID.
//
// # Timezones
//
// Mysql has some interesting quirks:
//   - Datetime types are internally stored in the timezone of the server, and then returned based on the timezone of the client.
//   - Timestamp types are internally stored in UTC and returned in the timezone of the client.
//
// This means that when querying based on DateTime fields, you need to be aware that the timezone will have shifted
// to the timezone of the server. Also, when transferring or syncing between databases in different timezones, you
// will need to account for the time differences.
//
// The easiest way to handle this is to set the timezone of the server to UTC, and make sure all time values are stored in UTC.
//
// When displaying times to users, change the timezone to that of the user, which may not be the Mysql server's
// timezone, and may not be the timezone of the computer running the go-sql-driver (which from MySQL's perspective is the client).
//
// The mysql-go-driver has the ability to set a default timezone in the Loc configuration parameter.
// It defaults to UTC.
// It appears to convert all times to this timezone before sending them
// to the database, and then when receiving times, it will set this as the timezone of the date.
// It is best to set this and your database to UTC, as this will make your database portable to other timezones.
//
// These issues are further compounded by the fact that MYSQL can initialize date and time values to what it
// believes is the current date and time in its server's timezone, but will not save the timezone itself.
// Because of that, you should initialize values in the application, rather than using MySQL's ability to
// set a default value.
//
// Set the ParseTime configuration parameter to TRUE so that the driver will parse the times into the correct
// timezone, navigating the GO server and database server timezones. Otherwise, we
// can only assume that the database is in UTC time, since we will not get any timezone info from the server.
//
// Be aware that when you view the data in SQL, it will appear in whatever timezone the MYSQL server is set to.
type DB struct {
	sql2.Base
	databaseName     string
	isMariaDB        bool
	defaultCollation string
}

// NewDB returns a new DB database object that you can add to the datastore.
// If connectionString is set, it will be used to create the configuration. Otherwise,
// use a config setting.
func NewDB(dbKey string,
	connectionString string,
	config *mysql.Config) (*DB, error) {
	if connectionString == "" && config == nil {
		return nil, fmt.Errorf("must specify how to connect to the database")

	}
	if connectionString == "" {
		connectionString = config.FormatDSN()
	}

	db3, err := sqldb.Open("mysql", connectionString)
	if err != nil {
		return nil, fmt.Errorf("could not open database: %w", err)
	}
	err = db3.Ping()
	if err != nil {
		return nil, fmt.Errorf("could not ping database: %w", err)
	}

	m := new(DB)
	m.Base = sql2.NewBase(dbKey, db3, m)
	if config != nil {
		m.databaseName = config.DBName // save off the database name for later use
	} else {
		cfg, err := mysql.ParseDSN(connectionString)
		if err != nil {
			return nil, fmt.Errorf("could not parse database DSN: %w", err)
		}
		m.databaseName = cfg.DBName
	}

	var version string

	if err = db3.QueryRow("SELECT VERSION()").Scan(&version); err != nil {
		return nil, err
	} else if strings.Contains(strings.ToLower(version), "mariadb") {
		m.isMariaDB = true
	}

	if err = db3.QueryRow("SELECT DEFAULT_COLLATION_NAME FROM information_schema.SCHEMATA WHERE SCHEMA_NAME = ?", m.databaseName).Scan(&m.defaultCollation); err != nil {
		return nil, err
	}

	return m, nil
}

// OverrideConfigSettings will use a map read in from a json file to modify
// the given config settings
func OverrideConfigSettings(config *mysql.Config, jsonContent map[string]interface{}) {
	for k, v := range jsonContent {
		switch k {
		case "database":
			config.DBName = v.(string)
		case "user":
			config.User = v.(string)
		case "password":
			config.Passwd = v.(string)
		case "net":
			config.Net = v.(string) // Typically, tcp or unix (for unix sockets).
		case "address":
			config.Addr = v.(string) // Note: if you set address, you MUST set net also.
		case "params":
			// Convert from map[string]any to map[string]string
			config.Params = anyutil.StringMap(v.(map[string]interface{}))
		case "collation":
			config.Collation = v.(string)
		case "max_allowed_packet":
			config.MaxAllowedPacket = int(v.(float64))
		case "server_pub_key":
			config.ServerPubKey = v.(string)
		case "tls_config":
			config.TLSConfig = v.(string)
		case "timeout":
			d, err := time.ParseDuration(v.(string))
			if err != nil {
				config.Timeout = d
			}
		case "read_timeout":
			d, err := time.ParseDuration(v.(string))
			if err != nil {
				config.ReadTimeout = d
			}
		case "write_timeout":
			d, err := time.ParseDuration(v.(string))
			if err != nil {
				config.WriteTimeout = d
			}
		case "allow_all_files":
			config.AllowAllFiles = v.(bool)
		case "allow_cleartext_passwords":
			config.AllowCleartextPasswords = v.(bool)
		case "allow_native_passwords":
			config.AllowNativePasswords = v.(bool)
		case "allow_old_passwords":
			config.AllowOldPasswords = v.(bool)
		}
	}

	// The other config options effect how queries work, and so should be set before
	// calling this function, as they will change how the GO code for these queries will
	// need to be written.
}

// QuoteIdentifier surrounds the given identifier with quote characters
// appropriate for mysql
func (m *DB) QuoteIdentifier(v string) string {
	var b strings.Builder
	b.WriteRune('`')
	b.WriteString(v)
	b.WriteRune('`')
	return b.String()
}

// FormatArgument formats the given argument number for embedding in a SQL statement.
// Mysql just uses a question mark as a placeholder.
func (m *DB) FormatArgument(_ int) string {
	return "?"
}

// DeleteUsesAlias indicates the database requires the alias of a table after
// a delete clause when using aliases in the delete.
func (m *DB) DeleteUsesAlias() bool {
	return true
}

// OperationSql provides Mysql specific SQL for certain operators.
func (m *DB) OperationSql(op Operator, operands []Node, operandStrings []string) (sql string) {
	switch op {
	case OpDateAddSeconds:
		// Modifying a datetime in the query
		// Only works on date, datetime and timestamps. Not times.
		s := operandStrings[0]
		s2 := operandStrings[1]
		sql = fmt.Sprintf(`DATE_ADD(%s, INTERVAL (%s) SECOND)`, s, s2)
	case OpXor:
		sOp := " " + op.String() + " "
		sql = " (" + strings.Join(operandStrings, sOp) + ") "
	}
	return
}

func (m *DB) SupportsForUpdate() bool {
	return true
}

// Insert inserts the given data as a new record in the database.
// It returns the record id of the new record.
func (m *DB) Insert(ctx context.Context, table string, _ string, fields map[string]interface{}) (string, error) {
	s, args := sql2.GenerateInsert(m, table, fields)
	if r, err := m.SqlExec(ctx, s, args...); err != nil {
		if me, ok := anyutil.As[*mysql.MySQLError](err); ok {
			// expected error situation to report to developer
			if me.Number == 1062 {
				// Since it is not possible to completely prevent a unique constraint error, except by implementing a separate
				// service to track and lock unique values that are in use (which is beyond the scope of the ORM), we need
				// to see if this is that kind of error and return it.
				return "", db.NewUniqueValueError(table, nil, err)
			}
		}
		return "", db.NewQueryError("SqlExec", s, args, err)
	} else {
		if id, err2 := r.LastInsertId(); err2 != nil {
			return "", db.NewQueryError("LastInsertId", s, args, err2)
		} else {
			return fmt.Sprint(id), nil
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
		if me, ok := anyutil.As[*mysql.MySQLError](err); ok {
			// expected error situation to report to developer. Unique value constraint was violated.
			if me.Number == 1062 {
				// Since it is not possible to completely prevent a unique constraint error, except by implementing a separate
				// service to track and lock unique values that are in use (which is beyond the scope of the ORM), we need
				// to see if this is that kind of error and return it, rather than panicking.
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

// SetConstraints turns constraints on or off.
// This operation MUST be done on the same connection as the operations depending on it.
func (m *DB) SetConstraints(ctx context.Context, on bool) error {
	arg := 0
	if on {
		arg = 1
	}
	sqlStr := fmt.Sprintf("SET FOREIGN_KEY_CHECKS = %d", arg)
	_, err := m.SqlDb().Exec(sqlStr)
	return err
}

type contextKey string

func (m *DB) constraintKey() contextKey {
	return contextKey("GoraddMysqlConstraint-" + m.DbKey())
}

func (m *DB) getConstraintsOff(ctx context.Context) bool {
	i := ctx.Value(m.constraintKey())
	return i != nil
}

// WithConstraintsOff makes sure operations in f occur with foreign key constraints turned off.
// As a byproduct of this, the operations will happen on the same pinned connection, meaning the
// operations should not be long-running so that the connection pool will not run dry.
// Nested calls will continue to operate with checks off, and the outermost call will turn them on.
func (m *DB) WithConstraintsOff(ctx context.Context, f func(ctx context.Context) error) (err error) {
	off := m.getConstraintsOff(ctx)
	if off {
		// constraints are already off, so just pass through
		err = f(ctx)
		return
	}

	err = m.WithSameConnection(ctx, func(ctx context.Context) (err error) {
		ctx = context.WithValue(ctx, m.constraintKey(), true)
		_, err = m.SqlExec(ctx, "SET FOREIGN_KEY_CHECKS = 0")
		if err != nil {
			return
		}
		defer func() {
			_, err = m.SqlExec(ctx, "SET FOREIGN_KEY_CHECKS = 1")
		}()
		return f(ctx)
	})

	return
}

// DestroySchema removes all the tables listed in the schema.
//
// Mysql automatically commits a transaction when dropping a table, so this operation
// cannot be done within a transaction, and is not reversible.
// It also handles turning off and on constraints, since that is session wide and so
// the connection must be controlled.
func (m *DB) DestroySchema(ctx context.Context, s schema.Database) {
	// gather table names to delete
	var tables []string

	for _, table := range s.EnumTables {
		tables = append(tables, table.QualifiedTableName())
	}
	for _, table := range s.Tables {
		tables = append(tables, table.QualifiedName())
	}
	for _, table := range s.AssociationTables {
		tables = append(tables, table.QualifiedTableName())
	}

	_ = m.WithConstraintsOff(ctx, func(ctx context.Context) (err error) {
		for _, table := range tables {
			_, _ = m.SqlExec(ctx, `DROP TABLE `+m.QuoteIdentifier(table))
		}
		return
	})
}
