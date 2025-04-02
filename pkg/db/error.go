package db

// OptimisticLockError reports errors related to optimistic locking. i.e. when the same record was changed by a different user
// prior to a save completing.
type OptimisticLockError struct {
	message string
	err     error // wrapped error coming from database driver if there is one
}

func (e OptimisticLockError) Error() string {
	return e.message
}

func (e OptimisticLockError) Unwrap() error {
	return e.err
}

func (e OptimisticLockError) Is(err error) bool {
	return e.err == err
}

// NewOptimisticLockError returns a new error related to optimistic locking.
// message should indicate what record was changed prior to a save completing.
func NewOptimisticLockError(msg string, err error) error {
	return OptimisticLockError{message: msg, err: err}
}

// RecordNotFoundError indicates a record was expected in the database, but was not found.
type RecordNotFoundError struct {
	message string
}

func (e RecordNotFoundError) Error() string {
	return e.message
}

// NewRecordNotFoundError returns a new error stating that a record was not found.
// The message should describe the search used that failed.
func NewRecordNotFoundError(msg string) error {
	return RecordNotFoundError{message: msg}
}

// DuplicateValueError indicates a record failed to save because a value in that
// record has a unique index and the value was found in another record.
type DuplicateValueError struct {
	message string
}

func (e DuplicateValueError) Error() string {
	return e.message
}

// NewDuplicateValueError returns a new error stating that a record could not be saved
// because a unique value in the new record was found in a different record.
func NewDuplicateValueError(msg string) error {
	return RecordNotFoundError{message: msg}
}
