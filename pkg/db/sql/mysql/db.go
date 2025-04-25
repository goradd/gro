package mysql

import (
	"context"
	sqldb "database/sql"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"github.com/goradd/all"
	"github.com/goradd/orm/pkg/db"
	sql2 "github.com/goradd/orm/pkg/db/sql"
	. "github.com/goradd/orm/pkg/query"
	"strings"
	"time"
)

// DB is the goradd driver for mysql databases. It works through the excellent go-sql-driver driver,
// to supply functionality above go's built in driver. To use it, call NewDB, but afterward,
// work through the DB parent interface so that the underlying database can be swapped out later if needed.
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
	sql2.DbHelper
	databaseName string
}

// NewDB returns a new DB database object that you can add to the datastore.
// If connectionString is set, it will be used to create the configuration. Otherwise,
// use a config setting.
func NewDB(dbKey string, connectionString string, config *mysql.Config) (*DB, error) {
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
	m.DbHelper = sql2.NewSqlHelper(dbKey, db3, m)
	if config != nil {
		m.databaseName = config.DBName // save off the database name for later use
	} else {
		cfg, err := mysql.ParseDSN(connectionString)
		if err != nil {
			return nil, fmt.Errorf("could not parse database DSN: %w", err)
		}
		m.databaseName = cfg.DBName
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
			config.Params = all.StringMap(v.(map[string]interface{}))
		case "collation":
			config.Collation = v.(string)
		case "maxAllowedPacket":
			config.MaxAllowedPacket = int(v.(float64))
		case "serverPubKey":
			config.ServerPubKey = v.(string)
		case "tlsConfig":
			config.TLSConfig = v.(string)
		case "timeout":
			config.Timeout = time.Duration(int(v.(float64))) * time.Second
		case "readTimeout":
			config.ReadTimeout = time.Duration(int(v.(float64))) * time.Second
		case "writeTimeout":
			config.WriteTimeout = time.Duration(int(v.(float64))) * time.Second
		case "allowAllFiles":
			config.AllowAllFiles = v.(bool)
		case "allowCleartextPasswords":
			config.AllowCleartextPasswords = v.(bool)
		case "allowNativePasswords":
			config.AllowNativePasswords = v.(bool)
		case "allowOldPasswords":
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
func (m *DB) OperationSql(op Operator, operandStrings []string) (sql string) {
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
		if me, ok := all.As[*mysql.MySQLError](err); ok {
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

// Update sets specific fields of a record that already exists in the database.
// optLockFieldName is the name of a version field that will implement an optimistic locking check before executing the update.
// Note that if the database is not currently in a transaction, then the optimistic lock
// will be a cursory check, but will not be able to definitively prevent a prior write.
// If optLockFieldName is provided, that field will be updated with a new version number value during the update process.
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
	newLock, err = m.DbHelper.CheckLock(ctx, table, pkName, pkValue, optLockFieldName, optLockFieldValue)
	if err != nil {
		return
	}
	if newLock != 0 {
		fields[optLockFieldName] = newLock
	}
	s, args := sql2.GenerateUpdate(m, table, fields, map[string]any{pkName: pkValue})
	_, err = m.SqlExec(ctx, s, args...)
	if err != nil {
		if me, ok := all.As[*mysql.MySQLError](err); ok {
			// expected error situation to report to developer
			if me.Number == 1062 {
				// Since it is not possible to completely prevent a unique constraint error, except by implementing a separate
				// service to track and lock unique values that are in use (which is beyond the scope of the ORM), we need
				// to see if this is that kind of error and return it, rather than panicking.
				return 0, db.NewUniqueValueError(table, nil, err)
			}
		}
		return 0, db.NewQueryError("SqlExec", s, args, err)
	}
	return
}
