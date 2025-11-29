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

// TestReverseUniqueLock tests insert and update of two linked records.
func TestReverseUniqueLock(t *testing.T) {
	ctx := context.Background()

	// Insert-insert
	r := goradd_unit2.NewRootUl()
	l := goradd_unit2.NewLeafUl()
	r.SetName("rootReverseUniqueLock")
	l.SetName("leafReverseUniqueLock")
	r.SetLeafUl(l)
	err := r.Save(ctx)
	require.NoError(t, err)

	var r2 *goradd_unit2.RootUl
	r2, err = goradd_unit2.LoadRootUl(ctx, r.ID(), node.RootUl().LeafUl())
	require.NoError(t, err)
	require.NotNilf(t, r2, "Object was nil based on ID %s", r.ID())
	assert.Equal(t, "rootReverseUniqueLock", r2.Name())
	assert.Equal(t, "leafReverseUniqueLock", r2.LeafUl().Name())

	// Update-update
	r.SetName("rootReverseUniqueLock2")
	r.LeafUl().SetName("leafReverseUniqueLock2")
	err = r.Save(ctx)
	assert.NoError(t, err)
	r2, err = goradd_unit2.LoadRootUl(ctx, r.ID(), node.RootUl().LeafUl())
	require.NoError(t, err)
	require.NotNilf(t, r2, "Object was nil based on ID %s", r.ID())
	assert.Equal(t, "rootReverseUniqueLock2", r2.Name())
	assert.Equal(t, "leafReverseUniqueLock2", r2.LeafUl().Name())

	// Insert-update
	r3 := goradd_unit2.NewRootUl()
	r3.SetName("rootReverseUniqueLock3")
	l.SetName("leafReverseUniqueLock3")
	r3.SetLeafUl(l)
	err = r3.Save(ctx)
	require.NoError(t, err)
	r2, err = goradd_unit2.LoadRootUl(ctx, r3.ID(), node.RootUl().LeafUl())
	require.NoError(t, err)
	require.NotNilf(t, r2, "Object was nil based on ID %s", r3.ID())
	assert.Equal(t, "rootReverseUniqueLock3", r2.Name())
	assert.Equal(t, "leafReverseUniqueLock3", r2.LeafUl().Name())

	// Update-insert
	l4 := goradd_unit2.NewLeafUl()
	r.SetName("rootReverseUniqueLock4")
	l4.SetName("leafReverseUniqueLock4")
	r.SetLeafUl(l4)
	err = r.Save(ctx)
	require.NoError(t, err)
	r2, err = goradd_unit2.LoadRootUl(ctx, r.ID(), node.RootUl().LeafUl())
	require.NoError(t, err)
	require.NotNilf(t, r2, "Object was nil based on ID %s", r.ID())
	assert.Equal(t, "rootReverseUniqueLock4", r2.Name())
	assert.Equal(t, "leafReverseUniqueLock4", r2.LeafUl().Name())

}

// TestReverseCollision tests saving two records that are changed at the same time.
func TestReverseUniqueLockCollision(t *testing.T) {
	ctx := context.Background()

	r := goradd_unit2.NewRootUl()
	l := goradd_unit2.NewLeafUl()
	r.SetName("rootReverseUniqueLockCollision")
	l.SetName("leafReverseUniqueLockCollision")
	r.SetLeafUl(l)
	err := r.Save(ctx)
	require.NoError(t, err)

	var r2 *goradd_unit2.RootUl
	r2, err = goradd_unit2.LoadRootUl(ctx, r.ID(), node.RootUl().LeafUl())
	require.NoError(t, err)

	r.SetName("rootReverseUniqueLockCollision2")
	r2.SetName("rootReverseUniqueLockCollision3")

	err = r.Save(ctx)
	err2 := r2.Save(ctx)
	assert.NoError(t, err)
	assert.Error(t, err2)
	assert.IsType(t, &db.OptimisticLockError{}, err2)

	// Level 2
	r2, _ = goradd_unit2.LoadRootUl(ctx, r.ID(), node.RootUl().LeafUl())

	r.LeafUl().SetName("leafReverseUniqueLockCollision2")
	r2.LeafUl().SetName("leafReverseUniqueLockCollision3")
	err = r.Save(ctx)
	err2 = r2.Save(ctx)
	assert.NoError(t, err)
	assert.Error(t, err2)
	assert.IsType(t, &db.OptimisticLockError{}, err2)

	// Delete
	r2, _ = goradd_unit2.LoadRootUl(ctx, r.ID(), node.RootUl().LeafUl())
	assert.NoError(t, r.Delete(ctx))
	r2.SetName("rootReverseUniqueLockCollision4")
	err2 = r2.Save(ctx)
	assert.IsType(t, &db.OptimisticLockError{}, err2)

}

func TestReverseUniqueLockNull(t *testing.T) {
	ctx := context.Background()

	r := goradd_unit2.NewRootUl()
	r.SetName("rootReverseUniqueLockNull")
	l := goradd_unit2.NewLeafUl()
	l.SetName("leafReverseUniqueLockNull")
	r.SetLeafUl(l)
	require.NoError(t, r.Save(ctx))

	r.SetLeafUl(nil)
	require.NoError(t, r.Save(ctx))

	r2, err := goradd_unit2.LoadRootUl(ctx, r.ID(), node.RootUl().LeafUl())
	require.NoError(t, err)
	assert.Nil(t, r2.LeafUl())

	l2, err := goradd_unit2.LoadLeafUl(ctx, l.ID())
	require.NoError(t, err)
	assert.Nil(t, l2) // reverse linked item that could not have a nil pointer was deleted
}

func TestReverseUniqueLockTwo(t *testing.T) {
	ctx := context.Background()
	r := goradd_unit2.NewRootUl()
	l := goradd_unit2.NewLeafUl()
	r.SetName("rootReverseUniqueLockTwo")
	l.SetName("leafReverseUniqueLockTwo")
	r.SetLeafUl(l)
	require.NoError(t, r.Save(ctx))

	l2 := goradd_unit2.NewLeafUl()
	l2.SetName("leafReverseUniqueLockTwo2")
	r.SetLeafUl(l2)
	require.NoError(t, r.Save(ctx)) // unique failure

	l2, err := goradd_unit2.LoadLeafUl(ctx, l.ID())
	require.NoError(t, err)
	assert.Nil(t, l2) // reverse linked item that could not have a nil pointer was deleted
}

func TestReverseUniqueLockDelete(t *testing.T) {
	ctx := context.Background()
	l := goradd_unit2.NewLeafUl()
	r := goradd_unit2.NewRootUl()
	l.SetName("leafReverseUniqueLockDelete")
	r.SetName("rootReverseUniqueLockDelete")
	r.SetLeafUl(l)
	require.NoError(t, r.Save(ctx))

	// Collision on shallow change
	r2, err := goradd_unit2.LoadRootUl(ctx, r.ID(), node.RootUl().LeafUl())
	require.NoError(t, err)
	r.SetName("rootReverseUniqueLockDelete2")
	_ = r.Save(ctx)
	err = r2.Delete(ctx)
	require.Error(t, err)
	assert.IsType(t, &db.OptimisticLockError{}, err)

	// No collision on deep Delete since it can't be detected
	r2, err = goradd_unit2.LoadRootUl(ctx, r.ID(), node.RootUl().LeafUl())
	require.NoError(t, err)
	err = r.LeafUl().Delete(ctx)
	require.NoError(t, err)
	err = r2.Delete(ctx)
	assert.NoError(t, err)
}
