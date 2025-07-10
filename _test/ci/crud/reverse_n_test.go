package crud

import (
	"github.com/goradd/orm/_test/gen/orm/goradd_unit"
	"github.com/goradd/orm/_test/gen/orm/goradd_unit/node"
	"github.com/goradd/orm/pkg/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

// TestReverse tests insert and update of two linked records.
func TestReverseNullable(t *testing.T) {
	ctx := context.Background()
	defer goradd_unit.ClearAll(ctx)

	// Insert-insert
	r := goradd_unit.NewRootN()
	l := goradd_unit.NewLeafN()
	r.SetName("root")
	l.SetName("leaf")
	r.SetLeafNs(l)
	err := r.Save(ctx)
	require.NoError(t, err)

	var r2 *goradd_unit.RootN
	r2, err = goradd_unit.LoadRootN(ctx, r.ID(), node.RootN().LeafNs())
	require.NoError(t, err)
	assert.Equal(t, "root", r2.Name())
	assert.Equal(t, "leaf", r2.LeafNs()[0].Name())

	// Update-update
	r.SetName("root2")
	r.LeafNs()[0].SetName("leaf2")
	err = r.Save(ctx)
	assert.NoError(t, err)
	r2, err = goradd_unit.LoadRootN(ctx, r.ID(), node.RootN().LeafNs())
	require.NoError(t, err)
	assert.Equal(t, "root2", r2.Name())
	assert.Equal(t, "leaf2", r2.LeafNs()[0].Name())

	// Insert-update
	r3 := goradd_unit.NewRootN()
	r3.SetName("root3")
	l.SetName("leaf3")
	r3.SetLeafNs(l)
	err = r3.Save(ctx)
	require.NoError(t, err)
	r2, err = goradd_unit.LoadRootN(ctx, r3.ID(), node.RootN().LeafNs())
	require.NoError(t, err)
	assert.Equal(t, "root3", r2.Name())
	assert.Equal(t, "leaf3", r2.LeafNs()[0].Name())

	// Update-insert
	l4 := goradd_unit.NewLeafN()
	r.SetName("root4")
	l4.SetName("leaf4")
	r.SetLeafNs(l4)
	err = r.Save(ctx)
	require.NoError(t, err)
	r2, err = goradd_unit.LoadRootN(ctx, r.ID(), node.RootN().LeafNs())
	require.NoError(t, err)
	assert.Equal(t, "root4", r2.Name())
	assert.Equal(t, "leaf4", r2.LeafNs()[0].Name())

}

// TestReverseCollision tests saving two records that are changed at the same time.
func TestReverseNullableCollision(t *testing.T) {
	ctx := context.Background()
	defer goradd_unit.ClearAll(ctx)

	r := goradd_unit.NewRootN()
	l := goradd_unit.NewLeafN()
	r.SetName("root")
	l.SetName("leaf")
	r.SetLeafNs(l)
	err := r.Save(ctx)
	require.NoError(t, err)

	var r2 *goradd_unit.RootN
	r2, err = goradd_unit.LoadRootN(ctx, r.ID(), node.RootN().LeafNs())
	require.NoError(t, err)

	// Update first
	r.SetName("root2")
	r.LeafNs()[0].SetName("leaf2")

	// Update second
	r2.SetName("root3")
	r2.LeafNs()[0].SetName("leaf3")

	// save first then second
	err = r.Save(ctx)
	err2 := r2.Save(ctx)
	assert.NoError(t, err)
	assert.NoError(t, err2)

	// Last save should win
	r3, err3 := goradd_unit.LoadRootN(ctx, r.ID(), node.RootN().LeafNs())
	assert.NoError(t, err3)
	assert.Equal(t, "root3", r3.Name())
	assert.Equal(t, "leaf3", r3.LeafNs()[0].Name())
}

func TestReverseNullableNull(t *testing.T) {
	ctx := context.Background()
	defer goradd_unit.ClearAll(ctx)

	r := goradd_unit.NewRootN()
	r.SetName("root")
	l := goradd_unit.NewLeafN()
	l.SetName("leaf")
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
	defer goradd_unit.ClearAll(ctx)
	r := goradd_unit.NewRootN()
	l := goradd_unit.NewLeafN()
	r.SetName("root")
	l.SetName("leaf")
	r.SetLeafNs(l)
	require.NoError(t, r.Save(ctx))

	l2 := goradd_unit.NewLeafN()
	l2.SetName("leaf2")
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
