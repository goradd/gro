package crud

import (
	"context"
	"testing"

	"github.com/goradd/gro/_test/gen/orm/goradd_unit"
	"github.com/goradd/gro/_test/gen/orm/goradd_unit/node"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestReverse tests insert and update of two linked records.
func TestReverseNullable(t *testing.T) {
	ctx := context.Background()

	// Insert-insert
	r := goradd_unit.NewRootN()
	l := goradd_unit.NewLeafN()
	r.SetName("rootReverseNullable")
	l.SetName("leafReverseNullable")
	r.SetLeafNs(l)
	err := r.Save(ctx)
	require.NoError(t, err)

	var r2 *goradd_unit.RootN
	r2, err = goradd_unit.LoadRootN(ctx, r.ID(), node.RootN().LeafNs())
	require.NoError(t, err)
	require.NotNilf(t, r2, "Object was nil based on ID %s", r.ID())
	assert.Equal(t, "rootReverseNullable", r2.Name())
	assert.Equal(t, "leafReverseNullable", r2.LeafNs()[0].Name())

	// Update-update
	r.SetName("rootReverseNullable2")
	r.LeafNs()[0].SetName("leafReverseNullable2")
	err = r.Save(ctx)
	assert.NoError(t, err)
	r2, err = goradd_unit.LoadRootN(ctx, r.ID(), node.RootN().LeafNs())
	require.NoError(t, err)
	require.NotNilf(t, r2, "Object was nil based on ID %s", r.ID())
	assert.Equal(t, "rootReverseNullable2", r2.Name())
	assert.Equal(t, "leafReverseNullable2", r2.LeafNs()[0].Name())

	// Insert-update
	r3 := goradd_unit.NewRootN()
	r3.SetName("rootReverseNullable3")
	l.SetName("leafReverseNullable3")
	r3.SetLeafNs(l)
	err = r3.Save(ctx)
	require.NoError(t, err)
	r2, err = goradd_unit.LoadRootN(ctx, r3.ID(), node.RootN().LeafNs())
	require.NoError(t, err)
	require.NotNilf(t, r2, "Object was nil based on ID %s", r3.ID())
	assert.Equal(t, "rootReverseNullable3", r2.Name())
	assert.Equal(t, "leafReverseNullable3", r2.LeafNs()[0].Name())

	// Update-insert
	l4 := goradd_unit.NewLeafN()
	r.SetName("rootReverseNullable4")
	l4.SetName("leafReverseNullable4")
	r.SetLeafNs(l4)
	err = r.Save(ctx)
	require.NoError(t, err)
	r2, err = goradd_unit.LoadRootN(ctx, r.ID(), node.RootN().LeafNs())
	require.NoError(t, err)
	require.NotNilf(t, r2, "Object was nil based on ID %s", r.ID())
	assert.Equal(t, "rootReverseNullable4", r2.Name())
	assert.Equal(t, "leafReverseNullable4", r2.LeafNs()[0].Name())

}

// TestReverseCollision tests saving two records that are changed at the same time.
func TestReverseNullableCollision(t *testing.T) {
	ctx := context.Background()

	r := goradd_unit.NewRootN()
	l := goradd_unit.NewLeafN()
	r.SetName("rootReverseNullableCollision")
	l.SetName("leafReverseNullableCollision")
	r.SetLeafNs(l)
	err := r.Save(ctx)
	require.NoError(t, err)

	var r2 *goradd_unit.RootN
	r2, err = goradd_unit.LoadRootN(ctx, r.ID(), node.RootN().LeafNs())
	require.NoError(t, err)

	// Update first
	r.SetName("rootReverseNullableCollision2")
	r.LeafNs()[0].SetName("leafReverseNullableCollision2")

	// Update second
	r2.SetName("rootReverseNullableCollision3")
	r2.LeafNs()[0].SetName("leafReverseNullableCollision3")

	// save first then second
	err = r.Save(ctx)
	err2 := r2.Save(ctx)
	assert.NoError(t, err)
	assert.NoError(t, err2)

	// Last save should win
	r3, err3 := goradd_unit.LoadRootN(ctx, r.ID(), node.RootN().LeafNs())
	assert.NoError(t, err3)
	assert.Equal(t, "rootReverseNullableCollision3", r3.Name())
	assert.Equal(t, "leafReverseNullableCollision3", r3.LeafNs()[0].Name())
}

func TestReverseNullableNull(t *testing.T) {
	ctx := context.Background()

	r := goradd_unit.NewRootN()
	r.SetName("rootReverseNullableNull")
	l := goradd_unit.NewLeafN()
	l.SetName("leafReverseNullableNull")
	r.SetLeafNs(l)
	require.NoError(t, r.Save(ctx))

	r.SetLeafNs()
	require.NoError(t, r.Save(ctx))

	r2, err := goradd_unit.LoadRootN(ctx, r.ID(), node.RootN().LeafNs())
	require.NoError(t, err)
	assert.Len(t, r2.LeafNs(), 0)

	l2, err := goradd_unit.LoadLeafN(ctx, l.ID())
	require.NoError(t, err)
	require.NotNil(t, l2)     // reverse linked item that could  have a nil pointer was retained
	assert.Nil(t, l2.RootN()) // old reference was updated to nil
}

func TestReverseNullableTwo(t *testing.T) {
	ctx := context.Background()
	r := goradd_unit.NewRootN()
	l := goradd_unit.NewLeafN()
	r.SetName("rootReverseNullableTwo")
	l.SetName("leafReverseNullableTwo")
	r.SetLeafNs(l)
	require.NoError(t, r.Save(ctx))

	l2 := goradd_unit.NewLeafN()
	l2.SetName("leafReverseNullableTwo2")
	r.SetLeafNs(l, l2)
	require.NoError(t, r.Save(ctx))

	r2, err := goradd_unit.LoadRootN(ctx, r.ID(), node.RootN().LeafNs())
	require.NoError(t, err)
	assert.Len(t, r2.LeafNs(), 2)

	r.SetLeafNs()
	require.NoError(t, r.Save(ctx))
	r2, err = goradd_unit.LoadRootN(ctx, r.ID(), node.RootN().LeafNs())
	require.NoError(t, err)
	assert.Len(t, r2.LeafNs(), 0)
}
