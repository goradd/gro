package crud

import (
	"github.com/goradd/orm/_test/gen/orm/goradd_unit"
	"github.com/goradd/orm/_test/gen/orm/goradd_unit/node"
	"github.com/goradd/orm/pkg/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

// TestForwardNullable tests insert and update of two linked records where the link is nullable.
func TestForwardNullableLock(t *testing.T) {
	ctx := db.NewContext(nil)
	defer goradd_unit.ClearAll(ctx)

	// Insert-insert
	l := goradd_unit.NewLeafNl()
	r := goradd_unit.NewRootNl()
	l.SetName("leaf")
	r.SetName("root")
	l.SetRootNl(r)
	err := l.Save(ctx)
	require.NoError(t, err)

	var l2 *goradd_unit.LeafNl
	l2, err = goradd_unit.LoadLeafNl(ctx, l.ID(), node.LeafNl().RootNl())
	require.NoError(t, err)
	assert.Equal(t, "leaf", l2.Name())
	assert.Equal(t, "root", l2.RootNl().Name())

	// Update-update
	l.SetName("leaf2")
	l.RootNl().SetName("root2")
	err = l.Save(ctx)
	assert.NoError(t, err)
	l2, err = goradd_unit.LoadLeafNl(ctx, l.ID(), node.LeafNl().RootNl())
	require.NoError(t, err)
	assert.Equal(t, "leaf2", l2.Name())
	assert.Equal(t, "root2", l2.RootNl().Name())

	// Insert-update
	l3 := goradd_unit.NewLeafNl()
	l3.SetName("leaf3")
	r.SetName("root3")
	l3.SetRootNl(r)
	err = l3.Save(ctx)
	require.NoError(t, err)
	l2, err = goradd_unit.LoadLeafNl(ctx, l3.ID(), node.LeafNl().RootNl())
	require.NoError(t, err)
	assert.Equal(t, "leaf3", l2.Name())
	assert.Equal(t, "root3", l2.RootNl().Name())

	// Update-insert
	r4 := goradd_unit.NewRootNl()
	l.SetName("leaf4")
	r4.SetName("root4")
	l.SetRootNl(r4)
	err = l.Save(ctx)
	require.NoError(t, err)
	l2, err = goradd_unit.LoadLeafNl(ctx, l.ID(), node.LeafNl().RootNl())
	require.NoError(t, err)
	assert.Equal(t, "leaf4", l2.Name())
	assert.Equal(t, "root4", l2.RootNl().Name())
}

// TestForwardNullableCollision tests saving two records that are changed at the same time.
func TestForwardNullableLockCollision(t *testing.T) {
	ctx := db.NewContext(nil)
	defer goradd_unit.ClearAll(ctx)

	l := goradd_unit.NewLeafNl()
	r := goradd_unit.NewRootNl()
	r.SetName("root")
	l.SetName("leaf")
	l.SetRootNl(r)
	err := l.Save(ctx)
	require.NoError(t, err)

	var l2 *goradd_unit.LeafNl
	l2, err = goradd_unit.LoadLeafNl(ctx, l.ID(), node.LeafNl().RootNl())
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
	l2, _ = goradd_unit.LoadLeafNl(ctx, l.ID(), node.LeafNl().RootNl())
	l.RootNl().SetName("root2")
	l2.RootNl().SetName("root3")

	err = l.Save(ctx)
	err2 = l2.Save(ctx)
	assert.NoError(t, err)
	assert.Error(t, err2)
	assert.IsType(t, &db.OptimisticLockError{}, err2)
}

func TestForwardNullableLockNull(t *testing.T) {
	ctx := db.NewContext(nil)
	defer goradd_unit.ClearAll(ctx)
	l := goradd_unit.NewLeafNl()
	l.SetName("leaf")
	assert.NoError(t, l.Save(ctx))

	l.SetRootNl(nil) // nullable
	assert.NoError(t, l.Save(ctx))
}

func TestForwardNullableLockTwo(t *testing.T) {
	ctx := db.NewContext(nil)
	defer goradd_unit.ClearAll(ctx)
	l := goradd_unit.NewLeafNl()
	r := goradd_unit.NewRootNl()
	l.SetName("leaf")
	r.SetName("root")
	l.SetRootNl(r)
	require.NoError(t, l.Save(ctx))

	l2 := goradd_unit.NewLeafNl()
	l2.SetName("leaf2")
	l2.SetRootNl(r)
	require.NoError(t, l2.Save(ctx))

	r2, err := goradd_unit.LoadRootNl(ctx, r.ID(), node.RootNl().LeafNls())
	require.NoError(t, err)
	assert.Len(t, r2.LeafNls(), 2)
}
