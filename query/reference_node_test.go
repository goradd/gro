package query

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReferenceNodeInterfaces(t *testing.T) {
	n := &ReferenceNode{
		ForeignKey: "dbCol",
		PrimaryKey: "dbPk",
		Field:      "obj",
	}

	assert.Implements(t, (*linker)(nil), n)
}
