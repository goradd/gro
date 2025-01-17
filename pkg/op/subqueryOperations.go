package op

import . "github.com/goradd/orm/pkg/query"

func Subquery(b BuilderI) *SubqueryNode {
	return NewSubqueryNode(b)
}
