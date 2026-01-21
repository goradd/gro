package config

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/go-sql-driver/mysql"
	"github.com/goradd/gro/db"
	mysql2 "github.com/goradd/gro/db/sql/mysql"
	"github.com/goradd/gro/db/sql/pgsql"
	"github.com/goradd/gro/db/sql/sqlite"
	"github.com/goradd/strings"
	"github.com/jackc/pgx/v5"
)

type ConfigFormat struct {
	Databases []map[string]any `json:"databases"`
}

func NewDatabase(config map[string]any) (database db.DatabaseI, err error) {
	typ := config["type"].(string)
	if typ == "" {
		slog.Error(`missing "type" value for database `)
	}
	key := config["key"].(string)
	if key == "" {
		slog.Error(`missing "key" value for database `)
	}

	switch typ {
	case db.DriverTypeMysql:
		database, err = initMysql(config)
	case db.DriverTypePostgres:
		database, err = initPgsql(config)
	case db.DriverTypeSQLite:
		database, err = initSQLite(config)

	}
	return
}

func initMysql(overrides map[string]any) (db1 db.DatabaseI, err error) {
	cfg := mysql.NewConfig()
	cfg.ParseTime = true
	mysql2.OverrideConfigSettings(cfg, overrides)
	key := overrides["key"].(string)

	db1, err = mysql2.NewDB(key, "", cfg)
	return db1, err
}

func initPgsql(overrides map[string]any) (db1 db.DatabaseI, err error) {
	cfg, _ := pgx.ParseConfig("")
	pgsql.OverrideConfigSettings(cfg, overrides)
	key := overrides["key"].(string)

	db1, err = pgsql.NewDB(key, "", cfg)
	return db1, err
}

func initSQLite(overrides map[string]any) (db1 db.DatabaseI, err error) {
	key := overrides["key"].(string)
	dsn := overrides["dsn"].(string)

	db1, err = sqlite.NewDB(key, dsn)
	return db1, err
}

func OpenConfigFile(path string) (databaseConfigs []map[string]any, err error) {
	var b []byte

	b, err = os.ReadFile(path)
	if err != nil {
		return
	}
	if err = json.Unmarshal(b, &databaseConfigs); err != nil {
		return
	}

	for i, dbConfig := range databaseConfigs {
		typ := dbConfig["type"].(string)
		if typ == "" {
			err = fmt.Errorf(`missing "type" value for database %d`, i)
			return
		}
		key := dbConfig["key"].(string)
		if key == "" {
			err = fmt.Errorf(`missing "key" value for database %d`, i)
			return
		}
		if !strings.IsSnake(key) {
			err = fmt.Errorf(`"key" value "%s" must be lower_snake_case`, key)
			return
		}
	}
	return
}

func InitDatastore(configs []map[string]interface{}) error {
	for _, c := range configs {
		db2, err := NewDatabase(c)
		if err != nil {
			return err
		}
		db.AddDatabase(db2, c["key"].(string))
	}
	return nil
}
