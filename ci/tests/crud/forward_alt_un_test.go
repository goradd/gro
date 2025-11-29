package crud

import (
	"context"
	"testing"

	"github.com/goradd/gro/ci/tests/gen/goradd_unit"
	goradd_unit2 "github.com/goradd/gro/ci/tests/gen/goradd_unit"
	"github.com/goradd/gro/ci/tests/gen/goradd_unit/node"
	"github.com/goradd/gro/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestForwardAltUniqueNullable tests insert and update of two linked records where the link is nullable
// and the foreign key is to a manual non-string and non-integer primary key.
func TestForwardAltUniqueNullable(t *testing.T) {
	ctx := context.Background()

	// Insert-insert
	l1 := goradd_unit2.NewAltLeafUn()
	r1 := goradd_unit.NewAltRootUn()
	r1.SetID(1.1)
	r1.SetName("rootForwardAltUniqueNullable")
	l1.SetName("leafForwardAltUniqueNullable")
	l1.SetAltRootUn(r1)
	require.NoError(t, l1.Save(ctx))

	// loading back
	l1b, err := goradd_unit2.LoadAltLeafUn(ctx, l1.ID(), node.AltLeafUn().AltRootUn())
	require.NoError(t, err)
	require.NotNilf(t, l1b, "Object was nil based on ID %s", l1.ID())
	assert.Equal(t, "leafForwardAltUniqueNullable", l1b.Name())
	assert.Equal(t, "rootForwardAltUniqueNullable", l1b.AltRootUn().Name())

	// Update-update
	l1.SetName("leafForwardAltUniqueNullable2")
	l1.AltRootUn().SetName("rootForwardAltUniqueNullable2")
	err = l1.Save(ctx)
	assert.NoError(t, err)
	l1b, err = goradd_unit2.LoadAltLeafUn(ctx, l1.ID(), node.AltLeafUn().AltRootUn())
	require.NoError(t, err)
	require.NotNilf(t, l1b, "Object was nil based on ID %s", l1.ID())
	assert.Equal(t, "leafForwardAltUniqueNullable2", l1b.Name())
	assert.Equal(t, "rootForwardAltUniqueNullable2", l1b.AltRootUn().Name())

	// Insert-update
	l2 := goradd_unit2.NewAltLeafUn()
	l2.SetName("leafForwardAltUniqueNullable3")
	r1.SetName("rootForwardAltUniqueNullable3")
	l2.SetAltRootUn(r1)
	err = l2.Save(ctx)
	assert.Error(t, err)
	assert.IsType(t, &db.UniqueValueError{}, err)

	// Update-insert
	r3 := goradd_unit.NewAltRootUn()
	l1.SetName("leafForwardAltUniqueNullable4")
	r3.SetName("rootForwardAltUniqueNullable4")
	r3.SetID(1.2)
	l1.SetAltRootUn(r3)
	err = l1.Save(ctx)
	require.NoError(t, err)
	l1b, err = goradd_unit2.LoadAltLeafUn(ctx, l1.ID(), node.AltLeafUn().AltRootUn())
	require.NoError(t, err)
	require.NotNilf(t, l1b, "Object was nil based on ID %s", l1.ID())
	assert.Equal(t, "leafForwardAltUniqueNullable4", l1b.Name())
	assert.Equal(t, "rootForwardAltUniqueNullable4", l1b.AltRootUn().Name())
}
