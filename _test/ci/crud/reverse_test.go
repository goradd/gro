package crud

import (
	"encoding/json"
	"github.com/goradd/orm/_test/gen/orm/goradd_unit"
	"github.com/goradd/orm/_test/gen/orm/goradd_unit/node"
	"github.com/goradd/orm/pkg/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

// TestReverse tests insert and update of two linked records.
func TestReverse(t *testing.T) {
	ctx := db.NewContext(nil)
	defer goradd_unit.ClearAll(ctx)

	// Insert-insert
	r := goradd_unit.NewRoot()
	l := goradd_unit.NewLeaf()
	r.SetName("root")
	l.SetName("leaf")
	r.SetLeafs(l)
	err := r.Save(ctx)
	require.NoError(t, err)

	var r2 *goradd_unit.Root
	r2, err = goradd_unit.LoadRoot(ctx, r.ID(), node.Root().Leafs())
	require.NoError(t, err)
	assert.Equal(t, "root", r2.Name())
	assert.Equal(t, "leaf", r2.Leafs()[0].Name())

	// Update-update
	r.SetName("root2")
	r.Leafs()[0].SetName("leaf2")
	err = r.Save(ctx)
	assert.NoError(t, err)
	r2, err = goradd_unit.LoadRoot(ctx, r.ID(), node.Root().Leafs())
	require.NoError(t, err)
	assert.Equal(t, "root2", r2.Name())
	assert.Equal(t, "leaf2", r2.Leafs()[0].Name())

	// Insert-update
	r3 := goradd_unit.NewRoot()
	r3.SetName("root3")
	l.SetName("leaf3")
	r3.SetLeafs(l)
	err = r3.Save(ctx)
	require.NoError(t, err)
	r2, err = goradd_unit.LoadRoot(ctx, r3.ID(), node.Root().Leafs())
	require.NoError(t, err)
	assert.Equal(t, "root3", r2.Name())
	assert.Equal(t, "leaf3", r2.Leafs()[0].Name())

	// Update-insert
	l4 := goradd_unit.NewLeaf()
	r.SetName("root4")
	l4.SetName("leaf4")
	r.SetLeafs(l4)
	err = r.Save(ctx)
	require.NoError(t, err)
	r2, err = goradd_unit.LoadRoot(ctx, r.ID(), node.Root().Leafs())
	require.NoError(t, err)
	assert.Equal(t, "root4", r2.Name())
	assert.Equal(t, "leaf4", r2.Leafs()[0].Name())

}

// TestReverseCollision tests saving two records that are changed at the same time.
func TestReverseCollision(t *testing.T) {
	ctx := db.NewContext(nil)
	defer goradd_unit.ClearAll(ctx)

	r := goradd_unit.NewRoot()
	l := goradd_unit.NewLeaf()
	r.SetName("root")
	l.SetName("leaf")
	r.SetLeafs(l)
	err := r.Save(ctx)
	require.NoError(t, err)

	var r2 *goradd_unit.Root
	r2, err = goradd_unit.LoadRoot(ctx, r.ID(), node.Root().Leafs())
	require.NoError(t, err)

	// Update first
	r.SetName("root2")
	r.Leafs()[0].SetName("leaf2")

	// Update second
	r2.SetName("root3")
	r2.Leafs()[0].SetName("leaf3")

	// save first then second
	err = r.Save(ctx)
	err2 := r2.Save(ctx)
	assert.NoError(t, err)
	assert.NoError(t, err2)

	// Last save should win
	r3, err3 := goradd_unit.LoadRoot(ctx, r.ID(), node.Root().Leafs())
	assert.NoError(t, err3)
	assert.Equal(t, "root3", r3.Name())
	assert.Equal(t, "leaf3", r3.Leafs()[0].Name())
}

func TestReverseNull(t *testing.T) {
	ctx := db.NewContext(nil)
	defer goradd_unit.ClearAll(ctx)

	r := goradd_unit.NewRoot()
	r.SetName("root")
	l := goradd_unit.NewLeaf()
	l.SetName("leaf")
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
	ctx := db.NewContext(nil)
	defer goradd_unit.ClearAll(ctx)
	r := goradd_unit.NewRoot()
	l := goradd_unit.NewLeaf()
	r.SetName("root")
	l.SetName("leaf")
	r.SetLeafs(l)
	require.NoError(t, r.Save(ctx))

	l2 := goradd_unit.NewLeaf()
	l2.SetName("leaf2")
	r.SetLeafs(l, l2)
	require.NoError(t, r.Save(ctx))

	r2, err := goradd_unit.LoadRoot(ctx, r.ID(), node.Root().Leafs())
	require.NoError(t, err)
	assert.Len(t, r2.Leafs(), 2)
}

func TestReverseJson(t *testing.T) {
	ctx := db.NewContext(nil)
	defer goradd_unit.ClearAll(ctx)
	r := goradd_unit.NewRoot()
	l := goradd_unit.NewLeaf()
	r.SetName("root")
	l.SetName("leaf")
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
	assert.Equal(t, "leaf", v2[0].(map[string]any)["name"].(string))

	var root goradd_unit.Root
	err = json.Unmarshal(j, &root)
	require.NoError(t, err)
	assert.Equal(t, "root", root.Name())
	assert.Equal(t, "leaf", root.Leafs()[0].Name())

}
