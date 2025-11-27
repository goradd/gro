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

// TestReverse tests insert and update of two linked records.
func TestReverse(t *testing.T) {
	ctx := context.Background()

	// Insert-insert
	r := goradd_unit.NewRoot()
	l := goradd_unit.NewLeaf()
	r.SetName("rootReverse")
	l.SetName("leafReverse")
	r.SetLeafs(l)
	err := r.Save(ctx)
	require.NoError(t, err)

	var r2 *goradd_unit.Root
	r2, err = goradd_unit.LoadRoot(ctx, r.ID(), node.Root().Leafs())
	require.NoError(t, err)
	require.NotNilf(t, r2, "Object was nil based on ID %s", r.ID())
	assert.Equal(t, "rootReverse", r2.Name())
	assert.Equal(t, "leafReverse", r2.Leafs()[0].Name())

	// Update-update
	r.SetName("rootReverse2")
	r.Leafs()[0].SetName("leafReverse2")
	err = r.Save(ctx)
	assert.NoError(t, err)
	r2, err = goradd_unit.LoadRoot(ctx, r.ID(), node.Root().Leafs())
	require.NoError(t, err)
	require.NotNilf(t, r2, "Object was nil based on ID %s", r.ID())
	assert.Equal(t, "rootReverse2", r2.Name())
	assert.Equal(t, "leafReverse2", r2.Leafs()[0].Name())

	// Insert-update
	r3 := goradd_unit.NewRoot()
	r3.SetName("rootReverse3")
	l.SetName("leafReverse3")
	r3.SetLeafs(l)
	err = r3.Save(ctx)
	require.NoError(t, err)
	r2, err = goradd_unit.LoadRoot(ctx, r3.ID(), node.Root().Leafs())
	require.NoError(t, err)
	require.NotNilf(t, r2, "Object was nil based on ID %s", r3.ID())
	assert.Equal(t, "rootReverse3", r2.Name())
	assert.Equal(t, "leafReverse3", r2.Leafs()[0].Name())

	// Update-insert
	l4 := goradd_unit.NewLeaf()
	r.SetName("rootReverse4")
	l4.SetName("leafReverse4")
	r.SetLeafs(l4)
	err = r.Save(ctx)
	require.NoError(t, err)
	r2, err = goradd_unit.LoadRoot(ctx, r.ID(), node.Root().Leafs())
	require.NoError(t, err)
	require.NotNilf(t, r2, "Object was nil based on ID %s", r.ID())
	assert.Equal(t, "rootReverse4", r2.Name())
	assert.Equal(t, "leafReverse4", r2.Leafs()[0].Name())

}

// TestReverseCollision tests saving two records that are changed at the same time.
func TestReverseCollision(t *testing.T) {
	ctx := context.Background()

	r := goradd_unit.NewRoot()
	l := goradd_unit.NewLeaf()
	r.SetName("rootReverseCollision")
	l.SetName("leafReverseCollision")
	r.SetLeafs(l)
	err := r.Save(ctx)
	require.NoError(t, err)

	var r2 *goradd_unit.Root
	r2, err = goradd_unit.LoadRoot(ctx, r.ID(), node.Root().Leafs())
	require.NoError(t, err)

	// Update first
	r.SetName("rootReverseCollision2")
	r.Leafs()[0].SetName("leafReverseCollision2")

	// Update second
	r2.SetName("rootReverseCollision3")
	r2.Leafs()[0].SetName("leafReverseCollision3")

	// save first then second
	err = r.Save(ctx)
	err2 := r2.Save(ctx)
	assert.NoError(t, err)
	assert.NoError(t, err2)

	// Last save should win
	r3, err3 := goradd_unit.LoadRoot(ctx, r.ID(), node.Root().Leafs())
	assert.NoError(t, err3)
	assert.Equal(t, "rootReverseCollision3", r3.Name())
	assert.Equal(t, "leafReverseCollision3", r3.Leafs()[0].Name())
}

func TestReverseNull(t *testing.T) {
	ctx := context.Background()

	r := goradd_unit.NewRoot()
	r.SetName("rootReverseCollision")
	l := goradd_unit.NewLeaf()
	l.SetName("leafReverseCollision")
	r.SetLeafs(l)
	require.NoError(t, r.Save(ctx))

	r.SetLeafs()
	require.NoError(t, r.Save(ctx))

	r2, err := goradd_unit.LoadRoot(ctx, r.ID(), node.Root().Leafs())
	require.NoError(t, err)
	assert.Len(t, r2.Leafs(), 0)

	l2, err := goradd_unit.LoadLeaf(ctx, l.ID())
	require.NoError(t, err)
	assert.Nil(t, l2) // reverse linked item that could not have a nil pointer was deleted
}

func TestReverseTwo(t *testing.T) {
	ctx := context.Background()
	r := goradd_unit.NewRoot()
	l := goradd_unit.NewLeaf()
	r.SetName("rootReverseTwo")
	l.SetName("leafReverseTwo")
	r.SetLeafs(l)
	require.NoError(t, r.Save(ctx))

	l2 := goradd_unit.NewLeaf()
	l2.SetName("leafReverseTwo2")
	r.SetLeafs(l, l2)
	require.NoError(t, r.Save(ctx))

	r2, err := goradd_unit.LoadRoot(ctx, r.ID(), node.Root().Leafs())
	require.NoError(t, err)
	assert.Len(t, r2.Leafs(), 2)
}

func TestReverseJson(t *testing.T) {
	ctx := context.Background()
	r := goradd_unit.NewRoot()
	l := goradd_unit.NewLeaf()
	r.SetName("rootReverseJson")
	l.SetName("leafReverseJson")
	r.SetLeafs(l)
	require.NoError(t, r.Save(ctx))

	j, err := r.MarshalJSON()
	require.NoError(t, err)

	var m map[string]any

	err = json.Unmarshal(j, &m)
	require.NoError(t, err)
	v, ok := m["leafs"]
	assert.True(t, ok)
	v2 := v.([]any)
	assert.Equal(t, "leafReverseJson", v2[0].(map[string]any)["name"].(string))

	var root goradd_unit.Root
	err = json.Unmarshal(j, &root)
	require.NoError(t, err)
	assert.Equal(t, "rootReverseJson", root.Name())
	assert.Equal(t, "leafReverseJson", root.Leafs()[0].Name())

}
