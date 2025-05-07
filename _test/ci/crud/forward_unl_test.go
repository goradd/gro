package crud

import (
	"github.com/goradd/orm/_test/gen/orm/goradd_unit"
	"github.com/goradd/orm/_test/gen/orm/goradd_unit/node"
	"github.com/goradd/orm/pkg/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

// TestForwardUniqueNullableLock tests insert and update of two linked records where the link is nullable.
func TestForwardUniqueNullableLock(t *testing.T) {
	ctx := db.NewContext(nil)
	defer goradd_unit.ClearAll(ctx)

	// Insert-insert
	l1 := goradd_unit.NewLeafUnl()
	r1 := goradd_unit.NewRootUnl()
	r1.SetName("root")
	l1.SetName("leaf")
	l1.SetRootUnl(r1)
	require.NoError(t, l1.Save(ctx))

	// loading back
	l1b, err := goradd_unit.LoadLeafUnl(ctx, l1.ID(), node.LeafUnl().RootUnl())
	require.NoError(t, err)
	assert.Equal(t, "leaf", l1b.Name())
	assert.Equal(t, "root", l1b.RootUnl().Name())

	// Update-update
	l1.SetName("leaf2")
	l1.RootUnl().SetName("root2")
	err = l1.Save(ctx)
	assert.NoError(t, err)
	l1b, err = goradd_unit.LoadLeafUnl(ctx, l1.ID(), node.LeafUnl().RootUnl())
	require.NoError(t, err)
	assert.Equal(t, "leaf2", l1b.Name())
	assert.Equal(t, "root2", l1b.RootUnl().Name())

	// Insert-update
	l2 := goradd_unit.NewLeafUnl()
	l2.SetName("leaf3")
	r1.SetName("root3")
	l2.SetRootUnl(r1)
	err = l2.Save(ctx)
	assert.Error(t, err)
	assert.IsType(t, &db.UniqueValueError{}, err)

	// Update-insert
	r3 := goradd_unit.NewRootUnl()
	l1.SetName("leaf4")
	r3.SetName("root4")
	l1.SetRootUnl(r3)
	err = l1.Save(ctx)
	require.NoError(t, err)
	l1b, err = goradd_unit.LoadLeafUnl(ctx, l1.ID(), node.LeafUnl().RootUnl())
	require.NoError(t, err)
	assert.Equal(t, "leaf4", l1b.Name())
	assert.Equal(t, "root4", l1b.RootUnl().Name())
}

// TestForwardNullableCollision tests saving two records that are changed at the same time.
func TestForwardUniqueNullableLockCollision(t *testing.T) {
	ctx := db.NewContext(nil)
	defer goradd_unit.ClearAll(ctx)

	l := goradd_unit.NewLeafUnl()
	r := goradd_unit.NewRootUnl()
	r.SetName("root")
	l.SetName("leaf")
	l.SetRootUnl(r)
	err := l.Save(ctx)
	require.NoError(t, err)

	var l2 *goradd_unit.LeafUnl
	l2, err = goradd_unit.LoadLeafUnl(ctx, l.ID(), node.LeafUnl().RootUnl())
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
	l2, _ = goradd_unit.LoadLeafUnl(ctx, l.ID(), node.LeafUnl().RootUnl())
	l.RootUnl().SetName("root2")
	l2.RootUnl().SetName("root3")

	err = l.Save(ctx)
	err2 = l2.Save(ctx)
	assert.NoError(t, err)
	assert.Error(t, err2)
	assert.IsType(t, &db.OptimisticLockError{}, err2)

	// Delete
	l2, _ = goradd_unit.LoadLeafUnl(ctx, l.ID(), node.LeafUnl().RootUnl())
	assert.NoError(t, l.RootUnl().Delete(ctx))
	l2.SetName("leaf4")
	err2 = l2.Save(ctx)
	assert.IsType(t, &db.OptimisticLockError{}, err2)
}

func TestForwardUniqueNullableLockNull(t *testing.T) {
	ctx := db.NewContext(nil)
	defer goradd_unit.ClearAll(ctx)
	l := goradd_unit.NewLeafUnl()
	l.SetName("leaf")
	assert.NoError(t, l.Save(ctx))

	l.SetRootUnl(nil) // nullable
	assert.NoError(t, l.Save(ctx))
}

func TestForwardUniqueNullableLockTwo(t *testing.T) {
	ctx := db.NewContext(nil)
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

	l2, err := goradd_unit.LoadLeafUnl(ctx, l.ID())
	require.NoError(t, err)
	assert.NotNil(t, l2) // reverse linked item that is allowed to have a nil pointer was not deleted
}

func TestForwardUniqueNullableLockDelete(t *testing.T) {
	ctx := db.NewContext(nil)
	defer goradd_unit.ClearAll(ctx)
	l := goradd_unit.NewLeafUnl()
	r := goradd_unit.NewRootUnl()
	l.SetName("leaf")
	r.SetName("root")
	l.SetRootUnl(r)
	require.NoError(t, l.Save(ctx))

	// Collision on Change
	l2, err := goradd_unit.LoadLeafUnl(ctx, l.ID())
	require.NoError(t, err)
	l.SetName("leaf2")
	_ = l.Save(ctx)
	err = l2.Delete(ctx)
	require.Error(t, err)
	assert.IsType(t, &db.OptimisticLockError{}, err)

	// Collision on deep Delete
	l2, err = goradd_unit.LoadLeafUnl(ctx, l.ID())
	require.NoError(t, err)
	err = l.RootUnl().Delete(ctx)
	require.NoError(t, err)
	err = l2.Delete(ctx)
	assert.Error(t, err)
	assert.IsType(t, &db.OptimisticLockError{}, err)

	// Deep delete set link to null
	l2, err = goradd_unit.LoadLeafUnl(ctx, l.ID())
	assert.NoError(t, err)
	assert.NotNil(t, l2)
}
