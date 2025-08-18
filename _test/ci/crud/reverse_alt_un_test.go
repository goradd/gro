package crud

import (
	"context"
	"testing"

	"github.com/goradd/orm/_test/gen/orm/goradd_unit"
	"github.com/goradd/orm/_test/gen/orm/goradd_unit/node"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestReverseAltUniqueNullable tests insert and update of two linked records.
func TestReverseAltUniqueNullable(t *testing.T) {
	ctx := context.Background()
	defer goradd_unit.ClearAll(ctx)

	// Insert-insert
	r := goradd_unit.NewAltRootUn()
	l := goradd_unit.NewAltLeafUn()
	r.SetName("root")
	r.SetID(1.1)
	l.SetName("leaf")
	r.SetAltLeafUn(l)
	err := r.Save(ctx)
	require.NoError(t, err)

	var r2 *goradd_unit.AltRootUn
	r2, err = goradd_unit.LoadAltRootUn(ctx, r.ID(), node.AltRootUn().AltLeafUn())
	require.NoError(t, err)
	require.NotNilf(t, r2, "Object was nil based on ID %s", r.ID())
	assert.Equal(t, "root", r2.Name())
	assert.Equal(t, "leaf", r2.AltLeafUn().Name())

	// Update-update
	r.SetName("root2")
	r.AltLeafUn().SetName("leaf2")
	err = r.Save(ctx)
	assert.NoError(t, err)
	r2, err = goradd_unit.LoadAltRootUn(ctx, r.ID(), node.AltRootUn().AltLeafUn())
	require.NoError(t, err)
	require.NotNilf(t, r2, "Object was nil based on ID %s", r.ID())
	assert.Equal(t, "root2", r2.Name())
	assert.Equal(t, "leaf2", r2.AltLeafUn().Name())

	// Insert-update
	r3 := goradd_unit.NewAltRootUn()
	r3.SetName("root3")
	r3.SetID(1.2)
	l.SetName("leaf3")
	r3.SetAltLeafUn(l)
	err = r3.Save(ctx)
	require.NoError(t, err)
	r2, err = goradd_unit.LoadAltRootUn(ctx, r3.ID(), node.AltRootUn().AltLeafUn())
	require.NoError(t, err)
	require.NotNilf(t, r2, "Object was nil based on ID %s", r3.ID())
	assert.Equal(t, "root3", r2.Name())
	assert.Equal(t, "leaf3", r2.AltLeafUn().Name())

	// Update-insert
	l4 := goradd_unit.NewAltLeafUn()
	r.SetName("root4")
	l4.SetName("leaf4")
	r.SetAltLeafUn(l4)
	err = r.Save(ctx)
	require.NoError(t, err)
	r2, err = goradd_unit.LoadAltRootUn(ctx, r.ID(), node.AltRootUn().AltLeafUn())
	require.NotNilf(t, r2, "Object was nil based on ID %s", r.ID())
	require.NoError(t, err)
	assert.Equal(t, "root4", r2.Name())
	assert.Equal(t, "leaf4", r2.AltLeafUn().Name())

}
