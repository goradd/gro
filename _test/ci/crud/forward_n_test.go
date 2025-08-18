package crud

import (
	"context"
	"testing"

	"github.com/goradd/orm/_test/gen/orm/goradd_unit"
	"github.com/goradd/orm/_test/gen/orm/goradd_unit/node"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestForwardNullable tests insert and update of two linked records where the link is nullable.
func TestForwardNullable(t *testing.T) {
	ctx := context.Background()
	defer goradd_unit.ClearAll(ctx)

	// Insert-insert
	l := goradd_unit.NewLeafN()
	r := goradd_unit.NewRootN()
	l.SetName("leaf")
	r.SetName("root")
	l.SetRootN(r)
	err := l.Save(ctx)
	require.NoError(t, err)

	var l2 *goradd_unit.LeafN
	l2, err = goradd_unit.LoadLeafN(ctx, l.ID(), node.LeafN().RootN())
	require.NoError(t, err)
	require.NotNilf(t, l2, "Object was nil based on ID %s", l.ID())
	assert.Equal(t, "leaf", l2.Name())
	assert.Equal(t, "root", l2.RootN().Name())

	// Update-update
	l.SetName("leaf2")
	l.RootN().SetName("root2")
	err = l.Save(ctx)
	assert.NoError(t, err)
	l2, err = goradd_unit.LoadLeafN(ctx, l.ID(), node.LeafN().RootN())
	require.NoError(t, err)
	require.NotNilf(t, l2, "Object was nil based on ID %s", l.ID())
	assert.Equal(t, "leaf2", l2.Name())
	assert.Equal(t, "root2", l2.RootN().Name())

	// Insert-update
	l3 := goradd_unit.NewLeafN()
	l3.SetName("leaf3")
	r.SetName("root3")
	l3.SetRootN(r)
	err = l3.Save(ctx)
	require.NoError(t, err)
	l2, err = goradd_unit.LoadLeafN(ctx, l3.ID(), node.LeafN().RootN())
	require.NoError(t, err)
	require.NotNilf(t, l2, "Object was nil based on ID %s", l3.ID())
	assert.Equal(t, "leaf3", l2.Name())
	assert.Equal(t, "root3", l2.RootN().Name())

	// Update-insert
	r4 := goradd_unit.NewRootN()
	l.SetName("leaf4")
	r4.SetName("root4")
	l.SetRootN(r4)
	err = l.Save(ctx)
	require.NoError(t, err)
	l2, err = goradd_unit.LoadLeafN(ctx, l.ID(), node.LeafN().RootN())
	require.NoError(t, err)
	require.NotNilf(t, l2, "Object was nil based on ID %s", l.ID())
	assert.Equal(t, "leaf4", l2.Name())
	assert.Equal(t, "root4", l2.RootN().Name())
}

// TestForwardNullableCollision tests saving two records that are changed at the same time.
func TestForwardNullableCollision(t *testing.T) {
	ctx := context.Background()
	defer goradd_unit.ClearAll(ctx)

	l := goradd_unit.NewLeafN()
	r := goradd_unit.NewRootN()
	r.SetName("root")
	l.SetName("leaf")
	l.SetRootN(r)
	err := l.Save(ctx)
	require.NoError(t, err)

	var l2 *goradd_unit.LeafN
	l2, err = goradd_unit.LoadLeafN(ctx, l.ID(), node.LeafN().RootN())
	require.NoError(t, err)

	// Update first
	l.SetName("leaf2")
	l.RootN().SetName("root2")

	// Update second
	l2.SetName("leaf3")
	l2.RootN().SetName("root3")

	// save first then second
	err = l.Save(ctx)
	err2 := l2.Save(ctx)
	assert.NoError(t, err)
	assert.NoError(t, err2)

	// Last save should win
	l3, err3 := goradd_unit.LoadLeafN(ctx, l.ID(), node.LeafN().RootN())
	assert.NoError(t, err3)
	assert.Equal(t, "leaf3", l3.Name())
	assert.Equal(t, "root3", l3.RootN().Name())
}

func TestForwardNullableNull(t *testing.T) {
	ctx := context.Background()
	defer goradd_unit.ClearAll(ctx)
	l := goradd_unit.NewLeafN()
	l.SetName("leaf")
	assert.NoError(t, l.Save(ctx))

	l.SetRootN(nil) // nullable
	assert.NoError(t, l.Save(ctx))
}

func TestForwardNullableTwo(t *testing.T) {
	ctx := context.Background()
	defer goradd_unit.ClearAll(ctx)
	l := goradd_unit.NewLeafN()
	r := goradd_unit.NewRootN()
	l.SetName("leaf")
	r.SetName("root")
	l.SetRootN(r)
	require.NoError(t, l.Save(ctx))

	l2 := goradd_unit.NewLeafN()
	l2.SetName("leaf2")
	l2.SetRootN(r)
	require.NoError(t, l2.Save(ctx))

	r2, err := goradd_unit.LoadRootN(ctx, r.ID(), node.RootN().LeafNs())
	require.NoError(t, err)
	assert.Len(t, r2.LeafNs(), 2)
}

func TestForwardNullableDelete(t *testing.T) {
	ctx := context.Background()
	defer goradd_unit.ClearAll(ctx)
	l := goradd_unit.NewLeafN()
	r := goradd_unit.NewRootN()
	l.SetName("leaf")
	r.SetName("root")
	l.SetRootN(r)
	require.NoError(t, l.Save(ctx))

	require.NoError(t, l.Delete(ctx))

	l2, err := goradd_unit.LoadLeafN(ctx, l.ID())
	require.NoError(t, err)
	assert.Nil(t, l2)

	r2, err := goradd_unit.LoadRootN(ctx, r.ID())
	require.NoError(t, err)
	assert.NotNil(t, r2)
}
