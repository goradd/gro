package query

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestReferenceNodeInterfaces(t *testing.T) {
	n := &ReferenceNode{
		ColumnQueryName: "dbCol",
		Identifier:      "Obj",
		ReceiverType:    ColTypeString,
	}

	assert.Implements(t, (*linker)(nil), n)
}
