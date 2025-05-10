// Package sql contains helper functions that connect a standard Go database/sql object
// to the GoRADD system.
//
// Most of the functionality in this package is used by database implementations. GoRADD users would
// not normally directly call functions in this package.
package sql

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/goradd/orm/pkg/db"
	"github.com/goradd/orm/pkg/db/jointree"
	. "github.com/goradd/orm/pkg/query"
	"io"
	"log/slog"
	"strings"
	"time"
)

// The DbI interface describes the interface that a sql database needs to implement so that
// it will work with the Builder object. If you know a DatabaseI object is a
// standard Go database/sql database, you can
// cast it to this type to gain access to the low level SQL driver and send it raw SQL commands.
type DbI interface {
	// SqlExec executes a query that does not expect to return values.
	// It will time out the query if contextTimeout is exceeded
	SqlExec(ctx context.Context, sql string, args ...interface{}) (r sql.Result, err error)
	// SqlQuery executes a SQL query that returns values.
	// It will time out the query if contextTimeout is exceeded
	SqlQuery(ctx context.Context, sql string, args ...interface{}) (r *sql.Rows, err error)
	// QuoteIdentifier will put quotes around an identifier in a database specific way.
	QuoteIdentifier(string) string
	// FormatArgument will return the placeholder string for the n'th argument
	// in a sql string. Some sqls just use "?" while others identify an argument
	// by location, like Postgres's $1, $2, etc.
	FormatArgument(n int) string
	// SupportsForUpdate will return true if it supports SELECT ... FOR UPDATE clauses for row level locking in a transaction
	SupportsForUpdate() bool
}

type contextKey string

// ProfileEntry contains the data collected during sql profiling
type ProfileEntry struct {
	DbKey     string
	BeginTime time.Time
	EndTime   time.Time
	Typ       string
	Sql       string
}

// sqlContext is what is stored in the current context to keep track of queries.
// You must save a copy of this in the
// current context with the sqlContext key before calling database functions in order to use transactions or
// database profiling, or anything else the context is required for. The framework does this for you, but you will need
// to do this yourself if using the orm without the framework.
type sqlContext struct {
	tx       *sql.Tx
	txCount  int // Keeps track of when to close a transaction
	profiles []ProfileEntry
}

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

func RowClose(c io.Closer) {
	if err := c.Close(); err != nil {
		slog.Warn("Error closing sql row cursor", slog.Any(db.LogError, err))
	}
}
