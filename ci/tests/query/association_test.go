package query

import (
	"context"
	"strconv"
	"testing"

	goradd2 "github.com/goradd/gro/ci/tests/gen/goradd"
	node2 "github.com/goradd/gro/ci/tests/gen/goradd/node"
	"github.com/goradd/gro/query/op"
	"github.com/stretchr/testify/assert"
)

func TestManyMany(t *testing.T) {
	ctx := context.Background()
	projects, err := goradd2.QueryProjects(ctx).
		Select(node2.Project().TeamMembers()).
		OrderBy(node2.Project().ID()).
		Load()

	assert.NoError(t, err)
	if len(projects[0].TeamMembers()) != 5 {
		t.Error("Did not find 5 team members in project 1. Found: " + strconv.Itoa(len(projects[0].TeamMembers())))
	}

}

func TestMany2(t *testing.T) {

	ctx := context.Background()

	// All People Who Are on a Project Managed by Karen Wolfe (Person Value #7)
	people, err := goradd2.QueryPeople(ctx).
		OrderBy(node2.Person().LastName(), node2.Person().FirstName()).
		Where(op.Equal(node2.Person().Projects().Manager().LastName(), "Wolfe")).
		Distinct().
		Select(node2.Person().LastName(), node2.Person().FirstName()).
		Load()
	assert.NoError(t, err)
	names := []string{}
	for _, p := range people {
		names = append(names, p.FirstName()+" "+p.LastName())
	}
	names2 := []string{
		"Brett Carlisle",
		"John Doe",
		"Samantha Jones",
		"Jacob Pratt",
		"Kendall Public",
		"Ben Robinson",
		"Alex Smith",
		"Wendy Smith",
		"Karen Wolfe",
	}

	assert.Equal(t, names2, names)
}

func TestManySelect(t *testing.T) {
	ctx := context.Background()

	people, err := goradd2.QueryPeople(ctx).
		OrderBy(node2.Person().LastName(), node2.Person().FirstName(), node2.Person().Projects().Name()).
		Where(op.Equal(node2.Person().Projects().Manager().LastName(), "Wolfe")).
		Select(node2.Person().LastName(), node2.Person().FirstName(), node2.Person().Projects().Name()).
		Load()
	assert.NoError(t, err)
	person := people[0]
	projects := person.Projects()
	name := projects[0].Name()

	assert.Equal(t, "ACME Payment System", name)
}

func Test2Nodes(t *testing.T) {
	ctx := context.Background()
	milestones, err := goradd2.QueryMilestones(ctx).
		Select(node2.Milestone().Project().Manager()).
		Where(op.Equal(node2.Milestone().ID(), "1")). // Filter out people who are not managers
		Load()
	assert.NoError(t, err)
	assert.True(t, milestones[0].NameIsLoaded(), "Milestone 1 has a name")
	assert.Equal(t, "Milestone A", milestones[0].Name(), "Milestone 1 has name of Milestone A")
	assert.False(t, milestones[0].Project().NameIsLoaded(), "Project 1 should not have a loaded name")
	assert.True(t, milestones[0].Project().Manager().FirstNameIsLoaded(), "Person 7 has a name")
	assert.Equal(t, "Karen", milestones[0].Project().Manager().FirstName(), "Person 7 has first name of Karen")
}

func TestForwardMany(t *testing.T) {
	ctx := context.Background()
	milestones, err := goradd2.QueryMilestones(ctx).
		Select(node2.Milestone().Project().TeamMembers()).
		OrderBy(node2.Milestone().Project().TeamMembers().LastName(), node2.Milestone().Project().TeamMembers().FirstName()).
		Where(op.Equal(node2.Milestone().ID(), "1")). // Filter out people who are not managers
		Load()
	assert.NoError(t, err)
	names := []string{}
	for _, p := range milestones[0].Project().TeamMembers() {
		names = append(names, p.FirstName()+" "+p.LastName())
	}
	names2 := []string{
		"Samantha Jones",
		"Kendall Public",
		"Alex Smith",
		"Wendy Smith",
		"Karen Wolfe",
	}
	assert.Equal(t, names2, names)

}

func TestManyForward(t *testing.T) {
	ctx := context.Background()
	people, err := goradd2.QueryPeople(ctx).
		OrderBy(node2.Person().ID(), node2.Person().Projects().Name()).
		Select(node2.Person().Projects().Manager().FirstName(), node2.Person().Projects().Manager().LastName()).
		Load()
	assert.NoError(t, err)
	names := []string{}
	var p *goradd2.Project
	for _, p = range people[0].Projects() {
		names = append(names, p.Manager().FirstName()+" "+p.Manager().LastName())
	}
	names2 := []string{
		"Karen Wolfe",
		"John Doe",
	}
	assert.Equal(t, names2, names)

}

func Test2ndLoad(t *testing.T) {
	ctx := context.Background()
	projects, err := goradd2.QueryProjects(ctx).
		OrderBy(node2.Project().Manager().FirstName()).
		Load()
	assert.NoError(t, err)
	assert.Nil(t, projects[0].Manager())
	mgr, err2 := projects[0].LoadManager(ctx)
	assert.NoError(t, err2)
	assert.NotNil(t, mgr)
	assert.NotNil(t, projects[0].Manager(), "Manager object was added to project by LoadManager")
	assert.True(t, mgr.LastNameIsLoaded())
}

func TestAssociationCalculation(t *testing.T) {
	ctx := context.Background()
	projects, err := goradd2.QueryProjects(ctx).
		Calculation(node2.Project(), "count", op.Count(node2.Project().TeamMembers())).
		GroupBy(node2.Project()).
		OrderBy(node2.Project().ID()).
		Load()
	assert.NoError(t, err)
	assert.Equal(t, 5, projects[0].GetAlias("count").Int())
}

func TestAssociationByPrimaryKeys(t *testing.T) {
	ctx := context.Background()
	person := goradd2.NewPerson()
	person.SetID("100")
	person.SetFirstName("Fox")
	person.SetLastName("In Box")
	person.SetProjectsByID("1", "2", "3")
	assert.NoError(t, person.Save(ctx))

	person2, err := goradd2.LoadPerson(ctx, person.ID(), node2.Person().Projects())
	assert.NoError(t, err)
	assert.Len(t, person2.Projects(), 3)

	assert.NoError(t, person2.Delete(ctx))
	project, err2 := goradd2.LoadProject(ctx, "1")
	assert.NoError(t, err2)
	assert.NotNil(t, project)
}
