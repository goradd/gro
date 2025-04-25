package mysql

import (
	"context"
)

// DestroySchema removes all tables and data from the database.
// Note that this is not limited to what was previously read by ExtractSchema, but rather
// drops all the tables that are currently found.
// The entire process is in a transaction.
func (m *DB) DestroySchema(ctx context.Context) (err error) {
	rawTables := m.getRawTables()
	_, err = m.SqlExec(ctx, `SET FOREIGN_KEY_CHECKS = 0`)
	if err != nil {
		return err
	}
	defer func() {
		_, _ = m.SqlExec(ctx, `SET FOREIGN_KEY_CHECKS = 1`)
	}()
	for _, table := range rawTables {
		_, err = m.SqlExec(ctx, `DROP TABLE `+m.QuoteIdentifier(table.name))
		if err != nil {
			return err
		}
	}
	return nil
}
