package sql

import (
	"database/sql"
	"fmt"
	"github.com/goradd/gro/db"
	"github.com/goradd/gro/db/jointree"
	"github.com/goradd/gro/query"
)

type sqlCursor struct {
	rows                 *sql.Rows
	columnTypes          []query.ReceiverType
	columnNames          []string
	joinTree             *jointree.JoinTree
	columnReceivers      []SqlReceiver
	columnValueReceivers []interface{}
	query                string
	args                 []any
}

// NewSqlCursor is a new cursor created from the result of a sql query.
// columnTypes are the receiver types in the order of the query.
// columnNames are optional, and if left off, will be taken from the rows value as described in the database.
// builder is optional, and is used for unpacking joined tables.
func NewSqlCursor(rows *sql.Rows,
	columnTypes []query.ReceiverType,
	columnNames []string,
	joinTree *jointree.JoinTree,
	sql string,
	args []any,
) query.CursorI {
	var err error

	if rows == nil {
		panic("rows cannot be nil")
	}

	if columnNames == nil {
		columnNames, err = rows.Columns()
		if err != nil {
			_ = rows.Close()
			panic(fmt.Errorf("error getting column names from sql result: %w", err)) // a framework error, rows were closed before we got here
		}
		if len(columnNames) < len(columnTypes) {
			_ = rows.Close()
			panic(fmt.Errorf("column names length mismatch, expected at least %d, got %d", len(columnTypes), len(columnNames)))
		}
	} else {
		if len(columnNames) < len(columnTypes) {
			_ = rows.Close()
			panic(fmt.Errorf("column names length mismatch, expected at least %d, got %d", len(columnTypes), len(columnNames)))
		}
	}

	cursor := sqlCursor{
		rows:                 rows,
		columnTypes:          columnTypes,
		columnNames:          columnNames,
		joinTree:             joinTree,
		columnReceivers:      make([]SqlReceiver, len(columnTypes)),
		columnValueReceivers: make([]interface{}, len(columnTypes)),
		query:                sql,
		args:                 args,
	}

	for i := range cursor.columnReceivers {
		cursor.columnValueReceivers[i] = &(cursor.columnReceivers[i].R)
	}
	return &cursor
}

// Next returns the values of the next row in the result set.
// Returns nil if there are no more rows in the result set.
//
// The returned map is keyed by column name, which is either the column names provided
// when the cursor was created, or taken from the database itself.
//
// If an error occurs, will panic with the error.
func (r *sqlCursor) Next() (map[string]interface{}, error) {
	var err error

	if r == nil || r.rows == nil {
		return nil, nil
	}
	if r.rows.Next() {
		if err = r.rows.Scan(r.columnValueReceivers...); err != nil {
			return nil, db.NewQueryError("Scan", r.query, r.args, err)
		}

		values := make(map[string]interface{}, len(r.columnReceivers))
		for j, vr := range r.columnReceivers {
			values[r.columnNames[j]] = vr.Unpack(r.columnTypes[j])
		}
		if r.joinTree != nil {
			v2 := unpack(r.joinTree, []map[string]interface{}{values})
			return v2[0], nil
		} else {
			return values, nil
		}
	} else {
		if err = r.rows.Err(); err != nil {
			return nil, db.NewQueryError("rows.Err", r.query, r.args, err)
		}
		return nil, nil
	}
}

// Close closes the cursor.
//
// Once you are done with the cursor, you MUST call Close, so it is
// probably best to put a defer Close statement ahead of using Next.
func (r *sqlCursor) Close() error {
	if r == nil || r.rows == nil {
		return nil
	}

	if err := r.rows.Close(); err != nil {
		return db.NewQueryError("Cursor Close", r.query, r.args, err)
	}
	return nil
}
