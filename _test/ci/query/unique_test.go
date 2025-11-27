package query

import (
	"context"
	"testing"

	"github.com/goradd/gro/_test/gen/orm/goradd"
	"github.com/goradd/gro/_test/gen/orm/goradd_unit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUniquePrimaryKey(t *testing.T) {
	ctx := context.Background()
	gift := goradd.NewGift()
	gift.SetNumber(1)
	gift.SetName("Conflict")
	err := gift.Save(ctx)
	assert.Error(t, err)
}

func TestUniqueValue(t *testing.T) {
	ctx := context.Background()
	login := goradd.NewLogin()
	login.SetID("300")
	login.SetUsername("system")
	err := login.Save(ctx)
	assert.Error(t, err)

	login, err = goradd.LoadLoginByUsername(ctx, "system")
	assert.NoError(t, err)
	login.SetUsername("jdoe")
	err = login.Save(ctx)
	assert.Error(t, err)
}

/*
func TestAlot(t *testing.T) {
	for i := 0; i < 1000; i++ {
		TestUnique2Value(t)

	}
}*/

func TestUnique2Value(t *testing.T) {
	ctx := context.Background()
	i := goradd_unit.NewDoubleIndex()
	i.SetID(1)
	i.SetFieldInt(1)
	i.SetFieldString("test")
	err := i.Save(ctx)
	assert.NoError(t, err)

	i2 := i.Copy()
	i2.SetID(2)
	err = i2.Save(ctx)
	return
	require.Error(t, err, "error on collision of insert with unique index on 2 columns")

	i3 := goradd_unit.NewDoubleIndex()
	i3.SetID(2)
	i3.SetFieldInt(2)
	i3.SetFieldString("test2")
	err = i3.Save(ctx)
	assert.NoError(t, err)
	i3.SetFieldInt(1)
	i3.SetFieldString("test")
	err = i3.Save(ctx)
	require.Error(t, err, "updating double-unique index detects collision")

	// Cleanup only if not failing so we can see error in db
	i.Delete(ctx)
	i3.Delete(ctx)
}
