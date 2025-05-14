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

// TestForwardLock tests insert and update of two linked records that have an optimistic lock.
func TestAssociationLock(t *testing.T) {
	ctx := db.NewContext(nil)
	defer goradd_unit.ClearAll(ctx)

	// Insert-insert
	l1 := goradd_unit.NewLeafNl()
	l2 := goradd_unit.NewLeafNl()
	l1.SetName("leaf1")
	l2.SetName("leaf2")
	l1.SetLeaf2s(l2)
	err := l1.Save(ctx)
	require.NoError(t, err)

	var l3 *goradd_unit.LeafNl
	l3, err = goradd_unit.LoadLeafNl(ctx, l1.ID(), node.LeafNl().Leaf2s())
	require.NoError(t, err)
	assert.Equal(t, "leaf1", l3.Name())
	assert.Equal(t, "leaf2", l3.Leaf2s()[0].Name())

	// Update-update
	l3.SetName("leaf11")
	l3.Leaf2s()[0].SetName("leaf22")
	err = l3.Save(ctx)
	assert.NoError(t, err)
	l3, err = goradd_unit.LoadLeafNl(ctx, l1.ID(), node.LeafNl().Leaf2s())
	require.NoError(t, err)
	assert.Equal(t, "leaf11", l3.Name())
	assert.Equal(t, "leaf22", l3.Leaf2s()[0].Name())

	// Insert-update
	l4 := goradd_unit.NewLeafNl()
	l4.SetName("leaf4")
	l3.SetName("leaf111")
	l4.SetLeaf2s(l3)
	err = l4.Save(ctx)
	require.NoError(t, err)
	l3, err = goradd_unit.LoadLeafNl(ctx, l4.ID(), node.LeafNl().Leaf2s())
	require.NoError(t, err)
	assert.Equal(t, "leaf4", l3.Name())
	assert.Equal(t, "leaf111", l3.Leaf2s()[0].Name())

	// Update-insert
	l5 := goradd_unit.NewLeafNl()
	l4.SetName("leaf44")
	l5.SetName("leaf5")
	l4.SetLeaf2s(l5)
	err = l4.Save(ctx)
	require.NoError(t, err)
	l3, err = goradd_unit.LoadLeafNl(ctx, l4.ID(), node.LeafNl().Leaf2s())
	require.NoError(t, err)
	assert.Equal(t, "leaf44", l3.Name())
	assert.Equal(t, "leaf5", l3.Leaf2s()[0].Name())
}

func TestAssociationLockCollision(t *testing.T) {
	ctx := db.NewContext(nil)
	defer goradd_unit.ClearAll(ctx)

	l1 := goradd_unit.NewLeafNl()
	l2 := goradd_unit.NewLeafNl()
	l1.SetName("leaf1")
	l2.SetName("leaf2")
	l1.SetLeaf2s(l2)
	err := l1.Save(ctx)
	require.NoError(t, err)

	var l3 *goradd_unit.LeafNl
	l3, err = goradd_unit.LoadLeafNl(ctx, l1.ID(), node.LeafNl().Leaf2s())
	require.NoError(t, err)

	// Update both
	l1.SetName("leaf11")
	l3.SetName("leaf12")

	err = l1.Save(ctx)
	err2 := l3.Save(ctx)
	assert.NoError(t, err)
	assert.Error(t, err2)
	assert.IsType(t, &db.OptimisticLockError{}, err2)

	// 2nd level
	l3, _ = goradd_unit.LoadLeafNl(ctx, l1.ID(), node.LeafNl().Leaf2s())
	l1.Leaf2s()[0].SetName("leaf21")
	l3.Leaf2s()[0].SetName("leaf22")

	err = l1.Save(ctx)
	err2 = l3.Save(ctx)
	assert.NoError(t, err)
	assert.Error(t, err2)
	assert.IsType(t, &db.OptimisticLockError{}, err2)
}

func TestAssociationLockNull(t *testing.T) {
	ctx := db.NewContext(nil)
	defer goradd_unit.ClearAll(ctx)

	l1 := goradd_unit.NewLeafNl()
	l2 := goradd_unit.NewLeafNl()
	l1.SetName("leaf1")
	l2.SetName("leaf2")
	l1.SetLeaf2s(l2)
	err := l1.Save(ctx)
	require.NoError(t, err)

	l3, _ := goradd_unit.LoadLeafNl(ctx, l1.ID(), node.LeafNl().Leaf2s())
	assert.Len(t, l3.Leaf2s(), 1)

	l1.SetLeaf2s()
	err = l1.Save(ctx)
	require.NoError(t, err)

	l3, _ = goradd_unit.LoadLeafNl(ctx, l1.ID(), node.LeafNl().Leaf2s())
	assert.Len(t, l3.Leaf2s(), 0)
}

func TestAssociationLockTwo(t *testing.T) {
	ctx := db.NewContext(nil)
	defer goradd_unit.ClearAll(ctx)

	l1 := goradd_unit.NewLeafNl()
	l2 := goradd_unit.NewLeafNl()
	l3 := goradd_unit.NewLeafNl()

	l1.SetName("leaf1")
	l2.SetName("leaf2")
	l3.SetName("leaf3")
	l1.SetLeaf2s(l2, l3)
	err := l1.Save(ctx)
	require.NoError(t, err)

	l4, err := goradd_unit.LoadLeafNl(ctx, l1.ID(), node.LeafNl().Leaf2s())
	require.NoError(t, err)
	assert.Len(t, l4.Leaf2s(), 2)
}

func TestAssociationLockDelete(t *testing.T) {
	ctx := db.NewContext(nil)
	defer goradd_unit.ClearAll(ctx)

	l1 := goradd_unit.NewLeafNl()
	l2 := goradd_unit.NewLeafNl()
	l3 := goradd_unit.NewLeafNl()

	l1.SetName("leaf1")
	l2.SetName("leaf2")
	l3.SetName("leaf3")
	l1.SetLeaf2s(l2, l3)
	err := l1.Save(ctx)
	require.NoError(t, err)

	err = l2.Delete(ctx)
	assert.NoError(t, err)

	l4, err := goradd_unit.LoadLeafNl(ctx, l1.ID(), node.LeafNl().Leaf2s())
	assert.Len(t, l4.Leaf2s(), 1)
}

func TestAssociationLockJson(t *testing.T) {
	ctx := db.NewContext(nil)
	defer goradd_unit.ClearAll(ctx)

	l1 := goradd_unit.NewLeafNl()
	l2 := goradd_unit.NewLeafNl()
	l3 := goradd_unit.NewLeafNl()

	l1.SetName("leaf1")
	l2.SetName("leaf2")
	l3.SetName("leaf3")
	l1.SetLeaf2s(l2, l3)
	require.NoError(t, l1.Save(ctx))

	j, err := l1.MarshalJSON()
	require.NoError(t, err)

	var m map[string]any

	err = json.Unmarshal(j, &m)
	require.NoError(t, err)
	v, ok := m["leaf2s"]
	assert.True(t, ok)
	v2 := v.([]any)
	assert.Equal(t, "leaf2", v2[0].(map[string]any)["name"].(string))

	var leaf goradd_unit.LeafNl
	err = json.Unmarshal(j, &leaf)
	require.NoError(t, err)
	assert.Equal(t, "leaf1", leaf.Name())
	assert.Equal(t, "leaf2", leaf.Leaf2s()[0].Name())
}
