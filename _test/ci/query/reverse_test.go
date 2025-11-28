package query

import (
	"context"
	"os"
	"testing"

	"github.com/davecgh/go-spew/spew"
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
	if len(names) == 0 {
		spew.Dump(people)
		os.Exit(1)
	}
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
