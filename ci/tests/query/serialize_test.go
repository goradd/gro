package query

import (
	"bytes"
	"context"
	"encoding/gob"
	"testing"

	"github.com/goradd/gro/ci/tests/gen/goradd"
	goradd2 "github.com/goradd/gro/ci/tests/gen/goradd"
	node2 "github.com/goradd/gro/ci/tests/gen/goradd/node"
	"github.com/goradd/gro/query"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// serialize and deserialize the node
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

func TestNodeSerializeReference(t *testing.T) {
	ctx := context.Background()

	var n query.Node = node2.Project().Manager()
	proj, err := goradd2.LoadProject(ctx, "1", n)
	assert.NoError(t, err)
	assert.Equal(t, proj.Manager().LastName(), "Wolfe")

	n2 := serNode(t, n)

	// can we still select a manager with the new node
	proj, err = goradd2.LoadProject(ctx, "1", n2)
	assert.NoError(t, err)
	assert.Equal(t, proj.Manager().LastName(), "Wolfe")
}

func TestNodeSerializeReverseReference(t *testing.T) {
	ctx := context.Background()
	var n query.Node = node2.Person().ManagerProjects()

	n2 := serNode(t, n)

	// can we still select a manager with the new node
	person, err := goradd2.LoadPerson(ctx, "1", n2)
	assert.NoError(t, err)
	assert.Len(t, person.ManagerProjects(), 1)
	assert.Equal(t, "3", person.ManagerProjects()[0].ID())
}

func TestNodeSerializeManyMany(t *testing.T) {
	ctx := context.Background()
	var n query.Node = node2.Person().Projects()

	n2 := serNode(t, n)

	// can we still select project as team member
	person, err := goradd2.LoadPerson(ctx, "1", n2)
	assert.NoError(t, err)
	assert.Len(t, person.Projects(), 2)
}

func serObject(t *testing.T, n interface{}) interface{} {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	err := enc.Encode(&n)
	require.NoError(t, err)

	var n2 interface{}
	dec := gob.NewDecoder(&buf)
	err = dec.Decode(&n2)
	require.NoError(t, err)
	return n2
}

func TestRecordSerializeComplex1(t *testing.T) {
	ctx := context.Background()
	person, err := goradd2.LoadPerson(ctx, "7",
		node2.Person().Projects(),        // many many
		node2.Person().ManagerProjects(), // reverse
		node2.Person().PersonType(),      // enum
		node2.Person().Login(),           // reverse unique
	)
	assert.NoError(t, err)

	// Serialize and deserialize
	person2 := serObject(t, person).(*goradd2.Person)
	assert.Len(t, person2.Projects(), 2)
	assert.Len(t, person2.ManagerProjects(), 2)
	assert.Equal(t, "kwolfe", person2.Login().Username())
}

func TestRecordSerializeComplex2(t *testing.T) {
	ctx := context.Background()
	login, err := goradd2.LoadLogin(ctx, "4",
		node2.Login().Person().Projects(),        // many many
		node2.Login().Person().ManagerProjects(), // reverse
		node2.Login().Person().Login(),           // reverse unique
	)
	assert.NoError(t, err)

	// Serialize and deserialize
	login2 := serObject(t, login).(*goradd.Login)
	assert.Len(t, login2.Person().Projects(), 2)
	assert.Len(t, login2.Person().ManagerProjects(), 2)
	assert.Equal(t, "kwolfe", login2.Person().Login().Username())
}
