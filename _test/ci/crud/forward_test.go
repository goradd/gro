package crud

import (
	"github.com/goradd/orm/_test/gen/orm/goradd_unit"
	"github.com/goradd/orm/_test/gen/orm/goradd_unit/node"
	"github.com/goradd/orm/pkg/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

// TestForward tests insert and update of two linked records.
func TestForward(t *testing.T) {
	ctx := db.NewContext(nil)
	defer goradd_unit.ClearAll(ctx)

	// Insert-insert
	l := goradd_unit.NewLeaf()
	r := goradd_unit.NewRoot()
	r.SetName("root")
	l.SetName("leaf")
	l.SetRoot(r)
	err := l.Save(ctx)
	require.NoError(t, err)

	var l2 *goradd_unit.Leaf
	l2, err = goradd_unit.LoadLeaf(ctx, l.ID(), node.Leaf().Root())
	require.NoError(t, err)
	assert.Equal(t, "leaf", l2.Name())
	assert.Equal(t, "root", l2.Root().Name())

	// Update-update
	l.SetName("leaf2")
	l.Root().SetName("root2")
	err = l.Save(ctx)
	assert.NoError(t, err)
	l2, err = goradd_unit.LoadLeaf(ctx, l.ID(), node.Leaf().Root())
	require.NoError(t, err)
	assert.Equal(t, "leaf2", l2.Name())
	assert.Equal(t, "root2", l2.Root().Name())

	// Insert-update
	l3 := goradd_unit.NewLeaf()
	l3.SetName("leaf3")
	r.SetName("root3")
	l3.SetRoot(r)
	err = l3.Save(ctx)
	require.NoError(t, err)
	l2, err = goradd_unit.LoadLeaf(ctx, l3.ID(), node.Leaf().Root())
	require.NoError(t, err)
	assert.Equal(t, "leaf3", l2.Name())
	assert.Equal(t, "root3", l2.Root().Name())

	// Update-insert
	r4 := goradd_unit.NewRoot()
	l.SetName("leaf4")
	r4.SetName("root4")
	l.SetRoot(r4)
	err = l.Save(ctx)
	require.NoError(t, err)
	l2, err = goradd_unit.LoadLeaf(ctx, l.ID(), node.Leaf().Root())
	require.NoError(t, err)
	assert.Equal(t, "leaf4", l2.Name())
	assert.Equal(t, "root4", l2.Root().Name())
}

// TestForwardCollision tests saving two records that are changed at the same time.
func TestForwardCollision(t *testing.T) {
	ctx := db.NewContext(nil)
	defer goradd_unit.ClearAll(ctx)

	l := goradd_unit.NewLeaf()
	r := goradd_unit.NewRoot()
	r.SetName("root")
	l.SetName("leaf")
	l.SetRoot(r)
	err := l.Save(ctx)
	require.NoError(t, err)

	var l2 *goradd_unit.Leaf
	l2, err = goradd_unit.LoadLeaf(ctx, l.ID(), node.Leaf().Root())
	require.NoError(t, err)

	// Update first
	l.SetName("leaf2")
	l.Root().SetName("root2")

	// Update second
	l2.SetName("leaf3")
	l2.Root().SetName("root3")

	// save first then second
	err = l.Save(ctx)
	err2 := l2.Save(ctx)
	assert.NoError(t, err)
	assert.NoError(t, err2)

	// Last save should win
	l3, err3 := goradd_unit.LoadLeaf(ctx, l.ID(), node.Leaf().Root())
	assert.NoError(t, err3)
	assert.Equal(t, "leaf3", l3.Name())
	assert.Equal(t, "root3", l3.Root().Name())
}

func TestForwardNull(t *testing.T) {
	ctx := db.NewContext(nil)
	defer goradd_unit.ClearAll(ctx)
	l := goradd_unit.NewLeaf()
	l.SetName("leaf")
	assert.Panics(t, func() {
		_ = l.Save(ctx) // root is required since RootID is not nullable
	})

	assert.Panics(t, func() {
		l.SetRoot(nil) // not nullable
	})
}

func TestForwardTwo(t *testing.T) {
	ctx := db.NewContext(nil)
	defer goradd_unit.ClearAll(ctx)
	l := goradd_unit.NewLeaf()
	r := goradd_unit.NewRoot()
	l.SetName("leaf")
	r.SetName("root")
	l.SetRoot(r)
	require.NoError(t, l.Save(ctx))

	l2 := goradd_unit.NewLeaf()
	l2.SetName("leaf2")
	l2.SetRoot(r)
	require.NoError(t, l2.Save(ctx))

	r2, err := goradd_unit.LoadRoot(ctx, r.ID(), node.Root().Leafs())
	require.NoError(t, err)
	assert.Len(t, r2.Leafs(), 2)
}

func TestForwardDelete(t *testing.T) {
	ctx := db.NewContext(nil)
	defer goradd_unit.ClearAll(ctx)
	l := goradd_unit.NewLeaf()
	r := goradd_unit.NewRoot()
	l.SetName("leaf")
	r.SetName("root")
	l.SetRoot(r)
	require.NoError(t, l.Save(ctx))

	require.NoError(t, l.Delete(ctx))

	l2, err := goradd_unit.LoadLeaf(ctx, l.ID())
	require.NoError(t, err)
	assert.Nil(t, l2)

	r2, err := goradd_unit.LoadRoot(ctx, r.ID())
	require.NoError(t, err)
	assert.NotNil(t, r2)
}
