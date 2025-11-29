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

// TestForwardUniqueNullableLock tests insert and update of two linked records where the link is nullable.
func TestForwardUniqueNullableLock(t *testing.T) {
	ctx := context.Background()

	// Insert-insert
	l1 := goradd_unit2.NewLeafUnl()
	r1 := goradd_unit2.NewRootUnl()
	r1.SetName("rootForwardUniqueNullableLock")
	l1.SetName("leafForwardUniqueNullableLock")
	l1.SetRootUnl(r1)
	require.NoError(t, l1.Save(ctx))

	// loading back
	l1b, err := goradd_unit2.LoadLeafUnl(ctx, l1.ID(), node.LeafUnl().RootUnl())
	require.NoError(t, err)
	require.NotNilf(t, l1b, "Object was nil based on ID %s", l1.ID())
	assert.Equal(t, "leafForwardUniqueNullableLock", l1b.Name())
	assert.Equal(t, "rootForwardUniqueNullableLock", l1b.RootUnl().Name())

	// Update-update
	l1.SetName("leafForwardUniqueNullableLock2")
	l1.RootUnl().SetName("rootForwardUniqueNullableLock2")
	err = l1.Save(ctx)
	assert.NoError(t, err)
	l1b, err = goradd_unit2.LoadLeafUnl(ctx, l1.ID(), node.LeafUnl().RootUnl())
	require.NoError(t, err)
	require.NotNilf(t, l1b, "Object was nil based on ID %s", l1.ID())
	assert.Equal(t, "leafForwardUniqueNullableLock2", l1b.Name())
	assert.Equal(t, "rootForwardUniqueNullableLock2", l1b.RootUnl().Name())

	// Insert-update
	l2 := goradd_unit2.NewLeafUnl()
	l2.SetName("leafForwardUniqueNullableLock3")
	r1.SetName("rootForwardUniqueNullableLock3")
	l2.SetRootUnl(r1)
	err = l2.Save(ctx)
	assert.Error(t, err)
	assert.IsType(t, &db.UniqueValueError{}, err)

	// Update-insert
	r3 := goradd_unit2.NewRootUnl()
	l1.SetName("leafForwardUniqueNullableLock4")
	r3.SetName("rootForwardUniqueNullableLock4")
	l1.SetRootUnl(r3)
	err = l1.Save(ctx)
	require.NoError(t, err)
	l1b, err = goradd_unit2.LoadLeafUnl(ctx, l1.ID(), node.LeafUnl().RootUnl())
	require.NoError(t, err)
	require.NotNilf(t, l1b, "Object was nil based on ID %s", l1.ID())
	assert.Equal(t, "leafForwardUniqueNullableLock4", l1b.Name())
	assert.Equal(t, "rootForwardUniqueNullableLock4", l1b.RootUnl().Name())
}

// TestForwardNullableCollision tests saving two records that are changed at the same time.
func TestForwardUniqueNullableLockCollision(t *testing.T) {
	ctx := context.Background()

	l := goradd_unit2.NewLeafUnl()
	r := goradd_unit2.NewRootUnl()
	r.SetName("rootForwardUniqueNullableLockCollision")
	l.SetName("leafForwardUniqueNullableLockCollision")
	l.SetRootUnl(r)
	err := l.Save(ctx)
	require.NoError(t, err)

	var l2 *goradd_unit2.LeafUnl
	l2, err = goradd_unit2.LoadLeafUnl(ctx, l.ID(), node.LeafUnl().RootUnl())
	require.NoError(t, err)

	// Update both
	l.SetName("leafForwardUniqueNullableLockCollision2")
	l2.SetName("leafForwardUniqueNullableLockCollision3")

	err = l.Save(ctx)
	err2 := l2.Save(ctx)
	assert.NoError(t, err)
	assert.Error(t, err2)
	assert.IsType(t, &db.OptimisticLockError{}, err2)

	// 2nd level
	l2, _ = goradd_unit2.LoadLeafUnl(ctx, l.ID(), node.LeafUnl().RootUnl())
	l.RootUnl().SetName("rootForwardUniqueNullableLockCollision2")
	l2.RootUnl().SetName("rootForwardUniqueNullableLockCollision3")

	err = l.Save(ctx)
	err2 = l2.Save(ctx)
	assert.NoError(t, err)
	assert.Error(t, err2)
	assert.IsType(t, &db.OptimisticLockError{}, err2)

	// Delete
	l2, _ = goradd_unit2.LoadLeafUnl(ctx, l.ID(), node.LeafUnl().RootUnl())
	assert.NoError(t, l.RootUnl().Delete(ctx))
	l2.SetName("leafForwardUniqueNullableLockCollision4")
	err2 = l2.Save(ctx)
	assert.IsType(t, &db.OptimisticLockError{}, err2)
}

func TestForwardUniqueNullableLockNull(t *testing.T) {
	ctx := context.Background()
	l := goradd_unit2.NewLeafUnl()
	l.SetName("leafForwardUniqueNullableLockNull")
	assert.NoError(t, l.Save(ctx))

	l.SetRootUnl(nil) // nullable
	assert.NoError(t, l.Save(ctx))
}

func TestForwardUniqueNullableLockTwo(t *testing.T) {
	ctx := context.Background()
	r := goradd_unit2.NewRootUnl()
	l := goradd_unit2.NewLeafUnl()
	r.SetName("rootForwardUniqueNullableLockTwo")
	l.SetName("leafForwardUniqueNullableLockTwo")
	r.SetLeafUnl(l)
	require.NoError(t, r.Save(ctx))

	l2 := goradd_unit2.NewLeafUnl()
	l2.SetName("leafForwardUniqueNullableLockTwo2")
	r.SetLeafUnl(l2)
	require.NoError(t, r.Save(ctx)) // unique failure

	l2, err := goradd_unit2.LoadLeafUnl(ctx, l.ID())
	require.NoError(t, err)
	assert.NotNil(t, l2) // reverse linked item that is allowed to have a nil pointer was not deleted
}

func TestForwardUniqueNullableLockDelete(t *testing.T) {
	ctx := context.Background()
	l := goradd_unit2.NewLeafUnl()
	r := goradd_unit2.NewRootUnl()
	l.SetName("leafForwardUniqueNullableLockDelete")
	r.SetName("rootForwardUniqueNullableLockDelete")
	l.SetRootUnl(r)
	require.NoError(t, l.Save(ctx))

	// Collision on Change
	l2, err := goradd_unit2.LoadLeafUnl(ctx, l.ID())
	require.NoError(t, err)
	l.SetName("leafForwardUniqueNullableLockDelete2")
	_ = l.Save(ctx)
	err = l2.Delete(ctx)
	require.Error(t, err)
	assert.IsType(t, &db.OptimisticLockError{}, err)

	// Collision on deep Delete
	l2, err = goradd_unit2.LoadLeafUnl(ctx, l.ID())
	require.NoError(t, err)
	err = l.RootUnl().Delete(ctx)
	require.NoError(t, err)
	err = l2.Delete(ctx)
	assert.Error(t, err)
	assert.IsType(t, &db.OptimisticLockError{}, err)

	// Deep delete set link to null
	l2, err = goradd_unit2.LoadLeafUnl(ctx, l.ID())
	assert.NoError(t, err)
	assert.NotNil(t, l2)
}
