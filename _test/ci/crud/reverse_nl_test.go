package crud

import (
	"github.com/goradd/orm/_test/gen/orm/goradd_unit"
	"github.com/goradd/orm/_test/gen/orm/goradd_unit/node"
	"github.com/goradd/orm/pkg/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

// TestReverseNullableLock tests insert and update of two linked records.
func TestReverseNullableLock(t *testing.T) {
	ctx := db.NewContext(nil)
	defer goradd_unit.ClearAll(ctx)

	// Insert-insert
	r := goradd_unit.NewRootNl()
	l := goradd_unit.NewLeafNl()
	r.SetName("root")
	l.SetName("leaf")
	r.SetLeafNls(l)
	err := r.Save(ctx)
	require.NoError(t, err)

	var r2 *goradd_unit.RootNl
	r2, err = goradd_unit.LoadRootNl(ctx, r.ID(), node.RootNl().LeafNls())
	require.NoError(t, err)
	assert.Equal(t, "root", r2.Name())
	assert.Equal(t, "leaf", r2.LeafNls()[0].Name())

	// Update-update
	r.SetName("root2")
	r.LeafNls()[0].SetName("leaf2")
	err = r.Save(ctx)
	assert.NoError(t, err)
	r2, err = goradd_unit.LoadRootNl(ctx, r.ID(), node.RootNl().LeafNls())
	require.NoError(t, err)
	assert.Equal(t, "root2", r2.Name())
	assert.Equal(t, "leaf2", r2.LeafNls()[0].Name())

	// Insert-update
	r3 := goradd_unit.NewRootNl()
	r3.SetName("root3")
	l.SetName("leaf3")
	r3.SetLeafNls(l)
	err = r3.Save(ctx)
	require.NoError(t, err)
	r2, err = goradd_unit.LoadRootNl(ctx, r3.ID(), node.RootNl().LeafNls())
	require.NoError(t, err)
	assert.Equal(t, "root3", r2.Name())
	assert.Equal(t, "leaf3", r2.LeafNls()[0].Name())

	// Update-insert
	l4 := goradd_unit.NewLeafNl()
	r.SetName("root4")
	l4.SetName("leaf4")
	r.SetLeafNls(l4)
	err = r.Save(ctx)
	require.NoError(t, err)
	r2, err = goradd_unit.LoadRootNl(ctx, r.ID(), node.RootNl().LeafNls())
	require.NoError(t, err)
	assert.Equal(t, "root4", r2.Name())
	assert.Equal(t, "leaf4", r2.LeafNls()[0].Name())

}

// TestReverseNullableLockCollision tests saving two records that are changed at the same time.
func TestReverseNullableLockCollision(t *testing.T) {
	ctx := db.NewContext(nil)
	defer goradd_unit.ClearAll(ctx)

	r := goradd_unit.NewRootNl()
	l := goradd_unit.NewLeafNl()
	r.SetName("root")
	l.SetName("leaf")
	r.SetLeafNls(l)
	err := r.Save(ctx)
	require.NoError(t, err)

	var r2 *goradd_unit.RootNl
	r2, err = goradd_unit.LoadRootNl(ctx, r.ID(), node.RootNl().LeafNls())
	require.NoError(t, err)

	r.SetName("root2")
	r2.SetName("root3")

	err = r.Save(ctx)
	err2 := r2.Save(ctx)
	assert.NoError(t, err)
	assert.Error(t, err2)
	assert.IsType(t, &db.OptimisticLockError{}, err2)

	// Level 2
	r2, _ = goradd_unit.LoadRootNl(ctx, r.ID(), node.RootNl().LeafNls())

	r.LeafNls()[0].SetName("leaf2")
	r2.LeafNls()[0].SetName("leaf3")
	err = r.Save(ctx)
	err2 = r2.Save(ctx)
	assert.NoError(t, err)
	assert.Error(t, err2)
	assert.IsType(t, &db.OptimisticLockError{}, err2)
}

func TestReverseNullableLockNull(t *testing.T) {
	ctx := db.NewContext(nil)
	defer goradd_unit.ClearAll(ctx)

	r := goradd_unit.NewRootNl()
	r.SetName("root")
	l := goradd_unit.NewLeafNl()
	l.SetName("leaf")
	r.SetLeafNls(l)
	require.NoError(t, r.Save(ctx))

	r.SetLeafNls()
	require.NoError(t, r.Save(ctx))

	r2, err := goradd_unit.LoadRootNl(ctx, r.ID(), node.RootNl().LeafNls())
	require.NoError(t, err)
	assert.Len(t, r2.LeafNls(), 0)

	l2, err := goradd_unit.LoadLeafNl(ctx, l.ID())
	require.NoError(t, err)
	require.NotNil(t, l2)      // reverse linked item that could  have a nil pointer was retained
	assert.Nil(t, l2.RootNl()) // old reference was updated to nil
}

func TestReverseNullableLockTwo(t *testing.T) {
	ctx := db.NewContext(nil)
	defer goradd_unit.ClearAll(ctx)
	r := goradd_unit.NewRootNl()
	l := goradd_unit.NewLeafNl()
	r.SetName("root")
	l.SetName("leaf")
	r.SetLeafNls(l)
	require.NoError(t, r.Save(ctx))

	l2 := goradd_unit.NewLeafNl()
	l2.SetName("leaf2")
	r.SetLeafNls(l, l2)
	require.NoError(t, r.Save(ctx))

	r2, err := goradd_unit.LoadRootNl(ctx, r.ID(), node.RootNl().LeafNls())
	require.NoError(t, err)
	assert.Len(t, r2.LeafNls(), 2)

	r.SetLeafNls()
	require.NoError(t, r.Save(ctx))
	r2, err = goradd_unit.LoadRootNl(ctx, r.ID(), node.RootNl().LeafNls())
	require.NoError(t, err)
	assert.Len(t, r2.LeafNls(), 0)
}
