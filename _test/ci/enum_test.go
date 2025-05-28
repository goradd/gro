package ci

import (
	"github.com/goradd/orm/_test/gen/orm/goradd"
	"github.com/goradd/orm/_test/gen/orm/goradd/node"
	"github.com/goradd/orm/pkg/db"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBasicEnum(t *testing.T) {
	ctx := db.NewContext(nil)
	projects, err := goradd.QueryProjects(ctx).
		OrderBy(node.Project().ID()).
		Load()
	assert.NoError(t, err)
	if projects[0].Status() != goradd.ProjectStatusCompleted {
		t.Error("Did not find correct project type.")
	}
}
