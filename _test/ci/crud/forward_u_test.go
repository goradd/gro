package crud

import (
	"context"
	"github.com/goradd/orm/_test/gen/orm/goradd_unit"
	"github.com/goradd/orm/_test/gen/orm/goradd_unit/node"
	"github.com/goradd/orm/pkg/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

// TestForwardUnique covers insert/update flows and enforces that only one LeafU may point to a given RootU.
func TestForwardUnique(t *testing.T) {
	ctx := context.Background()
	defer goradd_unit.ClearAll(ctx)

	// Insert-insert
	l1 := goradd_unit.NewLeafU()
	r1 := goradd_unit.NewRootU()
	r1.SetName("root")
	l1.SetName("leaf")
	l1.SetRootU(r1)
	require.NoError(t, l1.Save(ctx))

	// loading back
	l1b, err := goradd_unit.LoadLeafU(ctx, l1.ID(), node.LeafU().RootU())
	require.NoError(t, err)
	assert.Equal(t, "leaf", l1b.Name())
	assert.Equal(t, "root", l1b.RootU().Name())

	// Update-update
	l1.SetName("leaf2")
	l1.RootU().SetName("root2")
	err = l1.Save(ctx)
	assert.NoError(t, err)
	l1b, err = goradd_unit.LoadLeafU(ctx, l1.ID(), node.LeafU().RootU())
	require.NoError(t, err)
	assert.Equal(t, "leaf2", l1b.Name())
	assert.Equal(t, "root2", l1b.RootU().Name())

	// Insert-update
	l2 := goradd_unit.NewLeafU()
	l2.SetName("leaf3")
	r1.SetName("root3")
	l2.SetRootU(r1)
	err = l2.Save(ctx)
	assert.Error(t, err)
	assert.IsType(t, &db.UniqueValueError{}, err)

	// Update-insert
	r3 := goradd_unit.NewRootU()
	l1.SetName("leaf4")
	r3.SetName("root4")
	l1.SetRootU(r3)
	err = l1.Save(ctx)
	require.NoError(t, err)
	l1b, err = goradd_unit.LoadLeafU(ctx, l1.ID(), node.LeafU().RootU())
	require.NoError(t, err)
	assert.Equal(t, "leaf4", l1b.Name())
	assert.Equal(t, "root4", l1b.RootU().Name())
}

// TestForwardUniqueCollision tests saving two records that are changed at the same time.
func TestForwardUniqueCollision(t *testing.T) {
	ctx := context.Background()
	defer goradd_unit.ClearAll(ctx)

	l := goradd_unit.NewLeafU()
	r := goradd_unit.NewRootU()
	r.SetName("root")
	l.SetName("leaf")
	l.SetRootU(r)
	err := l.Save(ctx)
	require.NoError(t, err)

	var l2 *goradd_unit.LeafU
	l2, err = goradd_unit.LoadLeafU(ctx, l.ID(), node.LeafU().RootU())
	require.NoError(t, err)

	// Update first
	l.SetName("leaf2")
	l.RootU().SetName("root2")

	// Update second
	l2.SetName("leaf3")
	l2.RootU().SetName("root3")

	// save first then second
	err = l.Save(ctx)
	err2 := l2.Save(ctx)
	assert.NoError(t, err)
	assert.NoError(t, err2)

	// Last save should win
	l3, err3 := goradd_unit.LoadLeafU(ctx, l.ID(), node.LeafU().RootU())
	assert.NoError(t, err3)
	assert.Equal(t, "leaf3", l3.Name())
	assert.Equal(t, "root3", l3.RootU().Name())
}

func TestForwardUniqueNull(t *testing.T) {
	ctx := context.Background()
	defer goradd_unit.ClearAll(ctx)
	l := goradd_unit.NewLeafU()
	l.SetName("leaf")
	assert.Panics(t, func() {
		_ = l.Save(ctx) // root is required since RootID is not nullable
	})

	assert.Panics(t, func() {
		l.SetRootU(nil) // not nullable
	})
}
func TestForwardUniqueTwo(t *testing.T) {
	ctx := context.Background()
	defer goradd_unit.ClearAll(ctx)
	l := goradd_unit.NewLeafU()
	r := goradd_unit.NewRootU()
	l.SetName("leaf")
	r.SetName("root")
	l.SetRootU(r)
	require.NoError(t, l.Save(ctx))

	l2 := goradd_unit.NewLeafU()
	l2.SetName("leaf2")
	l2.SetRootU(r)
	require.Error(t, l2.Save(ctx)) // unique value collision error
}

func TestForwardUniqueDelete(t *testing.T) {
	ctx := context.Background()
	defer goradd_unit.ClearAll(ctx)
	l := goradd_unit.NewLeafU()
	r := goradd_unit.NewRootU()
	l.SetName("leaf")
	r.SetName("root")
	l.SetRootU(r)
	require.NoError(t, l.Save(ctx))

	require.NoError(t, l.Delete(ctx))

	l2, err := goradd_unit.LoadLeafU(ctx, l.ID())
	require.NoError(t, err)
	assert.Nil(t, l2)

	r2, err := goradd_unit.LoadRootU(ctx, r.ID())
	require.NoError(t, err)
	assert.NotNil(t, r2)
}
