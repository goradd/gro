package ci

import (
	"bytes"
	"encoding/gob"
	"github.com/goradd/orm/_test/gen/orm/goradd/node"
	"github.com/goradd/orm/pkg/query"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNodeSerialize(t *testing.T) {
	var n query.Node = node.Person().FirstName()

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	err := enc.Encode(&n)
	assert.NoError(t, err)

	var n2 query.Node
	dec := gob.NewDecoder(&buf)
	err = dec.Decode(&n2)
	assert.NoError(t, err)
}
