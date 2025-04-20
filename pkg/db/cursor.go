package db

import (
	"errors"
	"fmt"
)

type Cursor[T any] interface {
	Next() (*T, error)
	Close() error
}

// WalkCursor iterates through a cursor, calls the handler for each item, and ensures the cursor is closed.
// The handler receives both the item and its 0-based index.
func WalkCursor[T any](cursor Cursor[T], fn func(index int, item *T) error) (rerr error) {
	defer func() {
		if cerr := cursor.Close(); cerr != nil && rerr == nil {
			rerr = fmt.Errorf("failed to close cursor: %w", cerr)
		} else if cerr != nil && rerr != nil {
			rerr = errors.Join(rerr, fmt.Errorf("close failed: %w", cerr))
		}
	}()

	for i := 0; ; i++ {
		item, err := cursor.Next()
		if err != nil {
			return fmt.Errorf("cursor.Next failed: %w", err)
		}
		if item == nil {
			break
		}
		if err := fn(i, item); err != nil {
			return fmt.Errorf("handler error at index %d: %w", i, err)
		}
	}

	return nil
}
