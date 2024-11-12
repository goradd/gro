package op

import . "spekary/goradd/orm/pkg/query"

func Subquery(b QueryBuilderI) *SubqueryNode {
	return NewSubqueryNode(b)
}
