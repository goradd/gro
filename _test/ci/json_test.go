package ci

import (
	"context"
	"encoding/json"
	"github.com/goradd/orm/_test/gen/orm/goradd"
	"github.com/goradd/orm/_test/gen/orm/goradd/node"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestJsonMarshall1(t *testing.T) {
	ctx := context.Background()

	p, err := goradd.LoadProject(ctx, "1",
		node.Project().Name(),
		node.Project().Status(),
		node.Project().Manager().FirstName())
	assert.NoError(t, err)
	j, err2 := json.Marshal(p)
	assert.NoError(t, err2)
	m := make(map[string]interface{})
	err = json.Unmarshal(j, &m)
	assert.NoError(t, err)
	assert.Exactly(t, "ACME Website Redesign", m["name"])
	assert.Exactly(t, "Karen", m["manager"].(map[string]interface{})["firstName"])
}

func TestJsonUnmarshall1(t *testing.T) {
	p := goradd.NewProject()
	err := json.Unmarshal([]byte(
		`{
	"name":"ACME Website Redesign",
	"status":3,
	"num":14,
	"startDate":"2020-11-01T00:00:00Z"
}
`),
		&p)
	assert.NoError(t, err)
	assert.Exactly(t, "ACME Website Redesign", p.Name())
	assert.Exactly(t, goradd.ProjectStatusCompleted, p.Status())
	assert.Exactly(t, 14, p.Num())
	assert.Exactly(t, 2020, p.StartDate().Year())
}

func TestJsonMarshall2(t *testing.T) {
	ctx := context.Background()

	p, err := goradd.LoadPerson(ctx, "1",
		node.Person().FirstName(),
		node.Person().LastName(),
		node.Person().PersonType())
	assert.NoError(t, err)
	j, err2 := json.Marshal(p)
	assert.NoError(t, err2)
	m := make(map[string]interface{})
	err = json.Unmarshal(j, &m)
	assert.NoError(t, err)
	assert.Equal(t, "John", m["firstName"])
	assert.Equal(t, "Doe", m["lastName"])
	assert.Equal(t, "inactive", m["personType"])
}

func TestJsonUnmarshall2(t *testing.T) {
	p := goradd.NewPerson()
	err := json.Unmarshal([]byte(
		`{
	"firstName":"John",
	"lastName":"Doe",
	"personType":"inactive"
}
`),
		&p)
	assert.NoError(t, err)
	assert.Equal(t, "John", p.FirstName())
	assert.Equal(t, "Doe", p.LastName())
	assert.Equal(t, goradd.PersonTypeInactive, p.PersonType())
}

func TestJsonMarshallReferences(t *testing.T) {
	ctx := context.Background()
	project, err := goradd.LoadProject(ctx, "1", node.Project().Manager())
	assert.NoError(t, err)

	b, err2 := json.Marshal(project)
	assert.NoError(t, err2)

	project2 := goradd.NewProject()
	err = project2.UnmarshalJSON(b)
	assert.NoError(t, err)
	assert.Equal(t, project.ID(), project2.ID())
	assert.Equal(t, project.Manager().ID(), project2.Manager().ID())
}

func TestJsonMarshallReverse(t *testing.T) {
	ctx := context.Background()
	person, err := goradd.LoadPerson(ctx, "7", node.Person().ManagerProjects(), node.Person().EmployeeInfo())
	assert.NoError(t, err)
	b, err2 := json.Marshal(person)
	assert.NoError(t, err2)

	person2 := goradd.NewPerson()
	err = person2.UnmarshalJSON(b)
	assert.NoError(t, err)
	assert.Equal(t, person.ID(), person2.ID())
	assert.Equal(t, len(person.ManagerProjects()), len(person2.ManagerProjects()))
	assert.Equal(t, person.ManagerProjects()[0].ID(), person2.ManagerProjects()[0].ID())
	assert.Equal(t, 123, person.EmployeeInfo().EmployeeNumber())
}

func TestJsonMarshallManyManyReferences(t *testing.T) {
	ctx := context.Background()
	project, err := goradd.LoadProject(ctx, "1", node.Project().TeamMembers())

	b, err := json.Marshal(project)
	assert.NoError(t, err)

	project2 := goradd.NewProject()
	err = project2.UnmarshalJSON(b)
	assert.NoError(t, err)
	assert.Equal(t, project.ID(), project2.ID())
	assert.Equal(t, len(project.TeamMembers()), len(project2.TeamMembers()))
	assert.Equal(t, project.TeamMembers()[0].ID(), project2.TeamMembers()[0].ID())
}
