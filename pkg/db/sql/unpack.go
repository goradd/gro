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

// SqlReceiveRows gets data from a sql result set and returns it as a slice of maps.
//
// Each column is mapped to its column name.
// If you provide columnNames, those will be used in the map. Otherwise, it will get the column names out of the
// result set provided.
func SqlReceiveRows(rows *sql.Rows,
	columnTypes []query.ReceiverType,
	columnNames []string,
	joinTree *jointree.JoinTree,
) []map[string]any {

	var values []map[string]any

	cursor := NewSqlCursor(rows, columnTypes, columnNames, nil)
	defer cursor.Close()
	for v := cursor.Next(); v != nil; v = cursor.Next() {
		values = append(values, v)
	}
	if joinTree != nil {
		values = unpack(joinTree, values)
	}

	return values
}

/*
// sqlReceiveRows2 gets data from a sql result set and returns it as a slice of maps. Each column is mapped to its column name.
// If you provide column names, those will be used in the map. Otherwise it will get the column names out of the
// result set provided
// This unused code is here in case we need to jetison the cursor method above.
func sqlReceiveRows2(rows *sql.Rows, columnTypes []query.ReceiverType, columnNames []string) (values []map[string]interface{}) {
	var err error

	values = []map[string]interface{}{}

	columnReceivers := make([]SqlReceiver, len(columnTypes))
	columnValueReceivers := make([]interface{}, len(columnTypes))

	if columnNames == nil {
		columnNames, err = rows.Columns()
		if err != nil {
			log.Panic(err)
		}
	}

	for i, _ := range columnReceivers {
		columnValueReceivers[i] = &(columnReceivers[i].R)
	}

	for rows.Next() {
		err = rows.Scan(columnValueReceivers...)

		if err != nil {
			log.Panic(err)
		}

		v1 := make(map[string]interface{}, len(columnReceivers))
		for j, vr := range columnReceivers {
			v1[columnNames[j]] = vr.Unpack(columnTypes[j])
		}
		values = append(values, v1)

	}
	err = rows.Err()
	if err != nil {
		log.Panic(err)
	}
	return
}
*/

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

type oMapType = maps.SliceMap[string, any] // We need a map that preserves insertion order

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
	var o2 db.ValueMap

	oMap := new(oMapType)
	aliasMap := new(oMapType)

	// First we create a tree structure of the data that will mirror the node structure
	for _, row := range rows {
		rowId := u.unpackObject(u.jt.Root, row, oMap)
		u.unpackSpecialAliases(rowId, row, aliasMap)
	}

	// We then walk the tree and create the final data structure as arrays
	for key, value := range oMap.All() {
		out2 := u.expand(u.jt.Root, value.(*oMapType))
		// Add the Alias calculations specifically requested by the caller
		for _, o2 = range out2 {
			if m := aliasMap.Get(key); m != nil {
				o2[query.AliasResults] = m
			}
			out = append(out, o2)
		}
	}
	return out
}

// unpackObject finds the object that corresponds to parent in the row, and either adds it to the oMap, or if its
// already in the oMap, reuses the old one and adds more data to it. oMap should only contain objects of parent type.
// Returns the row id to use to refer to the row later.
func (u *unpacker) unpackObject(parent *jointree.Element, row db.ValueMap, result *oMapType) (rowId string) {
	var obj *oMapType
	var key string

	rowId = u.makeObjectKey(parent, row)

	if curObj := result.Get(rowId); curObj != nil {
		obj = curObj.(*oMapType)
	} else {
		obj = new(oMapType)
		result.Set(rowId, obj)
	}

	// recurse for all embedded objects
	for _, childItem := range parent.References {
		key = u.makeObjectKey(childItem, row)
		if !obj.Has(key) {
			// If this is the first time, create the group
			newValues := new(oMapType)
			obj.Set(key, newValues)
			u.unpackObject(childItem, row, newValues)
		} else {
			// Already have a group, so add to the group
			currentValues := obj.Get(key).(*oMapType)
			u.unpackObject(childItem, row, currentValues)
		}
	}
	for leafItem := range parent.Selects.All() {
		u.unpackLeaf(leafItem, row, obj)
	}
	return
}

func (u *unpacker) unpackLeaf(j *jointree.Element, row db.ValueMap, obj *oMapType) {
	if node, ok := j.QueryNode.(*query.ColumnNode); ok {
		key := j.Alias
		fieldName := node.QueryName
		obj.Set(fieldName, row[key])
	} else {
		panic("Unexpected node type.")
	}
}

// makeObjectKey makes the key for the object of the row.
// The key is used in subsequent calls to determine what row joined data belongs to.
func (u *unpacker) makeObjectKey(j *jointree.Element, row db.ValueMap) string {
	pk := j.PrimaryKey()

	if pk == nil || pk.Alias == "" {
		// We are not identifying the row by a PK because of one of the following:
		// 1) This is a distinct select, and we are not selecting pks to avoid affecting the results of the query
		// 2) This is a groupby clause, which forces us to select only the groupby items and we cannot add a PK to the row
		// We will therefore make up a unique key to identify the row
		u.rowId++
		return strconv.Itoa(u.rowId)
	}

	v := row[pk.Alias]
	if v == nil {
		panic(fmt.Sprintf("expected value for %s was not returned in the query", pk))
	}

	return pk.Alias + "." + fmt.Sprint(v)
}

func (u *unpacker) unpackSpecialAliases(rowId string, row db.ValueMap, aliasMap *oMapType) {
	if curObj := aliasMap.Get(rowId); curObj != nil {
		return // already added these to the row
	}

	obj := new(oMapType)
	for key, _ := range u.jt.Aliases {
		obj.Set(key, row[key])
	}

	if obj.Len() > 0 {
		aliasMap.Set(rowId, obj)
	}
}

// expand converts the omap into an array of maps.
// If j contains is expanded, then more than one item will be returned.
func (u *unpacker) expand(j *jointree.Element, nodeObject *oMapType) (outArray []db.ValueMap) {
	var item db.ValueMap
	var innerNodeObject *oMapType
	var copies []db.ValueMap
	var innerCopies []db.ValueMap
	var newArray []db.ValueMap

	outArray = append(outArray, db.NewValueMap())

	// order of reference or leaf processing is not important
	for el := range j.Selects.All() {
		for _, item = range outArray {
			if cn, ok := el.QueryNode.(*query.ColumnNode); ok {
				item[cn.QueryName] = nodeObject.Get(cn.QueryName)
			}
		}
	}

	for _, el := range j.References {
		copies = []db.ValueMap{}
		tableName := el.QueryNode.TableName_()

		for _, item = range outArray {
			switch el.QueryNode.NodeType_() {
			case query.ReferenceNodeType:
				// Should be a one or zero item array here
				om := nodeObject.Get(tableName).(*oMapType)
				if om.Len() > 1 {
					panic("Cannot have an array with more than one item here.")
				} else if om.Len() == 1 {
					innerNodeObject = nodeObject.Get(tableName).(*oMapType).GetAt(0).(*oMapType)
					innerCopies = u.expand(el, innerNodeObject)
					if len(innerCopies) > 1 {
						for _, cp2 := range innerCopies {
							nodeCopy := item.Copy().(db.ValueMap)
							nodeCopy[tableName] = cp2
							copies = append(copies, nodeCopy)
						}
					} else {
						item[tableName] = map[string]interface{}(innerCopies[0])
					}
				}
				// else we likely were not included because of a conditional join
			case query.ReverseNodeType:
				if el.Expanded { // unique reverse or single expansion many
					newArray = []db.ValueMap{}
					for _, value := range nodeObject.Get(tableName).(*oMapType).All() {
						innerNodeObject = value.(*oMapType)
						innerCopies = u.expand(el, innerNodeObject)
						for _, ic := range innerCopies {
							newArray = append(newArray, ic)
						}
					}
					for _, cp2 := range newArray {
						nodeCopy := item.Copy().(db.ValueMap)
						nodeCopy[tableName] = cp2
						copies = append(copies, nodeCopy)
					}
				} else {
					// From this point up, we should not be creating additional copies, since from this point down, we
					// are gathering an array
					newArray = []db.ValueMap{}
					for _, value := range nodeObject.Get(tableName).(*oMapType).All() {
						innerNodeObject = value.(*oMapType)
						innerCopies = u.expand(el, innerNodeObject)
						for _, ic := range innerCopies {
							newArray = append(newArray, ic)
						}
					}
					item[tableName] = newArray
				}

			case query.ManyManyNodeType:
				newArray = []db.ValueMap{}
				for _, value := range nodeObject.Get(tableName).(*oMapType).All() {
					innerNodeObject = value.(*oMapType)
					innerCopies = u.expand(el, innerNodeObject)
					for _, ic := range innerCopies {
						newArray = append(newArray, ic)
					}
				}
				if !el.Expanded {
					item[tableName] = newArray
				} else {
					for _, cp2 := range newArray {
						nodeCopy := item.Copy().(db.ValueMap)
						nodeCopy[tableName] = []db.ValueMap{cp2}
						copies = append(copies, nodeCopy)
					}
				}
			}

		}
		if len(copies) > 0 {
			outArray = copies
		}
	}
	return
}
