package crud

import (
	"context"
	"testing"

	"github.com/goradd/gro/ci/tests/gen/goradd_unit"
	goradd_unit2 "github.com/goradd/gro/ci/tests/gen/goradd_unit"
	"github.com/goradd/gro/ci/tests/gen/goradd_unit/node"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestReverseAltUniqueNullable tests insert and update of two linked records.
func TestReverseAltUniqueNullable(t *testing.T) {
	ctx := context.Background()

	// Insert-insert
	r := goradd_unit.NewAltRootUn()
	l := goradd_unit2.NewAltLeafUn()
	r.SetName("rootReverseAltUniqueNullable")
	r.SetID(1.1)
	l.SetName("leafReverseAltUniqueNullable")
	r.SetAltLeafUn(l)
	err := r.Save(ctx)
	require.NoError(t, err)

	var r2 *goradd_unit.AltRootUn
	r2, err = goradd_unit2.LoadAltRootUn(ctx, r.ID(), node.AltRootUn().AltLeafUn())
	require.NoError(t, err)
	require.NotNilf(t, r2, "Object was nil based on ID %s", r.ID())
	assert.Equal(t, "rootReverseAltUniqueNullable", r2.Name())
	assert.Equal(t, "leafReverseAltUniqueNullable", r2.AltLeafUn().Name())

	// Update-update
	r.SetName("rootReverseAltUniqueNullable2")
	r.AltLeafUn().SetName("leafReverseAltUniqueNullable2")
	err = r.Save(ctx)
	assert.NoError(t, err)
	r2, err = goradd_unit2.LoadAltRootUn(ctx, r.ID(), node.AltRootUn().AltLeafUn())
	require.NoError(t, err)
	require.NotNilf(t, r2, "Object was nil based on ID %s", r.ID())
	assert.Equal(t, "rootReverseAltUniqueNullable2", r2.Name())
	assert.Equal(t, "leafReverseAltUniqueNullable2", r2.AltLeafUn().Name())

	// Insert-update
	r3 := goradd_unit.NewAltRootUn()
	r3.SetName("rootReverseAltUniqueNullable3")
	r3.SetID(1.2)
	l.SetName("leafReverseAltUniqueNullable3")
	r3.SetAltLeafUn(l)
	err = r3.Save(ctx)
	require.NoError(t, err)
	r2, err = goradd_unit2.LoadAltRootUn(ctx, r3.ID(), node.AltRootUn().AltLeafUn())
	require.NoError(t, err)
	require.NotNilf(t, r2, "Object was nil based on ID %s", r3.ID())
	assert.Equal(t, "rootReverseAltUniqueNullable3", r2.Name())
	assert.Equal(t, "leafReverseAltUniqueNullable3", r2.AltLeafUn().Name())

	// Update-insert
	l4 := goradd_unit2.NewAltLeafUn()
	r.SetName("rootReverseAltUniqueNullable4")
	l4.SetName("leafReverseAltUniqueNullable4")
	r.SetAltLeafUn(l4)
	err = r.Save(ctx)
	require.NoError(t, err)
	r2, err = goradd_unit2.LoadAltRootUn(ctx, r.ID(), node.AltRootUn().AltLeafUn())
	require.NotNilf(t, r2, "Object was nil based on ID %s", r.ID())
	require.NoError(t, err)
	assert.Equal(t, "rootReverseAltUniqueNullable4", r2.Name())
	assert.Equal(t, "leafReverseAltUniqueNullable4", r2.AltLeafUn().Name())

}
