package crud

import (
	"context"
	"testing"

	"github.com/davecgh/go-spew/spew"
	goradd_unit2 "github.com/goradd/gro/ci/tests/gen/goradd_unit"
	"github.com/goradd/gro/ci/tests/gen/goradd_unit/node"
	"github.com/goradd/gro/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestForwardUnique covers insert/update flows and enforces that only one LeafU may point to a given RootU.
func TestForwardUnique(t *testing.T) {
	ctx := context.Background()

	// Insert-insert
	l1 := goradd_unit2.NewLeafU()
	r1 := goradd_unit2.NewRootU()
	r1.SetName("rootForwardUnique")
	l1.SetName("leafForwardUnique")
	l1.SetRootU(r1)
	require.NoError(t, l1.Save(ctx))

	// loading back
	l1b, err := goradd_unit2.LoadLeafU(ctx, l1.ID(), node.LeafU().RootU())
	require.NoError(t, err)
	require.NotNilf(t, l1b, "Object was nil based on ID %s", l1.ID())
	assert.Equal(t, "leafForwardUnique", l1b.Name())
	assert.Equal(t, "rootForwardUnique", l1b.RootU().Name())

	// Update-update
	l1.SetName("leafForwardUnique2")
	l1.RootU().SetName("rootForwardUnique2")
	err = l1.Save(ctx)
	assert.NoError(t, err)
	l1b, err = goradd_unit2.LoadLeafU(ctx, l1.ID(), node.LeafU().RootU())
	require.NoError(t, err)
	require.NotNilf(t, l1b, "Object was nil based on ID %s", l1.ID())
	assert.Equal(t, "leafForwardUnique2", l1b.Name())
	assert.Equal(t, "rootForwardUnique2", l1b.RootU().Name())

	// Insert-update
	l2 := goradd_unit2.NewLeafU()
	l2.SetName("leafForwardUnique3")
	r1.SetName("rootForwardUnique3")
	l2.SetRootU(r1)
	err = l2.Save(ctx)
	assert.Error(t, err)
	assert.IsType(t, &db.UniqueValueError{}, err)

	// Update-insert
	r3 := goradd_unit2.NewRootU()
	l1.SetName("leafForwardUnique4")
	r3.SetName("rootForwardUnique4")
	l1.SetRootU(r3)
	err = l1.Save(ctx)
	require.NoError(t, err)
	l1b, err = goradd_unit2.LoadLeafU(ctx, l1.ID(), node.LeafU().RootU())
	require.NoError(t, err)
	require.NotNilf(t, l1b, "Object was nil based on ID %s", l1.ID())
	assert.Equal(t, "leafForwardUnique4", l1b.Name())
	assert.Equal(t, "rootForwardUnique4", l1b.RootU().Name())
}

// TestForwardUniqueCollision tests saving two records that are changed at the same time.
func TestForwardUniqueCollision(t *testing.T) {
	ctx := context.Background()

	l := goradd_unit2.NewLeafU()
	r := goradd_unit2.NewRootU()
	r.SetName("rootForwardUniqueCollision")
	l.SetName("leafForwardUniqueCollision")
	l.SetRootU(r)
	err := l.Save(ctx)
	require.NoError(t, err)

	var l2 *goradd_unit2.LeafU
	l2, err = goradd_unit2.LoadLeafU(ctx, l.ID(), node.LeafU().RootU())
	require.NoError(t, err)

	// Update first
	l.SetName("leafForwardUniqueCollision2")
	l.RootU().SetName("rootForwardUniqueCollision2")

	// Update second
	l2.SetName("leafForwardUniqueCollision3")
	l2.RootU().SetName("rootForwardUniqueCollision3")

	// save first then second
	err = l.Save(ctx)
	err2 := l2.Save(ctx)
	assert.NoError(t, err)
	assert.NoError(t, err2)

	// Last save should win
	l3, err3 := goradd_unit2.LoadLeafU(ctx, l.ID(), node.LeafU().RootU())
	assert.NoError(t, err3)
	assert.Equal(t, "leafForwardUniqueCollision3", l3.Name())
	require.NotNil(t, "rootForwardUniqueCollision3", l3.RootU(), "Missing selected object: RootU "+spew.Sdump(l3))
	assert.Equal(t, "rootForwardUniqueCollision3", l3.RootU().Name())
}

func TestForwardUniqueNull(t *testing.T) {
	ctx := context.Background()
	l := goradd_unit2.NewLeafU()
	l.SetName("leafForwardUniqueNull")
	assert.Panics(t, func() {
		_ = l.Save(ctx) // root is required since RootID is not nullable
	})

	assert.Panics(t, func() {
		l.SetRootU(nil) // not nullable
	})
}
func TestForwardUniqueTwo(t *testing.T) {
	ctx := context.Background()
	l := goradd_unit2.NewLeafU()
	r := goradd_unit2.NewRootU()
	l.SetName("leafForwardUniqueTwo")
	r.SetName("rootForwardUniqueTwo")
	l.SetRootU(r)
	require.NoError(t, l.Save(ctx))

	l2 := goradd_unit2.NewLeafU()
	l2.SetName("leafForwardUniqueTwo2")
	l2.SetRootU(r)
	require.Error(t, l2.Save(ctx)) // unique value collision error
}

func TestForwardUniqueDelete(t *testing.T) {
	ctx := context.Background()
	l := goradd_unit2.NewLeafU()
	r := goradd_unit2.NewRootU()
	l.SetName("leafForwardUniqueDelete")
	r.SetName("rootForwardUniqueDelete")
	l.SetRootU(r)
	require.NoError(t, l.Save(ctx))

	require.NoError(t, l.Delete(ctx))

	l2, err := goradd_unit2.LoadLeafU(ctx, l.ID())
	require.NoError(t, err)
	assert.Nil(t, l2)

	r2, err := goradd_unit2.LoadRootU(ctx, r.ID())
	require.NoError(t, err)
	assert.NotNil(t, r2)
}
