package ci

import (
	"github.com/goradd/orm/_test/gen/orm/goradd"
	"github.com/goradd/orm/_test/gen/orm/goradd/node"
	"github.com/goradd/orm/pkg/db"
	"github.com/goradd/orm/pkg/op"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBasicEnum(t *testing.T) {
	ctx := db.NewContext(nil)
	projects, err := goradd.QueryProjects(ctx).
		OrderBy(node.Project().ID()).
		Load()
	assert.NoError(t, err)
	if projects[0].Status() != goradd.ProjectStatusCompleted {
		t.Error("Did not find correct project type.")
	}
}

func TestManyEnum(t *testing.T) {
	ctx := db.NewContext(nil)
	people, err := goradd.QueryPeople(ctx).
		OrderBy(node.Person().ID()).
		Load()
	assert.NoError(t, err)
	if people[0].Types().Len() != 2 {
		t.Error("Did not expand to 2 person types.")
	}

	if !people[0].Types().Has(goradd.PersonTypeInactive) {
		t.Error("Did not find correct person type.")
	}
}

func TestManyEnumSingles(t *testing.T) {
	ctx := db.NewContext(nil)
	people, err := goradd.QueryPeople(ctx).
		OrderBy(node.Person().ID()).
		Load()
	assert.NoError(t, err)
	if !people[4].Types().Has(goradd.PersonTypeWorksFromHome) {
		t.Error("Did not find correct person type.")
	}
}

func TestManyEnums(t *testing.T) {
	ctx := db.NewContext(nil)

	// All people who are inactive
	people, err := goradd.QueryPeople(ctx).
		OrderBy(node.Person().LastName(), node.Person().FirstName()).
		Where(op.Contains(node.Person().Types(), goradd.PersonTypeInactive)).
		Distinct().
		Select(node.Person().LastName(), node.Person().FirstName()).
		Load()
	assert.NoError(t, err)
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
