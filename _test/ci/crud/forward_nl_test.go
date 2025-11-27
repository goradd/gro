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

// TestForwardNullable tests insert and update of two linked records where the link is nullable.
func TestForwardNullableLock(t *testing.T) {
	ctx := context.Background()

	// Insert-insert
	l := goradd_unit.NewLeafNl()
	r := goradd_unit.NewRootNl()
	l.SetName("leafForwardNullableLock")
	r.SetName("rootForwardNullableLock")
	l.SetRootNl(r)
	err := l.Save(ctx)
	require.NoError(t, err)

	var l2 *goradd_unit.LeafNl
	l2, err = goradd_unit.LoadLeafNl(ctx, l.ID(), node.LeafNl().RootNl())
	require.NoError(t, err)
	require.NotNilf(t, l2, "Object was nil based on ID %s", l.ID())
	assert.Equal(t, "leafForwardNullableLock", l2.Name())
	assert.Equal(t, "rootForwardNullableLock", l2.RootNl().Name())

	// Update-update
	l.SetName("leafForwardNullableLock2")
	l.RootNl().SetName("rootForwardNullableLock2")
	err = l.Save(ctx)
	assert.NoError(t, err)
	l2, err = goradd_unit.LoadLeafNl(ctx, l.ID(), node.LeafNl().RootNl())
	require.NoError(t, err)
	require.NotNilf(t, l2, "Object was nil based on ID %s", l.ID())
	assert.Equal(t, "leafForwardNullableLock2", l2.Name())
	assert.Equal(t, "rootForwardNullableLock2", l2.RootNl().Name())

	// Insert-update
	l3 := goradd_unit.NewLeafNl()
	l3.SetName("leafForwardNullableLock3")
	r.SetName("rootForwardNullableLock3")
	l3.SetRootNl(r)
	err = l3.Save(ctx)
	require.NoError(t, err)
	l2, err = goradd_unit.LoadLeafNl(ctx, l3.ID(), node.LeafNl().RootNl())
	require.NoError(t, err)
	require.NotNilf(t, l2, "Object was nil based on ID %s", l3.ID())
	assert.Equal(t, "leafForwardNullableLock3", l2.Name())
	assert.Equal(t, "rootForwardNullableLock3", l2.RootNl().Name())

	// Update-insert
	r4 := goradd_unit.NewRootNl()
	l.SetName("leafForwardNullableLock4")
	r4.SetName("rootForwardNullableLock4")
	l.SetRootNl(r4)
	err = l.Save(ctx)
	require.NoError(t, err)
	l2, err = goradd_unit.LoadLeafNl(ctx, l.ID(), node.LeafNl().RootNl())
	require.NoError(t, err)
	require.NotNilf(t, l2, "Object was nil based on ID %s", l.ID())
	assert.Equal(t, "leafForwardNullableLock4", l2.Name())
	assert.Equal(t, "rootForwardNullableLock4", l2.RootNl().Name())
}

// TestForwardNullableCollision tests saving two records that are changed at the same time.
func TestForwardNullableLockCollision(t *testing.T) {
	ctx := context.Background()

	l := goradd_unit.NewLeafNl()
	r := goradd_unit.NewRootNl()
	r.SetName("rootForwardNullableLockCollision")
	l.SetName("leafForwardNullableLockCollision")
	l.SetRootNl(r)
	err := l.Save(ctx)
	require.NoError(t, err)

	var l2 *goradd_unit.LeafNl
	l2, err = goradd_unit.LoadLeafNl(ctx, l.ID(), node.LeafNl().RootNl())
	require.NoError(t, err)

	// Update both
	l.SetName("leafForwardNullableLockCollision2")
	l2.SetName("leafForwardNullableLockCollision3")

	err = l.Save(ctx)
	err2 := l2.Save(ctx)
	assert.NoError(t, err)
	assert.Error(t, err2)
	assert.IsType(t, &db.OptimisticLockError{}, err2)

	// 2nd level
	l2, _ = goradd_unit.LoadLeafNl(ctx, l.ID(), node.LeafNl().RootNl())
	l.RootNl().SetName("rootForwardNullableLockCollision2")
	l2.RootNl().SetName("rootForwardNullableLockCollision3")

	err = l.Save(ctx)
	err2 = l2.Save(ctx)
	assert.NoError(t, err)
	assert.Error(t, err2)
	assert.IsType(t, &db.OptimisticLockError{}, err2)

	// Delete
	l2, _ = goradd_unit.LoadLeafNl(ctx, l.ID(), node.LeafNl().RootNl())
	assert.NoError(t, l.RootNl().Delete(ctx))
	l2.SetName("leafForwardNullableLockCollision4")
	err2 = l2.Save(ctx)
	assert.IsType(t, &db.OptimisticLockError{}, err2)

}

func TestForwardNullableLockNull(t *testing.T) {
	ctx := context.Background()
	l := goradd_unit.NewLeafNl()
	l.SetName("leafForwardNullableLockNull")
	assert.NoError(t, l.Save(ctx))

	l.SetRootNl(nil) // nullable
	assert.NoError(t, l.Save(ctx))
}

func TestForwardNullableLockTwo(t *testing.T) {
	ctx := context.Background()
	defer goradd_unit.ClearAll(ctx)
	l := goradd_unit.NewLeafNl()
	r := goradd_unit.NewRootNl()
	l.SetName("leafForwardNullableLockTwo")
	r.SetName("rootForwardNullableLockTwo")
	l.SetRootNl(r)
	require.NoError(t, l.Save(ctx))

	l2 := goradd_unit.NewLeafNl()
	l2.SetName("leafForwardNullableLockTwo2")
	l2.SetRootNl(r)
	require.NoError(t, l2.Save(ctx))

	r2, err := goradd_unit.LoadRootNl(ctx, r.ID(), node.RootNl().LeafNls())
	require.NoError(t, err)
	assert.Len(t, r2.LeafNls(), 2)
}

func TestForwardNullableLockDelete(t *testing.T) {
	ctx := context.Background()
	l := goradd_unit.NewLeafNl()
	r := goradd_unit.NewRootNl()
	l.SetName("leafForwardNullableLockDelete")
	r.SetName("rootForwardNullableLockDelete")
	l.SetRootNl(r)
	require.NoError(t, l.Save(ctx))

	// Collision on Change
	l2, err := goradd_unit.LoadLeafNl(ctx, l.ID())
	require.NoError(t, err)
	l.SetName("leafForwardNullableLockDelete2")
	_ = l.Save(ctx)
	err = l2.Delete(ctx)
	require.Error(t, err)
	assert.IsType(t, &db.OptimisticLockError{}, err)

	// Collision on deep Delete
	l2, err = goradd_unit.LoadLeafNl(ctx, l.ID())
	require.NoError(t, err)
	err = l.RootNl().Delete(ctx)
	require.NoError(t, err)
	err = l2.Delete(ctx)
	assert.Error(t, err)
	assert.IsType(t, &db.OptimisticLockError{}, err)

	// Deep delete set link to null
	l2, err = goradd_unit.LoadLeafNl(ctx, l.ID())
	assert.NoError(t, err)
	assert.NotNil(t, l2)

}
