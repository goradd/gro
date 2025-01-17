package op

import . "github.com/goradd/orm/pkg/query"

// On some databases, these aggregate operations will only work if there is a GroupBy clause as well.

func Min(n Node) *OperationNode {
	return NewFunctionNode("MIN", n)
}

func Max(n Node) *OperationNode {
	return NewFunctionNode("MAX", n)
}

func Avg(n Node) *OperationNode {
	return NewFunctionNode("AVG", n)
}

func Sum(n Node) *OperationNode {
	return NewFunctionNode("SUM", n)
}

func Count(nodes ...Node) *OperationNode {
	return NewCountNode(nodes...)
}
