package pgsql

import (
	"context"
	"github.com/goradd/orm/pkg/db"
	"github.com/goradd/orm/pkg/schema"
	"log/slog"
)

// DestroySchema removes all tables and data from the tables found in the given schema s.
func (m *DB) DestroySchema(ctx context.Context, s schema.Database) (err error) {
	// gather table names to delete
	var tables []string

	// build in creation order
	for _, table := range s.EnumTables {
		tables = append(tables, table.QualifiedName())
	}
	for _, table := range s.Tables {
		tables = append(tables, table.QualifiedName())
	}
	for _, table := range s.AssociationTables {
		tables = append(tables, table.QualifiedName())
	}

	// iterate in the reverse order of creation
	for i := len(tables) - 1; i >= 0; i-- {
		table := tables[i]
		_, err := m.SqlExec(ctx, `DROP TABLE `+m.QuoteIdentifier(table))
		if err != nil {
			slog.Error("failed to drop table",
				slog.String(db.LogTable, table),
				slog.Any(db.LogError, err),
			)
		}
	}
	return nil
}
