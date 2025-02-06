// Package jointree supports the query buildNodeTree process.
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

// CountAlias is the alias that will be used for auto generated count operation nodes.
const CountAlias = "_count"
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
	isSubquery         bool
	Command            query.BuilderCommand
	Limits             query.LimitParams
	Condition          query.Node
	GroupBys           []query.Node
	OrderBys           []query.Sorter
	Having             query.Node
	hasSelects         bool
	hasCalcs           bool
	hasAggregates      bool
}

// NewJoinTree analyzes b and turns it into a JoinTree.
func NewJoinTree(b query.BuilderI) *JoinTree {
	builder := b.(*query.Builder)

	t := JoinTree{
		IsDistinct: builder.IsDistinct,
		Command:    builder.Command,
		Limits:     builder.Limits,
		GroupBys:   builder.GroupBys,
		OrderBys:   builder.OrderBys,
		Having:     builder.HavingNode,
		Root:       newElement(builder.Root),
	}

	if len(builder.Conditions) == 1 {
		t.Condition = builder.Conditions[0]
	} else if len(builder.Conditions) > 1 {
		t.Condition = op.And(all.MapSlice[any](builder.Conditions)...)
	}
	t.buildNodeTree(builder)
	t.buildCommand(builder)
	return &t
}

// buildNodeTree performs initial analysis and processing of the builder.
// In particular, it gathers and inserts all the nodes found in the builder, and then assigns aliases to the table nodes.
func (t *JoinTree) buildNodeTree(b *query.Builder) {
	nodes := b.Nodes()
	for _, n := range nodes {
		t.addNode(n)
	}
	if len(b.Calculations) > 0 {
		t.hasCalcs = true
	}
	if len(b.Selects) > 0 {
		t.hasSelects = true
	}

	t.assignTableAliases(t.Root)
}

// buildCommand performs command specific processing on the builder
func (t *JoinTree) buildCommand(b *query.Builder) {
	switch t.Command {
	case query.BuilderCommandLoad:
		t.addCalculations(b)
		t.addSelectedColumns(b)
		t.assignSelectAliases()
	case query.BuilderCommandLoadCursor:
		t.checkCursor(b)
		t.addCalculations(b)
		t.addSelectedColumns(b)
		t.assignSelectAliases()
	case query.BuilderCommandCount:
		t.addCalculations(b)
		t.addSelectedColumns(b)
		t.assignSelectAliases()

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

	if query.NodeIsAggregate(node) {
		t.hasAggregates = true
	}

	if sq, ok := node.(*query.SubqueryNode); ok {
		top = t.addSubqueryNode(sq)
		return
	}
	rootNode := query.RootNode(node)
	if rootNode == nil {
		return
	}
	tableName = rootNode.TableName_()

	if t.Root.QueryNode.TableName_() != tableName {
		// TODO: If this has a parent builder from a sub query, check up the parent builder chain for a matching root table
		panic("Attempting to add a node that is not starting at the table being queried.")
	}

	// walk the current node tree and find an insertion point
	rn := reverse(node)
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

// insertNode inserts rn into the join tree.
// It does not check to see if the node is present already.
func (t *JoinTree) insertNode(rn *reverseNode, parent *Element) (top *Element) {
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
	var hasRootSelect bool

	// Column nodes are added to the join tree element
	// Table nodes found will add all the column nodes in that table
	for _, n := range builder.Selects {
		e, _, found := t.findNode(n)
		if !found {
			panic("prior node was not found in the tree") // this would be a bug in the ORM
		}
		if tn, ok := n.(query.TableNodeI); ok {
			t.selectAllColumnNodes(tn)
		} else {
			if nodeMatch(t.Root.QueryNode, e.Parent.QueryNode) {
				hasRootSelect = true
			}
			e.Parent.SelectedColumns.Add(e)
		}
		// Primary key nodes of all parent nodes must be added in order to assemble the final data structure
		// provided doing so will not mess up the query.
		if len(builder.GroupBys) > 0 || t.IsDistinct || t.isSubquery {
			// Do not auto-add primary key nodes in these cases
		} else {
			t.selectParentPrimaryKeys(e)
		}
	}

	if !hasRootSelect && !(len(builder.GroupBys) > 0 || t.hasAggregates || t.IsDistinct || t.isSubquery || t.Command == query.BuilderCommandCount) {
		t.selectAllColumnNodes(t.Root.QueryNode.(query.TableNodeI))
	}

	if len(builder.GroupBys) > 0 {
		t.addGroupBySelects(builder)
	}

	// Having clauses MUST be selected so they can be post filtered
	for _, n := range query.ContainedNodes(builder.HavingNode) {
		e, _, found := t.findNode(n)
		if !found {
			panic("prior node was not found in the tree")
		}
		e.Parent.SelectedColumns.Add(e)
	}
}

// selectAllColumnNodes adds all the column nodes of the table node to the list of selected elements.
func (t *JoinTree) selectAllColumnNodes(tableNode query.TableNodeI) {
	cols := tableNode.ColumnNodes_()
	for _, n := range cols {
		e2 := t.addNode(n)
		e2.Parent.SelectedColumns.Add(e2)
	}
}

// selectParentPrimaryKeys walks the tree selecting all the parent primary keys of n.
func (t *JoinTree) selectParentPrimaryKeys(e *Element) {
	for parentElement := e.Parent; parentElement != nil; parentElement = parentElement.Parent {
		fkn := parentElement.QueryNode.(query.PrimaryKeyer).PrimaryKey()
		en := t.addNode(fkn)
		en.Parent.SelectedColumns.Add(en)
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
			e.Parent.SelectedColumns.Add(e)
		} else if a, ok := n.(query.AliasNodeI); ok {
			if _, ok2 := b.Calculations[a.Alias()]; !ok2 {
				panic(fmt.Sprintf("alias in group by not found in Calculations: %s", a.Alias()))
			}
		} else if n.NodeType_() == query.OperationNodeType {
			panic("operation nodes in group bys must be aliased with a Calculation")
		} else if n.NodeType_() == query.SubqueryNodeType {
			panic("subquery nodes in group bys must be aliased with a Calculation")
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
		p := e.Parent.SelectedColumns.Clone()
		for _, e2 := range elements {
			p.Delete(e2)
		}
		if p.Len() > 0 {
			panic(fmt.Sprintf("table %s has selects that are not part of the group by", e.Parent.QueryNode.TableName_()))
		}
	}
}

func (t *JoinTree) addCalculations(b *query.Builder) {
	for alias, values := range b.Calculations {
		e, _, found := t.findNode(values.BaseNode)
		if !found {
			panic("calculation base node was not found in the tree")
		}
		if e.Calculations == nil {
			e.Calculations = make(map[string]query.Node)
		}
		e.Calculations[alias] = values.Operation
		t.hasCalcs = true
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

func (t *JoinTree) CalculationsIter() iter.Seq2[string, query.Node] {
	return t.Root.CalculationsIter()
}

// FindAlias searches the join tree for the given manually assigned alias and returns the node corresponding to the alias.
func (t *JoinTree) FindAlias(alias string) query.Node {
	return t.Root.FindCalculation(alias)
}

func (t *JoinTree) checkCursor(builder *query.Builder) {
	for _, n := range builder.Nodes() {
		if query.NodeIsArray(n) {
			panic("you cannot query a cursor and also have an array type node in the query")
		}
	}
}

func (t *JoinTree) HasAggregates() bool {
	return t.hasAggregates
}

func (t *JoinTree) HasCalcs() bool {
	return t.hasCalcs
}

func (t *JoinTree) HasSelects() bool {
	return t.hasSelects
}
