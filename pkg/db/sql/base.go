package sql

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/goradd/anyutil"
	"github.com/goradd/orm/pkg/db"
	"github.com/goradd/orm/pkg/db/jointree"
	. "github.com/goradd/orm/pkg/query"
	"github.com/goradd/orm/pkg/schema"
	"log/slog"
	"strings"
	"time"
)

// Base is a mixin for SQL database drivers that implement the standard Go database/sql interface.
type Base struct {
	dbKey     string  // key of the database as used in the global database map
	db        *sql.DB // Internal copy of a Go database/sql object
	dbi       DbI
	profiling bool
}

// NewBase creates a default Base mixin.
func NewBase(dbKey string, db *sql.DB, dbi DbI) Base {
	s := Base{
		dbKey: dbKey,
		db:    db,
		dbi:   dbi,
	}
	return s
}

// SqlExec executes the given SQL code, without returning any result rows.
func (h *Base) SqlExec(ctx context.Context, sql string, args ...interface{}) (r sql.Result, err error) {
	var beginTime, endTime time.Time

	if h.profiling {
		beginTime = time.Now()
	}

	if tx := h.getTransaction(ctx); tx != nil {
		r, err = tx.ExecContext(ctx, sql, args...)
	} else if con := h.getConnection(ctx); con != nil {
		r, err = con.ExecContext(ctx, sql, args...)
	} else {
		r, err = h.db.ExecContext(ctx, sql, args...)
	}

	if h.profiling {
		endTime = time.Now()

		slog.Debug("SqlExec: ",
			slog.String(db.LogSql, sql),
			slog.Any(db.LogArgs, args),
			slog.Any(db.LogStartTime, beginTime),
			slog.Any(db.LogEndTime, endTime),
			slog.Any(db.LogDuration, endTime.Sub(beginTime)),
		)
	}

	return
}

// SqlQuery executes the given sql, and returns a row result set.
func (h *Base) SqlQuery(ctx context.Context, sql string, args ...interface{}) (r *sql.Rows, err error) {
	var beginTime, endTime time.Time

	if h.profiling {
		beginTime = time.Now()
	}

	if tx := h.getTransaction(ctx); tx != nil {
		r, err = tx.QueryContext(ctx, sql, args...)
	} else if con := h.getConnection(ctx); con != nil {
		r, err = con.QueryContext(ctx, sql, args...)
	} else {
		r, err = h.db.QueryContext(ctx, sql, args...)
	}
	if h.profiling {
		endTime = time.Now()

		slog.Debug("SqlQuery: ",
			slog.String(db.LogSql, sql),
			slog.Any(db.LogArgs, args),
			slog.Any(db.LogStartTime, beginTime),
			slog.Any(db.LogEndTime, endTime),
			slog.Any(db.LogDuration, endTime.Sub(beginTime)),
		)
	}

	return
}

// IsInTransaction returns true if the database is in the middle of a transaction.
func (h *Base) IsInTransaction(ctx context.Context) (inTx bool) {
	return h.getTransaction(ctx) != nil
}

// DbKey returns the key of the database in the global database store.
func (h *Base) DbKey() string {
	return h.dbKey
}

// SqlDb returns the underlying database/sql database object.
func (h *Base) SqlDb() *sql.DB {
	return h.db
}

// StartProfiling will start the database profiling process.
func (h *Base) StartProfiling() {
	h.profiling = true
}

// StopProfiling will start the database profiling process.
func (h *Base) StopProfiling() {
	h.profiling = false
}

// Query queries table for fields and returns a cursor that can be used to scan the result set.
// If where is provided, it will limit the result set to rows with fields that match the where values.
// If orderBy is provided, the result set will be sorted in ascending order by the fields indicated there.
// The returned cursor must eventually be closed.
func (h *Base) Query(ctx context.Context, table string, fields map[string]ReceiverType, where map[string]any, orderBy []string) (CursorI, error) {
	var fieldNames []string
	var receivers []ReceiverType

	for k, v := range fields {
		fieldNames = append(fieldNames, k)
		receivers = append(receivers, v)
	}
	s, args := GenerateSelect(h.dbi, table, fieldNames, where, orderBy)
	if rows, err := h.SqlQuery(ctx, s, args...); err != nil {
		if rows != nil {
			_ = rows.Close()
		}
		return nil, db.NewQueryError("SqlQuery", s, args, err)
	} else {
		return NewSqlCursor(rows, receivers, fieldNames, nil, s, args), nil
	}
}

// Delete deletes the indicated records from the database.
// Care should be exercised when calling this directly, since linked records are not modified in any way.
// If this record has linked records, the database structure may be corrupted.
func (h *Base) Delete(ctx context.Context, table string, colName string, colValue any, optLockFieldName string, optLockFieldValue int64) error {
	where := map[string]any{colName: colValue}
	if optLockFieldName != "" {
		// push where field down a level so it gets ANDed.
		where[optLockFieldName] = optLockFieldValue
	}
	s, args := GenerateDelete(h.dbi, table, where, false)
	result, e := h.SqlExec(ctx, s, args...)
	if e != nil {
		return db.NewQueryError("SqlExec", s, args, e)
	}
	if rows, _ := result.RowsAffected(); rows == 0 {
		if optLockFieldName != "" {
			return db.NewOptimisticLockError(table, colValue, nil)
		} else {
			return db.NewRecordNotFoundError(table, colValue)
		}
	}
	return nil
}

// DeleteAll deletes all the records from a table.
func (h *Base) DeleteAll(ctx context.Context, table string) error {
	s, args := GenerateDelete(h.dbi, table, nil, false)
	_, e := h.SqlExec(ctx, s, args...)
	if e != nil {
		return db.NewQueryError("SqlExec", s, args, e)
	}
	return nil
}

func (h *Base) CheckLock(ctx context.Context,
	table string,
	pkName string,
	pkValue any,
	optLockFieldName string,
	optLockFieldValue int64) (newLock int64, err error) {

	if optLockFieldName != "" {
		s, args := GenerateVersionLock(h.dbi, table, pkName, pkValue, optLockFieldName, h.IsInTransaction(ctx))
		var rows *sql.Rows
		if rows, err = h.SqlQuery(ctx, s, args...); err != nil {
			return 0, db.NewQueryError("SqlQuery", s, args, err)
		} else {
			var version int64
			defer RowClose(rows)
			if !rows.Next() {
				// The record was deleted prior to an update completing.
				return 0, db.NewRecordNotFoundError(table, pkValue)
			}
			if err = rows.Scan(&version); err != nil {
				return 0, db.NewQueryError("Scan", s, args, err)
			}
			if version != optLockFieldValue {
				// The record was changed prior to an update completing.
				err = db.NewOptimisticLockError(table, pkValue, nil)
				return
			}
			// If we get here, and we are in a transaction, the record has been locked until the end of the transaction and optimistic locking is valid.
			// If not in a transaction, we know that so far the record has not changed, but it still has a slight chance of changing between here
			// and the execution of the GenerateUpdate below.
			// Generate a new version number prior to saving.
			newLock = db.RecordVersion(optLockFieldValue)
		}
	}
	return
}

// BuilderQuery performs a complex query using a query builder.
// The data returned will depend on the command inside the builder.
// Be sure when using BuilderCommandLoadCursor you close the returned cursor, probably with a defer command.
func (h *Base) BuilderQuery(ctx context.Context, builder *Builder) (ret any, err error) {
	joinTree := jointree.NewJoinTree(builder)
	switch joinTree.Command {
	case BuilderCommandLoad:
		ret, err = h.joinTreeLoad(ctx, joinTree)
	case BuilderCommandLoadCursor:
		ret, err = h.joinTreeLoadCursor(ctx, joinTree)
	case BuilderCommandCount:
		ret, err = h.joinTreeCount(ctx, joinTree)
	}
	return
}

func (h *Base) joinTreeLoad(ctx context.Context, joinTree *jointree.JoinTree) ([]map[string]any, error) {
	g := newSqlGenerator(joinTree, h.dbi)
	s, args := g.generateSelectSql()

	rows, err := h.dbi.SqlQuery(ctx, s, args...)
	if err != nil {
		return nil, db.NewQueryError("SqlQuery", s, args, err)
	}
	defer RowClose(rows)

	var names []string
	names, err = rows.Columns()
	if err != nil {
		return nil, db.NewQueryError("Columns", s, args, err)
	}

	// prepare the selected columns for unpacking
	columnTypes := make([]ReceiverType, 0, len(names))
	for sel := range joinTree.SelectsIter() {
		t := sel.QueryNode.(*ColumnNode).ReceiverType
		columnTypes = append(columnTypes, t)
	}
	// add special aliases
	for i := len(columnTypes); i < len(names); i++ {
		columnTypes = append(columnTypes, ColTypeBytes) // These will be unpacked when they are retrieved
	}

	return ReceiveRows(rows, columnTypes, names, joinTree, s, args)
}

// The cursor returned must be closed by the caller.
func (h *Base) joinTreeLoadCursor(ctx context.Context, joinTree *jointree.JoinTree) (any, error) {
	g := newSqlGenerator(joinTree, h.dbi)
	s, args := g.generateSelectSql()
	rows, err := h.dbi.SqlQuery(ctx, s, args...)
	if err != nil {
		return nil, db.NewQueryError("SqlQuery", s, args, err)
	}

	names, _ := rows.Columns()
	columnTypes := make([]ReceiverType, 0, len(names))
	for sel := range joinTree.SelectsIter() {
		t := sel.QueryNode.(*ColumnNode).ReceiverType
		columnTypes = append(columnTypes, t)
	}
	// add special aliases
	for i := len(columnTypes); i < len(names); i++ {
		columnTypes = append(columnTypes, ColTypeBytes) // These will be unpacked when they are retrieved
	}
	return NewSqlCursor(rows, columnTypes, nil, joinTree, s, args), nil
}

func (h *Base) joinTreeCount(ctx context.Context, joinTree *jointree.JoinTree) (int, error) {
	g := newSqlGenerator(joinTree, h.dbi)
	s, args := g.generateCountSql()
	rows, err := h.dbi.SqlQuery(ctx, s, args...)
	if err != nil {
		return 0, db.NewQueryError("SqlQuery", s, args, err)
	}
	defer RowClose(rows)

	names, _ := rows.Columns()
	columnTypes := []ReceiverType{ColTypeInteger}
	var result []map[string]any
	result, err = ReceiveRows(rows, columnTypes, names, nil, s, args)
	if err != nil {
		return 0, err
	}
	ret := result[0][names[0]].(int)
	return ret, nil
}

func (h *Base) CreateSchema(ctx context.Context, s schema.Database) error {
	if err := h.buildEnums(ctx, &s, s.EnumTables); err != nil {
		return err
	}
	if err := h.buildTables(ctx, &s, s.Tables); err != nil {
		return err
	}
	if err := h.buildAssociations(ctx, &s, s.AssociationTables); err != nil {
		return err
	}

	return nil
}

func (h *Base) buildTables(ctx context.Context, d *schema.Database, tables []*schema.Table) (err error) {
	var extras string
	for _, table := range tables {
		s, e := h.tableSql(d, table)
		if e != "" {
			extras += e + ";\n"
		}
		if s == "" {
			continue // already reported error
		}
		_, err = h.dbi.SqlExec(ctx, s)
		if err != nil {
			slog.Error("SQL error in buildTables.",
				slog.String("sql", s),
				slog.Any("error", err))
		}
	}
	if err == nil && extras != "" {
		_, err = h.dbi.SqlExec(ctx, extras)
	}
	return err
}

func (h *Base) buildEnums(ctx context.Context, d *schema.Database, tables []*schema.EnumTable) (err error) {
	for _, table := range tables {
		var args []any
		s := h.enumTableSql(d, table)
		if s == "" {
			return fmt.Errorf("error in table `%s`", table.Name)
		}
		if _, err = h.dbi.SqlExec(ctx, s); err != nil {
			slog.Error("SQL error",
				slog.String("sql", s),
				slog.Any("error", err))

			return
		}

		fieldKeys := table.FieldKeys()
		for _, v := range table.Values {
			s, args = h.enumValueSql(table.Name, fieldKeys, table.Fields, v)
			if _, err = h.dbi.SqlExec(ctx, s, args...); err != nil {
				slog.Error("SQL error",
					slog.String("sql", s),
					slog.Any("error", err),
					slog.Any("args", args))

				return
			}
		}
	}
	return nil
}

func (h *Base) buildAssociations(ctx context.Context, d *schema.Database, table []*schema.AssociationTable) (err error) {
	for _, table := range table {
		s := h.associationSql(d, table)
		if s == "" {
			return fmt.Errorf("error in table `%s`", table.Table)
		}
		_, err = h.dbi.SqlExec(ctx, s)
		if err != nil {
			slog.Error("SQL error",
				slog.String("sql", s),
				slog.Any("error", err),
			)
		}
	}
	return err
}

func (h *Base) tableSql(d *schema.Database, table *schema.Table) (string, string) {
	return h.dbi.TableDefinitionSql(d, table)
}

// enumTableSql returns the sql to create an enum table.
func (h *Base) enumTableSql(d *schema.Database, et *schema.EnumTable) (s string) {
	// Build a schema table to create the enum table
	table := &schema.Table{
		Name:    et.Name,
		Schema:  et.Schema,
		Comment: et.Comment,
	}

	for i, k := range et.FieldKeys() {
		var size uint64
		for _, v := range et.Values {
			if et.Fields[k].Type == schema.ColTypeString ||
				et.Fields[k].Type == schema.ColTypeBytes {
				if s, ok := v[k].(string); ok {
					size = max(size, uint64(len(s)))
				}
			}
		}
		// build a column to send to the column builder
		col := &schema.Column{
			Name:       k,
			Type:       et.Fields[k].Type,
			Size:       size,
			Identifier: et.Fields[k].Identifier,
		}
		if i == 0 {
			table.Indexes = append(table.Indexes, schema.Index{IndexLevel: schema.IndexLevelPrimaryKey, Columns: []string{col.Name}})
		}
		table.Columns = append(table.Columns, col)
	}

	s, e := h.tableSql(d, table)
	return s + ";\n" + e
}

func (h *Base) enumValueSql(tableName string, fieldKeys []string, fields map[string]schema.EnumField, v map[string]any) (sql string, args []any) {
	var columns []string
	var placeholders []string
	for _, k := range fieldKeys {
		columns = append(columns, fmt.Sprintf("%s", h.dbi.QuoteIdentifier(k)))

		fieldType := fields[k].Type
		value := v[k]

		placeholders = append(placeholders, h.dbi.FormatArgument(len(placeholders)+1))

		switch fieldType {
		case schema.ColTypeString:
			if s, ok := value.(string); ok {
				args = append(args, s)
			} else {
				slog.Error("wrong type for enum value",
					slog.String(db.LogTable, tableName),
					slog.String(db.LogColumn, k))
				args = append(args, "")
			}

		case schema.ColTypeInt:
			if anyutil.IsInteger(value) {
				args = append(args, value)
			} else {
				slog.Error("wrong type for enum value",
					slog.String(db.LogTable, tableName),
					slog.String(db.LogColumn, k))
				args = append(args, 0)
			}

		case schema.ColTypeFloat:
			if anyutil.IsFloat(value) {
				args = append(args, value)
			} else {
				slog.Error("wrong type for enum value",
					slog.String(db.LogTable, tableName),
					slog.String(db.LogColumn, k))
				args = append(args, 0.0)
			}

		default:
			args = append(args, value)
		}
	}

	return fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		h.dbi.QuoteIdentifier(tableName),
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	), args
}

func (h *Base) associationSql(d *schema.Database, at *schema.AssociationTable) string {
	// Build a schema table to create the association table
	table := &schema.Table{
		Name:    at.Table,
		Schema:  at.Schema,
		Comment: at.Comment,
	}

	// Make columns to send to the column builder
	ref := &schema.Reference{
		Table:      at.Ref1.Table,
		Column:     at.Ref1.Column,
		IndexLevel: schema.IndexLevelIndexed, // individual indexes on columns, though its a composite primary key
	}
	table.References = append(table.References, ref)

	ref = &schema.Reference{
		Table:      at.Ref2.Table,
		Column:     at.Ref2.Column,
		IndexLevel: schema.IndexLevelIndexed, // individual indexes on columns, though its a composite primary key
	}
	table.References = append(table.References, ref)

	// multicolumn index for uniqueness and row id
	table.Indexes = []schema.Index{
		{[]string{at.Ref1.Column, at.Ref2.Column}, schema.IndexLevelPrimaryKey},
	}
	s, e := h.tableSql(d, table)
	return s + ";\n" + e
}

// DestroySchema removes all tables and data from the tables found in the given schema s.
func (h *Base) DestroySchema(ctx context.Context, s schema.Database) {
	// gather table names to delete
	var tables []string

	for _, table := range s.EnumTables {
		tables = append(tables, table.QualifiedTableName())
	}
	for _, table := range s.Tables {
		tables = append(tables, table.QualifiedName())
	}
	for _, table := range s.AssociationTables {
		tables = append(tables, table.QualifiedTableName())
	}

	_ = db.WithConstraintsOff(ctx, h.dbi.(db.DatabaseI), func(ctx2 context.Context) error {
		for _, table := range tables {
			_, err := h.SqlExec(ctx2, `DROP TABLE `+h.dbi.QuoteIdentifier(table))
			if err != nil {
				slog.Warn("failed to drop table",
					slog.String(db.LogTable, table),
					slog.Any(db.LogError, err),
				)
			}
		}
		return nil
	})
}

func (h *Base) connectionKey() contextKey {
	return contextKey("DbCon-" + h.DbKey())
}

func (h *Base) getConnection(ctx context.Context) *sql.Conn {
	i := ctx.Value(h.connectionKey())
	if i != nil {
		if c, ok := i.(*sql.Conn); ok {
			return c
		}
	}
	return nil
}

// WithSameConnection ensures that all database operations that happen inside the handler
// use the same database connection session.
//
// The Go SQL driver uses connection pooling. Each separate database operation may happen on a
// different connection in the pool, even if inside a transaction. However, this breaks some
// processes on particular databases. For example, in MySQL and SQLite, the process of turning off and
// on foreign key checks only works within the same connection, so trying to bracket that with some
// separate calls to the database will not work. The solution is here, to call WithSameConnection, which
// will pin a connection to the context within f. It is important to not have too many of these calls happening
// at the same time, or the connection pool may run out, causing subsequent database calls to block until one is freed.
//
// Nested calls will operate on the same connection.
func (h *Base) WithSameConnection(ctx context.Context, f func(ctx context.Context) error) (err error) {
	con := h.getConnection(ctx)
	if con == nil {
		tx := h.getTransaction(ctx)
		if tx != nil {
			panic("you cannot call WithSameConnection from within a transaction. Instead, call WithTransaction from within WithSameConnection")
		}
		con, err = h.db.Conn(ctx)
		if con != nil {
			defer func() {
				err = con.Close()
			}()
		}
		if err != nil {
			return err
		}
		ctx = context.WithValue(ctx, h.connectionKey(), con)
	}
	err = f(ctx)
	return
}

func (h *Base) transactionKey() contextKey {
	return contextKey("DbTx-" + h.DbKey())
}

func (h *Base) getTransaction(ctx context.Context) *sql.Tx {
	i := ctx.Value(h.transactionKey())
	if i != nil {
		if t, ok := i.(*sql.Tx); ok {
			return t
		}
	}
	return nil
}

// WithTransaction wraps the function f in a database transaction.
// While the ORM by default will wrap individual database calls with a timeout,
// it will not apply this timeout to a transaction. It is up to you to pass a context that
// has a timeout to prevent the overall transaction from hanging.
// Nested calls will operate within the same transaction, and the outermost call will determine
// when the transaction is finally committed.
func (h *Base) WithTransaction(ctx context.Context, f func(ctx context.Context) error) (err error) {
	tx := h.getTransaction(ctx)
	if tx == nil {
		con := h.getConnection(ctx)
		if con == nil {
			tx, err = h.db.BeginTx(ctx, nil)
		} else {
			tx, err = con.BeginTx(ctx, nil)
		}
		if err != nil {
			return
		}
		if tx == nil {
			return fmt.Errorf("no transaction available")
		}
		defer func() {
			err = tx.Rollback() // will be a no-op if the commit happens first
		}()
		ctx = context.WithValue(ctx, h.transactionKey(), tx)
	}
	err = f(ctx)
	if err != nil {
		return err
	}
	err = tx.Commit()
	return
}
