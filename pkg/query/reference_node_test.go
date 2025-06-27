package query

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestReferenceNodeInterfaces(t *testing.T) {
	n := &ReferenceNode{
		ForeignKey: "dbCol",
		PrimaryKey: "dbPk",
		Identifier: "Obj",
	}

	assert.Implements(t, (*linker)(nil), n)
}
