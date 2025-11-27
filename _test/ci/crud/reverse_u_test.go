package crud

import (
	"context"
	"testing"

	"github.com/goradd/gro/_test/gen/orm/goradd_unit"
	"github.com/goradd/gro/_test/gen/orm/goradd_unit/node"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestReverseUnique tests insert and update of two linked records.
func TestReverseUnique(t *testing.T) {
	ctx := context.Background()

	// Insert-insert
	r := goradd_unit.NewRootU()
	l := goradd_unit.NewLeafU()
	r.SetName("rootReverseUnique")
	l.SetName("leafReverseUnique")
	r.SetLeafU(l)
	err := r.Save(ctx)
	require.NoError(t, err)

	var r2 *goradd_unit.RootU
	r2, err = goradd_unit.LoadRootU(ctx, r.ID(), node.RootU().LeafU())
	require.NoError(t, err)
	require.NotNilf(t, r2, "Object was nil based on ID %s", r.ID())
	assert.Equal(t, "rootReverseUnique", r2.Name())
	assert.Equal(t, "leafReverseUnique", r2.LeafU().Name())

	// Update-update
	r.SetName("rootReverseUnique2")
	r.LeafU().SetName("leafReverseUnique2")
	err = r.Save(ctx)
	assert.NoError(t, err)
	r2, err = goradd_unit.LoadRootU(ctx, r.ID(), node.RootU().LeafU())
	require.NoError(t, err)
	require.NotNilf(t, r2, "Object was nil based on ID %s", r.ID())
	assert.Equal(t, "rootReverseUnique2", r2.Name())
	assert.Equal(t, "leafReverseUnique2", r2.LeafU().Name())

	// Insert-update
	r3 := goradd_unit.NewRootU()
	r3.SetName("rootReverseUnique3")
	l.SetName("leafReverseUnique3")
	r3.SetLeafU(l)
	err = r3.Save(ctx)
	require.NoError(t, err)
	r2, err = goradd_unit.LoadRootU(ctx, r3.ID(), node.RootU().LeafU())
	require.NoError(t, err)
	require.NotNilf(t, r2, "Object was nil based on ID %s", r3.ID())
	assert.Equal(t, "rootReverseUnique3", r2.Name())
	assert.Equal(t, "leafReverseUnique3", r2.LeafU().Name())

	// Update-insert
	l4 := goradd_unit.NewLeafU()
	r.SetName("rootReverseUnique4")
	l4.SetName("leafReverseUnique4")
	r.SetLeafU(l4)
	err = r.Save(ctx)
	require.NoError(t, err)
	r2, err = goradd_unit.LoadRootU(ctx, r.ID(), node.RootU().LeafU())
	require.NoError(t, err)
	require.NotNilf(t, r2, "Object was nil based on ID %s", r.ID())
	assert.Equal(t, "rootReverseUnique4", r2.Name())
	assert.Equal(t, "leafReverseUnique4", r2.LeafU().Name())

}

// TestReverseCollision tests saving two records that are changed at the same time.
func TestReverseUniqueCollision(t *testing.T) {
	ctx := context.Background()

	r := goradd_unit.NewRootU()
	l := goradd_unit.NewLeafU()
	r.SetName("rootReverseUniqueCollision")
	l.SetName("leafReverseUniqueCollision")
	r.SetLeafU(l)
	err := r.Save(ctx)
	require.NoError(t, err)

	var r2 *goradd_unit.RootU
	r2, err = goradd_unit.LoadRootU(ctx, r.ID(), node.RootU().LeafU())
	require.NoError(t, err)

	// Update first
	r.SetName("rootReverseUniqueCollision2")
	r.LeafU().SetName("leafReverseUniqueCollision2")

	// Update second
	r2.SetName("rootReverseUniqueCollision3")
	r2.LeafU().SetName("leafReverseUniqueCollision3")

	// save first then second
	err = r.Save(ctx)
	err2 := r2.Save(ctx)
	assert.NoError(t, err)
	assert.NoError(t, err2)

	// Last save should win
	r3, err3 := goradd_unit.LoadRootU(ctx, r.ID(), node.RootU().LeafU())
	assert.NoError(t, err3)
	assert.Equal(t, "rootReverseUniqueCollision3", r3.Name())
	assert.Equal(t, "leafReverseUniqueCollision3", r3.LeafU().Name())
}

func TestReverseUniqueNull(t *testing.T) {
	ctx := context.Background()

	r := goradd_unit.NewRootU()
	r.SetName("rootReverseUniqueNull")
	l := goradd_unit.NewLeafU()
	l.SetName("leafReverseUniqueNull")
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
	r := goradd_unit.NewRootU()
	l := goradd_unit.NewLeafU()
	r.SetName("rootReverseUniqueTwo")
	l.SetName("leafReverseUniqueTwo")
	r.SetLeafU(l)
	require.NoError(t, r.Save(ctx))

	l2 := goradd_unit.NewLeafU()
	l2.SetName("leafReverseUniqueTwo2")
	r.SetLeafU(l2)
	require.NoError(t, r.Save(ctx)) // unique failure

	l2, err := goradd_unit.LoadLeafU(ctx, l.ID())
	require.NoError(t, err)
	assert.Nil(t, l2) // reverse linked item that could not have a nil pointer was deleted
}
