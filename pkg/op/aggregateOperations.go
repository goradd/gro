package op

import . "github.com/goradd/orm/pkg/query"

// Aggregates require a group by clause if aggregating over particular columns.

func Min(n Node) *OperationNode {
	return NewAggregateFunctionNode("MIN", n)
}

func Max(n Node) *OperationNode {
	return NewAggregateFunctionNode("MAX", n)
}

func Avg(n Node) *OperationNode {
	return NewAggregateFunctionNode("AVG", n)
}

func Sum(n Node) *OperationNode {
	return NewAggregateFunctionNode("SUM", n)
}

func Count(nodes ...Node) *OperationNode {
	return NewCountNode(nodes...)
}
