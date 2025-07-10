package crud

import (
	"github.com/goradd/orm/_test/gen/orm/goradd_unit"
	"github.com/goradd/orm/_test/gen/orm/goradd_unit/node"
	"github.com/goradd/orm/pkg/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

// TestReverseUnique tests insert and update of two linked records.
func TestReverseUnique(t *testing.T) {
	ctx := context.Background()
	defer goradd_unit.ClearAll(ctx)

	// Insert-insert
	r := goradd_unit.NewRootU()
	l := goradd_unit.NewLeafU()
	r.SetName("root")
	l.SetName("leaf")
	r.SetLeafU(l)
	err := r.Save(ctx)
	require.NoError(t, err)

	var r2 *goradd_unit.RootU
	r2, err = goradd_unit.LoadRootU(ctx, r.ID(), node.RootU().LeafU())
	require.NoError(t, err)
	assert.Equal(t, "root", r2.Name())
	assert.Equal(t, "leaf", r2.LeafU().Name())

	// Update-update
	r.SetName("root2")
	r.LeafU().SetName("leaf2")
	err = r.Save(ctx)
	assert.NoError(t, err)
	r2, err = goradd_unit.LoadRootU(ctx, r.ID(), node.RootU().LeafU())
	require.NoError(t, err)
	assert.Equal(t, "root2", r2.Name())
	assert.Equal(t, "leaf2", r2.LeafU().Name())

	// Insert-update
	r3 := goradd_unit.NewRootU()
	r3.SetName("root3")
	l.SetName("leaf3")
	r3.SetLeafU(l)
	err = r3.Save(ctx)
	require.NoError(t, err)
	r2, err = goradd_unit.LoadRootU(ctx, r3.ID(), node.RootU().LeafU())
	require.NoError(t, err)
	assert.Equal(t, "root3", r2.Name())
	assert.Equal(t, "leaf3", r2.LeafU().Name())

	// Update-insert
	l4 := goradd_unit.NewLeafU()
	r.SetName("root4")
	l4.SetName("leaf4")
	r.SetLeafU(l4)
	err = r.Save(ctx)
	require.NoError(t, err)
	r2, err = goradd_unit.LoadRootU(ctx, r.ID(), node.RootU().LeafU())
	require.NoError(t, err)
	assert.Equal(t, "root4", r2.Name())
	assert.Equal(t, "leaf4", r2.LeafU().Name())

}

// TestReverseCollision tests saving two records that are changed at the same time.
func TestReverseUniqueCollision(t *testing.T) {
	ctx := context.Background()
	defer goradd_unit.ClearAll(ctx)

	r := goradd_unit.NewRootU()
	l := goradd_unit.NewLeafU()
	r.SetName("root")
	l.SetName("leaf")
	r.SetLeafU(l)
	err := r.Save(ctx)
	require.NoError(t, err)

	var r2 *goradd_unit.RootU
	r2, err = goradd_unit.LoadRootU(ctx, r.ID(), node.RootU().LeafU())
	require.NoError(t, err)

	// Update first
	r.SetName("root2")
	r.LeafU().SetName("leaf2")

	// Update second
	r2.SetName("root3")
	r2.LeafU().SetName("leaf3")

	// save first then second
	err = r.Save(ctx)
	err2 := r2.Save(ctx)
	assert.NoError(t, err)
	assert.NoError(t, err2)

	// Last save should win
	r3, err3 := goradd_unit.LoadRootU(ctx, r.ID(), node.RootU().LeafU())
	assert.NoError(t, err3)
	assert.Equal(t, "root3", r3.Name())
	assert.Equal(t, "leaf3", r3.LeafU().Name())
}

func TestReverseUniqueNull(t *testing.T) {
	ctx := context.Background()
	defer goradd_unit.ClearAll(ctx)

	r := goradd_unit.NewRootU()
	r.SetName("root")
	l := goradd_unit.NewLeafU()
	l.SetName("leaf")
	r.SetLeafU(l)
	require.NoError(t, r.Save(ctx))

	r.SetLeafU(nil)
	require.NoError(t, r.Save(ctx))

	r2, err := goradd_unit.LoadRootU(ctx, r.ID(), node.RootU().LeafU())
	require.NoError(t, err)
	assert.Nil(t, r2.LeafU())

	l2, err := goradd_unit.LoadLeafU(ctx, l.ID())
	require.NoError(t, err)
	assert.Nil(t, l2) // reverse linked item that could not have a nil pointer was deleted
}

func TestReverseUniqueTwo(t *testing.T) {
	ctx := context.Background()
	defer goradd_unit.ClearAll(ctx)
	r := goradd_unit.NewRootU()
	l := goradd_unit.NewLeafU()
	r.SetName("root")
	l.SetName("leaf")
	r.SetLeafU(l)
	require.NoError(t, r.Save(ctx))

	l2 := goradd_unit.NewLeafU()
	l2.SetName("leaf2")
	r.SetLeafU(l2)
	require.NoError(t, r.Save(ctx)) // unique failure

	l2, err := goradd_unit.LoadLeafU(ctx, l.ID())
	require.NoError(t, err)
	assert.Nil(t, l2) // reverse linked item that could not have a nil pointer was deleted
}
