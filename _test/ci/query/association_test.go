package query

import (
	"context"
	"github.com/goradd/orm/_test/gen/orm/goradd"
	"github.com/goradd/orm/_test/gen/orm/goradd/node"
	"github.com/goradd/orm/pkg/op"
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

// TODO:
/*
func TestConditionalJoin(t *testing.T) {
	ctx := context.Background()

	projects := goradd.QueryProjects(ctx).
		OrderBy(node.Project().Name()).
		Select(node.Project().Manager(), op.Equal(node.Project().Manager().LastName(), "Wolfe")).
		Select(node.Project().TeamMembers(), op.Equal(node.Project().TeamMembers().LastName(), "Smith")).
		Load()

	// Reverse references
	people := goradd.QueryPeople(ctx).
		Select(node.Person().Addresses(), op.Equal(node.Person().Addresses().City(), "New York")).
		Select(node.Person().ManagerProjects(), op.Equal(node.Person().ManagerProjects().Status(), goradd.ProjectStatusOpen)).
		Select(node.Person().ManagerProjects().Milestones()).
		Select(node.Person().Login(), op.Like(node.Person().Login().Username(), "b%")).
		OrderBy(node.Person().LastName(), node.Person().FirstName(), node.Person().ManagerProjects().Name()).
		Load()

	assert.Equal(t, "John", people[2].FirstName(), "John Doe is the 3rd Person.")
	assert.Len(t, people[2].ManagerProjects(), 1, "John Doe manages 1 Project.")
	assert.Len(t, people[2].ManagerProjects()[0].Milestones(), 1, "John Doe has 1 Milestone")

	// Groups that are not expanded by the conditional join are still created as empty arrays. NoSql databases will need to do this too.
	// This makes it a little easier to write code that uses it, becuase you don't have to test for nil
	assert.Len(t, people[0].ManagerProjects(), 0)

	// Check parallel reverse reference with condition
	assert.Len(t, people[7].Addresses(), 2, "Ben Robinson has 2 Addresses")
	assert.Len(t, people[2].Addresses(), 0, "John Doe has no Addresses")

	// Reverse reference unique
	assert.Equal(t, "brobinson", people[7].Login().Username(), "Ben Robinson's Login was selected")
	assert.Nil(t, people[2].Login(), "John Doe's Login was not selected")

	// Forward reference
	assert.Nil(t, projects[2].Manager(), "")
	assert.Equal(t, projects[0].Manager().FirstName(), "Karen")

	// Many-many
	assert.Len(t, projects[3].TeamMembers(), 2, "Project 4 has 2 team members with last name Smith")
	assert.Equal(t, "Smith", projects[3].TeamMembers()[0].LastName(), "The first team member from project 4 has a last name of smith")
}
*/

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
