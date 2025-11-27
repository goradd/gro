package crud

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/goradd/gro/_test/gen/orm/goradd_unit"
	"github.com/goradd/gro/_test/gen/orm/goradd_unit/node"
	"github.com/goradd/gro/pkg/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestForwardLock tests insert and update of two linked records that have an optimistic lock.
func TestForwardLock(t *testing.T) {
	ctx := context.Background()

	// Insert-insert
	l := goradd_unit.NewLeafL()
	r := goradd_unit.NewRootL()
	r.SetName("rootForwardLock")
	l.SetName("leafForwardLock")
	l.SetRootL(r)
	err := l.Save(ctx)
	require.NoError(t, err)

	var l2 *goradd_unit.LeafL
	l2, err = goradd_unit.LoadLeafL(ctx, l.ID(), node.LeafL().RootL())
	require.NoError(t, err)
	require.NotNilf(t, l2, "Object was nil based on ID %s", l.ID())
	assert.Equal(t, "leafForwardLock", l2.Name())
	assert.Equal(t, "rootForwardLock", l2.RootL().Name())

	// Update-update
	l.SetName("leafForwardLock2")
	l.RootL().SetName("rootForwardLock2")
	err = l.Save(ctx)
	assert.NoError(t, err)
	l2, err = goradd_unit.LoadLeafL(ctx, l.ID(), node.LeafL().RootL())
	require.NoError(t, err)
	require.NotNilf(t, l2, "Object was nil based on ID %s", l.ID())
	assert.Equal(t, "leafForwardLock2", l2.Name())
	assert.Equal(t, "rootForwardLock2", l2.RootL().Name())

	// Insert-update
	l3 := goradd_unit.NewLeafL()
	l3.SetName("leafForwardLock3")
	r.SetName("rootForwardLock3")
	l3.SetRootL(r)
	err = l3.Save(ctx)
	require.NoError(t, err)
	l2, err = goradd_unit.LoadLeafL(ctx, l3.ID(), node.LeafL().RootL())
	require.NoError(t, err)
	require.NotNilf(t, l2, "Object was nil based on ID %s", l3.ID())
	assert.Equal(t, "leafForwardLock3", l2.Name())
	assert.Equal(t, "rootForwardLock3", l2.RootL().Name())

	// Update-insert
	r4 := goradd_unit.NewRootL()
	l.SetName("leafForwardLock4")
	r4.SetName("rootForwardLock4")
	l.SetRootL(r4)
	err = l.Save(ctx)
	require.NoError(t, err)
	l2, err = goradd_unit.LoadLeafL(ctx, l.ID(), node.LeafL().RootL())
	require.NoError(t, err)
	require.NotNilf(t, l2, "Object was nil based on ID %s", l.ID())
	assert.Equal(t, "leafForwardLock4", l2.Name())
	assert.Equal(t, "rootForwardLock4", l2.RootL().Name())
}

// TestForwardCollision tests saving two records that are changed at the same time.
func TestForwardLockCollision(t *testing.T) {
	ctx := context.Background()

	l := goradd_unit.NewLeafL()
	r := goradd_unit.NewRootL()
	r.SetName("rootForwardLockCollision")
	l.SetName("leafForwardLockCollision")
	l.SetRootL(r)
	err := l.Save(ctx)
	require.NoError(t, err)

	var l2 *goradd_unit.LeafL
	l2, err = goradd_unit.LoadLeafL(ctx, l.ID(), node.LeafL().RootL())
	require.NoError(t, err)

	// Update both
	l.SetName("leafForwardLockCollision2")
	l2.SetName("leafForwardLockCollision3")

	err = l.Save(ctx)
	err2 := l2.Save(ctx)
	assert.NoError(t, err)
	assert.Error(t, err2)
	assert.IsType(t, &db.OptimisticLockError{}, err2)

	// 2nd level
	l2, _ = goradd_unit.LoadLeafL(ctx, l.ID(), node.LeafL().RootL())
	l.RootL().SetName("rootForwardLockCollision2")
	l2.RootL().SetName("rootForwardLockCollision3")
	err = l.Save(ctx)
	err2 = l2.Save(ctx)
	assert.NoError(t, err)
	assert.Error(t, err2)
	assert.IsType(t, &db.OptimisticLockError{}, err2)

	// Delete
	l2, _ = goradd_unit.LoadLeafL(ctx, l.ID(), node.LeafL().RootL())
	assert.NoError(t, l.RootL().Delete(ctx))
	l2.SetName("leafForwardLockCollision4")
	err2 = l2.Save(ctx)
	assert.IsType(t, &db.OptimisticLockError{}, err2)
}

func TestForwardLockNull(t *testing.T) {
	ctx := context.Background()
	l := goradd_unit.NewLeafL()
	l.SetName("leafForwardLockNull")
	assert.Panics(t, func() {
		_ = l.Save(ctx) // root is required since RootID is not nullable
	})

	assert.Panics(t, func() {
		l.SetRootL(nil) // not nullable
	})
}

func TestForwardLockTwo(t *testing.T) {
	ctx := context.Background()
	l := goradd_unit.NewLeafL()
	r := goradd_unit.NewRootL()
	l.SetName("leafForwardLockTwo")
	r.SetName("rootForwardLockTwo")
	l.SetRootL(r)
	require.NoError(t, l.Save(ctx))

	l2 := goradd_unit.NewLeafL()
	l2.SetName("leafForwardLockTwo2")
	l2.SetRootL(r)
	require.NoError(t, l2.Save(ctx))

	r2, err := goradd_unit.LoadRootL(ctx, r.ID(), node.RootL().LeafLs())
	require.NoError(t, err)
	assert.Len(t, r2.LeafLs(), 2)
}

func TestForwardLockDelete(t *testing.T) {
	ctx := context.Background()
	l := goradd_unit.NewLeafL()
	r := goradd_unit.NewRootL()
	l.SetName("leafForwardLockDelete")
	r.SetName("rootForwardLockDelete")
	l.SetRootL(r)
	require.NoError(t, l.Save(ctx))

	// Collision on shallow change
	l2, err := goradd_unit.LoadLeafL(ctx, l.ID(), node.LeafL().RootL())
	require.NoError(t, err)
	l.SetName("leafForwardLockDelete2")
	_ = l.Save(ctx)
	err = l2.Delete(ctx)
	require.Error(t, err)
	assert.IsType(t, &db.OptimisticLockError{}, err)

	// Collision on deep Delete
	l2, err = goradd_unit.LoadLeafL(ctx, l.ID())
	assert.NoError(t, err)
	require.NotNil(t, l2)
	err = l.RootL().Delete(ctx)
	require.NoError(t, err)
	err = l2.Delete(ctx)
	assert.Error(t, err)
	assert.IsType(t, &db.OptimisticLockError{}, err)

	// Deep delete deleted the linked record
	l2, err = goradd_unit.LoadLeafL(ctx, l.ID())
	assert.NoError(t, err)
	assert.Nil(t, l2)
}

func TestForwardLockJson(t *testing.T) {
	ctx := context.Background()
	l := goradd_unit.NewLeafL()
	r := goradd_unit.NewRootL()
	l.SetName("leafForwardLockJson")
	r.SetName("rootForwardLockJson")
	l.SetRootL(r)
	require.NoError(t, l.Save(ctx))

	j, err := l.MarshalJSON()
	require.NoError(t, err)

	var m map[string]any

	err = json.Unmarshal(j, &m)
	require.NoError(t, err)
	v, ok := m["rootL"]
	assert.True(t, ok)
	assert.Equal(t, "rootForwardLockJson", v.(map[string]any)["name"].(string))

	var leaf goradd_unit.LeafL
	err = json.Unmarshal(j, &leaf)
	require.NoError(t, err)
	assert.Equal(t, "leafForwardLockJson", leaf.Name())
	assert.Equal(t, "rootForwardLockJson", leaf.RootL().Name())
}
