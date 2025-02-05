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

// String implements the Stringer interface and returns the default label for the object as it appears in html lists.
// Typically you would change this to whatever was pertinent to your application.
func (o *EmployeeInfo) String() string {
	if o == nil {
		return ""
	}
	return fmt.Sprintf("EmployeeInfo %v", o.PrimaryKey())
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

// DeleteEmployeeInfo deletes a employee_info record from the database given its primary key.
// Note that you can also delete loaded EmployeeInfo objects by calling Delete on them.
// doc: type=EmployeeInfo
func DeleteEmployeeInfo(ctx context.Context, pk string) {
	deleteEmployeeInfo(ctx, pk)
}

func init() {
	gob.RegisterName("goraddEmployeeInfo", new(EmployeeInfo))
}
