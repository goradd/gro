package goradd

// This is the implementation file for the Project ORM object.
// This is where you build the api to your data model for your web application and potentially mobile apps.
// Your edits to this file will be preserved.

import (
	"context"
	"encoding/gob"
)

// Project represents an item in the project table in the database.
type Project struct {
	projectBase
}

// NewProject creates a new Project object and initializes it to default values.
func NewProject() *Project {
	o := new(Project)
	o.Initialize()
	return o
}

// Initialize will initialize or re-initialize a Project database object to default values.
func (o *Project) Initialize() {
	o.projectBase.Initialize()
	// Add your own initializations here
}

// String implements the Stringer interface and returns the default label for the object as it appears in html lists.
// Typically you would change this to whatever was pertinent to your application.
func (o *Project) String() string {
	if o == nil {
		return "" // Possibly - Select One -?
	}
	return o.name
}

// QueryProjects returns a new query builder.
func QueryProjects(ctx context.Context) ProjectBuilder {
	return queryProjects(ctx)
}

// queryProjects creates a new builder and is the central spot where all queries are directed.
// You can modify this function to enforce restrictions on queries, for example to make sure the user is authorized to
// access the data.
func queryProjects(ctx context.Context) ProjectBuilder {
	return newProjectBuilder(ctx)
}

// DeleteProject deletes a project record from the database given its primary key.
// Note that you can also delete loaded Project objects by calling Delete on them.
// doc: type=Project
func DeleteProject(ctx context.Context, pk string) {
	deleteProject(ctx, pk)
}

func init() {
	gob.RegisterName("goraddProject", new(Project))
}
