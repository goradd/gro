package ci

import (
	"github.com/goradd/orm/_test/gen/orm/goradd"
	"github.com/goradd/orm/_test/gen/orm/goradd/node"
	"github.com/goradd/orm/pkg/db"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestReference(t *testing.T) {
	ctx := db.NewContext(nil)
	projects := goradd.QueryProjects(ctx).
		Select(node.Project().Manager()).
		OrderBy(node.Project().ID()).
		Load()

	if projects[0].Manager().FirstName() != "Karen" {
		t.Error("Person found not Karen, found " + projects[0].Manager().FirstName())
	}
}

func TestReferenceUpdate(t *testing.T) {
	ctx := db.NewContext(nil)

	// Test updating only a referenced object, and then saving, making sure all updates get recorded

	// Update an already linked object
	project := goradd.LoadProject(ctx, "1", node.Project().Manager())
	manager := project.Manager()
	fn := manager.FirstName()
	manager.SetFirstName("abcd")
	assert.NoError(t, project.Save(ctx))
	p := goradd.LoadPerson(ctx, manager.ID())
	assert.Equal(t, "abcd", p.FirstName())
	p.SetFirstName(fn)
	p.Save(ctx)

	// Create a newly linked object
	addr := goradd.NewAddress()
	addr.SetCity("Panama City")
	addr.SetStreet("1 El Camino")
	addr.SetPersonID("1")
	addr.Save(ctx)
	defer addr.Delete(ctx)
	person := goradd.NewPerson()
	person.SetFirstName("Jorge")
	person.SetLastName("Gonzales")
	addr.SetPerson(person)
	addr.Save(ctx)
	defer person.Delete(ctx)
	assert.NotEqual(t, "", addr.PersonID())
	assert.Equal(t, person.ID(), addr.PersonID())
}
