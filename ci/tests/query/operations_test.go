package query

import (
	"context"
	"testing"
	"time"

	goradd2 "github.com/goradd/gro/ci/tests/gen/goradd"
	node2 "github.com/goradd/gro/ci/tests/gen/goradd/node"
	"github.com/goradd/gro/query"
	"github.com/goradd/gro/query/op"
	"github.com/stretchr/testify/assert"
)

func TestEqualBasic(t *testing.T) {
	ctx := context.Background()
	projects, err := goradd2.QueryProjects(ctx).
		Where(op.Equal(node2.Project().Num(), 2)).
		OrderBy(node2.Project().Num()).
		Load()
	assert.NoError(t, err)
	assert.EqualValues(t, 2, projects[0].Num(), "Did not find correct project.")

}

func TestMultiWhere(t *testing.T) {
	ctx := context.Background()
	projects, err := goradd2.QueryPeople(ctx).
		Where(op.Equal(node2.Person().LastName(), "Smith")).
		Where(op.Equal(node2.Person().FirstName(), "Alex")).
		Load()
	assert.NoError(t, err)
	assert.Len(t, projects, 1)
}

func TestLogical(t *testing.T) {
	type testCase struct {
		testNode   query.Node
		objectNum  int
		expectedId interface{}
		count      int
		desc       string
	}
	tests := []testCase{
		{op.GreaterThan(node2.Project().Num(), 3), 0, 4, 1, "Greater than uint test"},
		{op.GreaterThan(node2.Project().StartDate(), time.Date(2006, 1, 1, 0, 0, 0, 0, time.UTC)), 0, 2, 2, "Greater than datetime test"},
		// SQLite does not have arbitrary precision number support
		//		{op.GreaterThan(node.Project().Spent(), 10000), 1, 2, 2, "Greater than float test"},
		{op.LessThan(node2.Project().Num(), 3), 1, 2, 2, "Less than uint test"},
		{op.LessThan(node2.Project().EndDate(), time.Date(2006, 1, 1, 0, 0, 0, 0, time.UTC)), 1, 4, 2, "Less than date test"},
		{op.IsNull(node2.Project().EndDate()), 0, 2, 1, "Is Null test"},
		{op.IsNotNull(node2.Project().EndDate()), 0, 1, 3, "Is Not Null test"},
		{op.GreaterOrEqual(node2.Project().Status(), 2), 1, 4, 2, "Greater or Equal test"},
		{op.LessOrEqual(node2.Project().StartDate(), time.Date(2006, 2, 15, 0, 0, 0, 0, time.UTC)), 2, 4, 3, "Less or equal date test"},
		{op.Or(op.Equal(node2.Project().Num(), 1), op.Equal(node2.Project().Num(), 4)), 1, 4, 2, "Or test"},
		{op.Xor(op.Equal(node2.Project().Num(), 3), op.Equal(node2.Project().Status(), 1)), 0, 2, 1, "Xor test"},
		{op.Not(op.Xor(op.Equal(node2.Project().Num(), 3), op.Equal(node2.Project().Status(), 1))), 0, 1, 3, "Not test"},
		{op.Like(node2.Project().Name(), "%ACME%"), 1, 4, 2, "Like test"},
		{op.In(node2.Project().Num(), 2, 3, 4), 1, 3, 3, "In test"},
	}

	ctx := context.Background()

	for i, c := range tests {
		t.Run(c.desc, func(t *testing.T) {
			projects, err := goradd2.QueryProjects(ctx).
				Where(c.testNode).
				OrderBy(node2.Project().Num()).
				Load()
			assert.NoError(t, err)
			if len(projects) <= c.objectNum {
				t.Errorf("Test case produced out of range error. Test case #: %d", i)
			} else {
				assert.EqualValues(t, c.expectedId, projects[c.objectNum].Num(), c.desc)
				assert.EqualValues(t, c.count, len(projects), c.desc+" - count")
			}
		})
	}
}

func TestCount2(t *testing.T) {
	ctx := context.Background()
	count, err := goradd2.QueryPeople(ctx).
		Select(node2.Person().LastName()).
		Distinct().
		Count()
	assert.NoError(t, err)
	assert.EqualValues(t, 10, count)

}

func TestCalculations(t *testing.T) {
	type testCase struct {
		testNode      *query.OperationNode
		objectNum     int
		expectedValue interface{}
		desc          string
	}
	var intTests = []testCase{
		{op.Multiply(node2.Project().Num(), 3), 3, 12, "Multiply test"},
		{op.Mod(node2.Project().Num(), 2), 2, 1, "Mod test"},
		{op.Round(op.Divide(node2.Project().Num(), 2)), 3, 2, "Mod test"},
	}

	var floatTests = []testCase{
		{op.Add(node2.Project().Spent(), node2.Project().Budget()), 0, 19811.00, "Add test"},
		{op.Subtract(node2.Project().Spent(), 2000), 0, 8250.75, "Subtract test"},
	}

	ctx := context.Background()

	for _, c := range intTests {
		projects, err := goradd2.QueryProjects(ctx).
			Calculation(node2.Project(), "Value", c.testNode).
			OrderBy(node2.Project().Num()).
			Load()
		assert.NoError(t, err)
		assert.EqualValues(t, c.expectedValue, projects[c.objectNum].GetAlias("Value").Int(), c.desc)
	}

	for _, c := range floatTests {
		projects, err := goradd2.QueryProjects(ctx).
			Calculation(node2.Project(), "Value", c.testNode).
			OrderBy(node2.Project().Num()).
			Load()
		assert.NoError(t, err)
		assert.EqualValues(t, c.expectedValue, projects[c.objectNum].GetAlias("Value").Float(), c.desc)
	}

}

func TestAggregates(t *testing.T) {
	ctx := context.Background()
	projects, err := goradd2.QueryProjects(ctx).
		Calculation(node2.Project(), "sum", op.Sum(node2.Project().Spent())).
		OrderBy(node2.Project().Status()).
		GroupBy(node2.Project().Status()).
		Load()
	assert.NoError(t, err)
	assert.EqualValues(t, 77400.5, projects[0].GetAlias("sum").Float())

	projects2, err2 := goradd2.QueryProjects(ctx).
		Calculation(node2.Project(), "min", op.Min(node2.Project().Spent())).
		OrderBy(node2.Project().Status()).
		GroupBy(node2.Project().Status()).
		Load()
	assert.NoError(t, err2)
	assert.EqualValues(t, 4200.50, projects2[0].GetAlias("min").Float())

	// test aggregate over all items
	projects3, err3 := goradd2.QueryProjects(ctx).
		Calculation(node2.Project(), "max", op.Max(node2.Project().Spent())).
		Load()
	assert.NoError(t, err3)
	assert.False(t, projects3[0].NameIsLoaded(), "aggregate functions should not select fields automatically")
	assert.EqualValues(t, 73200.0, projects3[0].GetAlias("max").Float())
}

/* TODO:

func TestAliases(t *testing.T) {
	ctx := context.Background()
	nVoyel := node.Person().ManagerProjects().Milestones()
	nVoyel.SetAlias("voyel")
	nConson := node.Person().ManagerProjects().Milestones()
	nConson.SetAlias("conson")

	people := goradd.QueryPeople(ctx).
		OrderBy(node.Person().LastName(), node.Person().FirstName()).
		Where(op.IsNotNull(nConson)).
		Join(nVoyel, op.In(nVoyel.Name(), "Milestone A", "Milestone E", "Milestone I")).
		Join(nConson, op.NotIn(nConson.Name(), "Milestone A", "Milestone E", "Milestone I")).
		GroupBy(node.Person().Value(), node.Person().FirstName(), node.Person().LastName()).
		Calculation("min_voyel", op.Min(nVoyel.Name())).
		Calculation("min_conson", op.Min(nConson.Name())).
		Load()

	assert.EqualValues(t, 3, len(people))
	assert.Equal(t, "Doe", people[0].LastName())
	assert.Equal(t, "Ho", people[1].LastName())
	assert.Equal(t, "Wolfe", people[2].LastName())

	assert.True(t, people[0].GetAlias("min_voyel").IsNil())
	assert.Equal(t, "Milestone F", people[0].GetAlias("min_conson").String())

	assert.Equal(t, "Milestone E", people[1].GetAlias("min_voyel").String())
	assert.Equal(t, "Milestone D", people[1].GetAlias("min_conson").String())

	assert.Equal(t, "Milestone A", people[2].GetAlias("min_voyel").String())
	assert.Equal(t, "Milestone B", people[2].GetAlias("min_conson").String())
}
*/
