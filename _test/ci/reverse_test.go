package ci

import (
	"github.com/goradd/orm/_test/gen/orm/goradd"
	"github.com/goradd/orm/_test/gen/orm/goradd/node"
	"github.com/goradd/orm/pkg/db"
	"github.com/goradd/orm/pkg/op"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	assert.True(t, people[2].ManagerProjects()[0].IDIsLoaded())
	assert.False(t, people[2].ManagerProjects()[0].NumIsLoaded())
}

// Complex test finding all the team members of all the projects a person is managing, ordering by last name
func TestReverseManyLoad(t *testing.T) {
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

	// Test deep IsDirty and Save
	assert.False(t, people[6].IsDirty())
	id := people[6].ManagerProjects()[0].TeamMembers()[0].ID()
	fn := people[6].ManagerProjects()[0].TeamMembers()[0].FirstName()
	people[6].ManagerProjects()[0].TeamMembers()[0].SetFirstName("A")
	assert.True(t, people[6].IsDirty())
	assert.NoError(t, people[6].Save(ctx))
	p := goradd.LoadPerson(ctx, id)
	assert.Equal(t, "A", p.FirstName())
	assert.False(t, people[6].IsDirty())
	// restore
	p.SetFirstName(fn)
	p.Save(ctx)
}

func TestUniqueReverseLoad(t *testing.T) {
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

func TestReverseUniqueInsert(t *testing.T) {
	ctx := db.NewContext(nil)
	// Test insert
	person := goradd.NewPerson()
	person.SetFirstName("Sam")
	person.SetLastName("I Am")

	// Not null reverse unique
	empInfo := goradd.NewEmployeeInfo()
	empInfo.SetEmployeeNumber(55)
	person.SetEmployeeInfo(empInfo)

	// Nullable reverse unique
	login := goradd.NewLogin()
	login.SetUsername("sammy")
	person.SetLogin(login)

	person.Save(ctx)

	assert.NotZero(t, person.Login().PersonID())
	assert.Equal(t, person.ID(), person.Login().PersonID())
	assert.NotZero(t, person.EmployeeInfo().PersonID())
	assert.Equal(t, person.ID(), person.EmployeeInfo().PersonID())

	empInfo = goradd.LoadEmployeeInfo(ctx, person.EmployeeInfo().ID())
	assert.Equal(t, person.ID(), empInfo.PersonID())

	login = goradd.LoadLogin(ctx, person.Login().ID())
	assert.Equal(t, person.ID(), login.PersonID())

	person2 := goradd.NewPerson()
	person2.SetFirstName("Yertle")
	person2.SetLastName("The Turtle")
	person2.SetLogin(person.Login())
	person2.SetEmployeeInfo(person.EmployeeInfo())
	person2.Save(ctx)

	person2 = goradd.LoadPerson(ctx, person2.ID(), node.Person().EmployeeInfo(), node.Person().Login())
	assert.Equal(t, person2.ID(), person2.EmployeeInfo().PersonID())
	assert.Equal(t, person.EmployeeInfo().ID(), person2.EmployeeInfo().ID())
	assert.Equal(t, person2.ID(), person2.Login().PersonID())
	assert.Equal(t, person.Login().ID(), person2.Login().ID())

	person2.Delete(ctx)
	empInfo = goradd.LoadEmployeeInfo(ctx, person.EmployeeInfo().ID())
	assert.Nil(t, empInfo)
	login = goradd.LoadLogin(ctx, person.Login().ID())
	assert.Zero(t, login.PersonID())
	login.Delete(ctx)
	login = goradd.LoadLogin(ctx, person.Login().ID())
	assert.Nil(t, login)
	person.Delete(ctx)
	assert.Nil(t, goradd.LoadPerson(ctx, person.ID()))
}

func TestReverseManyNotNullInsert(t *testing.T) {
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

	person.SetAddresses(
		addr1, addr2,
	)
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

func TestReverseManyNullInsertNewObject(t *testing.T) {
	ctx := db.NewContext(nil)
	// Test insert
	project := goradd.NewProject()
	project.SetName("Big project")
	project.SetNum(100)
	project.SetStatus(goradd.ProjectStatusOpen)

	person := goradd.NewPerson()
	person.SetFirstName("Cat")
	person.SetLastName("In the Hat")

	project.SetManager(person)
	project.Save(ctx)

	project2 := goradd.LoadProject(ctx, project.ID(), node.Project().Manager())
	assert.Equal(t, "Big project", project2.Name())
	assert.Equal(t, "Cat", project2.Manager().FirstName())

	// Delete manager and see that projects get set to null manager
	person.Delete(ctx)
	project2 = goradd.LoadProject(ctx, project.ID(), node.Project().Manager())
	assert.True(t, project2.ManagerIDIsNull())

	// Set pre-existing manager via pk
	project2.SetManagerID("1")
	project2.Save(ctx)

	project2 = goradd.LoadProject(ctx, project.ID(), node.Project().Manager())
	assert.Equal(t, "Doe", project2.Manager().LastName())

	// Delete project and see manager still exists
	project2.Delete(ctx)
	person = goradd.LoadPerson(ctx, "1")
	assert.NotNil(t, person)
}

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

func TestReverseSelectByID(t *testing.T) {
	ctx := db.NewContext(nil)

	projects := goradd.QueryProjects(ctx).
		OrderBy(node.Project().Name().Descending()).
		Load()

	require.Len(t, projects, 4)
	id := projects[3].ID()

	// Reverse references
	people := goradd.QueryPeople(ctx).
		Select(node.Person().ManagerProjects()).
		Where(op.Equal(node.Person().LastName(), "Wolfe")).
		Load()

	p := people[0]
	require.NotNil(t, p)
	m := p.ManagerProject(id)
	require.NotNil(t, m, "Could not fine project as manager: "+id)
	assert.Equal(t, m.Name(), "ACME Payment System")
}

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
	person.SetManagerProjects(newProjects...)
	require.NoError(t, person.Save(ctx))

	personTest := goradd.QueryPeople(ctx).
		Where(op.Equal(node.Person().ID(), "7")).
		Select(node.Person().ManagerProjects()).
		OrderBy(node.Person().ManagerProjects().ID()).
		Get()
	projectsTest := personTest.ManagerProjects()
	assert.Len(t, projectsTest, 2)
	assert.Equal(t, "1", projectsTest[0].ID())
	assert.Equal(t, "2", projectsTest[1].ID())

	// Set none
	person.SetManagerProjects()
	require.NoError(t, person.Save(ctx))
	assert.Equal(t, 0, goradd.CountProjectsByManagerID(ctx, person.ID()))

	// restore
	person.SetManagerProjects(projects...)
	require.NoError(t, person.Save(ctx))

	person = goradd.QueryPeople(ctx).
		Where(op.Equal(node.Person().ID(), "7")).
		Select(node.Person().ManagerProjects()).
		OrderBy(node.Person().ManagerProjects().ID()).
		Get()
	projects = person.ManagerProjects()
	assert.Len(t, projects, 2)
	assert.Equal(t, "1", projects[0].ID())
	assert.Equal(t, "4", projects[1].ID())

	// Fix nil value caused by removal of project
	projectsTest[1].SetManagerID("4")
	projectsTest[1].Save(ctx)

}
