package crud

import (
	"context"
	"testing"

	goradd_unit2 "github.com/goradd/gro/ci/tests/gen/goradd_unit"
	"github.com/goradd/gro/ci/tests/gen/goradd_unit/node"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestReverseUniqueNullable tests insert and update of two linked records.
func TestReverseUniqueNullable(t *testing.T) {
	ctx := context.Background()

	// Insert-insert
	r := goradd_unit2.NewRootUn()
	l := goradd_unit2.NewLeafUn()
	r.SetName("rootReverseUniqueNullable")
	l.SetName("leafReverseUniqueNullable")
	r.SetLeafUn(l)
	err := r.Save(ctx)
	require.NoError(t, err)

	var r2 *goradd_unit2.RootUn
	r2, err = goradd_unit2.LoadRootUn(ctx, r.ID(), node.RootUn().LeafUn())
	require.NoError(t, err)
	require.NotNilf(t, r2, "Object was nil based on ID %s", r.ID())
	assert.Equal(t, "rootReverseUniqueNullable", r2.Name())
	assert.Equal(t, "leafReverseUniqueNullable", r2.LeafUn().Name())

	// Update-update
	r.SetName("rootReverseUniqueNullable2")
	r.LeafUn().SetName("leafReverseUniqueNullable2")
	err = r.Save(ctx)
	assert.NoError(t, err)
	r2, err = goradd_unit2.LoadRootUn(ctx, r.ID(), node.RootUn().LeafUn())
	require.NoError(t, err)
	require.NotNilf(t, r2, "Object was nil based on ID %s", r.ID())
	assert.Equal(t, "rootReverseUniqueNullable2", r2.Name())
	assert.Equal(t, "leafReverseUniqueNullable2", r2.LeafUn().Name())

	// Insert-update
	r3 := goradd_unit2.NewRootUn()
	r3.SetName("rootReverseUniqueNullable3")
	l.SetName("leafReverseUniqueNullable3")
	r3.SetLeafUn(l)
	err = r3.Save(ctx)
	require.NoError(t, err)
	r2, err = goradd_unit2.LoadRootUn(ctx, r3.ID(), node.RootUn().LeafUn())
	require.NoError(t, err)
	require.NotNilf(t, r2, "Object was nil based on ID %s", r3.ID())
	assert.Equal(t, "rootReverseUniqueNullable3", r2.Name())
	assert.Equal(t, "leafReverseUniqueNullable3", r2.LeafUn().Name())

	// Update-insert
	l4 := goradd_unit2.NewLeafUn()
	r.SetName("rootReverseUniqueNullable4")
	l4.SetName("leafReverseUniqueNullable4")
	r.SetLeafUn(l4)
	err = r.Save(ctx)
	require.NoError(t, err)
	r2, err = goradd_unit2.LoadRootUn(ctx, r.ID(), node.RootUn().LeafUn())
	require.NoError(t, err)
	require.NotNilf(t, r2, "Object was nil based on ID %s", r.ID())
	assert.Equal(t, "rootReverseUniqueNullable4", r2.Name())
	assert.Equal(t, "leafReverseUniqueNullable4", r2.LeafUn().Name())

}

// TestReverseCollision tests saving two records that are changed at the same time.
func TestReverseUniqueNullableCollision(t *testing.T) {
	ctx := context.Background()

	r := goradd_unit2.NewRootUn()
	l := goradd_unit2.NewLeafUn()
	r.SetName("rootReverseUniqueNullableCollision")
	l.SetName("leafReverseUniqueNullableCollision")
	r.SetLeafUn(l)
	err := r.Save(ctx)
	require.NoError(t, err)

	var r2 *goradd_unit2.RootUn
	r2, err = goradd_unit2.LoadRootUn(ctx, r.ID(), node.RootUn().LeafUn())
	require.NoError(t, err)

	// Update first
	r.SetName("rootReverseUniqueNullableCollision2")
	r.LeafUn().SetName("leafReverseUniqueNullableCollision2")

	// Update second
	r2.SetName("rootReverseUniqueNullableCollision3")
	r2.LeafUn().SetName("leafReverseUniqueNullableCollision3")

	// save first then second
	err = r.Save(ctx)
	err2 := r2.Save(ctx)
	assert.NoError(t, err)
	assert.NoError(t, err2)

	// Last save should win
	r3, err3 := goradd_unit2.LoadRootUn(ctx, r.ID(), node.RootUn().LeafUn())
	assert.NoError(t, err3)
	assert.Equal(t, "rootReverseUniqueNullableCollision3", r3.Name())
	assert.Equal(t, "leafReverseUniqueNullableCollision3", r3.LeafUn().Name())
}

func TestReverseUniqueNullableNull(t *testing.T) {
	ctx := context.Background()

	r := goradd_unit2.NewRootUn()
	r.SetName("rootReverseUniqueNullableNull")
	l := goradd_unit2.NewLeafUn()
	l.SetName("leafReverseUniqueNullableNull")
	r.SetLeafUn(l)
	require.NoError(t, r.Save(ctx))

	r.SetLeafUn(nil)
	require.NoError(t, r.Save(ctx))

	r2, err := goradd_unit2.LoadRootUn(ctx, r.ID(), node.RootUn().LeafUn())
	require.NoError(t, err)
	assert.Nil(t, r2.LeafUn())

	l2, err := goradd_unit2.LoadLeafUn(ctx, l.ID())
	require.NoError(t, err)
	assert.NotNil(t, l2) // reverse linked item that could have a nil pointer was not deleted
	assert.Nil(t, l2.RootUn())
}

func TestReverseUniqueNullableTwo(t *testing.T) {
	ctx := context.Background()
	r := goradd_unit2.NewRootUn()
	l := goradd_unit2.NewLeafUn()
	r.SetName("rootReverseUniqueNullableTwo")
	l.SetName("leafReverseUniqueNullableTwo")
	r.SetLeafUn(l)
	require.NoError(t, r.Save(ctx))

	l2 := goradd_unit2.NewLeafUn()
	l2.SetName("leafReverseUniqueNullableTwo2")
	r.SetLeafUn(l2)
	require.NoError(t, r.Save(ctx)) // unique failure

	// Confirm detached.
	l3, _ := goradd_unit2.LoadLeafUn(ctx, l.ID())
	assert.Nil(t, l3.RootUn())
}
