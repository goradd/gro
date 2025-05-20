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

// DbHelper is a mixin for SQL database drivers.
// It implements common code needed by all SQL database drivers and default implementations of database code.
type DbHelper struct {
	dbKey     string  // key of the database as used in the global database map
	db        *sql.DB // Internal copy of a Go database/sql object
	dbi       DbI
	profiling bool
}

// NewSqlHelper creates a default DbHelper mixin.
func NewSqlHelper(dbKey string, db *sql.DB, dbi DbI) DbHelper {
	s := DbHelper{
		dbKey: dbKey,
		db:    db,
		dbi:   dbi,
	}
	return s
}

// Begin starts a transaction. You should immediately defer a Rollback using the returned transaction id.
// If you Commit before the Rollback happens, no Rollback will occur. The Begin-Commit-Rollback pattern is nestable.
func (h *DbHelper) Begin(ctx context.Context) (txid db.TransactionID, err error) {
	c := h.getSqlContext(ctx)
	if c == nil {
		panic("Can't use transactions without pre-loading a context")
	}
	c.txCount++

	if c.txCount == 1 {
		c.tx, err = h.db.BeginTx(ctx, nil)
		if err != nil {
			c.txCount-- // transaction did not begin
			return db.TransactionID(-1), db.NewQueryError("Begin", "", nil, err)
		}
	}
	return db.TransactionID(c.txCount), nil
}

// Commit commits the transaction, and if an error occurs, will panic with the error.
func (h *DbHelper) Commit(ctx context.Context, txid db.TransactionID) error {
	c := h.getSqlContext(ctx)
	if c == nil {
		panic("Can't use transactions without pre-loading a context")
	}

	if c.txCount != int(txid) {
		panic("Invalid transaction ID. Probably did not call Rollback after calling Begin in a previous wrapper")
	}

	if c.txCount == 0 {
		panic("Called Commit without a matching Begin")
	}

	if c.txCount == 1 {
		err := c.tx.Commit()
		if err != nil {
			return db.NewQueryError("Commit", "", nil, err)
		}
		c.tx = nil
	}
	c.txCount--
	return nil
}

// Rollback will rollback the transaction if the transaction is still pointing to the given txid. This gives the effect
// that if you call Rollback on a transaction that has already been committed, no Rollback will happen. This makes it easier
// to implement a transaction management scheme, because you simply always defer a Rollback after a Begin. Pass the txid
// that you got from the Begin to the Rollback. To trigger a Rollback, simply panic or exit the function.
func (h *DbHelper) Rollback(ctx context.Context, txid db.TransactionID) error {
	c := h.getSqlContext(ctx)
	if c == nil {
		panic("Can't use transactions without pre-loading a context")
	}

	if c.txCount == int(txid) {
		err := c.tx.Rollback()
		c.txCount = 0
		c.tx = nil
		if err != nil {
			return db.NewQueryError("Rollback", "", nil, err)
		}
	}
	return nil
}

// SqlExec executes the given SQL code, without returning any result rows.
func (h *DbHelper) SqlExec(ctx context.Context, sql string, args ...interface{}) (r sql.Result, err error) {
	c := h.getSqlContext(ctx)
	slog.Debug("SqlExec: ",
		slog.String(db.LogSql, sql),
		slog.Any(db.LogArgs, args))

	var beginTime = time.Now()
	if c != nil && c.tx != nil {
		r, err = c.tx.ExecContext(ctx, sql, args...)
	} else {
		r, err = h.db.ExecContext(ctx, sql, args...)
	}

	var endTime = time.Now()

	if c != nil && h.profiling {
		if args != nil {
			for _, arg := range args {
				sql = strings.TrimSpace(sql)
				sql += fmt.Sprintf(",\n%#v", arg)
			}
		}
		c.profiles = append(c.profiles, ProfileEntry{DbKey: h.dbKey, BeginTime: beginTime, EndTime: endTime, Typ: "SqlExec", Sql: sql})
	}

	return
}

/*
func (s *DbHelper) Prepare(ctx context.Context, sql string) (r *sql.Stmt, err error) {
	var c *sqlContext
	i := ctx.Value(goradd.sqlContext)
	if i != nil {
		c = i.(*sqlContext)
	}

	var beginTime = time.Now()
	if c != nil && c.tx != nil {
		r, err = c.tx.Prepare(sql)
	} else {
		r, err = s.db.Prepare(sql)
	}
	var endTime = time.Now()
	if c != nil && s.profiling {
		c.profiles = append(c.profiles, ProfileEntry{Key: s.dbKey, BeginTime: beginTime, EndTime: endTime, Typ: "Prepare", Sql: sql})
	}

	return
}*/

// SqlQuery executes the given sql, and returns a row result set.
func (h *DbHelper) SqlQuery(ctx context.Context, sql string, args ...interface{}) (r *sql.Rows, err error) {
	c := h.getSqlContext(ctx)
	slog.Debug("SqlExec: ",
		slog.String(db.LogSql, sql),
		slog.Any(db.LogArgs, args))

	var beginTime = time.Now()
	if c != nil && c.tx != nil {
		r, err = c.tx.QueryContext(ctx, sql, args...)
	} else {
		r, err = h.db.QueryContext(ctx, sql, args...)
	}
	var endTime = time.Now()
	if c != nil && h.profiling {
		if args != nil {
			for _, arg := range args {
				sql = strings.TrimSpace(sql)
				sql += fmt.Sprintf(",\n%#v", arg)
			}
		}
		c.profiles = append(c.profiles, ProfileEntry{DbKey: h.dbKey, BeginTime: beginTime, EndTime: endTime, Typ: "SqlQuery", Sql: sql})
	}

	return
}

// NewContext puts a blank context into the context chain to track transactions and other
// special database situations.
func (h *DbHelper) NewContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, h.contextKey(), &sqlContext{})
}

func (h *DbHelper) contextKey() contextKey {
	return contextKey("db-" + h.DbKey())
}

func (h *DbHelper) getSqlContext(ctx context.Context) *sqlContext {
	i := ctx.Value(h.contextKey())
	if i != nil {
		if c, ok := i.(*sqlContext); ok {
			return c
		}
	}
	return nil
}

// IsInTransaction returns true if the database is in the middle of a transaction.
func (h *DbHelper) IsInTransaction(ctx context.Context) (inTx bool) {
	c := h.getSqlContext(ctx)
	if c != nil && c.txCount > 0 {
		inTx = true
	}
	return
}

// DbKey returns the key of the database in the global database store.
func (h *DbHelper) DbKey() string {
	return h.dbKey
}

// SqlDb returns the underlying database/sql database object.
func (h *DbHelper) SqlDb() *sql.DB {
	return h.db
}

// StartProfiling will start the database profiling process.
func (h *DbHelper) StartProfiling() {
	h.profiling = true
}

// GetProfiles returns currently collected profile information
// TODO: Move profiles to a session variable so we can access ajax queries too
func (h *DbHelper) GetProfiles(ctx context.Context) []ProfileEntry {
	c := h.getSqlContext(ctx)
	if c == nil {
		panic("Profiling requires a preloaded context.")
	}

	p := c.profiles
	c.profiles = nil
	return p
}

// Query queries table for fields and returns a cursor that can be used to scan the result set.
// If where is provided, it will limit the result set to rows with fields that match the where values.
// If orderBy is provided, the result set will be sorted in ascending order by the fields indicated there.
// The returned cursor must eventually be closed.
func (h *DbHelper) Query(ctx context.Context, table string, fields map[string]ReceiverType, where map[string]any, orderBy []string) (CursorI, error) {
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
func (h *DbHelper) Delete(ctx context.Context, table string, colName string, colValue any, optLockFieldName string, optLockFieldValue int64) error {
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
func (h *DbHelper) DeleteAll(ctx context.Context, table string) error {
	s, args := GenerateDelete(h.dbi, table, nil, false)
	_, e := h.SqlExec(ctx, s, args...)
	if e != nil {
		return db.NewQueryError("SqlExec", s, args, e)
	}
	return nil
}

func (h *DbHelper) CheckLock(ctx context.Context,
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
func (h *DbHelper) BuilderQuery(ctx context.Context, builder *Builder) (ret any, err error) {
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

func (h *DbHelper) joinTreeLoad(ctx context.Context, joinTree *jointree.JoinTree) ([]map[string]any, error) {
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
func (h *DbHelper) joinTreeLoadCursor(ctx context.Context, joinTree *jointree.JoinTree) (any, error) {
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

func (h *DbHelper) joinTreeCount(ctx context.Context, joinTree *jointree.JoinTree) (int, error) {
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

func (h *DbHelper) CreateSchema(ctx context.Context, s schema.Database) error {
	for _, table := range s.EnumTables {
		if err := h.buildEnum(ctx, &s, table); err != nil {
			return err
		}
	}
	for _, table := range s.Tables {
		if err := h.buildTable(ctx, &s, table); err != nil {
			return err
		}
	}
	for _, table := range s.AssociationTables {
		if err := h.buildAssociation(ctx, &s, table); err != nil {
			return err
		}
	}

	return nil
}

func (h *DbHelper) buildTable(ctx context.Context, d *schema.Database, table *schema.Table) (err error) {
	s := h.tableSql(d, table)
	if s == "" {
		return nil // already reported error as table skipped
	}
	_, err = h.dbi.SqlExec(ctx, s)
	if err != nil {
		slog.Error("SQL error in buildTable.",
			slog.String("sql", s),
			slog.Any("error", err))
	}
	return err
}

func (h *DbHelper) buildEnum(ctx context.Context, d *schema.Database, table *schema.EnumTable) (err error) {
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
	return err
}

func (h *DbHelper) buildAssociation(ctx context.Context, d *schema.Database, table *schema.AssociationTable) (err error) {
	s := h.associationSql(d, table)
	if s == "" {
		return fmt.Errorf("error in table `%s`", table.Name)
	}
	_, err = h.dbi.SqlExec(ctx, s)
	if err != nil {
		slog.Error("SQL error",
			slog.String("sql", s),
			slog.Any("error", err),
		)
	}
	return err
}

func (h *DbHelper) tableSql(d *schema.Database, table *schema.Table) string {
	return h.dbi.TableDefinitionSql(d, table)
}

// enumTableSql returns the sql to create an enum table.
func (h *DbHelper) enumTableSql(d *schema.Database, et *schema.EnumTable) (s string) {
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
			Name:             k,
			Type:             et.Fields[k].Type,
			Size:             size,
			Identifier:       et.Fields[k].Identifier,
			IdentifierPlural: et.Fields[k].IdentifierPlural,
		}
		if i == 0 {
			col.IndexLevel = schema.IndexLevelManualPrimaryKey
		}
		table.Columns = append(table.Columns, col)
	}

	return h.tableSql(d, table)
}

func (h *DbHelper) enumValueSql(tableName string, fieldKeys []string, fields map[string]schema.EnumField, v map[string]any) (sql string, args []any) {
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

func (h *DbHelper) associationSql(d *schema.Database, at *schema.AssociationTable) string {
	// Build a schema table to create the association table
	table := &schema.Table{
		Name:    at.Name,
		Schema:  at.Schema,
		Comment: at.Comment,
	}

	// Make columns to send to the column builder
	col := &schema.Column{
		Name: at.Column1,
		Type: schema.ColTypeReference,
		Reference: &schema.Reference{
			Table: at.Table1,
		},
	}
	table.Columns = append(table.Columns, col)

	col = &schema.Column{
		Name:       at.Column2,
		Type:       schema.ColTypeReference,
		IndexLevel: schema.IndexLevelIndexed, // individual indexes for fast lookups
		Reference: &schema.Reference{
			Table: at.Table2,
		},
	}
	table.Columns = append(table.Columns, col)
	// multicolumn index for uniqueness and row id
	table.MultiColumnIndexes = []schema.MultiColumnIndex{
		{[]string{at.Column1, at.Column2}, schema.IndexLevelManualPrimaryKey},
	}
	return h.tableSql(d, table)
}
