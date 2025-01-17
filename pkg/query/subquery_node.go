package query

import (
	"bytes"
	"encoding/gob"
	"log/slog"
	"strings"
)

type SubqueryCommand int

const (
	SubqueryCommandDefault SubqueryCommand = iota
	SubqueryCommandLoad
	SubqueryCommandGet
	SubqueryCommandCount
)

// A SubqueryNode represents a "select" subquery. Subqueries are not always portable to other databases, and are not
// easily checked for syntax errors, since a subquery can return a scalar, vector, or even an entire table.
// You generally do not create a subquery node directly, but rather you use the codegenerated models to start a
// query on a table, and then end the query with "Subquery()" which will turn the query into a usable subquery node
// that you can embed in other queries.
type SubqueryNode struct {
	nodeAlias
	b   BuilderI
	cmd SubqueryCommand
}

// NewSubqueryNode creates a new subquery
func NewSubqueryNode(b BuilderI) *SubqueryNode {
	n := &SubqueryNode{
		b: b,
	}
	return n
}

func (n *SubqueryNode) Load() *SubqueryNode {
	n.cmd = SubqueryCommandLoad
	return n
}

func (n *SubqueryNode) Get() *SubqueryNode {
	n.cmd = SubqueryCommandGet
	return n
}

func (n *SubqueryNode) Count() *SubqueryNode {
	n.cmd = SubqueryCommandCount
	return n
}

func (n *SubqueryNode) NodeType_() NodeType {
	return SubqueryNodeType
}

// equals is used internally by the framework to determine if two nodes are equal
func (n *SubqueryNode) equals(n2 Node) bool {
	if cn, ok := n2.(*SubqueryNode); ok {
		return cn.b == n.b
	}
	return false
}

/*
func (n *SubqueryNode) containedNodes() (nodes []QueryNode) {
	nodes = append(nodes, n) // Return the subquery node itself, because we need to do some work on it

	// must expand the returned nodes one more time
	for _,n2 := range n.b.nodes() {	// Refers back to db package, so do this differently
		if cn,_ := n2.(container); cn != nil {
			nodes = append(nodes, cn.containedNodes()...)
		} else {
			nodes = append(nodes, n2)
		}
	}
	return nodes
}
*/

func (n *SubqueryNode) TableName_() string {
	return ""
}

func (n *SubqueryNode) DatabaseKey_() string {
	return ""
}

func (n *SubqueryNode) log(level int) {
	tabs := strings.Repeat("\t", level)
	slog.Debug(tabs + "Subquery: ")
}

func (n *SubqueryNode) GobEncode() (data []byte, err error) {
	var buf bytes.Buffer
	e := gob.NewEncoder(&buf)

	if err = e.Encode(n.alias); err != nil {
		panic(err)
	}
	if err = e.Encode(n.b); err != nil {
		panic(err)
	}
	data = buf.Bytes()
	return
}

func (n *SubqueryNode) GobDecode(data []byte) (err error) {
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	if err = dec.Decode(&n.alias); err != nil {
		panic(err)
	}
	if err = dec.Decode(&n.b); err != nil {
		panic(err)
	}
	return
}

func init() {
	gob.Register(&SubqueryNode{})
}

// SubqueryBuilder is used internally by the framework to return the internal query builder of the subquery
func SubqueryBuilder(n *SubqueryNode) BuilderI {
	return n.b
}
func SubqueryCmd(n *SubqueryNode) SubqueryCommand {
	return n.cmd
}

func (n *SubqueryNode) id() string {
	return n.alias
}
