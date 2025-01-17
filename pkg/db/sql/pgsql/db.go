package pgsql

import (
	"context"
	sqldb "database/sql"
	"fmt"
	"github.com/goradd/all"
	sql2 "github.com/goradd/orm/pkg/db/sql"
	"github.com/goradd/orm/pkg/model"
	. "github.com/goradd/orm/pkg/query"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"strings"
)

// DB is the goradd driver for postgresql databases.
type DB struct {
	sql2.DbHelper
	Model *model.Database
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
	config *pgx.ConnConfig) *DB {
	if connectionString == "" && config == nil {
		panic("must specify how to connect to the database")
	}

	if connectionString == "" {
		connectionString = stdlib.RegisterConnConfig(config)
	}

	db3, err := sqldb.Open("pgx", connectionString)
	if err != nil {
		panic("Could not open database: " + err.Error())
	}
	err = db3.Ping()
	if err != nil {
		panic("Could not ping database " + dbKey + ":" + err.Error())
	}

	m := new(DB)
	m.DbHelper = sql2.NewSqlDb(dbKey, db3, m)
	return m
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
			config.RuntimeParams = all.StringMap(v.(map[string]interface{}))
		case "kerberosServerName":
			config.KerberosSrvName = v.(string)
		case "kerberosSPN":
			config.KerberosSpn = v.(string)
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
func (m *DB) OperationSql(op Operator, operandStrings []string) (sql string) {
	switch op {
	case OpDateAddSeconds:
		// Modifying a datetime in the query
		// Only works on date, datetime and timestamps. Not times.
		s := operandStrings[0]
		s2 := operandStrings[1]
		return fmt.Sprintf(`(%s + make_interval(seconds => %s))`, s, s2)
	}
	return
}

// Insert inserts the given data as a new record in the database.
// It returns the record id of the new record.
func (m *DB) Insert(ctx context.Context, table string, fields map[string]interface{}) string {
	sql, args := sql2.GenerateInsert(m, table, fields)
	sql += " RETURNING "
	sql += m.Model.Table(table).PrimaryKeyColumn().QueryName
	if rows, err := m.SqlQuery(ctx, sql, args...); err != nil {
		panic(err.Error())
	} else {
		var id string
		defer rows.Close()
		for rows.Next() {
			err = rows.Scan(&id)
		}
		if err != nil {
			panic(err.Error())
			return ""
		} else {
			return id
		}
	}
}
