package db

import "fmt"

// OptimisticLockError reports errors related to optimistic locking. i.e. when the same record was changed by a different user
// prior to a save completing.
type OptimisticLockError struct {
	Table   string
	PkValue any
	Err     error // wrapped error coming from database driver if there is one
}

func (e *OptimisticLockError) Error() string {
	return fmt.Sprintf("optimistic lock failed: table = %s, pk = %v", e.Table, e.PkValue)
}

func (e *OptimisticLockError) Unwrap() error {
	return e.Err
}

// NewOptimisticLockError returns a new error related to optimistic locking.
//
// Test and get the values using:
//
//	 if myerr, ok := all.As[*OptimisticLockError](err); ok {
//		// process error
//	 }
func NewOptimisticLockError(table string, pkValue any, err error) error {
	return &OptimisticLockError{table, pkValue, err}
}

// RecordNotFoundError indicates a record was expected in the database, but was not found.
// The record may have been deleted simultaneously by another process.
type RecordNotFoundError struct {
	Table   string
	PkValue any
}

func (e *RecordNotFoundError) Error() string {
	return fmt.Sprintf("record not found: table = %s, pk = %v", e.Table, e.PkValue)
}

// NewRecordNotFoundError returns a new error stating that a record was not found.
// The message should describe the search used that failed.
func NewRecordNotFoundError(table string, pkValue any) error {
	return &RecordNotFoundError{table, pkValue}
}

// UniqueValueError indicates a record failed to save because a value in that
// record has a unique index and the value was found in another record.
type UniqueValueError struct {
	Table  string
	Column string
	Value  any
	Err    error
}

func (e *UniqueValueError) Error() string {
	if e.Err == nil {
		return fmt.Sprintf("unique value conflict: table = %s, column = %s, value = %s", e.Table, e.Column, e.Value)
	}
	return fmt.Sprintf("unique value conflict: %s", e.Err.Error())
}

// NewUniqueValueError returns a new error stating that a record could not be saved
// because a unique value in the new record was found in a different record.
func NewUniqueValueError(table string, column string, value string, err error) error {
	return &UniqueValueError{table, column, value, err}
}

func (e *UniqueValueError) Unwrap() error {
	return e.Err
}

// QueryError indicates an error occurred while querying a database.
// This could mean a syntax error with the query, a problem with the database,
// a problem with the connection to the database, etc.
// Unique value collisions will be returned as a UniqueValueError.
type QueryError struct {
	// Operation is the call into the database, or database function that returned the error
	Operation string
	// Query is the query that was attempted
	Query string
	// Args are the arguments sent with the query
	Args []any
	// Error is the underlying error returned
	Err error
}

func (e *QueryError) Error() string {
	return fmt.Sprintf("%s query error\n%s]n%v\n%s", e.Operation, e.Query, e.Args, e.Err.Error())
}

func (e *QueryError) Unwrap() error {
	return e.Err
}

func NewQueryError(operation, query string, args []any, err error) error {
	return &QueryError{operation, query, args, err}
}
