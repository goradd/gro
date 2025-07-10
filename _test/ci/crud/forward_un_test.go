package crud

import (
	"github.com/goradd/orm/_test/gen/orm/goradd_unit"
	"github.com/goradd/orm/_test/gen/orm/goradd_unit/node"
	"github.com/goradd/orm/pkg/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

// TestForwardNullable tests insert and update of two linked records where the link is nullable.
func TestForwardUniqueNullable(t *testing.T) {
	ctx := context.Background()
	defer goradd_unit.ClearAll(ctx)

	// Insert-insert
	l1 := goradd_unit.NewLeafUn()
	r1 := goradd_unit.NewRootUn()
	r1.SetName("root")
	l1.SetName("leaf")
	l1.SetRootUn(r1)
	require.NoError(t, l1.Save(ctx))

	// loading back
	l1b, err := goradd_unit.LoadLeafUn(ctx, l1.ID(), node.LeafUn().RootUn())
	require.NoError(t, err)
	assert.Equal(t, "leaf", l1b.Name())
	assert.Equal(t, "root", l1b.RootUn().Name())

	// Update-update
	l1.SetName("leaf2")
	l1.RootUn().SetName("root2")
	err = l1.Save(ctx)
	assert.NoError(t, err)
	l1b, err = goradd_unit.LoadLeafUn(ctx, l1.ID(), node.LeafUn().RootUn())
	require.NoError(t, err)
	assert.Equal(t, "leaf2", l1b.Name())
	assert.Equal(t, "root2", l1b.RootUn().Name())

	// Insert-update
	l2 := goradd_unit.NewLeafUn()
	l2.SetName("leaf3")
	r1.SetName("root3")
	l2.SetRootUn(r1)
	err = l2.Save(ctx)
	assert.Error(t, err)
	assert.IsType(t, &db.UniqueValueError{}, err)

	// Update-insert
	r3 := goradd_unit.NewRootUn()
	l1.SetName("leaf4")
	r3.SetName("root4")
	l1.SetRootUn(r3)
	err = l1.Save(ctx)
	require.NoError(t, err)
	l1b, err = goradd_unit.LoadLeafUn(ctx, l1.ID(), node.LeafUn().RootUn())
	require.NoError(t, err)
	assert.Equal(t, "leaf4", l1b.Name())
	assert.Equal(t, "root4", l1b.RootUn().Name())
}

// TestForwardNullableCollision tests saving two records that are changed at the same time.
func TestForwardUniqueNullableCollision(t *testing.T) {
	ctx := context.Background()
	defer goradd_unit.ClearAll(ctx)

	l := goradd_unit.NewLeafUn()
	r := goradd_unit.NewRootUn()
	r.SetName("root")
	l.SetName("leaf")
	l.SetRootUn(r)
	err := l.Save(ctx)
	require.NoError(t, err)

	var l2 *goradd_unit.LeafUn
	l2, err = goradd_unit.LoadLeafUn(ctx, l.ID(), node.LeafUn().RootUn())
	require.NoError(t, err)

	// Update first
	l.SetName("leaf2")
	l.RootUn().SetName("root2")

	// Update second
	l2.SetName("leaf3")
	l2.RootUn().SetName("root3")

	// save first then second
	err = l.Save(ctx)
	err2 := l2.Save(ctx)
	assert.NoError(t, err)
	assert.NoError(t, err2)

	// Last save should win
	l3, err3 := goradd_unit.LoadLeafUn(ctx, l.ID(), node.LeafUn().RootUn())
	assert.NoError(t, err3)
	assert.Equal(t, "leaf3", l3.Name())
	assert.Equal(t, "root3", l3.RootUn().Name())
}

func TestForwardUniqueNullableNull(t *testing.T) {
	ctx := context.Background()
	defer goradd_unit.ClearAll(ctx)
	l := goradd_unit.NewLeafUn()
	l.SetName("leaf")
	assert.NoError(t, l.Save(ctx))

	l.SetRootUn(nil) // nullable
	assert.NoError(t, l.Save(ctx))
}

func TestForwardUniqueNullableTwo(t *testing.T) {
	ctx := context.Background()
	defer goradd_unit.ClearAll(ctx)
	r := goradd_unit.NewRootUn()
	l := goradd_unit.NewLeafUn()
	r.SetName("root")
	l.SetName("leaf")
	r.SetLeafUn(l)
	require.NoError(t, r.Save(ctx))

	l2 := goradd_unit.NewLeafUn()
	l2.SetName("leaf2")
	r.SetLeafUn(l2)
	require.NoError(t, r.Save(ctx)) // unique failure

	l2, err := goradd_unit.LoadLeafUn(ctx, l.ID())
	require.NoError(t, err)
	assert.NotNil(t, l2) // reverse linked item that is allowed to have a nil pointer was not deleted
}

func TestForwardUniqueNullableDelete(t *testing.T) {
	ctx := context.Background()
	defer goradd_unit.ClearAll(ctx)
	l := goradd_unit.NewLeafUn()
	r := goradd_unit.NewRootUn()
	l.SetName("leaf")
	r.SetName("root")
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
