package query

import (
	"testing"

	"github.com/goradd/orm/pkg/schema"
	"github.com/stretchr/testify/assert"
)

func TestColumnNodeInterfaces(t *testing.T) {
	n := &ColumnNode{QueryName: "dbName", Field: "goName", ReceiverType: ColTypeString, IsPrimaryKey: true}

	assert.Implements(t, (*ColumnNodeI)(nil), n)
}

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
