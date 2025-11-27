package crud

import (
	"context"
	"testing"

	"github.com/goradd/gro/_test/gen/orm/goradd_unit"
	"github.com/goradd/gro/_test/gen/orm/goradd_unit/node"
	"github.com/goradd/gro/pkg/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestForwardNullable tests insert and update of two linked records where the link is nullable.
func TestForwardUniqueNullable(t *testing.T) {
	ctx := context.Background()

	// Insert-insert
	l1 := goradd_unit.NewLeafUn()
	r1 := goradd_unit.NewRootUn()
	r1.SetName("rootForwardUniqueNullable")
	l1.SetName("leafForwardUniqueNullable")
	l1.SetRootUn(r1)
	require.NoError(t, l1.Save(ctx))

	// loading back
	l1b, err := goradd_unit.LoadLeafUn(ctx, l1.ID(), node.LeafUn().RootUn())
	require.NoError(t, err)
	require.NotNilf(t, l1b, "Object was nil based on ID %s", l1.ID())
	assert.Equal(t, "leafForwardUniqueNullable", l1b.Name())
	assert.Equal(t, "rootForwardUniqueNullable", l1b.RootUn().Name())

	// Update-update
	l1.SetName("leafForwardUniqueNullable2")
	l1.RootUn().SetName("rootForwardUniqueNullable2")
	err = l1.Save(ctx)
	assert.NoError(t, err)
	l1b, err = goradd_unit.LoadLeafUn(ctx, l1.ID(), node.LeafUn().RootUn())
	require.NoError(t, err)
	require.NotNilf(t, l1b, "Object was nil based on ID %s", l1.ID())
	assert.Equal(t, "leafForwardUniqueNullable2", l1b.Name())
	assert.Equal(t, "rootForwardUniqueNullable2", l1b.RootUn().Name())

	// Insert-update
	l2 := goradd_unit.NewLeafUn()
	l2.SetName("leafForwardUniqueNullable3")
	r1.SetName("rootForwardUniqueNullable3")
	l2.SetRootUn(r1)
	err = l2.Save(ctx)
	assert.Error(t, err)
	assert.IsType(t, &db.UniqueValueError{}, err)

	// Update-insert
	r3 := goradd_unit.NewRootUn()
	l1.SetName("leafForwardUniqueNullable4")
	r3.SetName("rootForwardUniqueNullable4")
	l1.SetRootUn(r3)
	err = l1.Save(ctx)
	require.NoError(t, err)
	l1b, err = goradd_unit.LoadLeafUn(ctx, l1.ID(), node.LeafUn().RootUn())
	require.NoError(t, err)
	require.NotNilf(t, l1b, "Object was nil based on ID %s", l1.ID())
	assert.Equal(t, "leafForwardUniqueNullable4", l1b.Name())
	assert.Equal(t, "rootForwardUniqueNullable4", l1b.RootUn().Name())
}

// TestForwardNullableCollision tests saving two records that are changed at the same time.
func TestForwardUniqueNullableCollision(t *testing.T) {
	ctx := context.Background()

	l := goradd_unit.NewLeafUn()
	r := goradd_unit.NewRootUn()
	r.SetName("rootForwardUniqueNullableCollision")
	l.SetName("leafForwardUniqueNullableCollision")
	l.SetRootUn(r)
	err := l.Save(ctx)
	require.NoError(t, err)

	var l2 *goradd_unit.LeafUn
	l2, err = goradd_unit.LoadLeafUn(ctx, l.ID(), node.LeafUn().RootUn())
	require.NoError(t, err)

	// Update first
	l.SetName("leafForwardUniqueNullableCollision2")
	l.RootUn().SetName("rootForwardUniqueNullableCollision2")

	// Update second
	l2.SetName("leafForwardUniqueNullableCollision3")
	l2.RootUn().SetName("rootForwardUniqueNullableCollision3")

	// save first then second
	err = l.Save(ctx)
	err2 := l2.Save(ctx)
	assert.NoError(t, err)
	assert.NoError(t, err2)

	// Last save should win
	l3, err3 := goradd_unit.LoadLeafUn(ctx, l.ID(), node.LeafUn().RootUn())
	assert.NoError(t, err3)
	assert.Equal(t, "leafForwardUniqueNullableCollision3", l3.Name())
	assert.Equal(t, "rootForwardUniqueNullableCollision3", l3.RootUn().Name())
}

func TestForwardUniqueNullableNull(t *testing.T) {
	ctx := context.Background()
	l := goradd_unit.NewLeafUn()
	l.SetName("leafForwardUniqueNullableNull")
	assert.NoError(t, l.Save(ctx))

	l.SetRootUn(nil) // nullable
	assert.NoError(t, l.Save(ctx))
}

func TestForwardUniqueNullableTwo(t *testing.T) {
	ctx := context.Background()
	r := goradd_unit.NewRootUn()
	l := goradd_unit.NewLeafUn()
	r.SetName("rootForwardUniqueNullableTwo")
	l.SetName("leafForwardUniqueNullableTwo")
	r.SetLeafUn(l)
	require.NoError(t, r.Save(ctx))

	l2 := goradd_unit.NewLeafUn()
	l2.SetName("leafForwardUniqueNullableTwo2")
	r.SetLeafUn(l2)
	require.NoError(t, r.Save(ctx)) // unique failure

	l2, err := goradd_unit.LoadLeafUn(ctx, l.ID())
	require.NoError(t, err)
	assert.NotNil(t, l2) // reverse linked item that is allowed to have a nil pointer was not deleted
}

func TestForwardUniqueNullableDelete(t *testing.T) {
	ctx := context.Background()
	l := goradd_unit.NewLeafUn()
	r := goradd_unit.NewRootUn()
	l.SetName("leafForwardUniqueNullableDelete")
	r.SetName("rootForwardUniqueNullableDelete")
	l.SetRootUn(r)
	require.NoError(t, l.Save(ctx))

	require.NoError(t, l.Delete(ctx))

	l2, err := goradd_unit.LoadLeafUn(ctx, l.ID())
	require.NoError(t, err)
	assert.Nil(t, l2)

	r2, err := goradd_unit.LoadRootUn(ctx, r.ID())
	require.NoError(t, err)
	assert.NotNil(t, r2)
}
