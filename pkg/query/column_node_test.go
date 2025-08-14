package query

import (
	"testing"

	"github.com/goradd/orm/pkg/schema"
	"github.com/stretchr/testify/assert"
)

func TestColumnNodeCreator(t *testing.T) {
	n := NewColumnNode("dbName",
		"goName",
		ColTypeString,
		schema.ColTypeString,
		schema.ColSubTypeNone,
		true,
		nil)

	assert.Implements(t, (*ColumnNodeI)(nil), n)
}
