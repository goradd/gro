package op

import . "github.com/goradd/orm/pkg/query"

func Subquery(b QueryBuilderI) *SubqueryNode {
	return NewSubqueryNode(b)
}
