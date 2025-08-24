package op

import . "github.com/goradd/gro/pkg/query"

// Aggregates require a group by clause if aggregating over particular columns.

// Min is an aggregate function that will find the minimum value of the combinations of nodes in the result set.
// The result set is all the various combinations of the joined results before they are combined into slices of objects.
// Since this is an aggregate operation, it will result in only one row and find the minimum value over all the rows in the result set, unless
// you also add a GroupBy statement.
func Min(n Node) *OperationNode {
	return NewAggregateFunctionNode("MIN", n)
}

// Max is an aggregate function that will find the maximum value of the combinations of nodes in the result set.
// The result set is all the various combinations of the joined results before they are combined into slices of objects.
// Since this is an aggregate operation, it will result in only one row and find the maximum value over all the rows in the result set, unless
// you also add a GroupBy statement.
func Max(n Node) *OperationNode {
	return NewAggregateFunctionNode("MAX", n)
}

// Avg is an aggregate function that will find the average value of the combinations of nodes in the result set.
// The result set is all the various combinations of the joined results before they are combined into slices of objects.
// Since this is an aggregate operation, it will result in only one row and find the average value over all the rows in the result set, unless
// you also add a GroupBy statement.
func Avg(n Node) *OperationNode {
	return NewAggregateFunctionNode("AVG", n)
}

// Sum is an aggregate function that will find the sum of the combinations of nodes in the result set.
// The result set is all the various combinations of the joined results before they are combined into slices of objects.
// Since this is an aggregate operation, it will result in only one row and find the average value over all the rows in the result set, unless
// you also add a GroupBy statement.
func Sum(n Node) *OperationNode {
	return NewAggregateFunctionNode("SUM", n)
}

// Count is an aggregate function that will count the number of occurrences of the combinations of nodes in the result set.
// The result set is all the various combinations of the joined results before they are combined into slices of objects.
// Since this is an aggregate operation, it will result in only one row and count over all the rows in the result set, unless
// you also add a GroupBy statement.
func Count(nodes ...Node) *OperationNode {
	return NewCountNode(nodes...)
}
