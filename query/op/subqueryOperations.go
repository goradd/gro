package op

import . "github.com/goradd/gro/query"

func Subquery(b BuilderI) *SubqueryNode {
	return NewSubqueryNode(b)
}
