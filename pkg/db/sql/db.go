// Package sql contains helper functions that connect a standard Go database/sql object
// to the GoRADD system.
//
// Most of the functionality in this package is used by database implementations. GoRADD users would
// not normally directly call functions in this package.
package sql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/goradd/orm/pkg/db"
	"github.com/goradd/orm/pkg/db/jointree"
	. "github.com/goradd/orm/pkg/query"
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
	SqlExec(ctx context.Context, sql string, args ...interface{}) (r sql.Result, err error)
	// SqlQuery executes a SQL query that returns values.
	SqlQuery(ctx context.Context, sql string, args ...interface{}) (r *sql.Rows, err error)
	// QuoteIdentifier will put quotes around an identifier in a database specific way.
	QuoteIdentifier(string) string
	// FormatArgument will return the placeholder string for the n'th argument
	// in a sql string. Some sqls just use "?" while others identify an argument
	// by location, like Postgres's $1, $2, etc.
	FormatArgument(n int) string
	// The driver should return true if it supports SELECT ... FOR UPDATE clauses for row level locking in a transaction
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

// NewSqlDb creates a default DbHelper mixin.
func NewSqlDb(dbKey string, db *sql.DB, dbi DbI) DbHelper {
	s := DbHelper{
		dbKey: dbKey,
		db:    db,
		dbi:   dbi,
	}
	return s
}

// Begin starts a transaction. You should immediately defer a Rollback using the returned transaction id.
// If you Commit before the Rollback happens, no Rollback will occur. The Begin-Commit-Rollback pattern is nestable.
// Pass true to forWrite if this is a transaction that will insert or update data. This will cause row locks to be
// write locks.
func (h *DbHelper) Begin(ctx context.Context) (txid db.TransactionID) {
	c := h.getContext(ctx)
	if c == nil {
		panic("Can't use transactions without pre-loading a context")
	}
	c.txCount++

	if c.txCount == 1 {
		var err error

		c.tx, err = h.db.Begin()
		if err != nil {
			_ = c.tx.Rollback()
			c.txCount-- // transaction did not begin
			panic(err.Error())
		}
	}
	return db.TransactionID(c.txCount)
}

// Commit commits the transaction, and if an error occurs, will panic with the error.
func (h *DbHelper) Commit(ctx context.Context, txid db.TransactionID) {
	c := h.getContext(ctx)
	if c == nil {
		panic("Can't use transactions without pre-loading a context")
	}

	if c.txCount != int(txid) {
		panic("Missing Rollback after previous Begin")
	}

	if c.txCount == 0 {
		panic("Called Commit without a matching Begin")
	}
	if c.txCount == 1 {
		err := c.tx.Commit()
		if err != nil {
			panic(err.Error())
		}
		c.tx = nil
	}
	c.txCount--
}

// Rollback will rollback the transaction if the transaction is still pointing to the given txid. This gives the effect
// that if you call Rollback on a transaction that has already been committed, no Rollback will happen. This makes it easier
// to implement a transaction management scheme, because you simply always defer a Rollback after a Begin. Pass the txid
// that you got from the Begin to the Rollback. To trigger a Rollback, simply panic.
func (h *DbHelper) Rollback(ctx context.Context, txid db.TransactionID) {
	c := h.getContext(ctx)
	if c == nil {
		panic("Can't use transactions without pre-loading a context")
	}

	if c.txCount == int(txid) {
		err := c.tx.Rollback()
		c.txCount = 0
		c.tx = nil
		if err != nil {
			panic(err.Error())
		}
	}
}

// SqlExec executes the given SQL code, without returning any result rows.
func (h *DbHelper) SqlExec(ctx context.Context, sql string, args ...interface{}) (r sql.Result, err error) {
	c := h.getContext(ctx)
	slog.Debug("SqlExec: ", sql, args)

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
	c := h.getContext(ctx)
	slog.Debug("SqlQuery: ", sql, args)

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

func (h *DbHelper) getContext(ctx context.Context) *sqlContext {
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
	c := h.getContext(ctx)
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
	c := h.getContext(ctx)
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
func (h *DbHelper) Query(ctx context.Context, table string, fields map[string]ReceiverType, where map[string]any, orderBy []string) CursorI {
	var fieldNames []string
	var receivers []ReceiverType
	for k, v := range fields {
		fieldNames = append(fieldNames, k)
		receivers = append(receivers, v)
	}
	s, args := GenerateSelect(h.dbi, table, fieldNames, where, orderBy)
	if rows, err := h.SqlQuery(ctx, s, args...); err != nil {
		panic(err.Error())
	} else {
		return NewSqlCursor(rows, receivers, fieldNames, nil)
	}
}

// Delete deletes the indicated record from the database.
// Care should be exercised when calling this directly, since related records are not modified in any way.
// If this record has related records, the database structure may be corrupted.
func (h *DbHelper) Delete(ctx context.Context, table string, where map[string]any) {
	s, args := GenerateDelete(h.dbi, table, where)
	_, e := h.SqlExec(ctx, s, args...)
	if e != nil {
		panic(e.Error())
	}
}

// Insert inserts the given data as a new record in the database.
// It returns the record id of the new record.
func (h *DbHelper) Insert(ctx context.Context, table string, fields map[string]interface{}) string {
	s, args := GenerateInsert(h.dbi, table, fields)
	if r, err := h.SqlExec(ctx, s, args...); err != nil {
		panic(err.Error())
	} else {
		// Not all database implementations support LastInsertId
		// If yours does not, you will need to override this implementation
		if id, err2 := r.LastInsertId(); err2 != nil {
			panic(err2.Error())
			return ""
		} else {
			return fmt.Sprint(id)
		}
	}
}

// Update sets specific fields of a record that already exists in the database.
// optLockFieldName is the name of a version field that will implement an optimistic locking check before executing the update.
// Note that if the database is not currently in a transaction, then the optimistic lock
// will be a cursory check, but will not be able to definitively prevent a prior write.
// If optLockFieldName is provided, that field will be updated with a new version number value during the update process.
func (h *DbHelper) Update(ctx context.Context,
	table string,
	pkName string,
	pkValue any,
	fields map[string]any,
	optLockFieldName string,
	optLockFieldValue int64,
) error {
	if optLockFieldName != "" {
		s, args := GenerateVersionLock(h.dbi, table, pkName, pkValue, optLockFieldName, h.IsInTransaction(ctx))
		if rows, err := h.SqlQuery(ctx, s, args...); err != nil {
			return err
		} else {
			var version int64
			defer rows.Close()
			if !rows.Next() {
				// The record was deleted prior to an update completing.
				return db.NewRecordNotFoundError(fmt.Sprintf("Record not found, table: %s, pk: %s", table, pkValue))
			}
			if err = rows.Scan(&version); err != nil {
				panic(err) // a database error, perhaps the version field does not exist in the database?
			}
			if version != optLockFieldValue {
				// The record was changed prior to an update completing.
				return db.NewOptimisticLockError(fmt.Sprintf("Optimistic lock error, table: %s, pk: %s", table, pkValue), nil)
			}
			// If we get here, and we are in a transaction, the record has been locked until the end of the transaction and optimistic locking is valid.
			// If not in a transaction, we know that so far the record has not changed, but it still has a slight chance of changing between here
			// and the execution of the GenerateUpdate below.

			// Generate a new version number prior to saving.
			fields[optLockFieldName] = db.RecordVersion(optLockFieldValue)
		}
	}
	s, args := GenerateUpdate(h.dbi, table, fields, map[string]any{pkName: pkValue})
	_, e := h.SqlExec(ctx, s, args...)
	if e != nil {
		panic(e) // some kind of database error, should be notified immediately
	}
	return nil
}

// BuilderQuery performs a complex query using a query builder.
// The data returned will depend on the command inside the builder.
func (h *DbHelper) BuilderQuery(builder *Builder) any {
	joinTree := jointree.NewJoinTree(builder)
	switch joinTree.Command {
	case BuilderCommandLoad:
		return h.joinTreeLoad(builder.Ctx, joinTree)
	case BuilderCommandLoadCursor:
		return h.joinTreeLoadCursor(builder.Ctx, joinTree)
	case BuilderCommandCount:
		return h.joinTreeCount(builder.Ctx, joinTree)
	}
	return nil
}

func (h *DbHelper) joinTreeLoad(ctx context.Context, joinTree *jointree.JoinTree) []map[string]any {
	g := newSqlGenerator(joinTree, h.dbi)
	s, args := g.generateSelectSql()

	rows, err := h.dbi.SqlQuery(ctx, s, args...)

	if err != nil {
		// This is possibly generating an error related to the sql itself, so put the sql in the error message.
		s := err.Error()
		s += "\nSql: " + s

		panic(errors.New(s))
	}

	names, _ := rows.Columns()

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

	result := ReceiveRows(rows, columnTypes, names, joinTree)

	return result
}

func (h *DbHelper) joinTreeLoadCursor(ctx context.Context, joinTree *jointree.JoinTree) any {
	g := newSqlGenerator(joinTree, h.dbi)
	s, args := g.generateSelectSql()
	rows, err := h.dbi.SqlQuery(ctx, s, args...)

	if err != nil {
		// This is possibly generating an error related to the sql itself, so put the sql in the error message.
		s := err.Error()
		s += "\nSql: " + s

		panic(errors.New(s))
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
	return NewSqlCursor(rows, columnTypes, nil, joinTree)
}

func (h *DbHelper) joinTreeCount(ctx context.Context, joinTree *jointree.JoinTree) int {
	g := newSqlGenerator(joinTree, h.dbi)
	s, args := g.generateCountSql()
	rows, err := h.dbi.SqlQuery(ctx, s, args...)

	if err != nil {
		// This is possibly generating an error related to the sql itself, so put the sql in the error message.
		s := err.Error()
		s += "\nSql: " + s

		panic(errors.New(s))
	}

	names, _ := rows.Columns()
	columnTypes := []ReceiverType{ColTypeInteger}
	result := ReceiveRows(rows, columnTypes, names, nil)
	ret := result[0][names[0]].(int)
	return ret

}
