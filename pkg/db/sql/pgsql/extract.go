package pgsql

import (
	"database/sql"
	"fmt"
	"github.com/goradd/iter"
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
	fkMap   map[string][]pgForeignKey
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
	constraintName       string
	tableName            string
	columnName           string
	referencedTableName  string
	referencedColumnName sql.NullString
}

func (m *pgTable) findForeignKeyGroupByColumn(col string) []pgForeignKey {
	for _, group := range m.fkMap {
		for _, fk := range group {
			if fk.columnName == col {
				return group
			}
		}
	}
	return nil
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

		// Place foreign keys by table, since they are database wide
		for fkName, fkGroup := range foreignKeys {
			if len(fkGroup) > 1 {
				slog.Warn("Multi-column foreign key skipped.",
					slog.String(db.LogTable, table.name),
					slog.String("name", fkName),
					slog.String(db.LogComponent, "extract"))
				continue
			}
			if fkGroup[0].tableName == table.name &&
				fkGroup[0].referencedColumnName.Valid &&
				fkGroup[0].referencedTableName != "" {
				table.fkMap[fkName] = fkGroup
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
			fkMap:   make(map[string][]pgForeignKey),
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
func (m *DB) getForeignKeys(schemas []string, defaultSchemaName string) (foreignKeys map[string][]pgForeignKey) {
	foreignKeys = make(map[string][]pgForeignKey)

	stmt := fmt.Sprintf(`
SELECT
    pc.conname AS constraint_name,
    cl.relname AS table_name,
    nsp.nspname AS table_schema,
    att.attname AS column_name,
    fcl.relname AS foreign_table_name,
    fnsp.nspname AS foreign_table_schema,
    fatt.attname AS foreign_column_name
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
			&fk.referencedColumnName)

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
		fks := foreignKeys[fk.constraintName]
		fks = append(fks, fk)
		foreignKeys[fk.constraintName] = fks
	}
	return
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

	for tableName, rawTable := range iter.KeySort(rawTables) {
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
	var referenceSchemas []*schema.Reference
	var multiColumnPK *schema.Index

	// Build the indexes
	indexes := make(map[string]*schema.Index)
	singleIndexes := make(map[string]schema.IndexLevel)

	for _, idx := range t.indexes {
		if i, ok := indexes[idx.name]; ok {
			// add a column to the previously found index
			i.Columns = append(i.Columns, idx.columnName)
		} else {
			// create a new index
			var level schema.IndexLevel
			if idx.primary {
				level = schema.IndexLevelPrimaryKey
			} else if idx.unique {
				level = schema.IndexLevelUnique
			} else {
				level = schema.IndexLevelIndexed
			}
			mci := &schema.Index{Columns: []string{idx.columnName}, IndexLevel: level}
			indexes[idx.name] = mci
		}
	}

	// Fill the singleIndexes set with all the columns that have a single index,
	// There might be multiple single indexes on the same column, but if there are
	// we prioritize by the value of the index level.
	// We don't support the esoteric index types yet.
	for _, idx := range indexes {
		if len(idx.Columns) == 1 {
			if level, ok := singleIndexes[idx.Columns[0]]; ok {
				if idx.IndexLevel > level {
					singleIndexes[idx.Columns[0]] = idx.IndexLevel
				}
			} else {
				singleIndexes[idx.Columns[0]] = idx.IndexLevel
			}
		} else if idx.IndexLevel == schema.IndexLevelPrimaryKey {
			// We have a multi-column primary key
			multiColumnPK = idx
		}
	}

	var pkCount int

	for _, col := range t.columns {
		cd, rd := m.getColumnSchema(t, col, singleIndexes[col.name], enumTableSuffix)

		if rd != nil {
			referenceSchemas = append(referenceSchemas, rd)
		} else if cd.Type == schema.ColTypeAutoPrimaryKey ||
			cd.IndexLevel == schema.IndexLevelPrimaryKey ||
			multiColumnPK != nil && slices.Contains(multiColumnPK.Columns, col.name) {
			// private keys go first
			columnSchemas = slices.Insert(columnSchemas, pkCount, cd)
			pkCount++
		} else {
			columnSchemas = append(columnSchemas, cd)
		}
	}

	td := schema.Table{
		Name:       t.name,
		Schema:     t.schema,
		Columns:    columnSchemas,
		Comment:    t.comment,
		References: referenceSchemas,
	}

	// Create the  index array
	for _, idx := range indexes {
		if len(idx.Columns) > 1 {
			// only do multi-column indexes, since single column indexes should be specified in the column definition
			td.Indexes = append(td.Indexes, *idx)
		}
	}

	// Keep the Indexes in a predictable order
	slices.SortFunc(td.Indexes, func(m1 schema.Index, m2 schema.Index) int {
		return slices.Compare(m1.Columns, m2.Columns)
	})

	return td
}

func (m *DB) getEnumTableSchema(t pgTable) (ed schema.EnumTable, err error) {
	td := m.getTableSchema(t, "")

	var columnNames []string
	var quotedColumns []string
	var receiverTypes []ReceiverType

	if len(td.Columns) < 2 {
		err = fmt.Errorf("error: An enum table must have at least 2 columns")
		return
	}

	ed.Name = td.Name
	ed.Fields = make(map[string]schema.EnumField)

	var hasValue bool
	var hasName bool

	if td.References != nil {
		err = fmt.Errorf("cannot have references in an enum table")
		return
	}

	for _, c := range td.Columns {
		if c.Name == schema.ValueKey {
			hasValue = true
		} else if c.Name == schema.NameKey {
			hasName = true
		}

		columnNames = append(columnNames, c.Name)
		quotedColumns = append(quotedColumns, m.QuoteIdentifier(c.Name))
		recType := ReceiverTypeFromSchema(c.Type, c.Size)
		typ := c.Type

		if c.Name == schema.ValueKey && c.Type == schema.ColTypeAutoPrimaryKey {
			recType = ColTypeInteger
			typ = schema.ColTypeInt
		} else if c.Type == schema.ColTypeUnknown {
			recType = ColTypeBytes
			typ = schema.ColTypeBytes
		}
		receiverTypes = append(receiverTypes, recType)
		ft := schema.EnumField{
			Type: typ,
		}
		ed.Fields[c.Name] = ft
	}

	if !hasValue {
		err = fmt.Errorf(`an enum table must have a "value" column`)
		return
	}
	if !hasName {
		err = fmt.Errorf(`an enum table must have a "name" `)
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
		strings.Join(quotedColumns, `,`),
		m.QuoteIdentifier(td.Name),
		quotedColumns[0])

	result, err := m.SqlDb().Query(stmt)
	if err != nil {
		panic(err)
	}
	defer sql2.RowClose(result)

	receiver, err2 := sql2.ReceiveRows(result, receiverTypes, columnNames, nil, stmt, nil)
	if err2 != nil {
		panic(err2)
	}
	for i, row := range receiver {
		values := make(map[string]any)
		for k := range ed.Fields {
			if k == schema.ValueKey {
				i2, _ := row[k]
				if i+1 != i2 {
					// only if value is not the default, then include it in the value map
					values[k] = i2
				}
			} else {
				values[k] = row[k]
			}
		}
		ed.Values = append(ed.Values, values)
	}

	ed.Comment = t.comment
	delete(ed.Fields, schema.ValueKey)
	delete(ed.Fields, schema.NameKey)
	if len(ed.Fields) == 0 {
		ed.Fields = nil
	}

	return
}

func (m *DB) getColumnSchema(table pgTable,
	column pgColumn,
	indexLevel schema.IndexLevel,
	enumTableSuffix string) (columnSchema *schema.Column, refSchema *schema.Reference) {

	columnSchema = &schema.Column{
		Name: column.name,
	}
	columnSchema.Type, columnSchema.Size, columnSchema.DefaultValue = processTypeInfo(column)

	// treat auto incrementing values as id values
	if column.isAutoIncrement && column.isIdentity {
		columnSchema.Type = schema.ColTypeAutoPrimaryKey
		columnSchema.Size = 0
	} else {
		columnSchema.IndexLevel = indexLevel
	}

	columnSchema.IsNullable = column.isNullable

	fkGroup := table.findForeignKeyGroupByColumn(columnSchema.Name)
	if len(fkGroup) > 1 {
		slog.Warn("Multi-column foreign keys are not currently supported.",
			slog.String("Constraint", fkGroup[0].constraintName))
	} else if len(fkGroup) == 1 {
		fk := fkGroup[0]
		if fk.referencedTableName == "" {
			slog.Error("Foreign key reference is empty.",
				slog.String("Constraint", fkGroup[0].constraintName))
		} else {
			if enumTableSuffix != "" && strings.HasSuffix(fk.referencedTableName, enumTableSuffix) {
				// assume enum table table exists
				columnSchema.Type = schema.ColTypeEnum
				columnSchema.Size = 0
				columnSchema.EnumTable = fk.referencedTableName
			} else {
				if indexLevel != schema.IndexLevelUnique {
					// IndexLevelIndexed is default for references, so setting to None will preserve that, but also simplify schema file.
					indexLevel = schema.IndexLevelNone
				}
				refSchema = &schema.Reference{
					Table:      fk.referencedTableName,
					Column:     fk.columnName,
					IndexLevel: indexLevel,
					IsNullable: columnSchema.IsNullable,
				}
				columnSchema = nil
			}
		}
	}

	if columnSchema != nil {
		columnSchema.Comment = column.comment
		if column.collationName.Valid && column.collationName.String != "" {
			columnSchema.DatabaseDefinition = map[string]map[string]any{db.DriverTypePostgres: {"collation": column.collationName.String}}
		}
	}

	return
}

func (m *DB) getAssociationSchema(t pgTable, enumTableSuffix string) (mm schema.AssociationTable, err error) {
	td := m.getTableSchema(t, enumTableSuffix)
	if len(td.References) != 2 {
		err = fmt.Errorf("association table must have 2 foreign keys")
		return
	}
	for _, ref := range td.References {
		if ref.IsNullable {
			err = fmt.Errorf("column " + ref.Column + " cannot be nullable.")
			return
		}
	}
	mm.Table = td.Name
	mm.Ref1.Table = td.References[0].Table
	mm.Ref1.Column = td.References[0].Column
	mm.Ref2.Table = td.References[1].Table
	mm.Ref2.Column = td.References[1].Column
	mm.Comment = t.comment
	return
}
