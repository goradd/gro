package crud

import (
	"context"
	"testing"

	"github.com/goradd/orm/_test/gen/orm/goradd_unit"
	"github.com/goradd/orm/_test/gen/orm/goradd_unit/node"
	"github.com/goradd/orm/pkg/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestReverseUniqueLock tests insert and update of two linked records.
func TestReverseUniqueLock(t *testing.T) {
	ctx := context.Background()
	defer goradd_unit.ClearAll(ctx)

	// Insert-insert
	r := goradd_unit.NewRootUl()
	l := goradd_unit.NewLeafUl()
	r.SetName("root")
	l.SetName("leaf")
	r.SetLeafUl(l)
	err := r.Save(ctx)
	require.NoError(t, err)

	var r2 *goradd_unit.RootUl
	r2, err = goradd_unit.LoadRootUl(ctx, r.ID(), node.RootUl().LeafUl())
	require.NoError(t, err)
	require.NotNilf(t, r2, "Object was nil based on ID %s", r.ID())
	assert.Equal(t, "root", r2.Name())
	assert.Equal(t, "leaf", r2.LeafUl().Name())

	// Update-update
	r.SetName("root2")
	r.LeafUl().SetName("leaf2")
	err = r.Save(ctx)
	assert.NoError(t, err)
	r2, err = goradd_unit.LoadRootUl(ctx, r.ID(), node.RootUl().LeafUl())
	require.NoError(t, err)
	require.NotNilf(t, r2, "Object was nil based on ID %s", r.ID())
	assert.Equal(t, "root2", r2.Name())
	assert.Equal(t, "leaf2", r2.LeafUl().Name())

	// Insert-update
	r3 := goradd_unit.NewRootUl()
	r3.SetName("root3")
	l.SetName("leaf3")
	r3.SetLeafUl(l)
	err = r3.Save(ctx)
	require.NoError(t, err)
	r2, err = goradd_unit.LoadRootUl(ctx, r3.ID(), node.RootUl().LeafUl())
	require.NoError(t, err)
	require.NotNilf(t, r2, "Object was nil based on ID %s", r3.ID())
	assert.Equal(t, "root3", r2.Name())
	assert.Equal(t, "leaf3", r2.LeafUl().Name())

	// Update-insert
	l4 := goradd_unit.NewLeafUl()
	r.SetName("root4")
	l4.SetName("leaf4")
	r.SetLeafUl(l4)
	err = r.Save(ctx)
	require.NoError(t, err)
	r2, err = goradd_unit.LoadRootUl(ctx, r.ID(), node.RootUl().LeafUl())
	require.NoError(t, err)
	require.NotNilf(t, r2, "Object was nil based on ID %s", r.ID())
	assert.Equal(t, "root4", r2.Name())
	assert.Equal(t, "leaf4", r2.LeafUl().Name())

}

// TestReverseCollision tests saving two records that are changed at the same time.
func TestReverseUniqueLockCollision(t *testing.T) {
	ctx := context.Background()
	defer goradd_unit.ClearAll(ctx)

	r := goradd_unit.NewRootUl()
	l := goradd_unit.NewLeafUl()
	r.SetName("root")
	l.SetName("leaf")
	r.SetLeafUl(l)
	err := r.Save(ctx)
	require.NoError(t, err)

	var r2 *goradd_unit.RootUl
	r2, err = goradd_unit.LoadRootUl(ctx, r.ID(), node.RootUl().LeafUl())
	require.NoError(t, err)

	r.SetName("root2")
	r2.SetName("root3")

	err = r.Save(ctx)
	err2 := r2.Save(ctx)
	assert.NoError(t, err)
	assert.Error(t, err2)
	assert.IsType(t, &db.OptimisticLockError{}, err2)

	// Level 2
	r2, _ = goradd_unit.LoadRootUl(ctx, r.ID(), node.RootUl().LeafUl())

	r.LeafUl().SetName("leaf2")
	r2.LeafUl().SetName("leaf3")
	err = r.Save(ctx)
	err2 = r2.Save(ctx)
	assert.NoError(t, err)
	assert.Error(t, err2)
	assert.IsType(t, &db.OptimisticLockError{}, err2)

	// Delete
	r2, _ = goradd_unit.LoadRootUl(ctx, r.ID(), node.RootUl().LeafUl())
	assert.NoError(t, r.Delete(ctx))
	r2.SetName("root4")
	err2 = r2.Save(ctx)
	assert.IsType(t, &db.OptimisticLockError{}, err2)

}

func TestReverseUniqueLockNull(t *testing.T) {
	ctx := context.Background()
	defer goradd_unit.ClearAll(ctx)

	r := goradd_unit.NewRootUl()
	r.SetName("root")
	l := goradd_unit.NewLeafUl()
	l.SetName("leaf")
	r.SetLeafUl(l)
	require.NoError(t, r.Save(ctx))

	r.SetLeafUl(nil)
	require.NoError(t, r.Save(ctx))

	r2, err := goradd_unit.LoadRootUl(ctx, r.ID(), node.RootUl().LeafUl())
	require.NoError(t, err)
	assert.Nil(t, r2.LeafUl())

	l2, err := goradd_unit.LoadLeafUl(ctx, l.ID())
	require.NoError(t, err)
	assert.Nil(t, l2) // reverse linked item that could not have a nil pointer was deleted
}

func TestReverseUniqueLockTwo(t *testing.T) {
	ctx := context.Background()
	defer goradd_unit.ClearAll(ctx)
	r := goradd_unit.NewRootUl()
	l := goradd_unit.NewLeafUl()
	r.SetName("root")
	l.SetName("leaf")
	r.SetLeafUl(l)
	require.NoError(t, r.Save(ctx))

	l2 := goradd_unit.NewLeafUl()
	l2.SetName("leaf2")
	r.SetLeafUl(l2)
	require.NoError(t, r.Save(ctx)) // unique failure

	l2, err := goradd_unit.LoadLeafUl(ctx, l.ID())
	require.NoError(t, err)
	assert.Nil(t, l2) // reverse linked item that could not have a nil pointer was deleted
}

func TestReverseUniqueLockDelete(t *testing.T) {
	ctx := context.Background()
	defer goradd_unit.ClearAll(ctx)
	l := goradd_unit.NewLeafUl()
	r := goradd_unit.NewRootUl()
	l.SetName("leaf")
	r.SetName("root")
	r.SetLeafUl(l)
	require.NoError(t, r.Save(ctx))

	// Collision on shallow change
	r2, err := goradd_unit.LoadRootUl(ctx, r.ID(), node.RootUl().LeafUl())
	require.NoError(t, err)
	r.SetName("root2")
	_ = r.Save(ctx)
	err = r2.Delete(ctx)
	require.Error(t, err)
	assert.IsType(t, &db.OptimisticLockError{}, err)

	// No collision on deep Delete since it can't be detected
	r2, err = goradd_unit.LoadRootUl(ctx, r.ID(), node.RootUl().LeafUl())
	require.NoError(t, err)
	err = r.LeafUl().Delete(ctx)
	require.NoError(t, err)
	err = r2.Delete(ctx)
	assert.NoError(t, err)
}
