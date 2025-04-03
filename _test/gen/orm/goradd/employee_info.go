package goradd

// This is the implementation file for the EmployeeInfo ORM object.
// This is where you build the api to your data model for your web application and potentially mobile apps.
// Your edits to this file will be preserved.

import (
	"context"
	"encoding/gob"
	"fmt"
)

// EmployeeInfo represents an item in the employee_info table in the database.
type EmployeeInfo struct {
	employeeInfoBase
}

// NewEmployeeInfo creates a new EmployeeInfo object and initializes it to default values.
func NewEmployeeInfo() *EmployeeInfo {
	o := new(EmployeeInfo)
	o.Initialize()
	return o
}

// Initialize will initialize or re-initialize a EmployeeInfo database object to default values.
func (o *EmployeeInfo) Initialize() {
	o.employeeInfoBase.Initialize()
	// Add your own initializations here
}

// String implements the Stringer interface and returns a description of the record, primarily for debugging.
func (o *EmployeeInfo) String() string {
	if o == nil {
		return ""
	}
	return fmt.Sprintf("EmployeeInfo %v", o.PrimaryKey())
}

// Key returns a unique key for the object, among a list of similar objects.
func (o *EmployeeInfo) Key() string {
	if o == nil {
		return ""
	}
	return fmt.Sprintf("%v", o.PrimaryKey())
}

// Label returns a human-readable label of the object.
// This would be what a user would see as a description of the object if choosing from a list.
func (o *EmployeeInfo) Label() string {
	if o == nil {
		return ""
	}
	return fmt.Sprintf("Employee Info %v", o.PrimaryKey())
}

// Save will update or insert the object, depending on the state of the object.
// If it has an auto-generated primary key, it will be changed after an insert.
// Database errors generally will be handled by a panic and not returned here,
// since those indicate a problem with a database driver or configuration.
//
// Save will return a db.OptimisticLockError if it detects a collision when two users
// are attempting to change the same database record.
//
// It will return a db.NewDuplicateValueError if it detects a collision when an attempt
// is made to add a record with a unique column that is given a value that is already in the database.
//
// Updating a record that has not changed will have no effect on the database.
func (o *EmployeeInfo) Save(ctx context.Context) error {
	return o.save(ctx)
}

// QueryEmployeeInfos returns a new query builder.
func QueryEmployeeInfos(ctx context.Context) EmployeeInfoBuilder {
	return queryEmployeeInfos(ctx)
}

// queryEmployeeInfos creates a new builder and is the central spot where all queries are directed.
// You can modify this function to enforce restrictions on queries, for example to make sure the user is authorized to
// access the data.
func queryEmployeeInfos(ctx context.Context) EmployeeInfoBuilder {
	return newEmployeeInfoBuilder(ctx)
}

// DeleteEmployeeInfo deletes the employee_info record with primary key pk from the database.
// Note that you can also delete loaded EmployeeInfo objects by calling Delete on them.
// doc: type=EmployeeInfo
func DeleteEmployeeInfo(ctx context.Context, pk string) {
	deleteEmployeeInfo(ctx, pk)
}

func init() {
	gob.RegisterName("goraddEmployeeInfo", new(EmployeeInfo))
}
