package crud

import (
	"context"
	"testing"

	"github.com/goradd/gro/ci/tests/gen/goradd_unit"
	goradd_unit2 "github.com/goradd/gro/ci/tests/gen/goradd_unit"
	"github.com/goradd/gro/ci/tests/gen/goradd_unit/node"
	"github.com/goradd/gro/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestReverseUniqueNullableLock tests insert and update of two linked records.
func TestReverseUniqueNullableLock(t *testing.T) {
	ctx := context.Background()

	// Insert-insert
	r := goradd_unit2.NewRootUnl()
	l := goradd_unit2.NewLeafUnl()
	r.SetName("rootReverseUniqueNullableLock")
	l.SetName("leafReverseUniqueNullableLock")
	r.SetLeafUnl(l)
	err := r.Save(ctx)
	require.NoError(t, err)

	var r2 *goradd_unit2.RootUnl
	r2, err = goradd_unit.LoadRootUnl(ctx, r.ID(), node.RootUnl().LeafUnl())
	require.NoError(t, err)
	require.NotNilf(t, r2, "Object was nil based on ID %s", r.ID())
	assert.Equal(t, "rootReverseUniqueNullableLock", r2.Name())
	assert.Equal(t, "leafReverseUniqueNullableLock", r2.LeafUnl().Name())

	// Update-update
	r.SetName("rootReverseUniqueNullableLock2")
	r.LeafUnl().SetName("leafReverseUniqueNullableLock2")
	err = r.Save(ctx)
	assert.NoError(t, err)
	r2, err = goradd_unit.LoadRootUnl(ctx, r.ID(), node.RootUnl().LeafUnl())
	require.NoError(t, err)
	require.NotNilf(t, r2, "Object was nil based on ID %s", r.ID())
	assert.Equal(t, "rootReverseUniqueNullableLock2", r2.Name())
	assert.Equal(t, "leafReverseUniqueNullableLock2", r2.LeafUnl().Name())

	// Insert-update
	r3 := goradd_unit2.NewRootUnl()
	r3.SetName("rootReverseUniqueNullableLock3")
	l.SetName("leafReverseUniqueNullableLock3")
	r3.SetLeafUnl(l)
	err = r3.Save(ctx)
	require.NoError(t, err)
	r2, err = goradd_unit.LoadRootUnl(ctx, r3.ID(), node.RootUnl().LeafUnl())
	require.NoError(t, err)
	require.NotNilf(t, r2, "Object was nil based on ID %s", r3.ID())
	assert.Equal(t, "rootReverseUniqueNullableLock3", r2.Name())
	assert.Equal(t, "leafReverseUniqueNullableLock3", r2.LeafUnl().Name())

	// Update-insert
	l4 := goradd_unit2.NewLeafUnl()
	r.SetName("rootReverseUniqueNullableLock4")
	l4.SetName("leafReverseUniqueNullableLock4")
	r.SetLeafUnl(l4)
	err = r.Save(ctx)
	require.NoError(t, err)
	r2, err = goradd_unit.LoadRootUnl(ctx, r.ID(), node.RootUnl().LeafUnl())
	require.NoError(t, err)
	require.NotNilf(t, r2, "Object was nil based on ID %s", r.ID())
	assert.Equal(t, "rootReverseUniqueNullableLock4", r2.Name())
	assert.Equal(t, "leafReverseUniqueNullableLock4", r2.LeafUnl().Name())

}

// TestReverseUniqueNullableLockCollision tests saving two records that are changed at the same time.
func TestReverseUniqueNullableLockCollision(t *testing.T) {
	ctx := context.Background()

	r := goradd_unit2.NewRootUnl()
	l := goradd_unit2.NewLeafUnl()
	r.SetName("rootReverseUniqueNullableLockCollision")
	l.SetName("leafReverseUniqueNullableLockCollision")
	r.SetLeafUnl(l)
	err := r.Save(ctx)
	require.NoError(t, err)

	var r2 *goradd_unit2.RootUnl
	r2, err = goradd_unit.LoadRootUnl(ctx, r.ID(), node.RootUnl().LeafUnl())
	require.NoError(t, err)

	r.SetName("rootReverseUniqueNullableLockCollision2")
	r2.SetName("rootReverseUniqueNullableLockCollision3")

	err = r.Save(ctx)
	err2 := r2.Save(ctx)
	assert.NoError(t, err)
	assert.Error(t, err2)
	assert.IsType(t, &db.OptimisticLockError{}, err2)

	// Level 2
	r2, _ = goradd_unit.LoadRootUnl(ctx, r.ID(), node.RootUnl().LeafUnl())
	r.LeafUnl().SetName("leafReverseUniqueNullableLockCollision2")
	r2.LeafUnl().SetName("leafReverseUniqueNullableLockCollision3")
	err = r.Save(ctx)
	err2 = r2.Save(ctx)
	assert.NoError(t, err)
	assert.Error(t, err2)
	assert.IsType(t, &db.OptimisticLockError{}, err2)

	// Delete
	r2, _ = goradd_unit.LoadRootUnl(ctx, r.ID(), node.RootUnl().LeafUnl())
	assert.NoError(t, r.Delete(ctx))
	r2.SetName("rootReverseUniqueNullableLockCollision4")
	err2 = r2.Save(ctx)
	assert.IsType(t, &db.OptimisticLockError{}, err2)

}

func TestReverseUniqueNullableLockNull(t *testing.T) {
	ctx := context.Background()

	r := goradd_unit2.NewRootUnl()
	r.SetName("rootReverseUniqueNullableLockNull")
	l := goradd_unit2.NewLeafUnl()
	l.SetName("leafReverseUniqueNullableLockNull")
	r.SetLeafUnl(l)
	require.NoError(t, r.Save(ctx))

	r.SetLeafUnl(nil)
	require.NoError(t, r.Save(ctx))

	r2, err := goradd_unit.LoadRootUnl(ctx, r.ID(), node.RootUnl().LeafUnl())
	require.NoError(t, err)
	assert.Nil(t, r2.LeafUnl())

	l2, err := goradd_unit2.LoadLeafUnl(ctx, l.ID())
	require.NoError(t, err)
	assert.NotNil(t, l2) // reverse linked item that could have a nil pointer was not deleted
	assert.Nil(t, l2.RootUnl())
}

func TestReverseUniqueNullableLockTwo(t *testing.T) {
	ctx := context.Background()
	r := goradd_unit2.NewRootUnl()
	l := goradd_unit2.NewLeafUnl()
	r.SetName("rootReverseUniqueNullableLockTwo")
	l.SetName("leafReverseUniqueNullableLockTwo")
	r.SetLeafUnl(l)
	require.NoError(t, r.Save(ctx))

	l2 := goradd_unit2.NewLeafUnl()
	l2.SetName("leafReverseUniqueNullableLockTwo2")
	r.SetLeafUnl(l2)
	require.NoError(t, r.Save(ctx)) // unique failure

	// Confirm detached.
	l3, _ := goradd_unit2.LoadLeafUnl(ctx, l.ID())
	assert.Nil(t, l3.RootUnl())
}

func TestReverseUniqueNullableLockDelete(t *testing.T) {
	ctx := context.Background()
	l := goradd_unit2.NewLeafUnl()
	r := goradd_unit2.NewRootUnl()
	l.SetName("leafReverseUniqueNullableLockDelete")
	r.SetName("rootReverseUniqueNullableLockDelete")
	r.SetLeafUnl(l)
	require.NoError(t, r.Save(ctx))

	// Collision on shallow change
	r2, err := goradd_unit.LoadRootUnl(ctx, r.ID(), node.RootUnl().LeafUnl())
	require.NoError(t, err)
	r.SetName("rootReverseUniqueNullableLockDelete2")
	_ = r.Save(ctx)
	err = r2.Delete(ctx)
	require.Error(t, err)
	assert.IsType(t, &db.OptimisticLockError{}, err)

	// No collision on deep Delete since it can't be detected
	r2, err = goradd_unit.LoadRootUnl(ctx, r.ID(), node.RootUnl().LeafUnl())
	require.NoError(t, err)
	err = r.LeafUnl().Delete(ctx)
	require.NoError(t, err)
	err = r2.Delete(ctx)
	assert.NoError(t, err)
}
