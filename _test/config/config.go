// Package config configures the database for tests.
// By default, the package is set up to provide local execution of the tests for developers
// of goradd-orm. See InitDB to choose the database that will be used in the tests.
//
// During the automated testing process, the database will be selected using a configuration file
// pointed to by a command line argument.
package config

import (
	"context"
	"flag"
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/go-sql-driver/mysql"
	"github.com/goradd/gro/pkg/config"
	"github.com/goradd/gro/pkg/db"
	mysql2 "github.com/goradd/gro/pkg/db/sql/mysql"
	"github.com/goradd/gro/pkg/db/sql/pgsql"
	"github.com/goradd/gro/pkg/db/sql/sqlite"
	"github.com/goradd/gro/pkg/schema"
	"github.com/jackc/pgx/v5"
)

// Default credentials for purposes of local development.
// Pass configuration overrides when doing CI testing or in production.
// DO NOT put live passwords here!
const defaultUser = "root"
const defaultPassword = "12345"
const goraddKey = "goradd"
const goraddDatabaseName = "goradd"
const goraddUnitKey = "goradd_unit"
const goraddUnitDatabaseName = "goradd_unit"

// InitDB initializes the database.
// It checks for a command line argument, and if present, treats it as a path to a configuration file
// with database settings.
func InitDB() {
	var configFile string
	flag.StringVar(&configFile, "c", "", "Path to database configuration file")
	flag.Parse()

	// If a config file is provided, use it instead
	if configFile != "" {
		if databaseConfigs, err := config.OpenConfigFile(configFile); err != nil {
			panic(err)
		} else if err := config.InitDatastore(databaseConfigs); err != nil {
			panic(err)
		}
		return
	}

	// pick a database to initialize here if no config file
	//initMysql()
	initPostgres()
	//initSQLite()
}

func initMysql() {
	cfg := mysql.NewConfig()
	cfg.ParseTime = true
	cfg.DBName = goraddDatabaseName
	cfg.User = defaultUser
	cfg.Passwd = defaultPassword
	cfg.Net = "tcp"
	cfg.Addr = "127.0.0.1:3307"

	database, err := mysql2.NewDB(goraddKey, "", cfg)
	if err != nil {
		panic(err)
	}
	db.AddDatabase(database, goraddKey)

	cfg.DBName = goraddUnitDatabaseName
	database, err = mysql2.NewDB(goraddUnitKey, "", cfg)
	if err != nil {
		panic(err)
	}
	db.AddDatabase(database, goraddUnitKey)
	database.StartProfiling()
}

func initPostgres() {
	cfg, _ := pgx.ParseConfig("")

	cfg.Host = "127.0.0.1"
	cfg.User = defaultUser
	cfg.Password = defaultPassword
	cfg.Database = goraddDatabaseName

	database, err := pgsql.NewDB(goraddKey, "", cfg)
	if err != nil {
		panic(err)
	}

	database.StartProfiling()

	db.AddDatabase(database, goraddKey)

	cfg.Database = goraddUnitDatabaseName
	database, err = pgsql.NewDB(goraddUnitKey, "", cfg)
	if err != nil {
		panic(err)
	}
	db.AddDatabase(database, goraddUnitKey)
}

func initSQLite() {
	database, err := sqlite.NewDB(goraddKey, ":memory:")
	if err != nil {
		panic(err)
	}
	db.AddDatabase(database, goraddKey)

	path := testDir()
	fmt.Println(path)

	s, err := schema.ReadJsonFile(filepath.Join(path, "schema", "goradd_schema.json"))
	if err != nil {
		panic(err)
	}
	ctx := context.Background()
	err = s.Clean()
	if err != nil {
		panic(err)
	}
	err = database.CreateSchema(ctx, *s)
	if err != nil {
		panic(err)
	}

	database, err = sqlite.NewDB(goraddUnitKey, ":memory:")
	if err != nil {
		panic(err)
	}
	db.AddDatabase(database, goraddUnitKey)

	s, err = schema.ReadJsonFile(filepath.Join(path, "schema", "goraddunit_schema.json"))
	if err != nil {
		panic(err)
	}
	err = s.Clean()
	if err != nil {
		panic(err)
	}
	err = database.CreateSchema(ctx, *s)
	if err != nil {
		panic(err)
	}
}

func testDir() string {
	// skip=0 means "this call site" (inside currentSourceDir)
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		panic("unable to get caller info")
	}
	t := filepath.Dir(file)
	return filepath.Dir(t)
}
