package crud

import (
	"context"
	"testing"

	goradd_unit2 "github.com/goradd/gro/ci/tests/gen/goradd_unit"
	node2 "github.com/goradd/gro/ci/tests/gen/goradd_unit/node"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestForwardNullable tests insert and update of two linked records where the link is nullable.
func TestForwardNullable(t *testing.T) {
	ctx := context.Background()

	// Insert-insert
	l := goradd_unit2.NewLeafN()
	r := goradd_unit2.NewRootN()
	l.SetName("leafForwardNullable")
	r.SetName("rootForwardNullable")
	l.SetRootN(r)
	err := l.Save(ctx)
	require.NoError(t, err)

	var l2 *goradd_unit2.LeafN
	l2, err = goradd_unit2.LoadLeafN(ctx, l.ID(), node2.LeafN().RootN())
	require.NoError(t, err)
	require.NotNilf(t, l2, "Object was nil based on ID %s", l.ID())
	assert.Equal(t, "leafForwardNullable", l2.Name())
	assert.Equal(t, "rootForwardNullable", l2.RootN().Name())

	// Update-update
	l.SetName("leafForwardNullable2")
	l.RootN().SetName("rootForwardNullable2")
	err = l.Save(ctx)
	assert.NoError(t, err)
	l2, err = goradd_unit2.LoadLeafN(ctx, l.ID(), node2.LeafN().RootN())
	require.NoError(t, err)
	require.NotNilf(t, l2, "Object was nil based on ID %s", l.ID())
	assert.Equal(t, "leafForwardNullable2", l2.Name())
	assert.Equal(t, "rootForwardNullable2", l2.RootN().Name())

	// Insert-update
	l3 := goradd_unit2.NewLeafN()
	l3.SetName("leafForwardNullable3")
	r.SetName("rootForwardNullable3")
	l3.SetRootN(r)
	err = l3.Save(ctx)
	require.NoError(t, err)
	l2, err = goradd_unit2.LoadLeafN(ctx, l3.ID(), node2.LeafN().RootN())
	require.NoError(t, err)
	require.NotNilf(t, l2, "Object was nil based on ID %s", l3.ID())
	assert.Equal(t, "leafForwardNullable3", l2.Name())
	assert.Equal(t, "rootForwardNullable3", l2.RootN().Name())

	// Update-insert
	r4 := goradd_unit2.NewRootN()
	l.SetName("leafForwardNullable4")
	r4.SetName("rootForwardNullable4")
	l.SetRootN(r4)
	err = l.Save(ctx)
	require.NoError(t, err)
	l2, err = goradd_unit2.LoadLeafN(ctx, l.ID(), node2.LeafN().RootN())
	require.NoError(t, err)
	require.NotNilf(t, l2, "Object was nil based on ID %s", l.ID())
	assert.Equal(t, "leafForwardNullable4", l2.Name())
	assert.Equal(t, "rootForwardNullable4", l2.RootN().Name())
}

// TestForwardNullableCollision tests saving two records that are changed at the same time.
func TestForwardNullableCollision(t *testing.T) {
	ctx := context.Background()

	l := goradd_unit2.NewLeafN()
	r := goradd_unit2.NewRootN()
	r.SetName("rootForwardNullableCollision")
	l.SetName("leafForwardNullableCollision")
	l.SetRootN(r)
	err := l.Save(ctx)
	require.NoError(t, err)

	var l2 *goradd_unit2.LeafN
	l2, err = goradd_unit2.LoadLeafN(ctx, l.ID(), node2.LeafN().RootN())
	require.NoError(t, err)

	// Update first
	l.SetName("leafForwardNullableCollision2")
	l.RootN().SetName("rootForwardNullableCollision2")

	// Update second
	l2.SetName("leafForwardNullableCollision3")
	l2.RootN().SetName("rootForwardNullableCollision3")

	// save first then second
	err = l.Save(ctx)
	err2 := l2.Save(ctx)
	assert.NoError(t, err)
	assert.NoError(t, err2)

	// Last save should win
	l3, err3 := goradd_unit2.LoadLeafN(ctx, l.ID(), node2.LeafN().RootN())
	assert.NoError(t, err3)
	assert.Equal(t, "leafForwardNullableCollision3", l3.Name())
	assert.Equal(t, "rootForwardNullableCollision3", l3.RootN().Name())
}

func TestForwardNullableNull(t *testing.T) {
	ctx := context.Background()
	l := goradd_unit2.NewLeafN()
	l.SetName("leafForwardNullableNull")
	assert.NoError(t, l.Save(ctx))

	l.SetRootN(nil) // nullable
	assert.NoError(t, l.Save(ctx))
}

func TestForwardNullableTwo(t *testing.T) {
	ctx := context.Background()
	l := goradd_unit2.NewLeafN()
	r := goradd_unit2.NewRootN()
	l.SetName("leafForwardNullableTwo")
	r.SetName("rootForwardNullableTwo")
	l.SetRootN(r)
	require.NoError(t, l.Save(ctx))

	l2 := goradd_unit2.NewLeafN()
	l2.SetName("leaf2")
	l2.SetRootN(r)
	require.NoError(t, l2.Save(ctx))

	r2, err := goradd_unit2.LoadRootN(ctx, r.ID(), node2.RootN().LeafNs())
	require.NoError(t, err)
	assert.Len(t, r2.LeafNs(), 2)
}

func TestForwardNullableDelete(t *testing.T) {
	ctx := context.Background()
	l := goradd_unit2.NewLeafN()
	r := goradd_unit2.NewRootN()
	l.SetName("leafForwardNullableDelete")
	r.SetName("rootForwardNullableDelete")
	l.SetRootN(r)
	require.NoError(t, l.Save(ctx))

	require.NoError(t, l.Delete(ctx))

	l2, err := goradd_unit2.LoadLeafN(ctx, l.ID())
	require.NoError(t, err)
	assert.Nil(t, l2)

	r2, err := goradd_unit2.LoadRootN(ctx, r.ID())
	require.NoError(t, err)
	assert.NotNil(t, r2)
}
