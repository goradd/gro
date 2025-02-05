package query

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAliasNodeInterfaces(t *testing.T) {
	n := AliasNode{"test"}

	assert.Equal(t, "test", n.Alias())
}
