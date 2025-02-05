package goradd

// This is the implementation file for the Milestone ORM object.
// This is where you build the api to your data model for your web application and potentially mobile apps.
// Your edits to this file will be preserved.

import (
	"context"
	"encoding/gob"
)

// Milestone represents an item in the milestone table in the database.
type Milestone struct {
	milestoneBase
}

// NewMilestone creates a new Milestone object and initializes it to default values.
func NewMilestone() *Milestone {
	o := new(Milestone)
	o.Initialize()
	return o
}

// Initialize will initialize or re-initialize a Milestone database object to default values.
func (o *Milestone) Initialize() {
	o.milestoneBase.Initialize()
	// Add your own initializations here
}

// String implements the Stringer interface and returns the default label for the object as it appears in html lists.
// Typically you would change this to whatever was pertinent to your application.
func (o *Milestone) String() string {
	if o == nil {
		return ""
	}
	return o.name
}

// QueryMilestones returns a new query builder.
func QueryMilestones(ctx context.Context) MilestoneBuilder {
	return queryMilestones(ctx)
}

// queryMilestones creates a new builder and is the central spot where all queries are directed.
// You can modify this function to enforce restrictions on queries, for example to make sure the user is authorized to
// access the data.
func queryMilestones(ctx context.Context) MilestoneBuilder {
	return newMilestoneBuilder(ctx)
}

// DeleteMilestone deletes a milestone record from the database given its primary key.
// Note that you can also delete loaded Milestone objects by calling Delete on them.
// doc: type=Milestone
func DeleteMilestone(ctx context.Context, pk string) {
	deleteMilestone(ctx, pk)
}

func init() {
	gob.RegisterName("goraddMilestone", new(Milestone))
}
