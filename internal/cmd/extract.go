package cmd

import (
	"fmt"

	db2 "github.com/goradd/gro/db"
	"github.com/goradd/gro/internal/config"
	"github.com/goradd/gro/schema"
)

func Extract(dbConfigFile, schemaFile, dbKey string) error {
	if databaseConfigs, err := config.OpenConfigFile(dbConfigFile); err != nil {
		return err
	} else if err = config.InitDatastore(databaseConfigs); err != nil {
		return err
	} else {
		for _, c := range databaseConfigs {
			if c["key"].(string) == dbKey {
				setDefaultConfigSettings(c)
				db := db2.GetDatabase(dbKey)
				if db == nil {
					return fmt.Errorf("database for key %s not found", dbKey)
				}
				if e, ok := db.(db2.SchemaExtractor); !ok {
					return fmt.Errorf("database for key %s is not a SchemaExtractor", dbKey)
				} else {
					s := e.ExtractSchema(c)
					s.Sort()
					return schema.WriteJsonFile(&s, schemaFile)
				}
			}
		}
	}
	return nil
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
