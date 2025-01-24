package jointree

import (
	"cmp"
	"github.com/goradd/maps"
	"github.com/goradd/orm/pkg/query"
	"iter"
	"slices"
)

// Element is used to build the join tree. The join tree creates a hierarchy of joined nodes that let us
// generate aliases, serialize the query, and afterwards unpack the results.
type Element struct {
	QueryNode     query.Node
	Parent        *Element
	References    []*Element // TableNodeI objects
	Columns       []*Element
	Selects       maps.Set[*Element] // pointers to elements in Columns that will be selected to be returned in the query
	JoinCondition query.Node
	Alias         string
	Expanded      bool
	IsPK          bool
}

func newElement(node query.Node) *Element {
	e := new(Element)
	e.QueryNode = node

	// cache some things from the node
	if c, ok := node.(*query.ColumnNode); ok {
		e.IsPK = c.IsPrimaryKey
	} else {
		if c := query.NodeCondition(node); c != nil {
			e.JoinCondition = c
		}
		if query.NodeIsExpanded(node) {
			e.Expanded = true
		}
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
			for _, c := range slices.SortedFunc(e.Selects.All(), func(e1, e2 *Element) int {
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

// ColumnIter iterates on all the columns in this elment and its sub elements.
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
		s += ":" + j.QueryNode.(query.Aliaser).Alias()
	}
	return s
}
