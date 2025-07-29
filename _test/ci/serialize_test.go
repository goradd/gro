package ci

import (
	"bytes"
	"context"
	"encoding/gob"
	"github.com/goradd/orm/_test/gen/orm/goradd"
	"github.com/goradd/orm/_test/gen/orm/goradd/node"
	"github.com/goradd/orm/pkg/query"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
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

	var n query.Node = node.Project().Manager()
	proj, err := goradd.LoadProject(ctx, "1", n)
	assert.NoError(t, err)
	assert.Equal(t, proj.Manager().LastName(), "Wolfe")

	n2 := serNode(t, n)

	// can we still select a manager with the new node
	proj, err = goradd.LoadProject(ctx, "1", n2)
	assert.NoError(t, err)
	assert.Equal(t, proj.Manager().LastName(), "Wolfe")
}

func TestNodeSerializeReverseReference(t *testing.T) {
	ctx := context.Background()
	var n query.Node = node.Person().ManagerProjects()

	n2 := serNode(t, n)

	// can we still select a manager with the new node
	person, err := goradd.LoadPerson(ctx, "1", n2)
	assert.NoError(t, err)
	assert.Len(t, person.ManagerProjects(), 1)
	assert.Equal(t, "3", person.ManagerProjects()[0].ID())
}

func TestNodeSerializeManyMany(t *testing.T) {
	ctx := context.Background()
	var n query.Node = node.Person().Projects()

	n2 := serNode(t, n)

	// can we still select project as team member
	person, err := goradd.LoadPerson(ctx, "1", n2)
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
	person, err := goradd.LoadPerson(ctx, "7",
		node.Person().Projects(),        // many many
		node.Person().ManagerProjects(), // reverse
		node.Person().PersonType(),      // enum
		node.Person().Login(),           // reverse unique
	)
	assert.NoError(t, err)

	// Serialize and deserialize
	person2 := serObject(t, person).(*goradd.Person)
	assert.Len(t, person2.Projects(), 2)
	assert.Len(t, person2.ManagerProjects(), 2)
	assert.Equal(t, "kwolfe", person2.Login().Username())
}

func TestRecordSerializeComplex2(t *testing.T) {
	ctx := context.Background()
	login, err := goradd.LoadLogin(ctx, "4",
		node.Login().Person().Projects(),        // many many
		node.Login().Person().ManagerProjects(), // reverse
		node.Login().Person().Login(),           // reverse unique
	)
	assert.NoError(t, err)

	// Serialize and deserialize
	login2 := serObject(t, login).(*goradd.Login)
	assert.Len(t, login2.Person().Projects(), 2)
	assert.Len(t, login2.Person().ManagerProjects(), 2)
	assert.Equal(t, "kwolfe", login2.Person().Login().Username())
}
