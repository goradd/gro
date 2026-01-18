package query

import (
	"context"
	"encoding/json"
	"testing"

	goradd2 "github.com/goradd/gro/ci/tests/gen/goradd"
	node3 "github.com/goradd/gro/ci/tests/gen/goradd/node"
	goradd_unit2 "github.com/goradd/gro/ci/tests/gen/goradd_unit"
	node2 "github.com/goradd/gro/ci/tests/gen/goradd_unit/node"
	"github.com/goradd/gro/query"
	"github.com/goradd/gro/query/op"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBasic(t *testing.T) {
	ctx := context.Background()

	people, err := goradd2.QueryPeople(ctx).
		OrderBy(node3.Person().ID()).
		Load()
	assert.NoError(t, err)
	assert.Len(t, people, 12)
	assert.Equal(t, "John", people[0].FirstName())
}

func TestLoad(t *testing.T) {
	ctx := context.Background()

	people, err := goradd2.QueryPeople(ctx).
		Where(op.Equal(node3.Person().ID(), "7")).
		Select(
			node3.Person().Projects(),        // many many
			node3.Person().ManagerProjects(), // reverse
			node3.Person().Login(),
		).
		Load()
	assert.NoError(t, err)
	person := people[0]
	assert.Equal(t, "Wolfe", person.LastName(), "Last name was not selected by default")
	assert.Len(t, person.Projects(), 2)
	assert.Len(t, person.ManagerProjects(), 2)
	assert.Equal(t, "kwolfe", person.Login().Username())

	person, err = goradd2.LoadPerson(ctx, "7",
		node3.Person().Projects(),        // many many
		node3.Person().ManagerProjects(), // reverse
		node3.Person().Login(),           // reverse unique
	)
	assert.NoError(t, err)
	// Serialize and deserialize
	assert.Len(t, person.Projects(), 2)
	assert.Len(t, person.ManagerProjects(), 2)
	assert.Equal(t, "kwolfe", person.Login().Username())
}

func TestSort(t *testing.T) {
	ctx := context.Background()
	people, err := goradd2.QueryPeople(ctx).
		OrderBy(node3.Person().LastName()).
		Load()
	assert.NoError(t, err)
	if people[0].LastName() != "Brady" {
		t.Error("Person found not Brady, found " + people[0].LastName())
	}

	people, err = goradd2.QueryPeople(ctx).
		OrderBy(node3.Person().FirstName()).
		Load()
	assert.NoError(t, err)
	if people[0].FirstName() != "Alex" {
		t.Error("Person found not Alex, found " + people[0].FirstName())
	}

	// Testing for regression bug with multiple sorts
	people, err = goradd2.QueryPeople(ctx).
		OrderBy(node3.Person().LastName().Descending(), node3.Person().FirstName().Ascending()).
		Load()
	assert.NoError(t, err)
	if people[0].FirstName() != "Karen" || people[0].LastName() != "Wolfe" {
		t.Error("Person found not Karen Wolfe, found " + people[0].FirstName() + " " + people[0].LastName())
	}
}

func TestWhere(t *testing.T) {
	ctx := context.Background()
	people, err := goradd2.QueryPeople(ctx).
		Where(op.Equal(node3.Person().LastName(), "Smith")).
		OrderBy(node3.Person().FirstName().Descending(), node3.Person().LastName()).
		Load()
	assert.NoError(t, err)
	if people[0].FirstName() != "Wendy" {
		t.Error("Person found not Wendy, found " + people[0].FirstName())
	}
}

func TestAlias(t *testing.T) {
	ctx := context.Background()
	projects, err := goradd2.QueryProjects(ctx).
		Where(op.Equal(node3.Project().ID(), "1")).
		Calculation(node3.Project(), "Difference", op.Subtract(node3.Project().Budget(), node3.Project().Spent())).
		Load()
	assert.NoError(t, err)
	v := projects[0].GetAlias("Difference").Float()
	assert.EqualValues(t, -690.5, v)
}

func TestCursor(t *testing.T) {
	ctx := context.Background()
	projectCursor, err := goradd2.QueryProjects(ctx).
		LoadCursor()
	assert.NoError(t, err)
	var projects []*goradd2.Project
	for {
		project, err := projectCursor.Next()
		assert.NoError(t, err)
		if project == nil {
			break
		}
		projects = append(projects, project)
	}

	assert.Len(t, projects, 4)
}

/*
	func TestAlias2(t *testing.T) {
		ctx := context.Background()
		projects := goradd.QueryProjects(ctx).
			Alias("a", node.Project().Num()).
			Alias("b", node.Project().Name()).
			Alias("c", node.Project().Spent()).
			Alias("d", node.Project().StartDate()).
			Alias("e", op.IsNull(node.Project().EndDate())).
			OrderBy(node.Project().Value()).
			Load()

		project := projects[0]
		assert.Equal(t, 1, project.GetAlias("a").Int())
		assert.Equal(t, 1, project.Num())
		assert.Equal(t, "ACME Website Redesign", project.GetAlias("b").String())
		assert.Equal(t, 10250.75, project.GetAlias("c").Float())
		d := time.FromSqlDateTime("2004-03-01")
		assert.EqualValues(t, d, project.GetAlias("d").Time())
		//assert.EqualValues(t, d, project.StartDate())
		assert.Equal(t, false, project.GetAlias("e").Bool())
	}
*/
func TestCount(t *testing.T) {
	ctx := context.Background()

	count, err := goradd2.QueryProjects(ctx).
		Count()
	assert.NoError(t, err)
	assert.EqualValues(t, 4, count)
}

func TestGroupBy(t *testing.T) {
	ctx := context.Background()

	projects, err := goradd2.QueryProjects(ctx).
		Calculation(node3.Project(), "teamMemberCount", op.Count(node3.Project().TeamMembers())).
		GroupBy(node3.Project()).
		OrderBy(node3.Project().ID()).
		Load()
	assert.NoError(t, err)
	assert.EqualValues(t, 5, projects[0].GetAlias("teamMemberCount").Int())
}

func TestSelect(t *testing.T) {
	ctx := context.Background()

	projects, err := goradd2.QueryProjects(ctx).
		Select(node3.Project().Name()).
		Load()
	assert.NoError(t, err)
	project := projects[0]
	assert.True(t, project.NameIsLoaded())
	assert.False(t, project.ManagerIDIsLoaded())
	assert.True(t, project.IDIsLoaded())
}

func TestLimit(t *testing.T) {
	ctx := context.Background()

	people, err := goradd2.QueryPeople(ctx).
		OrderBy(node3.Person().ID()).
		Limit(2, 3).
		Load()
	assert.NoError(t, err)
	assert.EqualValues(t, "Jacob", people[0].FirstName())
	assert.Len(t, people, 2)
}

func TestSingleEmpty(t *testing.T) {
	ctx := context.Background()

	people, err := goradd2.QueryPeople(ctx).
		Where(op.Equal(node3.Person().ID(), "12345")).
		Load()
	assert.NoError(t, err)
	assert.Len(t, people, 0)

}

func TestLazyLoad(t *testing.T) {
	ctx := context.Background()

	projects, err := goradd2.QueryProjects(ctx).
		Where(op.Equal(node3.Project().ID(), "1")).
		Load()
	assert.NoError(t, err)
	var mId string = projects[0].ID() // foreign keys are treated as strings for cross-database compatibility
	assert.Equal(t, "1", mId)

	manager, err2 := projects[0].LoadManager(ctx)
	assert.NoError(t, err2)
	assert.Equal(t, "7", manager.ID())
}

func TestHaving(t *testing.T) {
	// This particular test shows a quirk of SQL that requires:
	// 1) If you have an aggregate clause (like COUNT), you MUST have a GROUPBY clause, and
	// 2) If you have a GROUPBY, you MUST SELECT and only select the things you are grouping by.
	//
	// Sooo, when we see a GroupBy, we automatically also select the same nodes.
	ctx := context.Background()

	projects, err := goradd2.QueryProjects(ctx).
		GroupBy(node3.Project().ID(), node3.Project().Name()).
		OrderBy(node3.Project().ID()).
		Calculation(node3.Project(), "team_member_count", op.Count(node3.Project().TeamMembers())).
		Having(op.GreaterThan(query.Alias("team_member_count"), 5)).
		Load()
	assert.NoError(t, err)
	assert.Len(t, projects, 2)
	assert.Equal(t, "State College HR System", projects[0].Name())
	assert.Equal(t, 6, projects[0].GetAlias("team_member_count").Int())
}

func TestFailedSelects(t *testing.T) {
	ctx := context.Background()

	assert.Panics(t, func() { goradd2.QueryProjects(ctx).Select(node3.Person()) })
}

func TestFailedGroupBy(t *testing.T) {
	ctx := context.Background()

	assert.Panics(t, func() {
		goradd2.QueryProjects(ctx).
			GroupBy(node3.Project().Name()).
			Select(node3.Project().Name())
	})
}

// Test that we can get from an integer keyed database
func TestIntKey(t *testing.T) {
	ctx := context.Background()

	g, err := goradd2.LoadGift(ctx, 2)
	assert.NoError(t, err)
	assert.Equal(t, "Turtle doves", g.Name())
}

func TestMultiParent(t *testing.T) {
	ctx := context.Background()

	baby := goradd_unit2.NewMultiParent()
	baby.SetName("Baby")
	mom := goradd_unit2.NewMultiParent()
	mom.SetName("Mom")
	dad := goradd_unit2.NewMultiParent()
	dad.SetName("Dad")

	baby.SetParent1(mom)
	baby.SetParent2(dad)
	baby.Save(ctx)
	defer baby.Delete(ctx)
	defer mom.Delete(ctx)
	defer dad.Delete(ctx)

	mom2, err := goradd_unit2.LoadMultiParent(ctx, mom.ID(), node2.MultiParent().Parent1MultiParents())
	assert.NoError(t, err)
	require.NotNil(t, mom2, "Multi parent should exist %s", mom.ID())
	assert.Equal(t, mom2.Parent1MultiParents()[0].ID(), baby.ID())

	b, errw := json.Marshal(baby)
	assert.NoError(t, errw)

	baby2 := goradd_unit2.NewMultiParent()
	err = baby2.UnmarshalJSON(b)
	assert.NoError(t, err)
	assert.Equal(t, baby.ID(), baby2.ID())
	assert.Equal(t, baby.Parent1().ID(), baby2.Parent1().ID())
	assert.Equal(t, baby.Parent2().ID(), baby2.Parent2().ID())
}

func TestWriteTimeout(t *testing.T) {
	ctx := context.Background()

	obj := goradd_unit2.NewTimeoutTest()
	obj.SetName("test")
	assert.Error(t, obj.Save(ctx))
}
