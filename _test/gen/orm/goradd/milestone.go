package goradd

// This is the implementation file for the Milestone ORM object.
// This is where you build the api to your data model for your web application and potentially mobile apps.
// Your edits to this file will be preserved.

import (
	"context"
	"encoding/gob"
	"fmt"
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

// String implements the Stringer interface and returns a description of the record, primarily for debugging.
func (o *Milestone) String() string {
	if o == nil {
		return ""
	}
	return fmt.Sprintf("Milestone %v", o.PrimaryKey())
}

// Key returns a unique key for the object, among a list of similar objects.
func (o *Milestone) Key() string {
	if o == nil {
		return ""
	}
	return fmt.Sprintf("%v", o.PrimaryKey())
}

// Label returns a human readable label of the object.
// This would be what a user would see as a description of the object if choosing from a list.
func (o *Milestone) Label() string {
	if o == nil {
		return ""
	}
	return o.Name()
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

// DeleteMilestone deletes the milestone record wtih primary key pk from the database.
// Note that you can also delete loaded Milestone objects by calling Delete on them.
// doc: type=Milestone
func DeleteMilestone(ctx context.Context, pk string) {
	deleteMilestone(ctx, pk)
}

func init() {
	gob.RegisterName("goraddMilestone", new(Milestone))
}
