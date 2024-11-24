package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"spekary/goradd/orm/pkg/config"
	db2 "spekary/goradd/orm/pkg/db"
	"spekary/goradd/orm/pkg/model"
	"spekary/goradd/orm/pkg/schema"
)

func extract(dbConfigFile, outFile string) {
	databaseConfigs, err := config.OpenConfigFile(dbConfigFile)
	if err != nil {
		panic(err)
	}
	var schemas []*schema.Database
	for _, c := range databaseConfigs {
		db, err := config.DatabaseFromConfig(c)
		if err != nil {
			panic(err)
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

	outputSchemas(schemas, outFile)
	m := model.FromSchemas(schemas)
	_ = m
}

func outputSchemas(schemas []*schema.Database, outFile string) {
	// Serialize the struct to JSON with indentation for readability
	jsonData, err := json.MarshalIndent(schemas, "", "  ")
	if err != nil {
		log.Fatal("Error marshaling JSON:", err)
		return
	}

	// Create or open a file for writing
	file, err := os.Create(outFile)
	if err != nil {
		log.Fatal("Error creating file:", err)
		return
	}
	defer file.Close()

	// Write the JSON data to the file
	_, err = file.Write(jsonData)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}
}
