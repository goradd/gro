package db

import (
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

// --- Mock Cursor ---
type mockCursor[T any] struct {
	items   []T
	index   int
	closed  bool
	failAt  int          // index at which Next() should fail
	closeFn func() error // optional custom close behavior
}

func (m *mockCursor[T]) Next() (*T, error) {
	var zero *T
	if m.index >= len(m.items) {
		return zero, nil
	}
	if m.failAt > 0 && m.index == m.failAt {
		return zero, errors.New("mock Next failure")
	}
	item := m.items[m.index]
	m.index++
	return &item, nil
}

func (m *mockCursor[T]) Close() error {
	m.closed = true
	if m.closeFn != nil {
		return m.closeFn()
	}
	return nil
}

// --- Tests ---
func TestWalkCursor_ZeroIterations(t *testing.T) {
	cursor := &mockCursor[string]{items: []string{}}
	var count int

	err := WalkCursor(cursor, func(i int, s *string) error {
		count++
		return nil
	})

	require.NoError(t, err)
	assert.Equal(t, 0, count)
	assert.True(t, cursor.closed)
}

func TestWalkCursor_OneIteration(t *testing.T) {
	cursor := &mockCursor[int]{items: []int{42}}
	var seen []int

	err := WalkCursor(cursor, func(i int, v *int) error {
		assert.Equal(t, 0, i)
		seen = append(seen, *v)
		return nil
	})

	require.NoError(t, err)
	assert.Equal(t, []int{42}, seen)
	assert.True(t, cursor.closed)
}

func TestWalkCursor_TwoIterations(t *testing.T) {
	cursor := &mockCursor[string]{items: []string{"a", "b"}}
	var out []string

	err := WalkCursor(cursor, func(i int, v *string) error {
		out = append(out, fmt.Sprintf("%d:%s", i, *v))
		return nil
	})

	require.NoError(t, err)
	assert.Equal(t, []string{"0:a", "1:b"}, out)
	assert.True(t, cursor.closed)
}

func TestWalkCursor_HandlerError(t *testing.T) {
	cursor := &mockCursor[int]{items: []int{1, 2, 3}}
	var seen []int

	err := WalkCursor(cursor, func(i int, v *int) error {
		seen = append(seen, *v)
		if *v == 2 {
			return errors.New("handler failed")
		}
		return nil
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "handler failed")
	assert.Equal(t, []int{1, 2}, seen) // should stop after 2
	assert.True(t, cursor.closed)
}

func TestWalkCursor_NextError(t *testing.T) {
	cursor := &mockCursor[string]{items: []string{"x", "y"}, failAt: 1}
	var seen []string

	err := WalkCursor(cursor, func(i int, s *string) error {
		seen = append(seen, *s)
		return nil
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "cursor.Next failed")
	assert.Equal(t, []string{"x"}, seen)
	assert.True(t, cursor.closed)
}
