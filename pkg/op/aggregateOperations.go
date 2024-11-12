package op

import . "spekary/goradd/orm/pkg/query"

// On some databases, these aggregate operations will only work if there is a GroupBy clause as well.

func Min(n NodeI) *OperationNode {
	return NewFunctionNode("MIN", n)
}

func Max(n NodeI) *OperationNode {
	return NewFunctionNode("MAX", n)
}

func Avg(n NodeI) *OperationNode {
	return NewFunctionNode("AVG", n)
}

func Sum(n NodeI) *OperationNode {
	return NewFunctionNode("SUM", n)
}

func Count(nodes ...NodeI) *OperationNode {
	return NewCountNode(nodes...)
}
