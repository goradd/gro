package pgsql

import (
	"context"
)

// DestroySchema removes all tables and data from the database.
// Note that this is not limited to what was previously read by ExtractSchema, but rather
// drops all the tables that are currently found.
// The entire process is in a transaction.
func (m *DB) DestroySchema(ctx context.Context) (err error) {
	rawTables := m.getRawTables(nil)
	for _, table := range rawTables {
		_, err = m.SqlExec(ctx, `DROP TABLE `+m.QuoteIdentifier(table.name))
		if err != nil {
			return err
		}
	}
	return nil
}
