package pgsql

import (
	"database/sql"
	"fmt"
	iter2 "github.com/goradd/iter"
	"github.com/goradd/maps"
	strings2 "github.com/goradd/strings"
	"log"
	log2 "log/slog"
	"slices"
	"sort"
	sql2 "spekary/goradd/orm/pkg/db/sql"
	. "spekary/goradd/orm/pkg/query"
	"spekary/goradd/orm/pkg/schema"
	"strings"
)

/*
This file contains the code that parses the data structure found in a Postgresql database into
our own cross-platform internal database description object.
*/

type pgTable struct {
	name    string
	schema  string
	columns []pgColumn
	indexes []pgIndex
	fkMap   map[string]pgForeignKey
	comment string
	options map[string]interface{}
}

type pgColumn struct {
	name            string
	defaultValue    sql2.SqlReceiver
	isNullable      bool
	dataType        string
	charLen         int
	characterMaxLen sql.NullInt64
	collationName   sql.NullString
	isIdentity      bool
	comment         string
	options         map[string]interface{}
}

type pgIndex struct {
	name        string
	schema      string
	unique      bool
	primary     bool
	tableName   string
	tableSchema string
	columnName  string
}

type pgForeignKey struct {
	constraintName       string
	tableName            string
	columnName           string
	referencedTableName  string
	referencedColumnName sql.NullString
	updateRule           sql.NullString
	deleteRule           sql.NullString
}

func (m *DB) ExtractSchema(options map[string]any) schema.Database {
	rawTables := m.getRawTables(options)
	return m.schemaFromRawTables(rawTables, options)
}

func (m *DB) getRawTables(options map[string]any) map[string]pgTable {
	var tableMap = make(map[string]pgTable)

	defaultSchemaName := "public"
	if d, ok := options["default_schema"].(string); ok && d != "" {
		defaultSchemaName = d
	}

	schemas, _ := options["schemas"].([]string)
	tables, schemas2 := m.getTables(schemas, defaultSchemaName)

	indexes, err := m.getIndexes(schemas2, defaultSchemaName)
	if err != nil {
		return nil
	}

	foreignKeys, err := m.getForeignKeys(schemas2, defaultSchemaName)
	if err != nil {
		return nil
	}

	for _, table := range tables {
		tableIndex := table.name
		if table.schema != "" {
			tableIndex = table.schema + "." + table.name
		}

		// Do some processing on the foreign keys
		for _, fk := range foreignKeys[tableIndex] {
			if fk.referencedColumnName.Valid && fk.referencedTableName != "" {
				if _, ok := table.fkMap[fk.columnName]; ok {
					log2.Warn(fmt.Sprintf("Column %s:%s multi-table foreign keys are not supported.", table.name, fk.columnName))
					delete(table.fkMap, fk.columnName)
				} else {
					table.fkMap[fk.columnName] = fk
				}
			}
		}

		columns, err2 := m.getColumns(table.name, table.schema)
		if err2 != nil {
			return nil
		}

		table.indexes = indexes[tableIndex]
		table.columns = columns
		tableMap[tableIndex] = table
	}

	return tableMap

}

// Gets information for a table
func (m *DB) getTables(schemas []string, defaultSchemaName string) ([]pgTable, []string) {
	var tableName, tableSchema, tableComment string
	var tables []pgTable
	var schemaMap maps.Set[string]

	stmt := `
	SELECT
	t.table_name,
	t.table_schema,
	COALESCE(obj_description((table_schema||'.'||quote_ident(table_name))::regclass), '')
	FROM
	information_schema.tables t
	WHERE
	table_type <> 'VIEW'`

	if schemas != nil {
		stmt += fmt.Sprintf(` AND table_schema IN ('%s')`, strings.Join(schemas, `','`))
	} else {
		stmt += `AND table_schema NOT IN ('pg_catalog', 'information_schema')`
	}

	rows, err := m.SqlDb().Query(stmt)

	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&tableName, &tableSchema, &tableComment)
		if err != nil {
			log.Fatal(err)
		}
		log2.Info("Importing schema for table " + tableSchema + "." + tableName)
		schemaMap.Add(tableSchema)
		table := pgTable{
			name:    tableName,
			schema:  strings2.If(tableSchema == defaultSchemaName, "", tableSchema),
			comment: tableComment,
			columns: []pgColumn{},
			fkMap:   make(map[string]pgForeignKey),
			indexes: []pgIndex{},
		}

		tables = append(tables, table)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

	return tables, schemaMap.Values()
}

func (m *DB) getColumns(table string, schema string) (columns []pgColumn, err error) {

	s := fmt.Sprintf(`
	SELECT
	c.column_name,
	c.column_default,
	c.is_nullable,
	c.data_type,
	c.character_maximum_length,
	c.is_identity,
	pgd.description,
	c.collation_name
FROM
	information_schema.columns as c
JOIN 
	pg_catalog.pg_statio_all_tables as st
	on c.table_schema = st.schemaname
	and c.table_name = st.relname
LEFT JOIN 
	pg_catalog.pg_description pgd
	on pgd.objoid=st.relid
	and pgd.objsubid=c.ordinal_position
WHERE
	c.table_name = '%s' %s
ORDER BY
	c.ordinal_position;
	`, table, strings2.If(schema == "", "", fmt.Sprintf("AND c.table_schema = '%s'", schema)))

	rows, err := m.SqlDb().Query(s)

	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	var col pgColumn

	for rows.Next() {
		col = pgColumn{}
		var descr sql.NullString
		var nullable sql2.SqlReceiver
		var ident sql2.SqlReceiver

		err = rows.Scan(&col.name, &(col.defaultValue.R), &(nullable.R), &col.dataType, &col.characterMaxLen, &(ident.R), &descr, &col.collationName)
		if err != nil {
			log.Fatal(err)
			return nil, err
		}
		col.isNullable = nullable.BoolI().(bool)
		col.isIdentity = ident.BoolI().(bool)

		if descr.Valid {
			if col.options, col.comment, err = sql2.ExtractOptions(descr.String); err != nil {
				log2.Warn("Error in table comment options for table " + table + ":" + col.name + " - " + err.Error())
			}
		}

		if s, _ := col.defaultValue.StringI().(string); strings.Contains(s, "nextval") {
			col.isIdentity = true
		}
		columns = append(columns, col)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

	return columns, err
}

func (m *DB) getIndexes(schemas []string, defaultSchemaName string) (indexes map[string][]pgIndex, err2 error) {

	indexes = make(map[string][]pgIndex)

	sql := fmt.Sprintf(`
	select idx.relname as index_name, 
       insp.nspname as index_schema,
       tbl.relname as table_name,
       tnsp.nspname as table_schema,
	   pgi.indisunique,
	   pgi.indisprimary,
	   a.attname as column_name
from pg_index pgi
  join pg_class idx on idx.oid = pgi.indexrelid
  join pg_namespace insp on insp.oid = idx.relnamespace
  join pg_class tbl on tbl.oid = pgi.indrelid
  join pg_namespace tnsp on tnsp.oid = tbl.relnamespace
  join pg_attribute a on a.attrelid = idx.oid
where
  tnsp.nspname IN ('%s')
	`, strings.Join(schemas, "','"))

	rows, err := m.SqlDb().Query(sql)

	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	var index pgIndex

	for rows.Next() {
		index = pgIndex{}
		err = rows.Scan(&index.name, &index.schema, &index.tableName, &index.tableSchema, &index.unique, &index.primary, &index.columnName)
		if err != nil {
			log.Fatal(err)
			return nil, err
		}
		indexKey := index.tableName
		if index.schema != defaultSchemaName && index.schema != "" {
			indexKey = index.schema + "." + index.tableName
		}
		tableIndexes := indexes[indexKey]
		tableIndexes = append(tableIndexes, index)
		indexes[indexKey] = tableIndexes
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

	return indexes, err
}

// getForeignKeys gets information on the foreign keys.
//
// Note that querying the information_schema database is SLOW, so we want to do it as few times as possible.
func (m *DB) getForeignKeys(schemas []string, defaultSchemaName string) (foreignKeys map[string][]pgForeignKey, err error) {
	fkMap := make(map[string]pgForeignKey)

	stmt := fmt.Sprintf(`
SELECT
    tc.constraint_name, 
    tc.table_name, 
    tc.table_schema, 
    kcu.column_name, 
    ccu.table_schema AS foreign_table_schema,
    ccu.table_name AS foreign_table_name,
    ccu.column_name AS foreign_column_name, 
    pgc.confdeltype,
    pgc.confupdtype
FROM 
    information_schema.table_constraints AS tc 
    JOIN information_schema.key_column_usage AS kcu
      ON tc.constraint_name = kcu.constraint_name
      AND tc.table_schema = kcu.table_schema
    JOIN information_schema.constraint_column_usage AS ccu
      ON ccu.constraint_name = tc.constraint_name
      AND ccu.table_schema = tc.table_schema
    JOIN pg_constraint as pgc
      ON tc.constraint_name = pgc.conname
WHERE tc.constraint_type = 'FOREIGN KEY' AND
      tc.table_schema IN ('%s')
	`, strings.Join(schemas, "','"))

	rows, err := m.SqlDb().Query(stmt)
	if err != nil {
		log.Fatal(err)
	}

	defer rows.Close()

	var tableSchema string
	var referencedSchema sql.NullString
	var referencedTable sql.NullString

	for rows.Next() {
		fk := pgForeignKey{}
		err = rows.Scan(&fk.constraintName,
			&fk.tableName,
			&tableSchema,
			&fk.columnName,
			&referencedSchema,
			&referencedTable,
			&fk.referencedColumnName,
			&fk.updateRule,
			&fk.deleteRule)

		if err != nil {
			log.Fatal(err)
			return nil, err
		}
		fk.referencedTableName = referencedTable.String
		if tableSchema != "" && tableSchema != defaultSchemaName {
			fk.tableName = tableSchema + "." + fk.tableName
		}
		if referencedSchema.String != "" && referencedSchema.String != defaultSchemaName {
			fk.referencedTableName = referencedSchema.String + "." + fk.referencedTableName
		}
		if fk.referencedColumnName.Valid {
			fkMap[fk.constraintName] = fk
		}
	}

	rows.Close()

	foreignKeys = make(map[string][]pgForeignKey)
	for _, fk := range iter2.KeySort(fkMap) {
		i := fk.tableName
		tableKeys := foreignKeys[i]
		tableKeys = append(tableKeys, fk)
		foreignKeys[i] = tableKeys
	}
	return foreignKeys, err
}

// Convert the database native type to a more generic sql type, and a go table type.
func processTypeInfo(column pgColumn) (
	typ schema.ColumnType,
	maxLength uint64,
	defaultValue interface{},
	err error) {

	switch column.dataType {
	case "time without time zone":
		fallthrough
	case "time":
		fallthrough
	case "timestamp":
		fallthrough
	case "timestamp with time zone":
		fallthrough
	case "datetime":
		fallthrough
	case "timestamp without time zone":
		fallthrough
	case "date":
		typ = schema.ColTypeTime

	case "boolean":
		typ = schema.ColTypeBool

	case "integer":
		fallthrough
	case "int":
		typ = schema.ColTypeInt
		maxLength = 32
	case "smallint":
		typ = schema.ColTypeInt
		maxLength = 16
	case "bigint":
		typ = schema.ColTypeInt
		maxLength = 64
	case "real":
		typ = schema.ColTypeFloat
		maxLength = 32
	case "double precision":
		typ = schema.ColTypeFloat
		maxLength = 64
	case "character varying":
		typ = schema.ColTypeString
		maxLength = uint64(column.characterMaxLen.Int64)
	case "char":
		typ = schema.ColTypeString
		maxLength = uint64(column.characterMaxLen.Int64)
	case "bytea":
		typ = schema.ColTypeBytes
		maxLength = 65535

	case "text":
		typ = schema.ColTypeString
		maxLength = 65535

	case "numeric":
		// No native equivalent in Go.
		// See the shopspring/decimal package for support.
		// You will need to shepherd numbers into and out of string format to move data to the database.
		typ = schema.ColTypeString
		maxLength = uint64(column.characterMaxLen.Int64) + 3

	case "year":
		typ = schema.ColTypeInt

	default:
		typ = schema.ColTypeUnknown
	}
	defaultValue = column.defaultValue.UnpackDefaultValue(typ, int(maxLength))
	return
}

func (m *DB) schemaFromRawTables(rawTables map[string]pgTable, options map[string]any) schema.Database {
	dd := schema.Database{
		ReferenceSuffix: options["reference_suffix"].(string),
		EnumTableSuffix: options["enum_table_suffix"].(string),
		AssnTableSuffix: options["assn_table_suffix"].(string),
		Key:             m.DbKey(),
	}

	for tableName, rawTable := range rawTables {
		if strings.Contains(rawTable.name, ".") {
			log2.Warn("table " + rawTable.name + " cannot contain a period in its name. Skipping.")
			continue
		}
		if strings.Contains(rawTable.schema, ".") {
			log2.Warn("schema " + tableName + " cannot contain a period in its schema name. Skipping.")
			continue
		}

		if strings2.EndsWith(tableName, dd.EnumTableSuffix) {
			if t, err := m.getEnumTableSchema(rawTable); err != nil {
				log2.Error("Enum rawTable " + tableName + " skipped: " + err.Error())
			} else {
				dd.EnumTables = append(dd.EnumTables, &t)
			}
		} else if strings2.EndsWith(tableName, dd.AssnTableSuffix) {
			if mm, err := m.getAssociationSchema(rawTable, dd.EnumTableSuffix); err != nil {
				log2.Error("Association rawTable " + tableName + " skipped: " + err.Error())
			} else {
				dd.AssociationTables = append(dd.AssociationTables, &mm)
			}
		} else {
			t := m.getTableSchema(rawTable)
			dd.Tables = append(dd.Tables, &t)
		}
	}
	return dd
}

func (m *DB) getTableSchema(t pgTable) schema.Table {
	var columnSchemas []*schema.Column

	// Build the indexes
	indexes := make(map[string]*schema.MultiColumnIndex)
	pkColumns := maps.NewSet[string]()
	uniqueColumns := maps.NewSet[string]()
	singleIndexes := maps.NewSet[string]()

	// Fill pkColumns map with the column names of all the pk columns
	// Also file the indexes map with a list of columns for each index
	for _, idx := range t.indexes {
		if idx.primary {
			pkColumns.Add(idx.columnName)
		} else if i, ok2 := indexes[idx.name]; ok2 {
			i.Columns = append(i.Columns, idx.columnName)
			sort.Strings(i.Columns) // make sure this list stays in a predictable order each time
		} else {
			i = &schema.MultiColumnIndex{IsUnique: idx.unique, Columns: []string{idx.columnName}}
			indexes[idx.name] = i
		}
	}

	// Fill the uniqueColumns map with all the columns that have a single unique index,
	// including any PK columns. Single indexes are used to determine 1 to 1 relationships.
	for _, idx := range indexes {
		if len(idx.Columns) == 1 && idx.IsUnique {
			singleIndexes.Add(idx.Columns[0])
			if idx.IsUnique {
				uniqueColumns.Add(idx.Columns[0])
			}
		}
	}
	if pkColumns.Len() == 1 {
		uniqueColumns.Add(pkColumns.Values()[0])
	}

	var pkCount int
	for _, col := range t.columns {
		if strings.Contains(col.name, ".") {
			log2.Warn(`column "` + col.name + `" cannot contain a period in its name. Skipping.`)
			continue
		}

		cd := m.getColumnSchema(t, col, singleIndexes.Has(col.name), pkColumns.Has(col.name), uniqueColumns.Has(col.name))

		if cd.Type == schema.ColTypeAutoPrimaryKey || cd.IndexLevel == schema.IndexLevelManualPrimaryKey {
			// private keys go first
			columnSchemas = slices.Insert(columnSchemas, pkCount, &cd)
			pkCount++
		} else {
			columnSchemas = append(columnSchemas, &cd)
		}

	}

	td := schema.Table{
		Name:    t.name,
		Schema:  t.schema,
		Columns: columnSchemas,
		Key:     t.name,
	}

	// Create the multi-column index array
	for _, idx := range indexes {
		if len(idx.Columns) > 1 {
			slices.Sort(idx.Columns)
			td.MultiColumnIndexes = append(td.MultiColumnIndexes, *idx)
		}
	}
	if pkColumns.Len() == 2 {
		mc := schema.MultiColumnIndex{
			IsUnique: true,
			Columns:  pkColumns.Values(),
		}
		slices.Sort(mc.Columns)
		td.MultiColumnIndexes = append(td.MultiColumnIndexes, mc)
	}

	// Keep the MultiColumnIndexes in a predictable order
	slices.SortFunc(td.MultiColumnIndexes, func(m1 schema.MultiColumnIndex, m2 schema.MultiColumnIndex) int {
		return slices.Compare(m1.Columns, m2.Columns)
	})

	return td
}

func (m *DB) getEnumTableSchema(t pgTable) (ed schema.EnumTable, err error) {
	td := m.getTableSchema(t)

	var columnNames []string
	var quotedNames []string
	var receiverTypes []ReceiverType

	if len(td.Columns) < 2 {
		err = fmt.Errorf("error: An enum table must have at least 2 columns")
		return
	}

	if td.Columns[0].Type == schema.ColTypeAutoPrimaryKey ||
		(td.Columns[0].IndexLevel == schema.IndexLevelManualPrimaryKey &&
			(td.Columns[0].Type == schema.ColTypeInt || td.Columns[0].Type == schema.ColTypeUint)) {
	} else {
		err = fmt.Errorf("error: An enum table must have a single primary key that is an integer column")
		return
	}

	if td.Columns[1].Type == schema.ColTypeAutoPrimaryKey ||
		td.Columns[1].IndexLevel == schema.IndexLevelManualPrimaryKey {
		err = fmt.Errorf("error: An enum table must cannot have more than one primary key column")
		return
	}

	ed.Name = td.Name

	for i, c := range td.Columns {
		if c.Type == schema.ColTypeReference {
			err = fmt.Errorf("cannot have a reference column in an enum table")
			return
		}

		columnNames = append(columnNames, c.Name)
		quotedNames = append(quotedNames, iq(c.Name))
		recType := ReceiverTypeFromSchema(c.Type, c.Size)
		if i == 0 {
			recType = ColTypeInteger // Force first value to be treated like an integer
		}
		if c.Type == schema.ColTypeUnknown {
			recType = ColTypeBytes
		}
		receiverTypes = append(receiverTypes, recType)
		ft := schema.EnumField{
			Name: c.Name,
			Type: c.Type,
		}

		if ed.Name == "name" {
			if len(ed.Fields) == 0 {
				panic("1st field should be the id primary key field")
			}
			slices.Insert(ed.Fields, 1, &ft)
		} else {
			ed.Fields = append(ed.Fields, &ft)
		}
	}

	stmt := fmt.Sprintf(`
SELECT
	%s
FROM
    %s
ORDER BY
    %s
`,
		strings.Join(quotedNames, `,`),
		iq(td.Name),
		quotedNames[0])

	result, err := m.SqlDb().Query(stmt)

	if err != nil {
		log.Fatal(err)
	}

	receiver := sql2.SqlReceiveRows(result, receiverTypes, columnNames, nil)
	for _, row := range receiver {
		var values []interface{}
		for _, field := range ed.Fields {
			values = append(values, row[field.Name])
		}
		ed.Values = append(ed.Values, values)
	}
	return
}

func (m *DB) getColumnSchema(table pgTable, column pgColumn, isIndexed bool, isPk bool, isUnique bool) schema.Column {
	cd := schema.Column{
		Name: column.name,
	}
	var err error
	cd.Type, cd.Size, cd.DefaultValue, err = processTypeInfo(column)
	if err != nil {
		log2.Warn(err.Error() + ". Table = " + table.name + "; Column = " + column.name)
	}

	isAuto := strings.Contains(column.dataType, "serial")
	// treat auto incrementing values as id values
	if isAuto {
		cd.Type = schema.ColTypeAutoPrimaryKey
	} else if isPk {
		cd.IndexLevel = schema.IndexLevelManualPrimaryKey
	} else if isUnique {
		cd.IndexLevel = schema.IndexLevelUnique
	} else if isIndexed {
		cd.IndexLevel = schema.IndexLevelIndexed
	}

	cd.IsOptional = column.isNullable

	if fk, ok2 := table.fkMap[cd.Name]; ok2 {
		tableName := fk.referencedTableName

		cd.Reference = &schema.Reference{
			Table:  tableName,
			Column: fk.referencedColumnName.String,
		}
	}

	if strings.HasSuffix(column.collationName.String, "ks") {
		cd.CaseInsensitive = true
	}

	return cd
}

func (m *DB) getAssociationSchema(t pgTable, enumTableSuffix string) (mm schema.AssociationTable, err error) {
	td := m.getTableSchema(t)
	if len(td.Columns) != 2 {
		err = fmt.Errorf("table " + td.Name + " must have only 2 primary key columns.")
		return
	}
	if len(td.MultiColumnIndexes) != 1 ||
		!td.MultiColumnIndexes[0].IsUnique {
		err = fmt.Errorf("association table must have one multi-column index that is unique")
	}

	var typeIndex = -1
	for i, cd := range td.Columns {
		if cd.Reference == nil {
			err = fmt.Errorf("column " + cd.Name + " must be a foreign key.")
			return
		}

		if cd.IsOptional {
			err = fmt.Errorf("column " + cd.Name + " cannot be nullable.")
			return
		}

		if strings.HasSuffix(cd.Reference.Table, enumTableSuffix) {
			if typeIndex != -1 {
				err = fmt.Errorf("column " + cd.Name + " cannot have two foreign keys to enum tables.")
				return
			}
			typeIndex = i
		}
	}

	mm.Name = td.Name
	mm.Table1 = td.Columns[0].Reference.Table
	mm.Column1 = td.Columns[0].Name
	mm.Table2 = td.Columns[1].Reference.Table
	mm.Column2 = td.Columns[1].Name
	return
}
