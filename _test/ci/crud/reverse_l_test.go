package crud

import (
	"context"
	"testing"

	"github.com/goradd/gro/_test/gen/orm/goradd_unit"
	"github.com/goradd/gro/_test/gen/orm/goradd_unit/node"
	"github.com/goradd/gro/pkg/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestReverseLock tests insert and update of two linked records that use an optimistic lock.
func TestReverseLock(t *testing.T) {
	ctx := context.Background()

	// Insert-insert
	r := goradd_unit.NewRootL()
	l := goradd_unit.NewLeafL()
	r.SetName("rootReverseLock")
	l.SetName("leafReverseLock")
	r.SetLeafLs(l)
	err := r.Save(ctx)
	require.NoError(t, err)

	var r2 *goradd_unit.RootL
	r2, err = goradd_unit.LoadRootL(ctx, r.ID(), node.RootL().LeafLs())
	require.NoError(t, err)
	require.NotNilf(t, r2, "Object was nil based on ID %s", r.ID())
	assert.Equal(t, "rootReverseLock", r2.Name())
	assert.Equal(t, "leafReverseLock", r2.LeafLs()[0].Name())

	// Update-update
	r.SetName("rootReverseLock2")
	r.LeafLs()[0].SetName("leafReverseLock2")
	err = r.Save(ctx)
	assert.NoError(t, err)
	r2, err = goradd_unit.LoadRootL(ctx, r.ID(), node.RootL().LeafLs())
	require.NoError(t, err)
	require.NotNilf(t, r2, "Object was nil based on ID %s", r.ID())
	assert.Equal(t, "rootReverseLock2", r2.Name())
	assert.Equal(t, "leafReverseLock2", r2.LeafLs()[0].Name())

	// Insert-update
	r3 := goradd_unit.NewRootL()
	r3.SetName("rootReverseLock3")
	l.SetName("leafReverseLock3")
	r3.SetLeafLs(l)
	err = r3.Save(ctx)
	require.NoError(t, err)
	r2, err = goradd_unit.LoadRootL(ctx, r3.ID(), node.RootL().LeafLs())
	require.NoError(t, err)
	require.NotNilf(t, r2, "Object was nil based on ID %s", r3.ID())
	assert.Equal(t, "rootReverseLock3", r2.Name())
	assert.Equal(t, "leafReverseLock3", r2.LeafLs()[0].Name())

	// Update-insert
	l4 := goradd_unit.NewLeafL()
	r.SetName("rootReverseLock4")
	l4.SetName("leafReverseLock4")
	r.SetLeafLs(l4)
	err = r.Save(ctx)
	require.NoError(t, err)
	r2, err = goradd_unit.LoadRootL(ctx, r.ID(), node.RootL().LeafLs())
	require.NoError(t, err)
	require.NotNilf(t, r2, "Object was nil based on ID %s", r.ID())
	assert.Equal(t, "rootReverseLock4", r2.Name())
	assert.Equal(t, "leafReverseLock4", r2.LeafLs()[0].Name())

}

// TestReverseCollision tests saving two records that are changed at the same time.
func TestReverseLockCollision(t *testing.T) {
	ctx := context.Background()

	r := goradd_unit.NewRootL()
	l := goradd_unit.NewLeafL()
	r.SetName("rootReverseLockCollision")
	l.SetName("leafReverseLockCollision")
	r.SetLeafLs(l)
	err := r.Save(ctx)
	require.NoError(t, err)

	var r2 *goradd_unit.RootL
	r2, err = goradd_unit.LoadRootL(ctx, r.ID(), node.RootL().LeafLs())
	require.NoError(t, err)

	r.SetName("rootReverseLockCollision2")
	r2.SetName("rootReverseLockCollision3")

	err = r.Save(ctx)
	err2 := r2.Save(ctx)
	assert.NoError(t, err)
	assert.Error(t, err2)
	assert.IsType(t, &db.OptimisticLockError{}, err2)

	// Level 2
	r2, _ = goradd_unit.LoadRootL(ctx, r.ID(), node.RootL().LeafLs())
	r.LeafLs()[0].SetName("leafReverseLockCollision2")
	r2.LeafLs()[0].SetName("leafReverseLockCollision3")
	err = r.Save(ctx)
	err2 = r2.Save(ctx)
	assert.NoError(t, err)
	assert.Error(t, err2)
	assert.IsType(t, &db.OptimisticLockError{}, err2)

	// Delete
	r2, _ = goradd_unit.LoadRootL(ctx, r.ID(), node.RootL().LeafLs())
	assert.NoError(t, r.Delete(ctx))
	r2.SetName("rootReverseLockCollision4")
	err2 = r2.Save(ctx)
	assert.IsType(t, &db.OptimisticLockError{}, err2)
}

func TestReverseLockNull(t *testing.T) {
	ctx := context.Background()

	r := goradd_unit.NewRootL()
	r.SetName("rootReverseLockNull")
	l := goradd_unit.NewLeafL()
	l.SetName("leafReverseLockNull")
	r.SetLeafLs(l)
	require.NoError(t, r.Save(ctx))

	r.SetLeafLs()
	require.NoError(t, r.Save(ctx))

	r2, err := goradd_unit.LoadRootL(ctx, r.ID(), node.RootL().LeafLs())
	require.NoError(t, err)
	assert.Len(t, r2.LeafLs(), 0)

	l2, err := goradd_unit.LoadLeafL(ctx, l.ID())
	require.NoError(t, err)
	assert.Nil(t, l2) // reverse linked item that could not have a nil pointer was deleted
}

func TestReverseLockTwo(t *testing.T) {
	ctx := context.Background()
	r := goradd_unit.NewRootL()
	l := goradd_unit.NewLeafL()
	r.SetName("rootReverseLockTwo")
	l.SetName("leafReverseLockTwo")
	r.SetLeafLs(l)
	require.NoError(t, r.Save(ctx))

	l2 := goradd_unit.NewLeafL()
	l2.SetName("leafReverseLockTwo2")
	r.SetLeafLs(l, l2)
	require.NoError(t, r.Save(ctx))

	r2, err := goradd_unit.LoadRootL(ctx, r.ID(), node.RootL().LeafLs())
	require.NoError(t, err)
	assert.Len(t, r2.LeafLs(), 2)
}

func TestReverseLockDelete(t *testing.T) {
	ctx := context.Background()
	l := goradd_unit.NewLeafL()
	r := goradd_unit.NewRootL()
	l.SetName("leafReverseLockDelete")
	r.SetName("rootReverseLockDelete")
	r.SetLeafLs(l)
	require.NoError(t, r.Save(ctx))

	// Collision on shallow change
	r2, err := goradd_unit.LoadRootL(ctx, r.ID(), node.RootL().LeafLs())
	require.NoError(t, err)
	r.SetName("rootReverseLockDelete2")
	_ = r.Save(ctx)
	err = r2.Delete(ctx)
	require.Error(t, err)
	assert.IsType(t, &db.OptimisticLockError{}, err)

	// No collision on deep Delete since it can't be detected
	r2, err = goradd_unit.LoadRootL(ctx, r.ID(), node.RootL().LeafLs())
	require.NoError(t, err)
	err = r.LeafLs()[0].Delete(ctx)
	require.NoError(t, err)
	err = r2.Delete(ctx)
	assert.NoError(t, err)
}
