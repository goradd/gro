package main

import (
	"github.com/goradd/orm/pkg/config"
	db2 "github.com/goradd/orm/pkg/db"
	"github.com/goradd/orm/pkg/schema"
)

func extract(dbConfigFile, outFile, dbKey string) {
	if databaseConfigs, err := config.OpenConfigFile(dbConfigFile); err != nil {
		panic(err)
	} else if err := config.InitDatastore(databaseConfigs); err != nil {
		panic(err)
	} else {
		for _, c := range databaseConfigs {
			if c["key"].(string) == dbKey {
				setDefaultConfigSettings(c)
				db := db2.GetDatabase(c["key"].(string))
				if db == nil {
					panic("database not found")
				}
				if e, ok := db.(db2.SchemaExtractor); ok {
					s := e.ExtractSchema(c)
					s.Sort()
					schema.WriteJsonFile(&s, outFile)
					return
				}
			}
		}
	}
}

func setDefaultConfigSettings(c map[string]interface{}) {
	if v, ok := c["reference_suffix"].(string); !ok || v == "" {
		c["reference_suffix"] = "_id"
	}
	if v, ok := c["enum_table_suffix"].(string); !ok || v == "" {
		c["enum_table_suffix"] = "_enum"
	}
	if v, ok := c["assn_table_suffix"].(string); !ok || v == "" {
		c["assn_table_suffix"] = "_assn"
	}
}
