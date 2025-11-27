package crud

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/goradd/gro/_test/gen/orm/goradd_unit"
	"github.com/goradd/gro/_test/gen/orm/goradd_unit/node"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestForward tests insert and update of two linked records.
func TestForward(t *testing.T) {
	ctx := context.Background()

	// Insert-insert
	l := goradd_unit.NewLeaf()
	r := goradd_unit.NewRoot()
	r.SetName("rootForward")
	l.SetName("leafForward")
	l.SetRoot(r)
	err := l.Save(ctx)
	require.NoError(t, err)

	var l2 *goradd_unit.Leaf
	l2, err = goradd_unit.LoadLeaf(ctx, l.ID(), node.Leaf().Root())
	require.NoError(t, err)
	require.NotNilf(t, l2, "Object was nil based on ID %s", l.ID())
	assert.Equal(t, "leafForward", l2.Name())
	assert.Equal(t, "rootForward", l2.Root().Name())

	// Update-update
	l.SetName("leafForward2")
	l.Root().SetName("rootForward2")
	err = l.Save(ctx)
	assert.NoError(t, err)
	l2, err = goradd_unit.LoadLeaf(ctx, l.ID(), node.Leaf().Root())
	require.NoError(t, err)
	require.NotNilf(t, l2, "Object was nil based on ID %s", l.ID())
	assert.Equal(t, "leafForward2", l2.Name())
	assert.Equal(t, "rootForward2", l2.Root().Name())

	// Insert-update
	l3 := goradd_unit.NewLeaf()
	l3.SetName("leafForward3")
	r.SetName("rootForward3")
	l3.SetRoot(r)
	err = l3.Save(ctx)
	require.NoError(t, err)
	l2, err = goradd_unit.LoadLeaf(ctx, l3.ID(), node.Leaf().Root())
	require.NoError(t, err)
	require.NotNilf(t, l2, "Object was nil based on ID %s", l3.ID())
	assert.Equal(t, "leafForward3", l2.Name())
	assert.Equal(t, "rootForward3", l2.Root().Name())

	// Update-insert
	r4 := goradd_unit.NewRoot()
	l.SetName("leafForward4")
	r4.SetName("rootForward4")
	l.SetRoot(r4)
	err = l.Save(ctx)
	require.NoError(t, err)
	l2, err = goradd_unit.LoadLeaf(ctx, l.ID(), node.Leaf().Root())
	assert.NoError(t, err)
	require.NotNilf(t, l2, "Object was nil based on ID %s", l.ID())
	assert.Equal(t, "leafForward4", l2.Name())
	assert.Equal(t, "rootForward4", l2.Root().Name())
}

// TestForwardCollision tests saving two records that are changed at the same time.
func TestForwardCollision(t *testing.T) {
	ctx := context.Background()

	l := goradd_unit.NewLeaf()
	r := goradd_unit.NewRoot()
	r.SetName("rootForwardCollision")
	l.SetName("leafForwardCollision")
	l.SetRoot(r)
	err := l.Save(ctx)
	require.NoError(t, err)

	var l2 *goradd_unit.Leaf
	l2, err = goradd_unit.LoadLeaf(ctx, l.ID(), node.Leaf().Root())
	require.NoError(t, err)

	// Update first
	l.SetName("leafForwardCollision2")
	l.Root().SetName("rootForwardCollision2")

	// Update second
	l2.SetName("leafForwardCollision3")
	l2.Root().SetName("rootForwardCollision3")

	// save first then second
	err = l.Save(ctx)
	err2 := l2.Save(ctx)
	assert.NoError(t, err)
	assert.NoError(t, err2)

	// Last save should win
	l3, err3 := goradd_unit.LoadLeaf(ctx, l.ID(), node.Leaf().Root())
	assert.NoError(t, err3)
	assert.Equal(t, "leafForwardCollision3", l3.Name())
	assert.Equal(t, "rootForwardCollision3", l3.Root().Name())
}

func TestForwardNull(t *testing.T) {
	ctx := context.Background()
	l := goradd_unit.NewLeaf()
	l.SetName("leafForwardNull")
	assert.Panics(t, func() {
		_ = l.Save(ctx) // root is required since RootID is not nullable
	})

	assert.Panics(t, func() {
		l.SetRoot(nil) // not nullable
	})
}

func TestForwardTwo(t *testing.T) {
	ctx := context.Background()
	l := goradd_unit.NewLeaf()
	r := goradd_unit.NewRoot()
	l.SetName("leafForwardTwo")
	r.SetName("rootForwardTwo")
	l.SetRoot(r)
	require.NoError(t, l.Save(ctx))

	l2 := goradd_unit.NewLeaf()
	l2.SetName("leafForwardTwo2")
	l2.SetRoot(r)
	require.NoError(t, l2.Save(ctx))

	r2, err := goradd_unit.LoadRoot(ctx, r.ID(), node.Root().Leafs())
	require.NoError(t, err)
	assert.Len(t, r2.Leafs(), 2)
}

func TestForwardDelete(t *testing.T) {
	ctx := context.Background()
	l := goradd_unit.NewLeaf()
	r := goradd_unit.NewRoot()
	l.SetName("leafForwardDelete")
	r.SetName("rootForwardDelete")
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

func TestForwardJson(t *testing.T) {
	ctx := context.Background()
	l := goradd_unit.NewLeaf()
	r := goradd_unit.NewRoot()
	l.SetName("leafForwardJson")
	r.SetName("rootForwardJson")
	l.SetRoot(r)
	require.NoError(t, l.Save(ctx))

	j, err := l.MarshalJSON()
	require.NoError(t, err)

	var m map[string]any

	err = json.Unmarshal(j, &m)
	require.NoError(t, err)
	v, ok := m["root"]
	assert.True(t, ok)
	assert.Equal(t, "rootForwardJson", v.(map[string]any)["name"].(string))

	var leaf goradd_unit.Leaf
	err = json.Unmarshal(j, &leaf)
	require.NoError(t, err)
	assert.Equal(t, "leafForwardJson", leaf.Name())
	assert.Equal(t, "rootForwardJson", leaf.Root().Name())
}
