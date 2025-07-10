package ci

import (
	"github.com/goradd/orm/_test/gen/orm/goradd"
	"github.com/goradd/orm/_test/gen/orm/goradd_unit"
	"github.com/goradd/orm/pkg/db"
	"github.com/stretchr/testify/assert"
	"testing"
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
	login.SetUsername("system")
	err := login.Save(ctx)
	assert.Error(t, err)

	login, err = goradd.LoadLoginByUsername(ctx, "system")
	assert.NoError(t, err)
	login.SetUsername("jdoe")
	err = login.Save(ctx)
	assert.Error(t, err)
}

func TestUnique2Value(t *testing.T) {
	ctx := context.Background()
	i := goradd_unit.NewDoubleIndex()
	i.SetID(1)
	i.SetFieldInt(1)
	i.SetFieldString("blah")
	err := i.Save(ctx)
	assert.NoError(t, err)
	defer i.Delete(ctx)
	i2 := i.Copy()
	i2.SetID(2)
	err = i2.Save(ctx)
	assert.Error(t, err, "error on collision of insert with unique index on 2 columns")

	i3 := goradd_unit.NewDoubleIndex()
	i3.SetID(2)
	i3.SetFieldInt(2)
	i3.SetFieldString("blah2")
	err = i3.Save(ctx)
	assert.NoError(t, err)
	defer i3.Delete(ctx)
	i3.SetFieldInt(1)
	i3.SetFieldString("blah")
	err = i3.Save(ctx)
	assert.Error(t, err, "updating double-unique index detects collision")
}
