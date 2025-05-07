package crud

import (
	"github.com/goradd/orm/_test/gen/orm/goradd_unit"
	"github.com/goradd/orm/_test/gen/orm/goradd_unit/node"
	"github.com/goradd/orm/pkg/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

// TestForwardLock tests insert and update of two linked records that have an optimistic lock.
func TestForwardLock(t *testing.T) {
	ctx := db.NewContext(nil)
	defer goradd_unit.ClearAll(ctx)

	// Insert-insert
	l := goradd_unit.NewLeafL()
	r := goradd_unit.NewRootL()
	r.SetName("root")
	l.SetName("leaf")
	l.SetRootL(r)
	err := l.Save(ctx)
	require.NoError(t, err)

	var l2 *goradd_unit.LeafL
	l2, err = goradd_unit.LoadLeafL(ctx, l.ID(), node.LeafL().RootL())
	require.NoError(t, err)
	assert.Equal(t, "leaf", l2.Name())
	assert.Equal(t, "root", l2.RootL().Name())

	// Update-update
	l.SetName("leaf2")
	l.RootL().SetName("root2")
	err = l.Save(ctx)
	assert.NoError(t, err)
	l2, err = goradd_unit.LoadLeafL(ctx, l.ID(), node.LeafL().RootL())
	require.NoError(t, err)
	assert.Equal(t, "leaf2", l2.Name())
	assert.Equal(t, "root2", l2.RootL().Name())

	// Insert-update
	l3 := goradd_unit.NewLeafL()
	l3.SetName("leaf3")
	r.SetName("root3")
	l3.SetRootL(r)
	err = l3.Save(ctx)
	require.NoError(t, err)
	l2, err = goradd_unit.LoadLeafL(ctx, l3.ID(), node.LeafL().RootL())
	require.NoError(t, err)
	assert.Equal(t, "leaf3", l2.Name())
	assert.Equal(t, "root3", l2.RootL().Name())

	// Update-insert
	r4 := goradd_unit.NewRootL()
	l.SetName("leaf4")
	r4.SetName("root4")
	l.SetRootL(r4)
	err = l.Save(ctx)
	require.NoError(t, err)
	l2, err = goradd_unit.LoadLeafL(ctx, l.ID(), node.LeafL().RootL())
	require.NoError(t, err)
	assert.Equal(t, "leaf4", l2.Name())
	assert.Equal(t, "root4", l2.RootL().Name())
}

// TestForwardCollision tests saving two records that are changed at the same time.
func TestForwardLockCollision(t *testing.T) {
	ctx := db.NewContext(nil)
	defer goradd_unit.ClearAll(ctx)

	l := goradd_unit.NewLeafL()
	r := goradd_unit.NewRootL()
	r.SetName("root")
	l.SetName("leaf")
	l.SetRootL(r)
	err := l.Save(ctx)
	require.NoError(t, err)

	var l2 *goradd_unit.LeafL
	l2, err = goradd_unit.LoadLeafL(ctx, l.ID(), node.LeafL().RootL())
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
	l2, _ = goradd_unit.LoadLeafL(ctx, l.ID(), node.LeafL().RootL())
	l.RootL().SetName("root2")
	l2.RootL().SetName("root3")

	err = l.Save(ctx)
	err2 = l2.Save(ctx)
	assert.NoError(t, err)
	assert.Error(t, err2)
	assert.IsType(t, &db.OptimisticLockError{}, err2)
}

func TestForwardLockNull(t *testing.T) {
	ctx := db.NewContext(nil)
	defer goradd_unit.ClearAll(ctx)
	l := goradd_unit.NewLeafL()
	l.SetName("leaf")
	assert.Panics(t, func() {
		_ = l.Save(ctx) // root is required since RootID is not nullable
	})

	assert.Panics(t, func() {
		l.SetRootL(nil) // not nullable
	})
}

func TestForwardLockTwo(t *testing.T) {
	ctx := db.NewContext(nil)
	defer goradd_unit.ClearAll(ctx)
	l := goradd_unit.NewLeafL()
	r := goradd_unit.NewRootL()
	l.SetName("leaf")
	r.SetName("root")
	l.SetRootL(r)
	require.NoError(t, l.Save(ctx))

	l2 := goradd_unit.NewLeafL()
	l2.SetName("leaf2")
	l2.SetRootL(r)
	require.NoError(t, l2.Save(ctx))

	r2, err := goradd_unit.LoadRootL(ctx, r.ID(), node.RootL().LeafLs())
	require.NoError(t, err)
	assert.Len(t, r2.LeafLs(), 2)
}

func TestForwardLockDelete(t *testing.T) {
	ctx := db.NewContext(nil)
	defer goradd_unit.ClearAll(ctx)
	l := goradd_unit.NewLeafL()
	r := goradd_unit.NewRootL()
	l.SetName("leaf")
	r.SetName("root")
	l.SetRootL(r)
	require.NoError(t, l.Save(ctx))

	// iterate on delete and change collisions

	l2, err := goradd_unit.LoadLeafL(ctx, l.ID())
	require.NoError(t, err)
	l.SetName("leaf2")
	_ = l.Save(ctx)
	err = l2.Delete(ctx)
	require.Error(t, err)
	assert.IsType(t, &db.OptimisticLockError{}, err)

	l2, err = goradd_unit.LoadLeafL(ctx, l.ID())
	require.NoError(t, err)
	err = l.Delete(ctx)
	require.NoError(t, err)
	err = l2.Delete(ctx)
	assert.Error(t, err)
	assert.IsType(t, &db.OptimisticLockError{}, err)
}
