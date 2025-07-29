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

// TestReverseUniqueNullableLock tests insert and update of two linked records.
func TestReverseUniqueNullableLock(t *testing.T) {
	ctx := context.Background()
	defer goradd_unit.ClearAll(ctx)

	// Insert-insert
	r := goradd_unit.NewRootUnl()
	l := goradd_unit.NewLeafUnl()
	r.SetName("root")
	l.SetName("leaf")
	r.SetLeafUnl(l)
	err := r.Save(ctx)
	require.NoError(t, err)

	var r2 *goradd_unit.RootUnl
	r2, err = goradd_unit.LoadRootUnl(ctx, r.ID(), node.RootUnl().LeafUnl())
	require.NoError(t, err)
	assert.Equal(t, "root", r2.Name())
	assert.Equal(t, "leaf", r2.LeafUnl().Name())

	// Update-update
	r.SetName("root2")
	r.LeafUnl().SetName("leaf2")
	err = r.Save(ctx)
	assert.NoError(t, err)
	r2, err = goradd_unit.LoadRootUnl(ctx, r.ID(), node.RootUnl().LeafUnl())
	require.NoError(t, err)
	assert.Equal(t, "root2", r2.Name())
	assert.Equal(t, "leaf2", r2.LeafUnl().Name())

	// Insert-update
	r3 := goradd_unit.NewRootUnl()
	r3.SetName("root3")
	l.SetName("leaf3")
	r3.SetLeafUnl(l)
	err = r3.Save(ctx)
	require.NoError(t, err)
	r2, err = goradd_unit.LoadRootUnl(ctx, r3.ID(), node.RootUnl().LeafUnl())
	require.NoError(t, err)
	assert.Equal(t, "root3", r2.Name())
	assert.Equal(t, "leaf3", r2.LeafUnl().Name())

	// Update-insert
	l4 := goradd_unit.NewLeafUnl()
	r.SetName("root4")
	l4.SetName("leaf4")
	r.SetLeafUnl(l4)
	err = r.Save(ctx)
	require.NoError(t, err)
	r2, err = goradd_unit.LoadRootUnl(ctx, r.ID(), node.RootUnl().LeafUnl())
	require.NoError(t, err)
	assert.Equal(t, "root4", r2.Name())
	assert.Equal(t, "leaf4", r2.LeafUnl().Name())

}

// TestReverseUniqueNullableLockCollision tests saving two records that are changed at the same time.
func TestReverseUniqueNullableLockCollision(t *testing.T) {
	ctx := context.Background()
	defer goradd_unit.ClearAll(ctx)

	r := goradd_unit.NewRootUnl()
	l := goradd_unit.NewLeafUnl()
	r.SetName("root")
	l.SetName("leaf")
	r.SetLeafUnl(l)
	err := r.Save(ctx)
	require.NoError(t, err)

	var r2 *goradd_unit.RootUnl
	r2, err = goradd_unit.LoadRootUnl(ctx, r.ID(), node.RootUnl().LeafUnl())
	require.NoError(t, err)

	r.SetName("root2")
	r2.SetName("root3")

	err = r.Save(ctx)
	err2 := r2.Save(ctx)
	assert.NoError(t, err)
	assert.Error(t, err2)
	assert.IsType(t, &db.OptimisticLockError{}, err2)

	// Level 2
	r2, _ = goradd_unit.LoadRootUnl(ctx, r.ID(), node.RootUnl().LeafUnl())
	r.LeafUnl().SetName("leaf2")
	r2.LeafUnl().SetName("leaf3")
	err = r.Save(ctx)
	err2 = r2.Save(ctx)
	assert.NoError(t, err)
	assert.Error(t, err2)
	assert.IsType(t, &db.OptimisticLockError{}, err2)

	// Delete
	r2, _ = goradd_unit.LoadRootUnl(ctx, r.ID(), node.RootUnl().LeafUnl())
	assert.NoError(t, r.Delete(ctx))
	r2.SetName("root4")
	err2 = r2.Save(ctx)
	assert.IsType(t, &db.OptimisticLockError{}, err2)

}

func TestReverseUniqueNullableLockNull(t *testing.T) {
	ctx := context.Background()
	defer goradd_unit.ClearAll(ctx)

	r := goradd_unit.NewRootUnl()
	r.SetName("root")
	l := goradd_unit.NewLeafUnl()
	l.SetName("leaf")
	r.SetLeafUnl(l)
	require.NoError(t, r.Save(ctx))

	r.SetLeafUnl(nil)
	require.NoError(t, r.Save(ctx))

	r2, err := goradd_unit.LoadRootUnl(ctx, r.ID(), node.RootUnl().LeafUnl())
	require.NoError(t, err)
	assert.Nil(t, r2.LeafUnl())

	l2, err := goradd_unit.LoadLeafUnl(ctx, l.ID())
	require.NoError(t, err)
	assert.NotNil(t, l2) // reverse linked item that could have a nil pointer was not deleted
	assert.Nil(t, l2.RootUnl())
}

func TestReverseUniqueNullableLockTwo(t *testing.T) {
	ctx := context.Background()
	defer goradd_unit.ClearAll(ctx)
	r := goradd_unit.NewRootUnl()
	l := goradd_unit.NewLeafUnl()
	r.SetName("root")
	l.SetName("leaf")
	r.SetLeafUnl(l)
	require.NoError(t, r.Save(ctx))

	l2 := goradd_unit.NewLeafUnl()
	l2.SetName("leaf2")
	r.SetLeafUnl(l2)
	require.NoError(t, r.Save(ctx)) // unique failure

	// Confirm detached.
	l3, _ := goradd_unit.LoadLeafUnl(ctx, l.ID())
	assert.Nil(t, l3.RootUnl())
}

func TestReverseUniqueNullableLockDelete(t *testing.T) {
	ctx := context.Background()
	defer goradd_unit.ClearAll(ctx)
	l := goradd_unit.NewLeafUnl()
	r := goradd_unit.NewRootUnl()
	l.SetName("leaf")
	r.SetName("root")
	r.SetLeafUnl(l)
	require.NoError(t, r.Save(ctx))

	// Collision on shallow change
	r2, err := goradd_unit.LoadRootUnl(ctx, r.ID(), node.RootUnl().LeafUnl())
	require.NoError(t, err)
	r.SetName("root2")
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
