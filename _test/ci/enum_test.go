package ci

import (
	"github.com/goradd/orm/_test/gen/orm/goradd"
	"github.com/goradd/orm/_test/gen/orm/goradd/node"
	"github.com/goradd/orm/pkg/db"
	"github.com/goradd/orm/pkg/op"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestManyEnums(t *testing.T) {
	ctx := db.NewContext(nil)

	// All people who are inactive
	people := goradd.QueryPeople(ctx).
		OrderBy(node.Person().LastName(), node.Person().FirstName()).
		Where(op.Contains(node.Person().Types(), goradd.PersonTypeInactive)).
		Distinct().
		Select(node.Person().LastName(), node.Person().FirstName()).
		Load()

	names := []string{}
	for _, p := range people {
		names = append(names, p.FirstName()+" "+p.LastName())
	}
	names2 := []string{
		"Linda Brady",
		"John Doe",
		"Ben Robinson",
	}
	assert.Equal(t, names2, names)
}
