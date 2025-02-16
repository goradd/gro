package query

import (
	"context"
	"fmt"
)

type BuilderCommand int

const (
	BuilderCommandLoad = iota
	BuilderCommandDelete
	BuilderCommandCount
	BuilderCommandLoadCursor
)

// AliasResults is the index name that will be used for all calculations and other aliased results in the result set.
const AliasResults = "aliases_"

type CursorI interface {
	Next() map[string]interface{}
	Close() error
}

// LimitParams is the information needed to limit the rows being requested.
type LimitParams struct {
	MaxRowCount int
	Offset      int
}

// AreSet returns true if the limit paramter values are not zero.
func (l LimitParams) AreSet() bool {
	return l.MaxRowCount > 0
}

// BuilderI is the interface to the builder structure. Since the builder is directly interacted with by the developer,
// passing this interface instead of the Builder object makes it more clear what the developer should use to build queries.
type BuilderI interface {
	//Join(n Node, condition Node)
	Where(c Node)
	Having(c Node)
	OrderBy(nodes ...Sorter)
	GroupBy(nodes ...Node)
	Limit(maxRowCount int, offset int)
	Select(nodes ...Node)
	Distinct()
	Calculation(base TableNodeI, alias string, n OperationNodeI)
	//Subquery() *SubqueryNode
	//Context() context.Context
}

type calc struct {
	BaseNode  TableNodeI
	Operation OperationNodeI
}

// Builder is a mix-in to implement the BuilderI interface in various builder classes.
// It gathers the builder instructions as the query is built, then the final build process
// processes the query into a join tree that can be used by database drivers to generate
// the database specific query.
type Builder struct {
	Ctx     context.Context // The context that will be used in all the queries
	Command BuilderCommand
	Root    TableNodeI
	//Joins      []Node
	OrderBys     []Sorter
	Conditions   []Node
	IsDistinct   bool
	Calculations map[string]calc
	// Adds a COUNT(*) to the select list
	GroupBys   []Node
	Selects    []Node
	Limits     LimitParams
	HavingNode Node
	IsSubquery bool
}

func NewBuilder(ctx context.Context, rootNode TableNodeI) *Builder {
	if NodeParent(rootNode) != nil {
		panic("root node must be a top level node")
	}

	return &Builder{Ctx: ctx, Root: rootNode}
}

// Context returns the context.
func (b *Builder) Context() context.Context {
	return b.Ctx
}

/*
// Join will attach the given reference node to the builder.
func (b *Builder) Join(alias string, n Node, condition Node) {
	// TBD: This must include a condition and an alias!
	if !NodeIsTable(n) {
		panic(fmt.Errorf("node %s is not joinable", n))
	}

	if condition != nil {
		if c, ok := n.(Conditioner); !ok {
			panic("node cannot have conditions")
		} else {
			c.SetCondition(condition)
		}
	}

	b.Joins = append(b.Joins, n)
}

*/

// Calculation adds the aliased calculation node operation onto base.
func (b *Builder) Calculation(base TableNodeI, alias string, operation OperationNodeI) {
	if b.Calculations == nil {
		b.Calculations = make(map[string]calc)
	}
	if _, ok := b.Calculations[alias]; ok {
		panic("alias already exists") // aliases must be unique across the entire operation
	}
	b.Calculations[alias] = calc{base, operation}
}

// Where adds condition to the Where clause. Multiple calls to Condition will result in conditions joined with an And.
func (b *Builder) Where(condition Node) {
	b.Conditions = append(b.Conditions, condition)
}

// OrderBy adds the order by nodes. If these are table type nodes, the primary key of the table will be used.
// These nodes can be modified using Ascending and Descending calls.
func (b *Builder) OrderBy(nodes ...Sorter) {
	b.OrderBys = append(b.OrderBys, nodes...)
}

// Limit limits the query to returning maxRowCount rows at most, starting at row offset.
// In SQL queries, this limits the rows BEFORE assembling many type relationships,
// which usually is not what is wanted.
// Therefore, you cannot put limits on queries that have reverse or many-many
// relationships.
// Also note that SQL at least will perform the entire query before finding the offset, which could have performance
// issues. If paging through a large dataset, consider getting a list of just primary keys of the records you want, saving
// that list, and then lazy-loading the rest of the information with a separate query.
//
// Warning: Setting maxRowCount to zero will turn off the limit. Setting it to less than zero will panic.
// If you really want to return no information, do not call the query.
func (b *Builder) Limit(maxRowCount int, offset int) {
	if b.Limits.AreSet() {
		panic("query already has a limit")
	}
	if maxRowCount < 0 {
		panic(fmt.Sprintf("setting maxRowCount to %d", maxRowCount))
	}
	b.Limits.MaxRowCount = maxRowCount
	b.Limits.Offset = offset
}

// Select specifies what specific columns will be loaded with data.
// By default, all the columns of the root table will be queried and loaded.
// If columns from the root table are selected, that will limit the columns queried and loaded to only those columns.
// If related tables are specified, then all the columns from those tables are queried, selected and joined to the root.
// If columns in related tables are specified, then only those columns will be queried and loaded.
// Depending on the query, additional columns may automatically be added to the query. In particular, primary key columns
// will be added in most situations. The exception to this would be in distinct queries, group by queries, or subqueries.
func (b *Builder) Select(nodes ...Node) {
	if b.GroupBys != nil {
		panic("you cannot have Select and GroupBy statements in the same query. The GroupBy columns will automatically be selected")
	}
	for _, n := range nodes {
		switch n.NodeType_() {
		case ColumnNodeType:
			fallthrough
		case ReferenceNodeType:
			fallthrough
		case ManyManyNodeType:
			fallthrough
		case ReverseNodeType:
			b.Selects = append(b.Selects, n)
		default:
			panic("you cannot Select on this type of node: " + n.NodeType_().String())
		}
	}
}

// Distinct sets the distinct bit, causing the query to not return duplicates.
func (b *Builder) Distinct() {
	b.IsDistinct = true
}

// GroupBy sets the nodes that are grouped.
// GroupBy only makes sense if only those same columns are selected. Most SQL databases enforce this.
func (b *Builder) GroupBy(nodes ...Node) {
	if b.Selects != nil {
		panic("do not have Select and GroupBy statements in the same query. The GroupBy columns will automatically be selected.")
	}
	b.GroupBys = append(b.GroupBys, nodes...)
}

// Having adds a HAVING condition, which is a filter that acts on the results of a query.
// In particular, it is useful for filtering after aggregate functions have done their work.
func (b *Builder) Having(node Node) {
	b.HavingNode = node // should be a condition node?
}

// Subquery adds a subquery node, which is like a mini query builder that should result in a single value.
func (b *Builder) Subquery() *SubqueryNode {
	n := NewSubqueryNode(b)
	b.IsSubquery = true
	return n
}

// Nodes returns all the nodes referred to in the query.
func (b *Builder) Nodes() (nodes []Node) {
	// first pass
	topNodes := b.topNodes()

	// unpack container nodes
	for _, n := range topNodes {
		if sn, ok := n.(*SubqueryNode); ok {
			nodes = append(nodes, n) // Return the subquery node itself, because we need to do some work on it
			b2 := sn.b.(*Builder)
			// recurse
			nodes = append(nodes, b2.Nodes()...)
		} else if cn := ContainedNodes(n); cn != nil {
			nodes = append(nodes, cn...)
		} else {
			nodes = append(nodes, n)
		}
	}
	return
}

// HasAggregate returns true if the builder has an aggregate function in it somwhere.
func (b *Builder) HasAggregate() bool {
	// first pass
	topNodes := b.topNodes()

	// unpack container nodes
	for _, n := range topNodes {
		if NodeHasAggregate(n) {
			return true
		}
	}
	return false
}

// topNodes returns all the top level nodes referred to in the query.
// Some of the nodes returned may be container nodes.
func (b *Builder) topNodes() []Node {
	var nodes []Node

	/*
		for _, n := range b.Joins {
			nodes = append(nodes, n)
			if c := NodeCondition(n); c != nil {
				nodes = append(nodes, c)
			}
		}
	*/

	for _, n := range b.OrderBys {
		nodes = append(nodes, n)
	}

	nodes = append(nodes, b.Conditions...)

	for _, n := range b.GroupBys {
		if p, ok := n.(PrimaryKeyer); ok {
			n = p.PrimaryKey() // Allow table nodes, but then actually have them be the pk in this context
		}
		nodes = append(nodes, n)
	}

	if b.HavingNode != nil {
		nodes = append(nodes, b.HavingNode)
	}
	nodes = append(nodes, b.Selects...)

	for _, n := range b.Calculations {
		nodes = append(nodes, n.BaseNode)
		nodes = append(nodes, n.Operation.(Node))
	}

	return nodes
}
