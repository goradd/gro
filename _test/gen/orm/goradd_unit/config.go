package goradd_unit

// Though this file is generated initially, if it exists, it will not be altered.
// Feel free to modify this file to suit your database configuration needs.

import (
	"encoding/json"
	"flag"
	"os"

	"github.com/go-sql-driver/mysql"
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
const key = "goradd_unit"
const databaseName = "goradd_unit"

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
// It checks for a "goradd" command line argument, and if present, treats it as a path to a configuration file
// with database settings.
func InitDB() {
	var configFile string
	flag.StringVar(&configFile, "goradd_unit", "", "Path to goradd_unit database configuration file")
	flag.Parse()

	var overrides map[string]any

	if configFile != "" {
		b, err := os.ReadFile(configFile)
		if err != nil {
			panic(err)
		}
		if err = json.Unmarshal(b, &overrides); err != nil {
			panic(err)
		}
	}

	// pick a database to initialize here
	initMysql(overrides)
	//initPostgres(overrides)
}
