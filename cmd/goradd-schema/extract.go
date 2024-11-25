package main

import (
	"spekary/goradd/orm/pkg/config"
	db2 "spekary/goradd/orm/pkg/db"
	"spekary/goradd/orm/pkg/model"
	"spekary/goradd/orm/pkg/schema"
)

func extract(dbConfigFile, outFile string) {
	if databaseConfigs, err := config.OpenConfigFile(dbConfigFile); err != nil {
		panic(err)
	} else if err := config.InitDatastore(databaseConfigs); err != nil {
		panic(err)
	} else {
		var schemas []*schema.Database
		for _, c := range databaseConfigs {
			db := db2.GetDatabase(c["key"].(string))
			if db == nil {
				panic("database not found")
			}
			if e, ok := db.(db2.SchemaExtractor); ok {
				if v, ok := c["reference_suffix"].(string); !ok || v == "" {
					c["reference_suffix"] = "_id"
				}
				if v, ok := c["enum_table_suffix"].(string); !ok || v == "" {
					c["enum_table_suffix"] = "_enum"
				}
				if v, ok := c["assn_table_suffix"].(string); !ok || v == "" {
					c["assn_table_suffix"] = "_assn"
				}
				s := e.ExtractSchema(c)
				s.FillDefaults()
				schemas = append(schemas, &s)
			}
		}

		schema.WriteJsonFile(schemas, outFile)
		m := model.FromSchemas(schemas)
		_ = m
	}
}
