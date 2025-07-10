package ci

import (
	"context"
	"github.com/goradd/orm/_test/gen/orm/goradd"
	"github.com/goradd/orm/_test/gen/orm/goradd/node"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBasicEnum(t *testing.T) {
	ctx := context.Background()
	projects, err := goradd.QueryProjects(ctx).
		OrderBy(node.Project().ID()).
		Load()
	assert.NoError(t, err)
	if projects[0].Status() != goradd.ProjectStatusCompleted {
		t.Error("Did not find correct project type.")
	}
}
