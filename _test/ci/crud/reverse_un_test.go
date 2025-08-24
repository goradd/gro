package crud

import (
	"context"
	"testing"

	"github.com/goradd/gro/_test/gen/orm/goradd_unit"
	"github.com/goradd/gro/_test/gen/orm/goradd_unit/node"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestReverseUniqueNullable tests insert and update of two linked records.
func TestReverseUniqueNullable(t *testing.T) {
	ctx := context.Background()
	defer goradd_unit.ClearAll(ctx)

	// Insert-insert
	r := goradd_unit.NewRootUn()
	l := goradd_unit.NewLeafUn()
	r.SetName("root")
	l.SetName("leaf")
	r.SetLeafUn(l)
	err := r.Save(ctx)
	require.NoError(t, err)

	var r2 *goradd_unit.RootUn
	r2, err = goradd_unit.LoadRootUn(ctx, r.ID(), node.RootUn().LeafUn())
	require.NoError(t, err)
	require.NotNilf(t, r2, "Object was nil based on ID %s", r.ID())
	assert.Equal(t, "root", r2.Name())
	assert.Equal(t, "leaf", r2.LeafUn().Name())

	// Update-update
	r.SetName("root2")
	r.LeafUn().SetName("leaf2")
	err = r.Save(ctx)
	assert.NoError(t, err)
	r2, err = goradd_unit.LoadRootUn(ctx, r.ID(), node.RootUn().LeafUn())
	require.NoError(t, err)
	require.NotNilf(t, r2, "Object was nil based on ID %s", r.ID())
	assert.Equal(t, "root2", r2.Name())
	assert.Equal(t, "leaf2", r2.LeafUn().Name())

	// Insert-update
	r3 := goradd_unit.NewRootUn()
	r3.SetName("root3")
	l.SetName("leaf3")
	r3.SetLeafUn(l)
	err = r3.Save(ctx)
	require.NoError(t, err)
	r2, err = goradd_unit.LoadRootUn(ctx, r3.ID(), node.RootUn().LeafUn())
	require.NoError(t, err)
	require.NotNilf(t, r2, "Object was nil based on ID %s", r3.ID())
	assert.Equal(t, "root3", r2.Name())
	assert.Equal(t, "leaf3", r2.LeafUn().Name())

	// Update-insert
	l4 := goradd_unit.NewLeafUn()
	r.SetName("root4")
	l4.SetName("leaf4")
	r.SetLeafUn(l4)
	err = r.Save(ctx)
	require.NoError(t, err)
	r2, err = goradd_unit.LoadRootUn(ctx, r.ID(), node.RootUn().LeafUn())
	require.NoError(t, err)
	require.NotNilf(t, r2, "Object was nil based on ID %s", r.ID())
	assert.Equal(t, "root4", r2.Name())
	assert.Equal(t, "leaf4", r2.LeafUn().Name())

}

// TestReverseCollision tests saving two records that are changed at the same time.
func TestReverseUniqueNullableCollision(t *testing.T) {
	ctx := context.Background()
	defer goradd_unit.ClearAll(ctx)

	r := goradd_unit.NewRootUn()
	l := goradd_unit.NewLeafUn()
	r.SetName("root")
	l.SetName("leaf")
	r.SetLeafUn(l)
	err := r.Save(ctx)
	require.NoError(t, err)

	var r2 *goradd_unit.RootUn
	r2, err = goradd_unit.LoadRootUn(ctx, r.ID(), node.RootUn().LeafUn())
	require.NoError(t, err)

	// Update first
	r.SetName("root2")
	r.LeafUn().SetName("leaf2")

	// Update second
	r2.SetName("root3")
	r2.LeafUn().SetName("leaf3")

	// save first then second
	err = r.Save(ctx)
	err2 := r2.Save(ctx)
	assert.NoError(t, err)
	assert.NoError(t, err2)

	// Last save should win
	r3, err3 := goradd_unit.LoadRootUn(ctx, r.ID(), node.RootUn().LeafUn())
	assert.NoError(t, err3)
	assert.Equal(t, "root3", r3.Name())
	assert.Equal(t, "leaf3", r3.LeafUn().Name())
}

func TestReverseUniqueNullableNull(t *testing.T) {
	ctx := context.Background()
	defer goradd_unit.ClearAll(ctx)

	r := goradd_unit.NewRootUn()
	r.SetName("root")
	l := goradd_unit.NewLeafUn()
	l.SetName("leaf")
	r.SetLeafUn(l)
	require.NoError(t, r.Save(ctx))

	r.SetLeafUn(nil)
	require.NoError(t, r.Save(ctx))

	r2, err := goradd_unit.LoadRootUn(ctx, r.ID(), node.RootUn().LeafUn())
	require.NoError(t, err)
	assert.Nil(t, r2.LeafUn())

	l2, err := goradd_unit.LoadLeafUn(ctx, l.ID())
	require.NoError(t, err)
	assert.NotNil(t, l2) // reverse linked item that could have a nil pointer was not deleted
	assert.Nil(t, l2.RootUn())
}

func TestReverseUniqueNullableTwo(t *testing.T) {
	ctx := context.Background()
	defer goradd_unit.ClearAll(ctx)
	r := goradd_unit.NewRootUn()
	l := goradd_unit.NewLeafUn()
	r.SetName("root")
	l.SetName("leaf")
	r.SetLeafUn(l)
	require.NoError(t, r.Save(ctx))

	l2 := goradd_unit.NewLeafUn()
	l2.SetName("leaf2")
	r.SetLeafUn(l2)
	require.NoError(t, r.Save(ctx)) // unique failure

	// Confirm detached.
	l3, _ := goradd_unit.LoadLeafUn(ctx, l.ID())
	assert.Nil(t, l3.RootUn())
}
