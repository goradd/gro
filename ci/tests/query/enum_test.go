package query

import (
	"context"
	"testing"

	"github.com/goradd/gro/ci/tests/gen/goradd"
	goradd2 "github.com/goradd/gro/ci/tests/gen/goradd"
	"github.com/goradd/gro/ci/tests/gen/goradd/node"
	"github.com/stretchr/testify/assert"
)

func TestBasicEnum(t *testing.T) {
	ctx := context.Background()
	projects, err := goradd2.QueryProjects(ctx).
		OrderBy(node.Project().ID()).
		Load()
	assert.NoError(t, err)
	if projects[0].Status() != goradd.ProjectStatusCompleted {
		t.Error("Did not find correct project type.")
	}
}
