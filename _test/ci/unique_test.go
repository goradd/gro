package ci

import (
	"github.com/goradd/orm/_test/gen/orm/goradd"
	"github.com/goradd/orm/pkg/db"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUniquePrimaryKey(t *testing.T) {
	ctx := db.NewContext(nil)
	gift := goradd.NewGift()
	gift.SetNumber(1)
	gift.SetName("Conflict")
	err := gift.Save(ctx)
	assert.Error(t, err)
}
