package query

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestReverseReferenceNodeInterfaces(t *testing.T) {
	n := &ReverseNode{
		ForeignKey: "col",
		PrimaryKey: "pk",
		Identifier: "Obj",
	}

	assert.Implements(t, (*linker)(nil), n)
}
