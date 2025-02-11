package ci

import (
	"github.com/goradd/orm/_test/gen/orm/goradd"
	"github.com/goradd/orm/_test/gen/orm/goradd/node"
	"github.com/goradd/orm/pkg/db"
	"github.com/goradd/orm/pkg/op"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestReverseConditionalSelect(t *testing.T) {
	ctx := db.NewContext(nil)
	people := goradd.QueryPeople(ctx).
		OrderBy(node.Person().ID(), node.Person().ManagerProjects().Name()).
		Where(op.IsNotNull(node.Person().ManagerProjects().ID())). // Filter out people who are not managers
		Select(node.Person().ManagerProjects().Name()).
		Load()

	if len(people[2].ManagerProjects()) != 2 {
		t.Error("Did not find 2 ManagerProjects.")
	}

	assert.Len(t, people[2].ManagerProjects(), 2)
	assert.Equal(t, people[2].ManagerProjects()[0].Name(), "ACME Payment System")
	assert.True(t, people[2].ManagerProjects()[0].IDIsValid())
	assert.False(t, people[2].ManagerProjects()[0].NumIsValid())
}

// Complex test finding all the team members of all the projects a person is managing, ordering by last name
func TestReverseMany(t *testing.T) {
	ctx := db.NewContext(nil)
	people := goradd.QueryPeople(ctx).
		OrderBy(node.Person().ID(), node.Person().ManagerProjects().TeamMembers().LastName(), node.Person().ManagerProjects().TeamMembers().FirstName()).
		Select(node.Person().ManagerProjects().TeamMembers().FirstName(), node.Person().ManagerProjects().TeamMembers().LastName()).
		Load()

	var names []string
	for _, p := range people[0].ManagerProjects()[0].TeamMembers() {
		names = append(names, p.FirstName()+" "+p.LastName())
	}
	names2 := []string{
		"John Doe",
		"Mike Ho",
		"Samantha Jones",
		"Jennifer Smith",
		"Wendy Smith",
	}
	assert.Equal(t, names2, names)

	names = []string{}
	for _, pr := range people[6].ManagerProjects() {
		for _, p := range pr.TeamMembers() {
			names = append(names, p.FirstName()+" "+p.LastName())
		}
	}
	assert.Len(t, names, 12) // Includes duplicates. If we ever get Distinct to manually remove duplicates, we should fix this.

	// Test deep IsDirty
	assert.False(t, people[6].IsDirty())
	people[6].ManagerProjects()[0].TeamMembers()[0].SetFirstName("A")
	assert.True(t, people[6].IsDirty())
}

func TestUniqueReverse(t *testing.T) {
	ctx := db.NewContext(nil)
	person := goradd.QueryPeople(ctx).
		Where(op.Equal(node.Person().LastName(), "Doe")).
		Get()
	assert.Nil(t, person.Login())

	person = goradd.QueryPeople(ctx).
		Where(op.Equal(node.Person().LastName(), "Doe")).
		Select(node.Person().Login()).
		Load()[0]
	assert.Equal(t, "jdoe", person.Login().Username())
}

/*
// TestReverseReferenceManySave is testing save and delete for a reverse reference that cannot be null.
func TestReverseReferenceManySave(t *testing.T) {
	ctx := db.NewContext(nil)
	// Test insert
	person := goradd.NewPerson()
	person.SetFirstName("Sam")
	person.SetLastName("I Am")

	addr1 := goradd.NewAddress()
	addr1.SetCity("Here")
	addr1.SetStreet("There")

	addr2 := goradd.NewAddress()
	addr2.SetCity("Near")
	addr2.SetStreet("Far")

	person.SetAddresses([]*goradd.Address{
		addr1, addr2,
	})

	person.Save(ctx)
	id := person.ID()

	addr1Id := addr1.ID()
	assert.NotEmpty(t, addr1Id)

	addr3 := person.Address(addr1Id)
	assert.Equal(t, "There", addr3.Street(), "Successfully attached the new addresses onto the person object.")

	person2 := goradd.LoadPerson(ctx, id, node.Person().Addresses())

	assert.Equal(t, "Sam", person2.FirstName(), "Retrieved the correct person")
	assert.Equal(t, 2, len(person2.Addresses()), "Retrieved the addresses attached to the person")

	person2.Delete(ctx)

	person3 := goradd.LoadPerson(ctx, id, node.Person().Addresses())
	assert.Nil(t, person3, "Successfully deleted the new person")

	addr4 := goradd.LoadAddress(ctx, addr1Id)
	assert.Nil(t, addr4, "Successfully deleted the address attached to the person")

}

// Testing a reverse reference with a unique index, which will cause a one-to-one relationship.
// This tests save and delete
func TestReverseReferenceUniqueSave(t *testing.T) {
	ctx := db.NewContext(nil)

	person := goradd.NewPerson()
	person.SetFirstName("Sam")
	person.SetLastName("I Am")

	e1 := goradd.NewEmployeeInfo()
	e1.SetEmployeeNumber(12345)
	person.SetEmployeeInfo(e1)

	person.Save(ctx)
	id := person.ID()

	e1Id := e1.ID()
	assert.NotEmpty(t, e1Id)

	e2 := person.EmployeeInfo()
	assert.Equal(t, e1Id, e2.ID(), "Successfully attached the new employee info object onto the person object.")

	person2 := goradd.LoadPerson(ctx, id, node.Person().EmployeeInfo())

	assert.Equal(t, "Sam", person2.FirstName(), "Retrieved the correct person")
	assert.Equal(t, e1Id, person2.EmployeeInfo().ID(), "Retrieved the employee info attached to the person")

	person2.Delete(ctx)

	person3 := goradd.LoadPerson(ctx, id, node.Person().EmployeeInfo())
	assert.Nil(t, person3, "Successfully deleted the new person")

	e4 := goradd.LoadEmployeeInfo(ctx, e1Id)
	assert.Nil(t, e4, "Successfully deleted the employee info attached to the person")

}
*/

func TestReverseReferenceCount(t *testing.T) {
	ctx := db.NewContext(nil)

	person := goradd.LoadPerson(ctx, "3")
	ct := person.CountAddresses(ctx)
	assert.Equal(t, 2, ct)

}

func TestReverseLoad(t *testing.T) {
	ctx := db.NewContext(nil)

	project := goradd.LoadProject(ctx, "1")
	project.LoadMilestones(ctx)
	milestone := project.Milestone("3")
	assert.NotNil(t, milestone)
	assert.Equal(t, "3", milestone.ID())
}

func TestReverseLoadUnsaved(t *testing.T) {
	ctx := db.NewContext(nil)

	project := goradd.LoadProject(ctx, "1")
	project.LoadMilestones(ctx)
	milestone := project.Milestone("3")
	milestone.SetName("A new name")
	assert.Panics(t, func() {
		project.LoadMilestones(ctx)
	})
}

/*
func TestReverseSet(t *testing.T) {
	ctx := db.NewContext(nil)

	person := goradd.QueryPeople(ctx).
		Where(op.Equal(node.Person().ID(), "7")).
		Select(node.Person().ManagerProjects()).
		OrderBy(node.Person().ManagerProjects().ID()).
		Get()
	projects := person.ManagerProjects()
	assert.Len(t, projects, 2)
	assert.Equal(t, "1", projects[0].ID())
	assert.Equal(t, "4", projects[1].ID())

	newProjects := goradd.QueryProjects(ctx).
		Where(op.In(node.Project().ID(), "1", "2")).
		Load()
	person.SetManagerProjects(newProjects)
	person.Save(ctx)

	personTest := goradd.QueryPeople(ctx).
		Where(op.Equal(node.Person().ID(), "7")).
		Select(node.Person().ManagerProjects()).
		OrderBy(node.Person().ManagerProjects().ID()).
		Get()
	projectsTest := personTest.ManagerProjects()
	assert.Len(t, projectsTest, 2)
	assert.Equal(t, "1", projectsTest[0].ID())
	assert.Equal(t, "2", projectsTest[1].ID())

	person.SetManagerProjects(projects)
	person.Save(ctx)

	person = goradd.QueryPeople(ctx).
		Where(op.Equal(node.Person().ID(), "7")).
		Select(node.Person().ManagerProjects()).
		OrderBy(node.Person().ManagerProjects().ID()).
		Get()
	projects = person.ManagerProjects()
	assert.Len(t, projects, 2)
	assert.Equal(t, "1", projects[0].ID())
	assert.Equal(t, "4", projects[1].ID())

	projectsTest[1].SetManagerID("4")
	projectsTest[1].Save(ctx)

}
*/
