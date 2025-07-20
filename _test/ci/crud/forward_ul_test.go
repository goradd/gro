package crud

import (
	"context"
	"github.com/goradd/orm/_test/gen/orm/goradd_unit"
	"github.com/goradd/orm/_test/gen/orm/goradd_unit/node"
	"github.com/goradd/orm/pkg/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

// TestForwardUnique covers insert/update flows and enforces that only one LeafUl may point to a given RootUl.
func TestForwardUniqueLock(t *testing.T) {
	ctx := context.Background()
	defer goradd_unit.ClearAll(ctx)

	// Insert-insert
	l1 := goradd_unit.NewLeafUl()
	r1 := goradd_unit.NewRootUl()
	r1.SetName("root")
	l1.SetName("leaf")
	l1.SetRootUl(r1)
	require.NoError(t, l1.Save(ctx))

	// loading back
	l1b, err := goradd_unit.LoadLeafUl(ctx, l1.ID(), node.LeafUl().RootUl())
	require.NoError(t, err)
	assert.Equal(t, "leaf", l1b.Name())
	assert.Equal(t, "root", l1b.RootUl().Name())

	// Update-update
	l1.SetName("leaf2")
	l1.RootUl().SetName("root2")
	err = l1.Save(ctx)
	assert.NoError(t, err)
	l1b, err = goradd_unit.LoadLeafUl(ctx, l1.ID(), node.LeafUl().RootUl())
	require.NoError(t, err)
	assert.Equal(t, "leaf2", l1b.Name())
	assert.Equal(t, "root2", l1b.RootUl().Name())

	// Insert-update
	l2 := goradd_unit.NewLeafUl()
	l2.SetName("leaf3")
	r1.SetName("root3")
	l2.SetRootUl(r1)
	err = l2.Save(ctx)
	assert.Error(t, err)
	assert.IsType(t, &db.UniqueValueError{}, err)

	// Update-insert
	r3 := goradd_unit.NewRootUl()
	l1.SetName("leaf4")
	r3.SetName("root4")
	l1.SetRootUl(r3)
	err = l1.Save(ctx)
	require.NoError(t, err)
	l1b, err = goradd_unit.LoadLeafUl(ctx, l1.ID(), node.LeafUl().RootUl())
	require.NoError(t, err)
	assert.Equal(t, "leaf4", l1b.Name())
	assert.Equal(t, "root4", l1b.RootUl().Name())
}

// TestForwardUniqueCollision tests saving two records that are changed at the same time.
func TestForwardUniqueLockCollision(t *testing.T) {
	ctx := context.Background()
	defer goradd_unit.ClearAll(ctx)

	l := goradd_unit.NewLeafUl()
	r := goradd_unit.NewRootUl()
	r.SetName("root")
	l.SetName("leaf")
	l.SetRootUl(r)
	err := l.Save(ctx)
	require.NoError(t, err)

	var l2 *goradd_unit.LeafUl
	l2, err = goradd_unit.LoadLeafUl(ctx, l.ID(), node.LeafUl().RootUl())
	require.NoError(t, err)

	// Update both
	l.SetName("leaf2")
	l2.SetName("leaf3")

	err = l.Save(ctx)
	err2 := l2.Save(ctx)
	assert.NoError(t, err)
	assert.Error(t, err2)
	assert.IsType(t, &db.OptimisticLockError{}, err2)

	// 2nd level
	l2, _ = goradd_unit.LoadLeafUl(ctx, l.ID(), node.LeafUl().RootUl())
	l.RootUl().SetName("root2")
	l2.RootUl().SetName("root3")

	err = l.Save(ctx)
	err2 = l2.Save(ctx)
	assert.NoError(t, err)
	assert.Error(t, err2)
	assert.IsType(t, &db.OptimisticLockError{}, err2)

	// Delete
	l2, _ = goradd_unit.LoadLeafUl(ctx, l.ID(), node.LeafUl().RootUl())
	assert.NoError(t, l.RootUl().Delete(ctx))
	l2.SetName("leaf4")
	err2 = l2.Save(ctx)
	assert.IsType(t, &db.OptimisticLockError{}, err2)
}

func TestForwardUniqueLockNull(t *testing.T) {
	ctx := context.Background()
	defer goradd_unit.ClearAll(ctx)
	l := goradd_unit.NewLeafUl()
	l.SetName("leaf")
	assert.Panics(t, func() {
		_ = l.Save(ctx) // root is required since RootID is not nullable
	})

	assert.Panics(t, func() {
		l.SetRootUl(nil) // not nullable
	})
}
func TestForwardUniqueLockTwo(t *testing.T) {
	ctx := context.Background()
	defer goradd_unit.ClearAll(ctx)
	l := goradd_unit.NewLeafUl()
	r := goradd_unit.NewRootUl()
	l.SetName("leaf")
	r.SetName("root")
	l.SetRootUl(r)
	require.NoError(t, l.Save(ctx))

	l2 := goradd_unit.NewLeafUl()
	l2.SetName("leaf2")
	l2.SetRootUl(r)
	require.Error(t, l2.Save(ctx)) // unique value collision error
}

func TestForwardUniqueLockDelete(t *testing.T) {
	ctx := context.Background()
	defer goradd_unit.ClearAll(ctx)
	l := goradd_unit.NewLeafUl()
	r := goradd_unit.NewRootUl()
	l.SetName("leaf")
	r.SetName("root")
	l.SetRootUl(r)
	require.NoError(t, l.Save(ctx))

	// Collision on Change
	l2, err := goradd_unit.LoadLeafUl(ctx, l.ID())
	require.NoError(t, err)
	l.SetName("leaf2")
	_ = l.Save(ctx)
	err = l2.Delete(ctx)
	require.Error(t, err)
	assert.IsType(t, &db.OptimisticLockError{}, err)

	// Collision on deep Delete
	l2, err = goradd_unit.LoadLeafUl(ctx, l.ID())
	require.NoError(t, err)
	err = l.RootUl().Delete(ctx)
	require.NoError(t, err)
	err = l2.Delete(ctx)
	assert.Error(t, err)
	assert.IsType(t, &db.OptimisticLockError{}, err)

	// Deep delete deleted the linked record
	l2, err = goradd_unit.LoadLeafUl(ctx, l.ID())
	assert.NoError(t, err)
	assert.Nil(t, l2)
}
