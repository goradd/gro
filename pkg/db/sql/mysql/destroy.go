package mysql

import (
	"context"
	"github.com/goradd/orm/pkg/schema"
)

// DestroySchema removes all tables and data from the tables found in the given schema s.
func (m *DB) DestroySchema(ctx context.Context, s schema.Database) (err error) {
	// gather table names to delete
	var tables []string

	for _, table := range s.AssociationTables {
		tables = append(tables, table.Name)
	}

	for _, table := range s.EnumTables {
		tables = append(tables, table.Name)
	}

	for _, table := range s.Tables {
		tables = append(tables, table.Name)
	}

	_, err = m.SqlExec(ctx, `SET FOREIGN_KEY_CHECKS = 0`)
	if err != nil {
		return err
	}
	defer func() {
		_, _ = m.SqlExec(ctx, `SET FOREIGN_KEY_CHECKS = 1`)
	}()
	for _, table := range tables {
		_, _ = m.SqlExec(ctx, `DROP TABLE `+m.QuoteIdentifier(table))
	}
	return nil
}
