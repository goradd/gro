package jointree

import (
	"cmp"
	iter2 "github.com/goradd/iter"
	"github.com/goradd/maps"
	"github.com/goradd/orm/pkg/query"
	"iter"
	"slices"
)

// Element is used to build the join tree. The join tree creates a hierarchy of joined nodes that let us
// generate aliases, serialize the query, and afterward, unpack the results.
type Element struct {
	QueryNode       query.Node
	Parent          *Element
	References      []*Element         // TableNodeI objects
	Columns         []*Element         // All columns that will be used to build the query, including those in Where, OrderBy and other clauses
	SelectedColumns maps.Set[*Element] // Pointers to elements in Columns that will be returned in the query
	//JoinCondition   query.Node
	Alias        string                // computed or assigned alias
	Calculations map[string]query.Node // calculations attached to this node by alias
	IsPK         bool
}

func newElement(node query.Node) *Element {
	e := new(Element)
	e.QueryNode = node

	// cache some things from the node
	if c, ok := node.(*query.ColumnNode); ok {
		e.IsPK = c.IsPrimaryKey
	}
	return e
}

// PrimaryKey will return the primary key join tree item attached to this item, or nil if none exists.
// If the element is not the kind of element that can have a primary key, it will panic.
func (j *Element) PrimaryKey() *Element {
	if _, ok := j.QueryNode.(query.PrimaryKeyer); !ok {
		panic("not a primary keyer")
	}
	if j.Columns != nil &&
		j.Columns[0].IsPK {
		return j.Columns[0]
	} else {
		return nil
	}
}

// SelectsIter iterates on all the selects in this element and its sub elements.
func (j *Element) SelectsIter() iter.Seq[*Element] {
	return func(yield func(*Element) bool) {
		var cols func(*Element) bool
		cols = func(e *Element) bool {
			for _, c := range slices.SortedFunc(e.SelectedColumns.All(), func(e1, e2 *Element) int {
				return cmp.Compare(e1.Alias, e2.Alias)
			}) {
				if !yield(c) {
					return false
				}
			}
			for _, r := range e.References {
				if !cols(r) {
					return false
				}
			}
			return false
		}
		cols(j)
	}
}

// CalculationsIter iterates on all the calculations in this element and its sub elements.
func (j *Element) CalculationsIter() iter.Seq2[string, query.Node] {
	return func(yield func(string, query.Node) bool) {
		for k, v := range iter2.KeySort(j.Calculations) {
			if !yield(k, v) {
				return
			}
		}
		for _, r := range j.References {
			for k, v := range r.CalculationsIter() {
				if !yield(k, v) {
					return
				}
			}
		}
	}
}

// ColumnIter iterates on all the columns in this element and its sub elements.
func (j *Element) ColumnIter() iter.Seq[*Element] {
	return func(yield func(*Element) bool) {
		var cols func(*Element) bool
		cols = func(e *Element) bool {
			for _, c := range e.Columns {
				if !yield(c) {
					return false
				}
			}
			for _, r := range e.References {
				if !cols(r) {
					return false
				}
			}
			return false
		}
		cols(j)
	}
}

// String shows information about the node for debugging.
func (j *Element) String() string {
	s := j.QueryNode.NodeType_().String()
	if tn := j.QueryNode.TableName_(); tn != "" {
		s += ":" + tn
	}
	if c, ok := j.QueryNode.(*query.ColumnNode); ok {
		s += ":" + c.QueryName
	}
	if j.Alias != "" {
		s += ":" + j.Alias
	}
	if j.QueryNode.NodeType_() == query.AliasNodeType {
		s += ":" + j.QueryNode.(query.AliasNodeI).Alias()
	}
	return s
}

// IsArray returns true if the enclosed query node is an array type node.
func (j *Element) IsArray() bool {
	return query.NodeIsArray(j.QueryNode)
}

func (j *Element) FindCalculation(alias string) query.Node {
	if calc, ok := j.Calculations[alias]; ok {
		return calc
	}
	for _, e := range j.References {
		n := e.FindCalculation(alias)
		if n != nil {
			return n
		}
	}
	return nil
}
