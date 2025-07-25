package query

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestManyManyNodeInterfaces(t *testing.T) {
	n := &ManyManyNode{
		AssnTableQueryName: "table",
		ParentForeignKey:   "col1",
		ParentPrimaryKey:   "col2",
		Field:              "Field1",
		RefForeignKey:      "col2",
		RefPrimaryKey:      "col1",
	}

	assert.Implements(t, (*linker)(nil), n)
}
