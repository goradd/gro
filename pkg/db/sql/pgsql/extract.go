package pgsql

import (
	"database/sql"
	"fmt"
	iter2 "github.com/goradd/iter"
	"github.com/goradd/maps"
	"github.com/goradd/orm/pkg/db"
	sql2 "github.com/goradd/orm/pkg/db/sql"
	. "github.com/goradd/orm/pkg/query"
	"github.com/goradd/orm/pkg/schema"
	strings2 "github.com/goradd/strings"
	"log"
	"log/slog"
	"slices"
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
	isAutoIncrement bool
	comment         string
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
	constraintName          string
	tableName               string
	columnName              string
	referencedTableName     string
	referencedColumnName    sql.NullString
	updateRule              sql.NullString
	deleteRule              sql.NullString
	referencedColumnKeyType sql.NullString
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

	schemas, ok := options["schemas"].([]string)
	if !ok {
		schemas = []string{defaultSchemaName}
	}
	tables, schemas2 := m.getTables(schemas, defaultSchemaName)

	indexes := m.getIndexes(schemas2, defaultSchemaName)

	foreignKeys := m.getForeignKeys(schemas2, defaultSchemaName)

	for _, table := range tables {
		tableIndex := table.name
		if table.schema != "" {
			tableIndex = table.schema + "." + table.name
		}

		// Do some processing on the foreign keys
		for _, fk := range foreignKeys[tableIndex] {
			if fk.referencedColumnName.Valid && fk.referencedTableName != "" {
				if _, ok := table.fkMap[fk.columnName]; ok {
					slog.Warn("Multi-table foreign keys are not supported.",
						slog.String(db.LogTable, table.name),
						slog.String(db.LogColumn, fk.columnName))
					delete(table.fkMap, fk.columnName)
				} else {
					table.fkMap[fk.columnName] = fk
				}
			}
		}

		columns := m.getColumns(table.name, table.schema)

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
		panic(err)
	}
	defer sql2.RowClose(rows)
	for rows.Next() {
		err = rows.Scan(&tableName, &tableSchema, &tableComment)
		if err != nil {
			log.Fatal(err)
		}
		slog.Info("Importing schema for table "+tableSchema+"."+tableName,
			slog.String(db.LogTable, tableName))
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
		panic(err)
	}

	return tables, schemaMap.Values()
}

func (m *DB) getColumns(table string, schema string) (columns []pgColumn) {

	s := fmt.Sprintf(`
	SELECT
	c.column_name,
	c.column_default,
	c.is_nullable,
	c.data_type,
	c.character_maximum_length,
	c.is_identity,
	c.identity_generation,
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
		panic(err)
	}
	defer sql2.RowClose(rows)
	var col pgColumn

	for rows.Next() {
		col = pgColumn{}
		var descr sql.NullString
		var nullable sql2.SqlReceiver
		var ident sql2.SqlReceiver
		var identGen sql2.SqlReceiver

		err = rows.Scan(
			&col.name,
			&(col.defaultValue.R),
			&(nullable.R),
			&col.dataType,
			&col.characterMaxLen,
			&(ident.R),
			&(identGen.R),
			&descr,
			&col.collationName,
		)
		if err != nil {
			panic(err)
		}
		col.isNullable = nullable.BoolI().(bool)
		col.isIdentity = ident.BoolI().(bool)
		if descr.Valid {
			col.comment = descr.String
		}

		if s, _ := col.defaultValue.StringI().(string); strings.Contains(s, "nextval") {
			col.isIdentity = true
			col.isAutoIncrement = true
		} else if identGen.StringI() != nil && identGen.StringI().(string) != "" {
			if identGen.StringI().(string) == "ALWAYS" {
				slog.Warn("Identity column was created using GENERATED ALWAYS. This will prevent the ORM from being able to import records. Create the column with GENERATED BY DEFAULT instead.",
					slog.String(db.LogTable, table),
					slog.String(db.LogColumn, col.name))
			}
			col.isAutoIncrement = true
		}
		columns = append(columns, col)
	}
	err = rows.Err()
	if err != nil {
		panic(err)
	}

	return columns
}

func (m *DB) getIndexes(schemas []string, defaultSchemaName string) (indexes map[string][]pgIndex) {

	indexes = make(map[string][]pgIndex)

	s := fmt.Sprintf(`
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

	rows, err := m.SqlDb().Query(s)

	if err != nil {
		panic(err)
	}
	defer sql2.RowClose(rows)
	var index pgIndex

	for rows.Next() {
		index = pgIndex{}
		err = rows.Scan(&index.name, &index.schema, &index.tableName, &index.tableSchema, &index.unique, &index.primary, &index.columnName)
		if err != nil {
			panic(err)
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
		panic(err)
	}

	return indexes
}

// getForeignKeys gets information on the foreign keys.
//
// Note that querying the information_schema database is SLOW, so we want to do it as few times as possible.
func (m *DB) getForeignKeys(schemas []string, defaultSchemaName string) (foreignKeys map[string][]pgForeignKey) {
	fkMap := make(map[string]pgForeignKey)

	stmt := fmt.Sprintf(`
SELECT
    pc.conname AS constraint_name,
    cl.relname AS table_name,
    nsp.nspname AS table_schema,
    att.attname AS column_name,
    fcl.relname AS foreign_table_name,
    fnsp.nspname AS foreign_table_schema,
    fatt.attname AS foreign_column_name,
    pc.confdeltype,
    pc.confupdtype,
    fpc.contype AS foreign_column_contype
FROM 
    pg_constraint pc
-- Join to get the table and schema of the constraint
JOIN pg_class cl 
    ON cl.oid = pc.conrelid
JOIN pg_namespace nsp 
    ON nsp.oid = cl.relnamespace
-- Join to get the constrained column names
JOIN unnest(pc.conkey) WITH ORDINALITY AS conkey (attnum, ord) 
    ON TRUE
JOIN pg_attribute att
    ON att.attnum = conkey.attnum
    AND att.attrelid = cl.oid
-- Join to get the foreign table and column if it is a foreign key
LEFT JOIN pg_class fcl 
    ON fcl.oid = pc.confrelid
LEFT JOIN pg_namespace fnsp 
    ON fnsp.oid = fcl.relnamespace
LEFT JOIN unnest(pc.confkey) WITH ORDINALITY AS confkey (attnum, ord) 
    ON confkey.ord = conkey.ord
LEFT JOIN pg_attribute fatt
    ON fatt.attnum = confkey.attnum
    AND fatt.attrelid = fcl.oid
-- Join to get the constraint type of the foreign key column
LEFT JOIN pg_constraint fpc
    ON fpc.conrelid = pc.confrelid
    AND confkey.attnum = ANY(fpc.conkey)
WHERE 
    pc.contype = 'f' -- Foreign keys only
    AND nsp.nspname IN ('%s'); 
	`, strings.Join(schemas, "','"))

	rows, err := m.SqlDb().Query(stmt)
	if err != nil {
		panic(err)
	}

	defer sql2.RowClose(rows)

	var tableSchema string
	var referencedSchema sql.NullString
	var referencedTable sql.NullString

	for rows.Next() {
		fk := pgForeignKey{}
		err = rows.Scan(&fk.constraintName,
			&fk.tableName,
			&tableSchema,
			&fk.columnName,
			&referencedTable,
			&referencedSchema,
			&fk.referencedColumnName,
			&fk.updateRule,
			&fk.deleteRule,
			&fk.referencedColumnKeyType)

		if err != nil {
			panic(err)
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

	foreignKeys = make(map[string][]pgForeignKey)
	for _, fk := range iter2.KeySort(fkMap) {
		i := fk.tableName
		tableKeys := foreignKeys[i]
		tableKeys = append(tableKeys, fk)
		foreignKeys[i] = tableKeys
	}
	return foreignKeys
}

// Convert the database native type to a more generic sql type, and a go table type.
func processTypeInfo(column pgColumn) (
	typ schema.ColumnType,
	maxLength uint64,
	defaultValue interface{}) {

	switch column.dataType {
	case "time without time zone",
		"time", "timestamp", "timestamp with time zone",
		"datetime", "timestamp without time zone", "date":
		typ = schema.ColTypeTime

	case "boolean":
		typ = schema.ColTypeBool

	case "integer", "int":
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
		// no max length

	case "text":
		typ = schema.ColTypeString
		// no max length

	case "numeric":
		// No native equivalent in Go.
		// See the shopspring/decimal package for support.
		// You will need to shepherd numbers into and out of []byte format to move data to the database.
		typ = schema.ColTypeUnknown
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
		EnumTableSuffix: options["enum_table_suffix"].(string),
		AssnTableSuffix: options["assn_table_suffix"].(string),
		Key:             m.DbKey(),
	}

	// Database wide setting to limit database write times through a context timeout in generated code
	if v, ok := options["context_write_timeout"]; ok {
		dd.WriteTimeout = v.(string)
	}
	// Database wide setting to limit database read times through a context timeout in generated code
	if v, ok := options["context_read_timeout"]; ok {
		dd.ReadTimeout = v.(string)
	}

	for tableName, rawTable := range rawTables {
		if strings.Contains(rawTable.name, ".") {
			slog.Warn("Table name "+rawTable.name+"cannot contain a period in its name. Skipping.",
				slog.String(db.LogTable, tableName))
			continue
		}
		if strings.Contains(rawTable.schema, ".") {
			slog.Warn("schema "+rawTable.schema+" cannot contain a period in its schema name. Skipping.",
				slog.String(db.LogTable, tableName))
			continue
		}

		if strings2.EndsWith(tableName, dd.EnumTableSuffix) {
			if t, err := m.getEnumTableSchema(rawTable); err != nil {
				slog.Warn("Error in enum rawTable. Skipping.",
					slog.String(db.LogTable, tableName),
					slog.Any(db.LogTable, err.Error()))
			} else {
				dd.EnumTables = append(dd.EnumTables, &t)
			}
		} else if strings2.EndsWith(tableName, dd.AssnTableSuffix) {
			if mm, err := m.getAssociationSchema(rawTable, dd.EnumTableSuffix); err != nil {
				slog.Warn("Error in association rawTable. Skipping.",
					slog.String(db.LogTable, tableName),
					slog.Any(db.LogError, err))
			} else {
				dd.AssociationTables = append(dd.AssociationTables, &mm)
			}
		} else {
			t := m.getTableSchema(rawTable, dd.EnumTableSuffix)
			dd.Tables = append(dd.Tables, &t)
		}
	}
	return dd
}

func (m *DB) getTableSchema(t pgTable, enumTableSuffix string) schema.Table {
	var columnSchemas []*schema.Column
	var multiColumnPK *schema.MultiColumnIndex

	// Build the indexes
	indexes := make(map[string]*schema.MultiColumnIndex)
	singleIndexes := make(map[string]schema.IndexLevel)

	for _, idx := range t.indexes {
		if i, ok := indexes[idx.name]; ok {
			// add a column to the previously found index
			i.Columns = append(i.Columns, idx.columnName)
		} else {
			// create a new index
			var level schema.IndexLevel
			if idx.primary {
				level = schema.IndexLevelManualPrimaryKey
			} else if idx.unique {
				level = schema.IndexLevelUnique
			} else {
				level = schema.IndexLevelIndexed
			}
			mci := &schema.MultiColumnIndex{Columns: []string{idx.columnName}, IndexLevel: level}
			indexes[idx.name] = mci
		}
	}

	// Fill the singleIndexes set with all the columns that have a single index,
	// There should not be multiple single indexes on the same column, but if there are
	// we prioritize by the value of the index level.
	for _, idx := range indexes {
		if len(idx.Columns) == 1 {
			if level, ok := singleIndexes[idx.Columns[0]]; ok {
				if idx.IndexLevel > level {
					singleIndexes[idx.Columns[0]] = idx.IndexLevel
				}
			} else {
				singleIndexes[idx.Columns[0]] = idx.IndexLevel
			}
		} else if idx.IndexLevel == schema.IndexLevelManualPrimaryKey {
			// We have a multi-column primary key
			multiColumnPK = idx
		}
	}

	var pkCount int

	// Since we don't support multi-column primary keys in the ORM, we convert to
	// an auto-generated primary key plus a unique index.
	if multiColumnPK != nil {
		cd := &schema.Column{
			Name:    "_id",
			Type:    schema.ColTypeAutoPrimaryKey,
			Comment: "Replacing multi-column primary key",
		}
		columnSchemas = append(columnSchemas, cd)
		multiColumnPK.IndexLevel = schema.IndexLevelUnique
		pkCount = 1
		// these columns should already be identified as not nullable
	}

	for _, col := range t.columns {
		if strings.Contains(col.name, ".") {
			slog.Warn("Column cannot contain a period in its name. Skipping.",
				slog.String(db.LogTable, t.name),
				slog.String(db.LogColumn, col.name))
			continue
		}

		cd := m.getColumnSchema(t, col, singleIndexes[col.name], enumTableSuffix)

		if cd.Type == schema.ColTypeAutoPrimaryKey ||
			cd.IndexLevel == schema.IndexLevelManualPrimaryKey ||
			multiColumnPK != nil && slices.Contains(multiColumnPK.Columns, col.name) {
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
		Comment: t.comment,
	}

	// Create the multi-column index array
	for _, idx := range indexes {
		if len(idx.Columns) > 1 {
			td.MultiColumnIndexes = append(td.MultiColumnIndexes, *idx)
		}
	}

	// Keep the MultiColumnIndexes in a predictable order
	slices.SortFunc(td.MultiColumnIndexes, func(m1 schema.MultiColumnIndex, m2 schema.MultiColumnIndex) int {
		return slices.Compare(m1.Columns, m2.Columns)
	})

	return td
}

func (m *DB) getEnumTableSchema(t pgTable) (ed schema.EnumTable, err error) {
	td := m.getTableSchema(t, "")

	var columnNames []string
	var quotedNames []string
	var receiverTypes []ReceiverType

	if len(td.Columns) < 2 {
		err = fmt.Errorf("error: An enum table must have at least 2 columns")
		return
	}

	ed.Name = td.Name
	ed.Fields = make(map[string]schema.EnumField)

	var hasConst bool
	var hasLabelOrIdentifier bool

	for _, c := range td.Columns {
		if c.Name == schema.ConstKey {
			hasConst = true
		} else if c.Name == schema.LabelKey {
			hasLabelOrIdentifier = true
		} else if c.Name == schema.IdentifierKey {
			hasLabelOrIdentifier = true
		}
		if c.Type == schema.ColTypeReference {
			err = fmt.Errorf("cannot have a reference column in an enum table")
			return
		}

		columnNames = append(columnNames, c.Name)
		quotedNames = append(quotedNames, m.QuoteIdentifier(c.Name))
		recType := ReceiverTypeFromSchema(c.Type, c.Size)
		if c.Name == schema.ConstKey && c.Type == schema.ColTypeAutoPrimaryKey {
			recType = ColTypeInteger
		} else if c.Type == schema.ColTypeUnknown {
			recType = ColTypeBytes
		}
		receiverTypes = append(receiverTypes, recType)
		ft := schema.EnumField{
			Type: c.Type,
		}
		ed.Fields[c.Name] = ft
	}

	if !hasConst {
		err = fmt.Errorf(`an enum table must have a "const"" column`)
		return
	}
	if !hasLabelOrIdentifier {
		err = fmt.Errorf(`an enum table must have a "label" of "identifier" column`)
		return
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
		m.QuoteIdentifier(td.Name),
		quotedNames[0])

	result, err := m.SqlDb().Query(stmt)
	if err != nil {
		panic(err)
	}
	defer sql2.RowClose(result)

	receiver, err2 := sql2.ReceiveRows(result, receiverTypes, columnNames, nil, stmt, nil)
	if err2 != nil {
		panic(err2)
	}
	for _, row := range receiver {
		values := make(map[string]any)
		for k := range ed.Fields {
			values[k] = row[k]
		}
		ed.Values = append(ed.Values, values)
	}
	return
}

func (m *DB) getColumnSchema(table pgTable,
	column pgColumn,
	indexLevel schema.IndexLevel,
	enumTableSuffix string) schema.Column {
	cd := schema.Column{
		Name: column.name,
	}
	cd.Type, cd.Size, cd.DefaultValue = processTypeInfo(column)

	// treat auto incrementing values as id values
	if column.isAutoIncrement && column.isIdentity {
		cd.Type = schema.ColTypeAutoPrimaryKey
		cd.Size = 0
	} else {
		cd.IndexLevel = indexLevel
	}

	cd.IsNullable = column.isNullable

	if enumTableSuffix != "" && strings.HasSuffix(column.name, enumTableSuffix) { // handle enum type
		if cd.Type == schema.ColTypeInt || cd.Type == schema.ColTypeUint {
			cd.Type = schema.ColTypeEnum
			if fk, ok := table.fkMap[column.name]; ok { // if it is a foreign key to the enum table, use that
				cd.Reference = &schema.Reference{
					Table: fk.referencedTableName,
				}
			} else {
				slog.Warn("Could not find foreign key to enum table.",
					slog.String(db.LogComponent, "extract"),
					slog.String(db.LogTable, table.name),
					slog.String(db.LogColumn, column.name))
			}
		} else {
			slog.Error("Enum type should be an integer",
				slog.String(db.LogComponent, "extract"),
				slog.String(db.LogTable, table.name),
				slog.String(db.LogColumn, column.name))
		}
		cd.Size = 0 // use default size
	} else if fk, ok2 := table.fkMap[cd.Name]; ok2 {
		tableName := fk.referencedTableName
		if fk.referencedColumnKeyType.String != "p" {
			slog.Warn("Foreign key appears to NOT be pointing to a primary key. Only primary key foreign keys are supported.",
				slog.String(db.LogTable, table.name),
				slog.String(db.LogColumn, column.name))
		}
		cd.Reference = &schema.Reference{
			Table: tableName,
		}
		cd.Type = schema.ColTypeReference
		cd.Size = 0
	}

	cd.Comment = column.comment
	if column.collationName.Valid && column.collationName.String != "" {
		cd.DatabaseDefinition = map[string]map[string]any{db.DriverTypePostgres: {"collation": column.collationName.String}}
	}

	return cd
}

func (m *DB) getAssociationSchema(t pgTable, enumTableSuffix string) (mm schema.AssociationTable, err error) {
	td := m.getTableSchema(t, enumTableSuffix)
	if len(td.Columns) == 0 {
		err = fmt.Errorf("association table must have 2 columns")
		return
	}
	if td.Columns[0].Name == "_id" {
		// This column was added because of a multi-column primary key.
		td.Columns = slices.Delete(td.Columns, 0, 1)
	}
	if len(td.Columns) != 2 {
		err = fmt.Errorf("association table must have only 2 columns")
		return
	}
	for _, cd := range td.Columns {
		if cd.Reference == nil {
			err = fmt.Errorf("column " + cd.Name + " must be a foreign key.")
			return
		}

		if cd.IsNullable {
			err = fmt.Errorf("column " + cd.Name + " cannot be nullable.")
			return
		}
	}

	mm.Name = td.Name
	mm.Table1 = td.Columns[0].Reference.Table
	mm.Name1 = td.Columns[0].Name
	mm.Table2 = td.Columns[1].Reference.Table
	mm.Name2 = td.Columns[1].Name
	return
}
