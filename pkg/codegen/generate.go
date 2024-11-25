package codegen

import (
	"github.com/goradd/goradd/pkg/log"
	"os"
	"spekary/goradd/orm/pkg/model"
	"spekary/goradd/orm/pkg/schema"
)

func Generate(schemas []*schema.Database) {
	m := model.FromSchemas(schemas)
	for _, db := range m {
		genDB(db)
	}
}

func genDB(db *model.Database) {
	for _, t := range templates {
		filename := t.FileName()
		if !t.Overwrite() && fileExists(filename) {
			continue
		}
		f, err := os.Create(filename)
		if err != nil {
			log.Error(err)
			continue
		}

		if d, ok := t.(DatabaseGenerator); ok {
			d.GenerateDatabase(db, f)
		}
		err = f.Close()
		if err != nil {
			log.Error(err)
		}
	}
}

func fileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return false
	}
	return err == nil
}
