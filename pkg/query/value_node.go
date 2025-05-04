package query

import (
	"bytes"
	"encoding/gob"
	"reflect"
	"time"
)

// ValueNode represents a value for a built-in type that is to be used in a query.
// For example,
type ValueNode struct {
	value interface{}
}

type Identifierer interface {
	Identifier() string
}

// Value is a shortcut for converting a constant value to a node
func Value(i interface{}) Node {
	return NewValueNode(i)
}

// NewValueNode returns a new ValueNode that wraps the given value.
func NewValueNode(i interface{}) Node {
	n := &ValueNode{
		value: i,
	}

	switch v := i.(type) {
	// do nothing
	case string:
	case int:
	case uint:
	case uint64:
	case int64:
	case float64:
	case float32:
	case time.Time:

		// casts
	case []byte:
		n.value = string(v[:])
	case nil:
		panic("You cannot use nil as an operator. If you are testing for a NULL, use the IsNull function.")
	default:
		// Use reflection to do various conversions
		typ := reflect.TypeOf(v)
		k := typ.Kind()
		val := reflect.ValueOf(v)

		switch k {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			n.value = int(val.Int())
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			n.value = uint(val.Uint())
		case reflect.Bool:
			n.value = val.Bool()
		case reflect.Float32:
			// converting float32 to float64 might cause problems in a sql statement, so we leave the type as float32
			n.value = float32(val.Float())
		case reflect.Float64:
			n.value = val.Float()
		case reflect.Slice, reflect.Array:
			var ary []Node
			for i2 := 0; i2 < val.Len(); i2++ {
				// TODO: Handle QueryNode's here too? Prevent more than one level deep?
				ary = append(ary, NewValueNode(val.Index(i2).Interface()))
			}
			n.value = ary
		case reflect.String:
			n.value = val.String()
		default:
			panic("Can't use this type as a value node.")
		}
	}
	return n
}

/*
// equals returns whether the given node is equal to this node.
func (n *ValueNode) equals(n2 Node) bool {
	if cn, ok := n2.(*ValueNode); ok {
		if an2, ok2 := cn.value.([]Node); ok2 {
			if an1, ok3 := n.value.([]Node); !ok3 {
				return false
			} else if len(an2) != len(an1) {
				return false
			} else {
				for i, n3 := range an1 {
					if !NodeIsEqual(n3, an2[i]) {
						return false
					}
				}
			}
			return true
		}
		return cn.value == n.value
	}
	return false
}
*/

func (n *ValueNode) TableName_() string {
	return ""
}

func (n *ValueNode) DatabaseKey_() string {
	return ""
}

/*
func (n *ValueNode) log(level int) {
	tabs := strings.Repeat("\t", level)
	slog.Debug(tabs + "Val: " + fmt.Sprint(n.value))
}
*/

// ValueNodeGetValue is used internally by the framework to get the node's internal value.
func ValueNodeGetValue(n *ValueNode) interface{} {
	return n.value
}

func (n *ValueNode) NodeType_() NodeType {
	return ValueNodeType
}

// GobEncode encodes the node for storage and retrieval
func (n *ValueNode) GobEncode() (data []byte, err error) {
	var buf bytes.Buffer
	e := gob.NewEncoder(&buf)

	if err = e.Encode(&n.value); err != nil {
		panic(err)
	}
	data = buf.Bytes()
	return
}

// GobDecode retrieves the node from storage
func (n *ValueNode) GobDecode(data []byte) (err error) {
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	if err = dec.Decode(&n.value); err != nil {
		panic(err)
	}
	return
}

func init() {
	gob.Register(&ValueNode{})
}
