package config

import (
	"encoding/json"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"github.com/goradd/goradd/pkg/log"
	"github.com/goradd/orm/pkg/db"
	mysql2 "github.com/goradd/orm/pkg/db/sql/mysql"
	"github.com/goradd/orm/pkg/db/sql/pgsql"
	"github.com/goradd/strings"
	"github.com/jackc/pgx/v5"
	"os"
)

type ConfigFormat struct {
	Databases []map[string]any `json:"databases"`
}

func NewDatabase(config map[string]any) (database db.DatabaseI, err error) {
	typ := config["type"].(string)
	if typ == "" {
		log.Error(`missing "type" value for database `)
	}
	key := config["key"].(string)
	if key == "" {
		log.Error(`missing "key" value for database `)
	}

	switch typ {
	case "mysql":
		database = initMysql(config)
	case "pgsql":
		database = initPgsql(config)
	}
	return
}

func initMysql(overrides map[string]any) db.DatabaseI {
	cfg := mysql.NewConfig()
	cfg.ParseTime = true
	mysql2.OverrideConfigSettings(cfg, overrides)
	key := overrides["key"].(string)

	db1 := mysql2.NewDB(key, "", cfg)
	return db1
}

func initPgsql(overrides map[string]any) db.DatabaseI {
	cfg, _ := pgx.ParseConfig("")
	pgsql.OverrideConfigSettings(cfg, overrides)
	key := overrides["key"].(string)

	db1 := pgsql.NewDB(key, "", cfg)
	return db1
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
