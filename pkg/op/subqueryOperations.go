package op

import . "github.com/goradd/gro/pkg/query"

func Subquery(b BuilderI) *SubqueryNode {
	return NewSubqueryNode(b)
}
