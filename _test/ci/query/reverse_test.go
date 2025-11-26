package query

import (
	"context"
	"testing"

	"github.com/goradd/gro/_test/gen/orm/goradd"
	"github.com/goradd/gro/_test/gen/orm/goradd/node"
	"github.com/goradd/gro/pkg/op"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReverseReference(t *testing.T) {
	ctx := context.Background()
	people, err := goradd.QueryPeople(ctx).
		Select(node.Person().ManagerProjects()).
		OrderBy(node.Person().LastName()).
		Load()
	assert.NoError(t, err)
	if people[2].FirstName() != "John" {
		t.Error("Did not find John.")
	}

	if len(people[2].ManagerProjects()) != 1 {
		t.Error("Did not find 1 ManagerProjects.")
	}

}

func TestReverseConditionalSelect(t *testing.T) {
	ctx := context.Background()
	people, err := goradd.QueryPeople(ctx).
		OrderBy(node.Person().ID(), node.Person().ManagerProjects().Name()).
		Where(op.IsNotNull(node.Person().ManagerProjects().ID())). // Filter out people who are not managers
		Select(node.Person().ManagerProjects().Name()).
		Load()
	assert.NoError(t, err)
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
	ctx := context.Background()
	people, err := goradd.QueryPeople(ctx).
		OrderBy(node.Person().LastName(), node.Person().ManagerProjects().TeamMembers().LastName(), node.Person().ManagerProjects().TeamMembers().FirstName()).
		Select(node.Person().ManagerProjects().TeamMembers().FirstName(), node.Person().ManagerProjects().TeamMembers().LastName()).
		Load()
	assert.NoError(t, err)
	var names []string
	for _, p := range people[2].ManagerProjects()[0].TeamMembers() {
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
	person := people[11]
	for _, pr := range person.ManagerProjects() {
		for _, p := range pr.TeamMembers() {
			names = append(names, p.FirstName()+" "+p.LastName())
		}
	}
	assert.Len(t, names, 12) // Includes duplicates. If we ever get Distinct to manually remove duplicates, we should fix this.

	// Test deep IsDirty and Save

	assert.False(t, person.IsDirty())
	id := person.ManagerProjects()[0].TeamMembers()[0].ID()
	fn := person.ManagerProjects()[0].TeamMembers()[0].FirstName()
	person.ManagerProjects()[0].TeamMembers()[0].SetFirstName("A")
	assert.True(t, person.IsDirty())
	assert.NoError(t, person.Save(ctx))
	p, err2 := goradd.LoadPerson(ctx, id)
	assert.NoError(t, err2)
	assert.Equal(t, "A", p.FirstName())
	assert.False(t, people[6].IsDirty())
	// restore
	p.SetFirstName(fn)
	assert.NoError(t, p.Save(ctx))
}

func TestUniqueReverseLoad(t *testing.T) {
	ctx := context.Background()
	person, err := goradd.QueryPeople(ctx).
		Where(op.Equal(node.Person().LastName(), "Doe")).
		Get()
	assert.NoError(t, err)
	assert.Nil(t, person.Login())

	people, err2 := goradd.QueryPeople(ctx).
		Where(op.Equal(node.Person().LastName(), "Doe")).
		Select(node.Person().Login()).
		Load()
	assert.NoError(t, err2)
	person = people[0]
	assert.Equal(t, "jdoe", person.Login().Username())
}

func TestReverseUniqueInsert(t *testing.T) {
	ctx := context.Background()
	var err error

	// Test insert
	person := goradd.NewPerson()
	person.SetID("106")
	person.SetFirstName("Sam")
	person.SetLastName("I Am")

	// Not null reverse unique
	empInfo := goradd.NewEmployeeInfo()
	empInfo.SetID("107")
	empInfo.SetEmployeeNumber(55)
	person.SetEmployeeInfo(empInfo)

	// Nullable reverse unique
	login := goradd.NewLogin()
	login.SetID("105")
	login.SetUsername("sammy")
	person.SetLogin(login)

	assert.NoError(t, person.Save(ctx))

	assert.NotZero(t, person.Login().PersonID())
	assert.Equal(t, person.ID(), person.Login().PersonID())
	assert.NotZero(t, person.EmployeeInfo().PersonID())
	assert.Equal(t, person.ID(), person.EmployeeInfo().PersonID())

	empInfo, err = goradd.LoadEmployeeInfo(ctx, person.EmployeeInfo().ID())
	assert.NoError(t, err)
	assert.Equal(t, person.ID(), empInfo.PersonID())

	login, err = goradd.LoadLogin(ctx, person.Login().ID())
	assert.NoError(t, err)
	assert.Equal(t, person.ID(), login.PersonID())

	person2 := goradd.NewPerson()
	person2.SetID("109")
	person2.SetFirstName("Yertle")
	person2.SetLastName("The Turtle")
	person2.SetLogin(person.Login())
	person2.SetEmployeeInfo(person.EmployeeInfo())
	assert.NoError(t, person2.Save(ctx))

	person2, err = goradd.LoadPerson(ctx, person2.ID(), node.Person().EmployeeInfo(), node.Person().Login())
	assert.NoError(t, err)
	assert.Equal(t, person2.ID(), person2.EmployeeInfo().PersonID())
	assert.Equal(t, person.EmployeeInfo().ID(), person2.EmployeeInfo().ID())
	assert.Equal(t, person2.ID(), person2.Login().PersonID())
	assert.Equal(t, person.Login().ID(), person2.Login().ID())

	assert.NoError(t, person2.Delete(ctx))
	empInfo, err = goradd.LoadEmployeeInfo(ctx, person.EmployeeInfo().ID())
	assert.NoError(t, err)
	assert.Nil(t, empInfo)
	login, err = goradd.LoadLogin(ctx, person.Login().ID())
	assert.NoError(t, err)
	assert.Zero(t, login.PersonID())
	assert.NoError(t, login.Delete(ctx))
	login, err = goradd.LoadLogin(ctx, person.Login().ID())
	assert.NoError(t, err)
	assert.Nil(t, login)
	assert.NoError(t, person.Delete(ctx))
	person, err = goradd.LoadPerson(ctx, person.ID())
	assert.NoError(t, err)
	assert.Nil(t, person)
}

func TestReverseManyNotNullInsert(t *testing.T) {
	ctx := context.Background()
	// Test insert
	person := goradd.NewPerson()
	person.SetID("110")
	person.SetFirstName("Sam")
	person.SetLastName("I Am")

	addr1 := goradd.NewAddress()
	addr1.SetID("111")
	addr1.SetCity("Here")
	addr1.SetStreet("There")

	addr2 := goradd.NewAddress()
	addr2.SetID("112")
	addr2.SetCity("Near")
	addr2.SetStreet("Far")

	person.SetAddresses(
		addr1, addr2,
	)
	assert.NoError(t, person.Save(ctx))

	id := person.ID()

	addr1Id := addr1.ID()
	assert.NotEmpty(t, addr1Id)

	addr3 := person.Address(addr1Id)
	assert.Equal(t, "There", addr3.Street(), "Successfully attached the new addresses onto the person object.")

	person2, err := goradd.LoadPerson(ctx, id, node.Person().Addresses())
	assert.NoError(t, err)

	assert.Equal(t, "Sam", person2.FirstName(), "Retrieved the correct person")
	assert.Equal(t, 2, len(person2.Addresses()), "Retrieved the addresses attached to the person")

	assert.NoError(t, person2.Delete(ctx))

	person3, err2 := goradd.LoadPerson(ctx, id, node.Person().Addresses())
	assert.NoError(t, err2)
	assert.Nil(t, person3, "Successfully deleted the new person")

	addr4, err3 := goradd.LoadAddress(ctx, addr1Id)
	assert.NoError(t, err3)
	assert.Nil(t, addr4, "Successfully deleted the address attached to the person")
}

func TestReverseManyNullInsertNewObject(t *testing.T) {
	ctx := context.Background()
	// Test insert
	project := goradd.NewProject()
	project.SetID("200")
	project.SetName("Big project")
	project.SetNum(100)
	project.SetStatus(goradd.ProjectStatusOpen)

	person := goradd.NewPerson()
	person.SetID("201")
	person.SetFirstName("Cat")
	person.SetLastName("In the Hat")

	project.SetManager(person)
	assert.NoError(t, project.Save(ctx))

	project2, err := goradd.LoadProject(ctx, project.ID(), node.Project().Manager())
	assert.NoError(t, err)
	assert.Equal(t, "Big project", project2.Name())
	assert.Equal(t, "Cat", project2.Manager().FirstName())

	// Delete manager and see that projects get set to null manager
	assert.NoError(t, person.Delete(ctx))
	project2, err = goradd.LoadProject(ctx, project.ID(), node.Project().Manager())
	assert.NoError(t, err)
	assert.True(t, project2.ManagerIDIsNull())

	// Set pre-existing manager via pk
	project2.SetManagerID("1")
	assert.NoError(t, project2.Save(ctx))

	project2, err = goradd.LoadProject(ctx, project.ID(), node.Project().Manager())
	assert.NoError(t, err)
	assert.Equal(t, "Doe", project2.Manager().LastName())

	// Delete project and see manager still exists
	assert.NoError(t, project2.Delete(ctx))
	person, err = goradd.LoadPerson(ctx, "1")
	assert.NoError(t, err)
	assert.NotNil(t, person)
}

func TestReverseReferenceCount(t *testing.T) {
	ctx := context.Background()

	person, err := goradd.LoadPerson(ctx, "3")
	assert.NoError(t, err)
	ct, err2 := person.CountAddresses(ctx)
	assert.NoError(t, err2)
	assert.Equal(t, 2, ct)

}

func TestReverseLoad(t *testing.T) {
	ctx := context.Background()

	project, err := goradd.LoadProject(ctx, "1")
	assert.NoError(t, err)
	_, err = project.LoadMilestones(ctx)
	assert.NoError(t, err)
	milestone := project.Milestone("3")
	assert.NotNil(t, milestone)
	assert.Equal(t, "3", milestone.ID())
}

func TestReverseLoadUnsaved(t *testing.T) {
	ctx := context.Background()

	project, err := goradd.LoadProject(ctx, "1")
	assert.NoError(t, err)
	_, err = project.LoadMilestones(ctx)
	assert.NoError(t, err)
	milestone := project.Milestone("3")
	milestone.SetName("A new name")
	assert.Panics(t, func() {
		_, _ = project.LoadMilestones(ctx)
	})
}

func TestReverseSelectByID(t *testing.T) {
	ctx := context.Background()

	projects, err := goradd.QueryProjects(ctx).
		OrderBy(node.Project().Name().Descending()).
		Load()
	assert.NoError(t, err)

	require.Len(t, projects, 4)
	id := projects[3].ID()

	// Reverse references
	people, err2 := goradd.QueryPeople(ctx).
		Select(node.Person().ManagerProjects()).
		Where(op.Equal(node.Person().LastName(), "Wolfe")).
		Load()
	assert.NoError(t, err2)

	p := people[0]
	require.NotNil(t, p)
	m := p.ManagerProject(id)
	require.NotNil(t, m, "Could not fine project as manager: "+id)
	assert.Equal(t, m.Name(), "ACME Payment System")
}

func TestReverseSet(t *testing.T) {
	ctx := context.Background()

	person, err := goradd.QueryPeople(ctx).
		Where(op.Equal(node.Person().ID(), "7")).
		Select(node.Person().ManagerProjects()).
		OrderBy(node.Person().ManagerProjects().ID()).
		Get()
	assert.NoError(t, err)
	projects := person.ManagerProjects()
	assert.Len(t, projects, 2)
	assert.Equal(t, "1", projects[0].ID())
	assert.Equal(t, "4", projects[1].ID())

	newProjects, err2 := goradd.QueryProjects(ctx).
		Where(op.In(node.Project().ID(), "1", "2")).
		Load()
	assert.NoError(t, err2)
	person.SetManagerProjects(newProjects...)
	require.NoError(t, person.Save(ctx))

	personTest, err3 := goradd.QueryPeople(ctx).
		Where(op.Equal(node.Person().ID(), "7")).
		Select(node.Person().ManagerProjects()).
		OrderBy(node.Person().ManagerProjects().ID()).
		Get()
	assert.NoError(t, err3)
	projectsTest := personTest.ManagerProjects()
	assert.Len(t, projectsTest, 2)
	assert.Equal(t, "1", projectsTest[0].ID())
	assert.Equal(t, "2", projectsTest[1].ID())

	// Set none
	person.SetManagerProjects()
	require.NoError(t, person.Save(ctx))
	c, err4 := goradd.CountProjectsByManagerID(ctx, person.ID())
	assert.NoError(t, err4)
	assert.Equal(t, 0, c)

	// restore
	person.SetManagerProjects(projects...)
	require.NoError(t, person.Save(ctx))

	person, err = goradd.QueryPeople(ctx).
		Where(op.Equal(node.Person().ID(), "7")).
		Select(node.Person().ManagerProjects()).
		OrderBy(node.Person().ManagerProjects().ID()).
		Get()
	assert.NoError(t, err)
	projects = person.ManagerProjects()
	assert.Len(t, projects, 2)
	assert.Equal(t, "1", projects[0].ID())
	assert.Equal(t, "4", projects[1].ID())

	// Fix nil value caused by removal of project
	projectsTest[1].SetManagerID("4")
	assert.NoError(t, projectsTest[1].Save(ctx))

}
