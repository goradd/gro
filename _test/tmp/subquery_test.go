package tmp

import (
	"github.com/goradd/orm/_test/gen/orm/goradd"
	"github.com/goradd/orm/_test/gen/orm/goradd/node"
	"github.com/goradd/orm/pkg/db"
	"github.com/goradd/orm/pkg/op"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSubquery(t *testing.T) {
	ctx := db.NewContext(nil)
	people := goradd.QueryPeople(ctx).
		Alias("manager_count",
			goradd.QueryProjects(ctx).
				Alias("", op.Count(node.Project().ManagerID())).
				Where(op.Equal(node.Project().ManagerID(), node.Person().ID())).
				Subquery()).
		Where(op.Equal(node.Person().LastName(), "Wolfe")).
		Load()
	assert.Equal(t, 2, people[0].GetAlias("manager_count").Int(), "Karen Wolfe manages 2 projects.")
}

func TestSubquery2(t *testing.T) {
	ctx := db.NewContext(nil)
	people := goradd.QueryPeople(ctx).
		Alias("manager_count",
			goradd.QueryProjects(ctx).
				Alias("", op.Count(node.Project().ManagerID())).
				Where(op.Equal(node.Project().ManagerID(), node.Person().ID())).
				Subquery()).
		Where(op.Equal(node.Person().LastName(), "Wolfe")).
		Get()
	assert.Equal(t, 2, people.GetAlias("manager_count").Int(), "Karen Wolfe manages 2 projects.")
}

// TODO: Test multi-level subquery
