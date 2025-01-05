// Package ci provides continuous integration tests for goradd-orm.
// By default, the package is set up to provide local execution of the tests for developers
// of goradd-orm. See InitDB to choose the database that will be used in the tests.
//
// During the automated testing process, the database will be selected using a configuration file
// pointed to by a command line argument.
package ci

import (
	"flag"
	"github.com/go-sql-driver/mysql"
	"github.com/goradd/orm/pkg/config"
	"github.com/goradd/orm/pkg/db"
	mysql2 "github.com/goradd/orm/pkg/db/sql/mysql"
	"github.com/goradd/orm/pkg/db/sql/pgsql"
	"github.com/jackc/pgx/v5"
)

// Default credentials for purposes of local development.
// Pass configuration overrides when doing CI testing or in production.
// DO NOT put live passwords here!
const defaultUser = "root"
const defaultPassword = "12345"
const key = "goradd"
const databaseName = "goradd"

func initMysql(overrides map[string]any) {
	cfg := mysql.NewConfig()
	cfg.ParseTime = true
	cfg.DBName = databaseName
	cfg.User = defaultUser
	cfg.Passwd = defaultPassword
	mysql2.OverrideConfigSettings(cfg, overrides)

	database := mysql2.NewDB(key, "", cfg)
	db.AddDatabase(database, key)
}

func initPostgres(overrides map[string]any) {
	cfg, _ := pgx.ParseConfig("")

	cfg.Host = "localhost"
	cfg.User = defaultUser
	cfg.Password = defaultPassword
	cfg.Database = databaseName

	pgsql.OverrideConfigSettings(cfg, overrides)
	database := pgsql.NewDB(key, "", cfg)
	db.AddDatabase(database, key)
}

// InitDB initializes the database.
// It checks for a command line argument, and if present, treats it as a path to a configuration file
// with database settings.
func InitDB() {
	var configFile string
	flag.StringVar(&configFile, "c", "", "Path to database configuration file")
	flag.Parse()

	var overrides map[string]any

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
	initMysql(overrides)
	//initPostgres(overrides)
}
