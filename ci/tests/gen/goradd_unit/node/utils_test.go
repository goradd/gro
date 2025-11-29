package node

import (
	"bytes"
	"encoding/gob"
	"testing"

	"github.com/goradd/gro/query"
	"github.com/stretchr/testify/assert"
)

func serNode(t *testing.T, n query.Node) query.Node {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	err := enc.Encode(&n)
	assert.NoError(t, err)

	var n2 query.Node
	dec := gob.NewDecoder(&buf)
	err = dec.Decode(&n2)
	assert.NoError(t, err)
	return n2
}
