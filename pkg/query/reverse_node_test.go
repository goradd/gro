package query

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReverseReferenceNodeInterfaces(t *testing.T) {
	n := &ReverseNode{
		ForeignKey: "col",
		PrimaryKey: "pk",
		Field:      "obj",
	}

	assert.Implements(t, (*linker)(nil), n)
}
