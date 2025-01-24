// Package jointree supports the query build process.
package jointree

import (
	"fmt"
	"github.com/goradd/all"
	"github.com/goradd/orm/pkg/op"
	"github.com/goradd/orm/pkg/query"
	"iter"
	"slices"
	"strconv"
)

const countAlias = "_count"
const columnAliasPrefix = "c_"
const tableAliasPrefix = "t_"

// JoinTree is used by various goradd-orm database drivers to convert a query.Builder object into a query understandable
// by the database.
// Developers do not normally need to call code here unless they are making a custom database driver.
//
// It analyzes a query.QueryBuilder object, combines all the nodes into a single tree structure,
// adds extra nodes implied by the query, and assigns aliases to the columns that will be selected.
type JoinTree struct {
	Root               *Element
	SubPrefix          string
	tableCounter       int
	selectAliasCounter int
	IsDistinct         bool
	hasGroupBys        bool
	isSubquery         bool
	Command            query.BuilderCommand
	Aliases            map[string]query.Node
	Limits             query.LimitParams
	Condition          query.Node
	GroupBys           []query.Node
	OrderBys           []query.Sorter
	Having             query.Node
}

// NewJoinTree analyzes b and turns it into a JoinTree.
func NewJoinTree(b query.BuilderI) *JoinTree {
	builder := b.(*query.Builder)
	t := JoinTree{
		IsDistinct:  builder.IsDistinct,
		Command:     builder.Command,
		hasGroupBys: len(builder.GroupBys) > 0,
		Aliases:     make(map[string]query.Node),
		Limits:      builder.Limits,
		GroupBys:    builder.GroupBys,
		OrderBys:    builder.OrderBys,
		Having:      builder.HavingNode,
	}
	if len(builder.Conditions) == 1 {
		t.Condition = builder.Conditions[0]
	} else if len(builder.Conditions) > 1 {
		t.Condition = op.And(all.MapSlice[any](builder.Conditions)...)
	}
	t.build(builder)
	t.buildCommand(builder)
	return &t
}

// build performs initial analysis and processing of the builder.
// In particular, it gathers and inserts all the nodes found in the builder, and then assigns aliases to the table nodes.
func (t *JoinTree) build(b *query.Builder) {
	nodes := b.Nodes()
	for _, n := range nodes {
		t.addNode(n)
	}
	t.assignTableAliases(t.Root)
}

// buildCommand performs command specific processing on the builder
func (t *JoinTree) buildCommand(b *query.Builder) {
	switch t.Command {
	case query.BuilderCommandLoad:
		fallthrough
	case query.BuilderCommandLoadCursor:
		t.addSelectedColumns(b)
		t.addAliases(b.AliasNodes)
		t.assignSelectAliases()
	case query.BuilderCommandCount:
		n := query.NewCountNode(b.Selects...)
		b.Calculation(countAlias, n)
		t.addAliases(b.AliasNodes)

	default:
		// do nothing more
	}
}

// addNode adds a node to the join tree, if the node is not already there,
// Returns the element added, or the element found if the node already exists.
// node should not be a container node.
func (t *JoinTree) addNode(node query.Node) (top *Element) {
	var tableName string
	//var hasSubquery bool // Turns off the check to make sure all nodes come from the same table, since subqueries might have different tables

	if sq, ok := node.(*query.SubqueryNode); ok {
		top = t.addSubqueryNode(sq)
		return
	}
	rootNode := query.RootNode(node)
	if rootNode == nil {
		return
	}
	tableName = rootNode.TableName_()

	if t.Root != nil {
		if t.Root.QueryNode.TableName_() != tableName {
			// TODO: If this has a parent builder from a sub query, check up the parent builder chain for a matching root table
			panic("Attempting to add a node that is not starting at the table being queried.")
		}
	}

	// walk the current node tree and find an insertion point
	rn := reverse(node)
	if t.Root == nil {
		top = t.insertNode(rn, nil) // seed the root tree
		return
	}
	e, rn, found := t.findNode(node)
	if !found {
		top = t.insertNode(rn, e)
	} else {
		top = e
	}
	return
}

// findNode searches the tree for the Element e matching node.
// If it is found, e will be the Element containing node, and rn will also point to the node.
// If it is not found, e will be the Element where node should be inserted, and rn will be the spot in node that
// should be inserted. In other words, the parent of rn will match the Node in e.
func (t *JoinTree) findNode(node query.Node) (e *Element, rn *reverseNode, found bool) {
	e = t.Root
	rn = reverse(node)
	if !nodeMatch(e.QueryNode, rn.node) {
		return nil, nil, false // root nodes do not match
	}

reverseNodeLoop:
	for {
		if rn.child == nil {
			found = nodeMatch(e.QueryNode, rn.node)
			if !found {
				e = e.Parent
			}
			return
		}

		for _, r := range e.References {
			if nodeMatch(r.QueryNode, rn.child.node) {
				rn = rn.child
				e = r
				continue reverseNodeLoop
			}
		}
		for _, c := range e.Columns {
			if nodeMatch(c.QueryNode, rn.child.node) {
				return c, rn.child, true
			}
		}
		// no match
		return e, rn.child, false
	}
}

// FindElement will return the element matching node, or nil if not found.
func (t *JoinTree) FindElement(node query.Node) *Element {
	e, _, found := t.findNode(node)
	if found {
		return e
	}
	return nil
}

func (t *JoinTree) addSubqueryNode(node query.Node) *Element {
	/*
		if sq, ok := node.(*query.SubqueryNode); ok {
			hasSubquery = true
			b.SubqueryCounter++
			b2 := SubqueryBuilder(sq).(*Builder)
			b2.SubPrefix = strconv.Itoa(b.SubqueryCounter) + "_"
			b2.ParentBuilder = b
			b2.buildJoinTree()
			continue
		}
	*/
	return nil
}

// insertNode inserts that node in the join tree.
// Does not check to see if the node is present already.
func (t *JoinTree) insertNode(rn *reverseNode, parent *Element) (top *Element) {
	if t.Root == nil {
		t.Root = newElement(rn.node)
		parent = t.Root
		rn = rn.child
		top = t.Root
	}

	for rn != nil {
		e := newElement(rn.node)
		if top == nil {
			top = e
		}
		if rn.node.NodeType_() == query.ColumnNodeType {
			if e.IsPK {
				// PKs go to the front
				parent.Columns = slices.Insert(parent.Columns, 0, e)
			} else {
				parent.Columns = append(parent.Columns, e)
			}
		} else {
			parent.References = append(parent.References, e)
		}
		e.Parent = parent
		parent = e
		rn = rn.child
	}
	return
}

// assignTableAliases will assign aliases to the item and all children that are tables.
// Call this with the root to assign all the table aliases.
func (t *JoinTree) assignTableAliases(item *Element) {
	t.tableCounter++
	item.Alias = tableAliasPrefix + t.SubPrefix + strconv.Itoa(t.tableCounter)
	for _, item2 := range item.References {
		t.assignTableAliases(item2)
	}
}

// addSelectedColumns determines which columns should be selected and adds them to the join tree.
func (t *JoinTree) addSelectedColumns(builder *query.Builder) {
	// First process explicit selects
	for _, n := range builder.Selects {
		e, _, found := t.findNode(n)
		if !found {
			panic("prior node was not found in the tree")
		}
		e.Parent.Selects.Add(e)
	}
	t.selectForeignKeys(t.Root) // do this here to make sure GroupBy will fail if they don't match

	if len(builder.GroupBys) > 0 {
		t.addGroupBySelects(builder)
	}
	t.selectRelatedTableColumns(t.Root)

	// Having clauses MUST be selected so they can be post filtered
	for _, n := range query.ContainedNodes(builder.HavingNode) {
		e, _, found := t.findNode(n)
		if !found {
			panic("prior node was not found in the tree")
		}
		e.Parent.Selects.Add(e)
	}

	// Check for extra group by columns here
}

// selectForeignKeys walks the tree looking for forward references and makes sure
// the foreign key column for that reference is selected.
func (t *JoinTree) selectForeignKeys(e *Element) {
	for _, er := range e.References {
		if er.QueryNode.NodeType_() == query.ReferenceNodeType {
			fkn := er.QueryNode.(query.PrimaryKeyer).PrimaryKeyNode()
			en := t.addNode(fkn)
			en.Parent.Selects.Add(en)
		}
		t.selectForeignKeys(er)
	}
}

// selectRelatedTableColumns will select columns that are not currently selected that are implied by
// the query or required to assemble the results.
func (t *JoinTree) selectRelatedTableColumns(e *Element) {
	if e.Selects.Len() == 0 {
		// nothing selected, so select everything
		cols := e.QueryNode.(query.TableNodeI).ColumnNodes_()
		for _, n := range cols {
			e2 := t.addNode(n)
			e2.Parent.Selects.Add(e2)
		}
	} else {
		// make sure primary key nodes are included if required
		if t.hasGroupBys || t.IsDistinct || t.isSubquery {
			// adding primary key node in these situations will mess up the query
		} else {
			n := e.QueryNode.(query.PrimaryKeyer).PrimaryKeyNode()
			e2 := t.addNode(n)
			e2.Parent.Selects.Add(e2)
		}
	}
	for _, e := range e.References {
		t.selectRelatedTableColumns(e)
	}
}

// addGroupBySelects handles adding groupby auto selects.
// This is very tricky, in that sql databases do not allow selecting of columns from a table that are not in the
// group by clause, since there may be multiple possibilities for the values of those extra columns.
// Forward references can only work if the foreign key is in the group by.
// Reverse references and ManyMany references do not work, since the foreign key would be required to be in the group by
// which defeats the purpose of a group by.
func (t *JoinTree) addGroupBySelects(b *query.Builder) {
	// GroupBy nodes that are columns not in operations must be explicitly selected.
	// Selecting here will prevent automatically adding select clauses to these tables.
	// We also check to make sure non-group by columns are not being selected, since that will cause an error if so.

	for _, n := range b.GroupBys {
		if n.NodeType_() == query.ColumnNodeType {
			e, _, found := t.findNode(n)
			if !found {
				panic("group by node not found in the tree")
			}
			e.Parent.Selects.Add(e)
		} else if a, ok := n.(query.Aliaser); ok {
			if offset := slices.IndexFunc(b.AliasNodes, func(a2 query.Aliaser) bool {
				return a2.Alias() == a.Alias()
			}); offset == -1 {
				panic(fmt.Sprintf("alias in group by not found: %s", a.Alias()))
			}
		} else if n.NodeType_() == query.OperationNodeType {
			panic("operation nodes in group bys must be aliased with a Computation node")
		} else if n.NodeType_() == query.SubqueryNodeType {
			panic("subquery nodes in group bys must be aliased with a Computation node")
		} else if n.NodeType_() == query.ReferenceNodeType {
			panic("group by a foreign key column rather than a reference")
		} else if n.NodeType_() == query.ReverseNodeType ||
			n.NodeType_() == query.ManyManyNodeType {
			panic("cannot group by on this node type")
		}
	}
}

func (t *JoinTree) checkGroupBySelects(b *query.Builder) {
	// GroupBy nodes that are columns not in operations must be explicitly selected.
	// Selecting here will prevent automatically adding select clauses to these tables.
	// We also check to make sure non-group by columns are not being selected, since that will cause an error if so.

	var elements []*Element
	for _, n := range b.GroupBys {
		e, _, _ := t.findNode(n)
		elements = append(elements, e)
	}
	// make sure no other selects are currently included in the related order by tables
	for _, e := range elements {
		p := e.Parent.Selects.Clone()
		for _, e2 := range elements {
			p.Delete(e2)
		}
		if p.Len() > 0 {
			panic(fmt.Sprintf("table %s has selects that are not part of the group by", e.Parent.QueryNode.TableName_()))
		}
	}
}

func (t *JoinTree) addAliases(aliases []query.Aliaser) {
	var n query.Node
	for _, a := range aliases {
		n = a
		t.Aliases[a.Alias()] = n
	}
}

func (t *JoinTree) assignSelectAliases() {
	for ec := range t.SelectsIter() {
		t.selectAliasCounter++
		ec.Alias = columnAliasPrefix + t.SubPrefix + strconv.Itoa(t.selectAliasCounter)
	}
}

func (t *JoinTree) SelectsIter() iter.Seq[*Element] {
	return t.Root.SelectsIter()
}
