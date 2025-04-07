package sql

import (
	"database/sql"
	"fmt"
	"github.com/goradd/maps"
	"github.com/goradd/orm/pkg/db"
	"github.com/goradd/orm/pkg/db/jointree"
	"github.com/goradd/orm/pkg/query"
	"strconv"
)

// ReceiveRows gets data from a sql result set and returns it as a slice of maps.
//
// Each column is mapped to its column name.
// If you provide columnNames, those will be used in the map. Otherwise, it will get the column names out of the
// result set provided.
func ReceiveRows(rows *sql.Rows,
	columnTypes []query.ReceiverType,
	columnNames []string,
	joinTree *jointree.JoinTree,
	sql string,
	args []any,
) (values []map[string]any, err error) {
	var cursor query.CursorI

	cursor = NewSqlCursor(rows, columnTypes, columnNames, nil, sql, args)
	defer func() {
		cerr := cursor.Close()
		if err != nil {
			err = cerr
		}
	}()

	var v map[string]any
	for v, err = cursor.Next(); v != nil; v, err = cursor.Next() {
		if err != nil {
			return nil, err
		}
		values = append(values, v)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	if joinTree != nil {
		values = unpack(joinTree, values)
	}

	return values, nil
}

/*
Notes on the unpacking process:
This is quite tricky. Depending on the node structure, you may get repeated branches, or repeated entire structures with
individual differences.

After getting sql rows full of aliases for individual columns, we let the node structure direct how to unpack it.
We are going to do it in steps:
1) Create objects keyed by join table alias and id number. Foreign keys and Unique Reverse Fks will be a key to an object.
Reverse FKs and ManyMany relationships will be an ordered map of keys.
2) Walk the node map, assembling the structure
	a) If we arrive at a toMany relationship that is specified not to assemble as an array, we will duplicate the entire
	   structure each time.
	b) If we arrive at a toMany relationship that is arrayed, we pull in the individual items and keep walking
3) Return the assembled structure

Note that the order matters, so we put the whole thing in an OrderedMap so we can walk the whole thing in the order
that each object arrives, but then look for items in order.
*/

type objListType = maps.SliceMap[string, db.ValueMap] // We need a map that preserves insertion order

type unpacker struct {
	rowId int
	jt    *jointree.JoinTree
}

func unpack(jt *jointree.JoinTree, rows []map[string]interface{}) (out []map[string]interface{}) {
	u := unpacker{
		jt: jt,
	}
	return u.unpackResult(rows)
}

// unpackResult takes a flattened result set from the database that is a series of values keyed by alias, and turns them
// into a hierarchical result set that is keyed by join table alias and key.
func (u *unpacker) unpackResult(rows []map[string]interface{}) (out []map[string]any) {
	objectList := new(objListType)

	for _, row := range rows {
		u.unpackObjectArray(u.jt.Root, row, objectList)
	}

	out = u.unpackObjectList(objectList)
	return
}

func (u *unpacker) unpackObjectArray(el *jointree.Element, row db.ValueMap, result *objListType) {
	var obj db.ValueMap

	key := u.makeObjectKey(el, row)
	if key == "" {
		return // there are no objects in the array
	}

	i := result.Get(key)
	if i != nil {
		obj = i
	} else {
		obj = db.NewValueMap()
		result.Set(key, obj)
	}
	u.unpackObject(el, row, obj)

}

// unpackObject adds data from row corresponding to element to the data in object.
// object may already have data in it, in which case the row will have repeated data, and we are here to
// find a sub-object that is not repeated.
func (u *unpacker) unpackObject(el *jointree.Element, row db.ValueMap, object db.ValueMap) {
	var isNew bool

	if len(object) == 0 {
		isNew = true
	}

	for _, childElement := range el.References {
		key := query.NodeIdentifier(childElement.QueryNode)

		i := object[key]
		if childElement.IsArray() {
			var childList *objListType

			if i != nil {
				childList = i.(*objListType)
			} else {
				childList = new(objListType)
				object[key] = childList
			}
			u.unpackObjectArray(childElement, row, childList)
		} else {
			var childItem db.ValueMap
			if i != nil {
				childItem = i.(db.ValueMap)
			} else {
				childItem = db.NewValueMap()
				object[key] = childItem
			}
			u.unpackObject(childElement, row, childItem)
		}
	}
	if isNew {
		for leafItem := range el.SelectedColumns.All() {
			u.unpackLeaf(leafItem, row, object)
		}
		u.unpackCalculationAliases(el.Calculations, row, object)
	}
	return
}

func (u *unpacker) unpackLeaf(j *jointree.Element, row db.ValueMap, obj db.ValueMap) {
	if node, ok := j.QueryNode.(*query.ColumnNode); ok {
		key := j.Alias
		fieldName := node.QueryName
		obj[fieldName] = row[key]
	} else {
		panic("Unexpected node type.") // this is a framework error, should not happen
	}
}

// makeObjectKey makes a key for the element, such that when multiple rows for the same top object are
// in the result set, they can be grouped together within the parent object.
// The key is used in subsequent calls to determine what row joined data belongs to.
func (u *unpacker) makeObjectKey(tableElement *jointree.Element, row db.ValueMap) string {
	pk := tableElement.PrimaryKey()

	if pk == nil || pk.Alias == "" {
		// We are not identifying the row by a PK because of one of the following:
		// 1) This is a distinct select, and we are not selecting pks to avoid affecting the results of the query
		// 2) This is a groupby clause, which forces us to select only the groupby items and means we cannot add a PK to the row
		// We will therefore make up a unique key to identify the row such that none of the rows are grouped.
		u.rowId++
		return strconv.Itoa(u.rowId)
	}

	v := row[pk.Alias]
	if v == nil {
		return "" // the object we are looking for does not exist in the database
	}

	return fmt.Sprint(v)
}

func (u *unpacker) unpackCalculationAliases(calcNodes map[string]query.Node, row db.ValueMap, result db.ValueMap) {
	var aliasMap map[string]any // using map[string]any instead of db.ValueMap serves two purposes:
	// 1) allows us just to pass it through and
	// 2) signals to later unpacking operations that it is not an object

	if i := result[query.AliasResults]; i == nil {
		aliasMap = make(map[string]any)
		result[query.AliasResults] = aliasMap
	} else {
		aliasMap = i.(map[string]any)
	}
	for alias := range calcNodes {
		if _, ok := aliasMap[alias]; ok {
			continue // already added the item to the object, this is a repeat
		}
		aliasMap[alias] = row[alias]
	}
}

// unpackObjectLists converts the unpacking structure into a basic map[string]any structure that is delivered as the query result.
func (u *unpacker) unpackObjectList(objList *objListType) (outMap []map[string]any) {
	for _, dbMap := range objList.All() {
		outMap = append(outMap, u.unpackObjectMap(dbMap))
	}
	return
}

func (u *unpacker) unpackObjectMap(dbMap db.ValueMap) (outMap map[string]any) {
	outMap = make(map[string]any)
	for k, val := range dbMap {
		if v, ok := val.(db.ValueMap); ok {
			outMap[k] = u.unpackObjectMap(v)
		} else if v, ok := val.(*objListType); ok {
			outMap[k] = u.unpackObjectList(v)
		} else {
			outMap[k] = val
		}
	}
	return
}
