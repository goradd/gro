package query

import (
	"context"
	"github.com/goradd/gro/_test/gen/orm/goradd"
	"github.com/goradd/gro/_test/gen/orm/goradd/node"
	"github.com/goradd/gro/pkg/op"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
)

func TestManyMany(t *testing.T) {
	ctx := context.Background()
	projects, err := goradd.QueryProjects(ctx).
		Select(node.Project().TeamMembers()).
		OrderBy(node.Project().ID()).
		Load()

	assert.NoError(t, err)
	if len(projects[0].TeamMembers()) != 5 {
		t.Error("Did not find 5 team members in project 1. Found: " + strconv.Itoa(len(projects[0].TeamMembers())))
	}

}

func TestMany2(t *testing.T) {

	ctx := context.Background()

	// All People Who Are on a Project Managed by Karen Wolfe (Person Value #7)
	people, err := goradd.QueryPeople(ctx).
		OrderBy(node.Person().LastName(), node.Person().FirstName()).
		Where(op.Equal(node.Person().Projects().Manager().LastName(), "Wolfe")).
		Distinct().
		Select(node.Person().LastName(), node.Person().FirstName()).
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

	people, err := goradd.QueryPeople(ctx).
		OrderBy(node.Person().LastName(), node.Person().FirstName(), node.Person().Projects().Name()).
		Where(op.Equal(node.Person().Projects().Manager().LastName(), "Wolfe")).
		Select(node.Person().LastName(), node.Person().FirstName(), node.Person().Projects().Name()).
		Load()
	assert.NoError(t, err)
	person := people[0]
	projects := person.Projects()
	name := projects[0].Name()

	assert.Equal(t, "ACME Payment System", name)
}

func Test2Nodes(t *testing.T) {
	ctx := context.Background()
	milestones, err := goradd.QueryMilestones(ctx).
		Select(node.Milestone().Project().Manager()).
		Where(op.Equal(node.Milestone().ID(), 1)). // Filter out people who are not managers
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
	milestones, err := goradd.QueryMilestones(ctx).
		Select(node.Milestone().Project().TeamMembers()).
		OrderBy(node.Milestone().Project().TeamMembers().LastName(), node.Milestone().Project().TeamMembers().FirstName()).
		Where(op.Equal(node.Milestone().ID(), 1)). // Filter out people who are not managers
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
	people, err := goradd.QueryPeople(ctx).
		OrderBy(node.Person().ID(), node.Person().Projects().Name()).
		Select(node.Person().Projects().Manager().FirstName(), node.Person().Projects().Manager().LastName()).
		Load()
	assert.NoError(t, err)
	names := []string{}
	var p *goradd.Project
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
	projects, err := goradd.QueryProjects(ctx).
		OrderBy(node.Project().Manager().FirstName()).
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
	projects, err := goradd.QueryProjects(ctx).
		Calculation(node.Project(), "count", op.Count(node.Project().TeamMembers())).
		GroupBy(node.Project()).
		OrderBy(node.Project().ID()).
		Load()
	assert.NoError(t, err)
	assert.Equal(t, 5, projects[0].GetAlias("count").Int())
}

func TestAssociationByPrimaryKeys(t *testing.T) {
	ctx := context.Background()
	person := goradd.NewPerson()
	person.SetFirstName("Fox")
	person.SetLastName("In Box")
	person.SetProjectsByID("1", "2", "3")
	assert.NoError(t, person.Save(ctx))

	person2, err := goradd.LoadPerson(ctx, person.ID(), node.Person().Projects())
	assert.NoError(t, err)
	assert.Len(t, person2.Projects(), 3)

	assert.NoError(t, person2.Delete(ctx))
	project, err2 := goradd.LoadProject(ctx, "1")
	assert.NoError(t, err2)
	assert.NotNil(t, project)
}
