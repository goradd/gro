package crud

import (
	"github.com/goradd/orm/_test/gen/orm/goradd_unit"
	"github.com/goradd/orm/_test/gen/orm/goradd_unit/node"
	"github.com/goradd/orm/pkg/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

// TestForwardAltUniqueNullable tests insert and update of two linked records where the link is nullable
// and the foreign key is to a manual non-string and non-integer primary key.
func TestForwardAltUniqueNullable(t *testing.T) {
	ctx := context.Background()
	defer goradd_unit.ClearAll(ctx)

	// Insert-insert
	l1 := goradd_unit.NewAltLeafUn()
	r1 := goradd_unit.NewAltRootUn()
	r1.SetID(1.1)
	r1.SetName("root")
	l1.SetName("leaf")
	l1.SetAltRootUn(r1)
	require.NoError(t, l1.Save(ctx))

	// loading back
	l1b, err := goradd_unit.LoadAltLeafUn(ctx, l1.ID(), node.AltLeafUn().AltRootUn())
	require.NoError(t, err)
	assert.Equal(t, "leaf", l1b.Name())
	assert.Equal(t, "root", l1b.AltRootUn().Name())

	// Update-update
	l1.SetName("leaf2")
	l1.AltRootUn().SetName("root2")
	err = l1.Save(ctx)
	assert.NoError(t, err)
	l1b, err = goradd_unit.LoadAltLeafUn(ctx, l1.ID(), node.AltLeafUn().AltRootUn())
	require.NoError(t, err)
	assert.Equal(t, "leaf2", l1b.Name())
	assert.Equal(t, "root2", l1b.AltRootUn().Name())

	// Insert-update
	l2 := goradd_unit.NewAltLeafUn()
	l2.SetName("leaf3")
	r1.SetName("root3")
	l2.SetAltRootUn(r1)
	err = l2.Save(ctx)
	assert.Error(t, err)
	assert.IsType(t, &db.UniqueValueError{}, err)

	// Update-insert
	r3 := goradd_unit.NewAltRootUn()
	l1.SetName("leaf4")
	r3.SetName("root4")
	r3.SetID(1.2)
	l1.SetAltRootUn(r3)
	err = l1.Save(ctx)
	require.NoError(t, err)
	l1b, err = goradd_unit.LoadAltLeafUn(ctx, l1.ID(), node.AltLeafUn().AltRootUn())
	require.NoError(t, err)
	assert.Equal(t, "leaf4", l1b.Name())
	assert.Equal(t, "root4", l1b.AltRootUn().Name())
}
