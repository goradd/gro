package crud

import (
	"context"
	"testing"

	goradd_unit2 "github.com/goradd/gro/ci/tests/gen/goradd_unit"
	"github.com/goradd/gro/ci/tests/gen/goradd_unit/node"
	"github.com/goradd/gro/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestForwardUnique covers insert/update flows and enforces that only one LeafUl may point to a given RootUl.
func TestForwardUniqueLock(t *testing.T) {
	ctx := context.Background()

	// Insert-insert
	l1 := goradd_unit2.NewLeafUl()
	r1 := goradd_unit2.NewRootUl()
	r1.SetName("rootForwardUniqueLock")
	l1.SetName("leafForwardUniqueLock")
	l1.SetRootUl(r1)
	require.NoError(t, l1.Save(ctx))

	// loading back
	l1b, err := goradd_unit2.LoadLeafUl(ctx, l1.ID(), node.LeafUl().RootUl())
	require.NoError(t, err)
	require.NotNilf(t, l1b, "Object was nil based on ID %s", l1.ID())
	assert.Equal(t, "leafForwardUniqueLock", l1b.Name())
	assert.Equal(t, "rootForwardUniqueLock", l1b.RootUl().Name())

	// Update-update
	l1.SetName("leafForwardUniqueLock2")
	l1.RootUl().SetName("rootForwardUniqueLock2")
	err = l1.Save(ctx)
	assert.NoError(t, err)
	l1b, err = goradd_unit2.LoadLeafUl(ctx, l1.ID(), node.LeafUl().RootUl())
	require.NoError(t, err)
	require.NotNilf(t, l1b, "Object was nil based on ID %s", l1.ID())
	assert.Equal(t, "leafForwardUniqueLock2", l1b.Name())
	assert.Equal(t, "rootForwardUniqueLock2", l1b.RootUl().Name())

	// Insert-update
	l2 := goradd_unit2.NewLeafUl()
	l2.SetName("leafForwardUniqueLock3")
	r1.SetName("rootForwardUniqueLock3")
	l2.SetRootUl(r1)
	err = l2.Save(ctx)
	assert.Error(t, err)
	assert.IsType(t, &db.UniqueValueError{}, err)

	// Update-insert
	r3 := goradd_unit2.NewRootUl()
	l1.SetName("leafForwardUniqueLock4")
	r3.SetName("rootForwardUniqueLock4")
	l1.SetRootUl(r3)
	err = l1.Save(ctx)
	require.NoError(t, err)
	l1b, err = goradd_unit2.LoadLeafUl(ctx, l1.ID(), node.LeafUl().RootUl())
	require.NoError(t, err)
	require.NotNilf(t, l1b, "Object was nil based on ID %s", l1.ID())
	assert.Equal(t, "leafForwardUniqueLock4", l1b.Name())
	assert.Equal(t, "rootForwardUniqueLock4", l1b.RootUl().Name())
}

// TestForwardUniqueCollision tests saving two records that are changed at the same time.
func TestForwardUniqueLockCollision(t *testing.T) {
	ctx := context.Background()

	l := goradd_unit2.NewLeafUl()
	r := goradd_unit2.NewRootUl()
	r.SetName("rootForwardUniqueLockCollision")
	l.SetName("leafForwardUniqueLockCollision")
	l.SetRootUl(r)
	err := l.Save(ctx)
	require.NoError(t, err)

	var l2 *goradd_unit2.LeafUl
	l2, err = goradd_unit2.LoadLeafUl(ctx, l.ID(), node.LeafUl().RootUl())
	require.NoError(t, err)

	// Update both
	l.SetName("leafForwardUniqueLockCollision2")
	l2.SetName("leafForwardUniqueLockCollision3")

	err = l.Save(ctx)
	err2 := l2.Save(ctx)
	assert.NoError(t, err)
	assert.Error(t, err2)
	assert.IsType(t, &db.OptimisticLockError{}, err2)

	// 2nd level
	l2, _ = goradd_unit2.LoadLeafUl(ctx, l.ID(), node.LeafUl().RootUl())
	l.RootUl().SetName("rootForwardUniqueLockCollision2")
	l2.RootUl().SetName("rootForwardUniqueLockCollision3")

	err = l.Save(ctx)
	err2 = l2.Save(ctx)
	assert.NoError(t, err)
	assert.Error(t, err2)
	assert.IsType(t, &db.OptimisticLockError{}, err2)

	// Delete
	l2, _ = goradd_unit2.LoadLeafUl(ctx, l.ID(), node.LeafUl().RootUl())
	assert.NoError(t, l.RootUl().Delete(ctx))
	l2.SetName("leafForwardUniqueLockCollision4")
	err2 = l2.Save(ctx)
	assert.IsType(t, &db.OptimisticLockError{}, err2)
}

func TestForwardUniqueLockNull(t *testing.T) {
	ctx := context.Background()
	l := goradd_unit2.NewLeafUl()
	l.SetName("leafForwardUniqueLockNull")
	assert.Panics(t, func() {
		_ = l.Save(ctx) // root is required since RootID is not nullable
	})

	assert.Panics(t, func() {
		l.SetRootUl(nil) // not nullable
	})
}
func TestForwardUniqueLockTwo(t *testing.T) {
	ctx := context.Background()
	l := goradd_unit2.NewLeafUl()
	r := goradd_unit2.NewRootUl()
	l.SetName("leafForwardUniqueLockTwo")
	r.SetName("rootForwardUniqueLockTwo")
	l.SetRootUl(r)
	require.NoError(t, l.Save(ctx))

	l2 := goradd_unit2.NewLeafUl()
	l2.SetName("leafForwardUniqueLockTwo2")
	l2.SetRootUl(r)
	require.Error(t, l2.Save(ctx)) // unique value collision error
}

func TestForwardUniqueLockDelete(t *testing.T) {
	ctx := context.Background()
	l := goradd_unit2.NewLeafUl()
	r := goradd_unit2.NewRootUl()
	l.SetName("leafForwardUniqueLockDelete")
	r.SetName("rootForwardUniqueLockDelete")
	l.SetRootUl(r)
	require.NoError(t, l.Save(ctx))

	// Collision on Change
	l2, err := goradd_unit2.LoadLeafUl(ctx, l.ID())
	require.NoError(t, err)
	l.SetName("leafForwardUniqueLockDelete2")
	_ = l.Save(ctx)
	err = l2.Delete(ctx)
	require.Error(t, err)
	assert.IsType(t, &db.OptimisticLockError{}, err)

	// Collision on deep Delete
	l2, err = goradd_unit2.LoadLeafUl(ctx, l.ID())
	require.NoError(t, err)
	err = l.RootUl().Delete(ctx)
	require.NoError(t, err)
	err = l2.Delete(ctx)
	assert.Error(t, err)
	assert.IsType(t, &db.OptimisticLockError{}, err)

	// Deep delete deleted the linked record
	l2, err = goradd_unit2.LoadLeafUl(ctx, l.ID())
	assert.NoError(t, err)
	assert.Nil(t, l2)
}
