package cmd

import (
	"context"
	"fmt"
	"log/slog"

	db2 "github.com/goradd/gro/db"
	"github.com/goradd/gro/internal/config"
	"github.com/goradd/gro/schema"
)

func Rebuild(dbConfigFile, schemaFile, dbKey string) error {
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
					return fmt.Errorf("database not found for key %s", dbKey)
				}
				if e, ok := db.(db2.SchemaRebuilder); ok {
					s, err2 := schema.ReadJsonFile(schemaFile)
					if err2 != nil {
						return err2
					}
					ctx := context.Background()
					err = db2.WithConstraintsOff(ctx, db, func(ctx context.Context) error {
						err2 := e.DestroySchema(ctx, *s)
						if err2 != nil {
							return err2
						}
						err2 = s.Clean()
						if err2 != nil {
							return err2
						}
						return e.CreateSchema(ctx, *s)
					})
					if err != nil {
						return err
					}
					return nil
				} else {
					slog.Error("Database cannot rebuild a schema.",
						slog.String(db2.LogDatabase, dbKey))
					return fmt.Errorf("database for key %s does not have Rebuild capabilities", dbKey)
				}
			}
		}
	}
	return nil
}
