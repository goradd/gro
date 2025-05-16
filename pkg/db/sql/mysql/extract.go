package mysql

import (
	"database/sql"
	"fmt"
	"github.com/goradd/iter"
	"github.com/goradd/orm/pkg/db"
	sql2 "github.com/goradd/orm/pkg/db/sql"
	. "github.com/goradd/orm/pkg/query"
	"github.com/goradd/orm/pkg/schema"
	strings2 "github.com/goradd/strings"
	"log"
	"log/slog"
	"math"
	"slices"
	"strings"
)

/*
This file contains the code that parses the data structure found in a MySQL database into
our own cross-platform internal database description object.
*/

type mysqlTable struct {
	name                string
	columns             []mysqlColumn
	indexes             []mysqlIndex
	fkMap               map[string]mysqlForeignKey
	comment             string
	supportsForeignKeys bool
}

type mysqlColumn struct {
	name            string
	defaultValue    sql2.SqlReceiver
	isNullable      string
	dataType        string
	dataLen         int
	characterMaxLen sql.NullInt64
	columnType      string
	collation       sql.NullString
	key             string
	extra           string
	comment         string
}

type mysqlIndex struct {
	name       string
	nonUnique  bool
	tableName  string
	columnName string
}

type mysqlForeignKey struct {
	constraintName            string
	tableName                 string
	columnName                string
	referencedTableName       sql.NullString
	referencedColumnName      sql.NullString
	referencedColumnIndexName sql.NullString
	updateRule                sql.NullString
	deleteRule                sql.NullString
}

func (m *DB) ExtractSchema(options map[string]any) schema.Database {
	rawTables := m.getRawTables()
	return m.schemaFromRawTables(rawTables, options)
}

func (m *DB) getRawTables() map[string]mysqlTable {
	var tableMap = make(map[string]mysqlTable)

	indexes, err := m.getIndexes()
	if err != nil {
		return nil
	}

	foreignKeys, err := m.getForeignKeys()
	if err != nil {
		return nil
	}

	tables := m.getTables()
	for _, table := range tables {
		// Do some processing on the foreign keys
		for _, fk := range foreignKeys[table.name] {
			if fk.referencedColumnName.Valid && fk.referencedTableName.Valid {
				if _, ok := table.fkMap[fk.columnName]; ok {
					slog.Warn("Multi-column foreign keys are not supported.",
						slog.String(db.LogTable, table.name),
						slog.String(db.LogColumn, fk.columnName),
						slog.String(db.LogComponent, "extract"))
					delete(table.fkMap, fk.columnName)
				} else {
					table.fkMap[fk.columnName] = fk
				}
			}
		}

		columns, err2 := m.getColumns(table.name)
		if err2 != nil {
			return nil
		}

		table.indexes = indexes[table.name]
		table.columns = columns
		tableMap[table.name] = table
	}

	return tableMap

}

// Gets information for a table
func (m *DB) getTables() []mysqlTable {
	var tableName, tableComment, tableEngine string
	var tables []mysqlTable

	// Use the MySQL5 Information Schema to get a list of all the tables in this database
	// (excluding views, etc.)
	dbName := m.databaseName

	rows, err := m.SqlDb().Query(fmt.Sprintf(`
	SELECT
	table_name,
	table_comment,
	engine
	FROM
	information_schema.tables
	WHERE
	table_type <> 'VIEW' AND
	table_schema = '%s';
	`, dbName))

	if err != nil {
		panic(err)
	}
	defer sql2.RowClose(rows)
	for rows.Next() {
		var supportsForeignKeys bool

		err = rows.Scan(&tableName, &tableComment, &tableEngine)
		if err != nil {
			panic(err)
		}
		if tableEngine == "InnoDB" {
			supportsForeignKeys = true
		}
		slog.Info("Importing schema",
			slog.String(db.LogTable, tableName))
		table := mysqlTable{
			name:                tableName,
			comment:             tableComment,
			columns:             []mysqlColumn{},
			fkMap:               make(map[string]mysqlForeignKey),
			indexes:             []mysqlIndex{},
			supportsForeignKeys: supportsForeignKeys,
		}
		tables = append(tables, table)
	}
	err = rows.Err()
	if err != nil {
		panic(err)
	}

	return tables
}

func (m *DB) getColumns(table string) (columns []mysqlColumn, err error) {
	dbName := m.databaseName

	rows, err := m.SqlDb().Query(fmt.Sprintf(`
	SELECT
	column_name,
	column_default,
	is_nullable,
	data_type,
	character_maximum_length,
	column_type,
	column_key,
	extra,
	column_comment,
	collation_name
	FROM
	information_schema.columns
	WHERE
	table_name = '%s' AND
	table_schema = '%s'
	ORDER BY
	ordinal_position;
	`, table, dbName))

	if err != nil {
		panic(err)
	}
	defer sql2.RowClose(rows)
	var col mysqlColumn

	for rows.Next() {
		col = mysqlColumn{}
		err = rows.Scan(&col.name, &col.defaultValue.R, &col.isNullable, &col.dataType, &col.characterMaxLen, &col.columnType, &col.key, &col.extra, &col.comment, &col.collation)
		if err != nil {
			panic(err)
		}

		columns = append(columns, col)
	}
	err = rows.Err()
	if err != nil {
		panic(err)
	}

	return columns, err
}

func (m *DB) getIndexes() (indexes map[string][]mysqlIndex, err error) {

	dbName := m.databaseName
	indexes = make(map[string][]mysqlIndex)

	rows, err := m.SqlDb().Query(fmt.Sprintf(`
	SELECT
	index_name,
	non_unique,
	table_name,
	column_name
	FROM
	information_schema.statistics
	WHERE
	table_schema = '%s';
	`, dbName))

	if err != nil {
		panic(err)
	}
	defer sql2.RowClose(rows)
	var index mysqlIndex

	for rows.Next() {
		index = mysqlIndex{}
		err = rows.Scan(&index.name, &index.nonUnique, &index.tableName, &index.columnName)
		if err != nil {
			panic(err)
		}
		tableIndexes := indexes[index.tableName]
		tableIndexes = append(tableIndexes, index)
		indexes[index.tableName] = tableIndexes
	}
	err = rows.Err()
	if err != nil {
		panic(err)
	}

	return indexes, err
}

// getForeignKeys gets information on the foreign keys.
//
// Note that querying the information_schema database is SLOW, so we want to do it as few times as possible.
func (m *DB) getForeignKeys() (foreignKeys map[string][]mysqlForeignKey, err error) {
	dbName := m.databaseName
	fkMap := make(map[string]mysqlForeignKey)

	rows, err := m.SqlDb().Query(fmt.Sprintf(`
SELECT
    t1.constraint_name,
    t1.table_name,
    t1.column_name,
    t1.referenced_table_name,
    t1.referenced_column_name,
    t2.index_name
FROM
    information_schema.key_column_usage AS t1
JOIN
    information_schema.statistics AS t2
    ON t2.table_name = t1.referenced_table_name
    AND t2.column_name = t1.referenced_column_name
WHERE
    t1.constraint_schema = '%s'
ORDER BY
    t1.ordinal_position
`, dbName))
	if err != nil {
		panic(err)
	}

	func() {
		defer sql2.RowClose(rows)

		for rows.Next() {
			fk := mysqlForeignKey{}
			err = rows.Scan(&fk.constraintName, &fk.tableName, &fk.columnName, &fk.referencedTableName, &fk.referencedColumnName, &fk.referencedColumnIndexName)
			if err != nil {
				panic(err)
			}
			if fk.referencedColumnName.Valid {
				fkMap[fk.constraintName] = fk
			}
		}
		err = rows.Err()
		if err != nil {
			panic(err)
		}
	}()

	rows, err = m.SqlDb().Query(fmt.Sprintf(`
	SELECT
	constraint_name,
	update_rule,
	delete_rule
	FROM
	information_schema.referential_constraints
	WHERE
	constraint_schema = '%s';
	`, dbName))
	if err != nil {
		panic(err)
	}

	defer sql2.RowClose(rows)
	for rows.Next() {
		var constraintName string
		var updateRule, deleteRule sql.NullString
		err = rows.Scan(&constraintName, &updateRule, &deleteRule)
		if err != nil {
			log.Fatal(err)
		}
		if fk, ok := fkMap[constraintName]; ok {
			fk.updateRule = updateRule
			fk.deleteRule = deleteRule
			fkMap[constraintName] = fk
		}
	}
	err = rows.Err()
	if err != nil {
		panic(err)
	}

	foreignKeys = make(map[string][]mysqlForeignKey)
	for _, fk := range iter.KeySort(fkMap) {
		tableKeys := foreignKeys[fk.tableName]
		tableKeys = append(tableKeys, fk)
		foreignKeys[fk.tableName] = tableKeys
	}
	return foreignKeys, err
}

// Convert the database native type to a more generic sql type, and a go table type.
func (m *DB) processTypeInfo(column mysqlColumn) (
	typ schema.ColumnType,
	subType schema.ColumnSubType,
	maxLength uint64,
	defaultValue interface{},
	extra map[string]interface{},
	err error) {
	dataLen, dataSubLen := sql2.GetDataDefLength(column.columnType)
	isUnsigned := strings.Contains(column.columnType, "unsigned")

	switch column.dataType {
	case "time":
		typ = schema.ColTypeTime
		subType = schema.ColSubTypeTimeOnly
	case "date":
		typ = schema.ColTypeTime
		subType = schema.ColSubTypeDateOnly
	case "timestamp":
		typ = schema.ColTypeTime
	case "datetime":
		typ = schema.ColTypeTime
	case "tinyint":
		if dataLen == 1 {
			typ = schema.ColTypeBool
		} else {
			maxLength = 8
			if isUnsigned {
				typ = schema.ColTypeUint
			} else {
				typ = schema.ColTypeInt
			}
		}

	case "int":
		maxLength = 32 // mysql standard int has a 32-bit limit even in 64-bit implementations
		if isUnsigned {
			typ = schema.ColTypeUint
		} else {
			typ = schema.ColTypeInt
		}

	case "smallint":
		maxLength = 16
		if isUnsigned {
			typ = schema.ColTypeUint
		} else {
			typ = schema.ColTypeInt
		}
		extra = map[string]interface{}{"type": column.columnType}

	case "mediumint":
		maxLength = 24
		if isUnsigned {
			typ = schema.ColTypeUint
		} else {
			typ = schema.ColTypeInt
		}
		extra = map[string]interface{}{"type": column.columnType}

	case "bigint":
		maxLength = 64
		if isUnsigned {
			typ = schema.ColTypeUint
		} else {
			typ = schema.ColTypeInt
			if column.name == schema.GroTimestampColumnName {
				subType = schema.ColSubTypeTimestamp
			} else if column.name == schema.GroLockColumnName {
				subType = schema.ColSubTypeLock
			}
		}

	case "float":
		typ = schema.ColTypeFloat
		maxLength = 32
	case "double":
		typ = schema.ColTypeFloat
		maxLength = 64

	case "varchar":
		typ = schema.ColTypeString
		maxLength = uint64(dataLen)
		if column.collation.String != m.defaultCollation {
			extra = map[string]interface{}{"collation": column.collation.String}
		}

	case "char":
		typ = schema.ColTypeString
		maxLength = uint64(dataLen)
		// default is a varchar, so capture the type to preserve the mysql type
		extra = map[string]interface{}{"type": column.columnType}
		if column.collation.String != m.defaultCollation {
			extra["collation"] = column.collation.String
		}

	case "varbinary":
		typ = schema.ColTypeBytes
		maxLength = uint64(dataLen)
	case "blob":
		typ = schema.ColTypeBytes
		maxLength = 0 // default
	case "tinyblob":
		typ = schema.ColTypeBytes
		maxLength = 255
	case "mediumblob":
		typ = schema.ColTypeBytes
		maxLength = 16777215
	case "longblob":
		typ = schema.ColTypeBytes
		maxLength = math.MaxUint32

	case "text":
		typ = schema.ColTypeString
		maxLength = 0 // text type is unsized. The limit is 65535 bytes (not characters) and so the limit on chars depends on the charset and what is actually being stored.
		if column.collation.String != m.defaultCollation {
			extra["collation"] = column.collation.String
		}

	case "tinytext":
		typ = schema.ColTypeString
		maxLength = 255
		extra = map[string]interface{}{"type": column.columnType}
		if column.collation.String != m.defaultCollation {
			extra["collation"] = column.collation.String
		}

	case "mediumtext":
		typ = schema.ColTypeString
		maxLength = 16777215
		extra = map[string]interface{}{"type": column.columnType}
		if column.collation.String != m.defaultCollation {
			extra["collation"] = column.collation.String
		}

	case "longtext":
		typ = schema.ColTypeString
		maxLength = 1073741823
		if column.collation.String != m.defaultCollation {
			extra = map[string]interface{}{"collation": column.collation.String}
		}

	case "decimal":
		// No native equivalent in Go.
		// See the shopspring/decimal or math/big packages for possible support.
		typ = schema.ColTypeString
		// pack the two length values to be unpacked in Go
		maxLength = uint64(dataLen) + uint64(dataSubLen<<16)
		subType = schema.ColSubTypeNumeric

	case "year":
		typ = schema.ColTypeInt
		extra = map[string]interface{}{"type": column.columnType}

	case "set":
		typ = schema.ColTypeUnknown
		maxLength = uint64(column.characterMaxLen.Int64)
		extra = map[string]interface{}{"type": column.columnType}
		if column.collation.String != m.defaultCollation {
			extra["collation"] = column.collation.String
		}

	case "enum":
		typ = schema.ColTypeUnknown
		maxLength = uint64(column.characterMaxLen.Int64)
		extra = map[string]interface{}{"type": column.columnType}
		if column.collation.String != m.defaultCollation {
			extra["collation"] = column.collation.String
		}

	case "json":
		// this might be an enum array field
		if strings.Contains(column.comment, `"enum_table"`) {
			typ = schema.ColTypeEnumArray
		} else {
			typ = schema.ColTypeJSON
		}

	default:
		typ = schema.ColTypeUnknown
		extra = map[string]interface{}{"type": column.columnType}
	}

	si := column.defaultValue.StringI()

	if typ == schema.ColTypeTime {
		if strings.Contains(strings.ToUpper(column.extra), "ON UPDATE") {
			defaultValue = "update"
		} else if si != nil && strings.Contains(strings.ToUpper(si.(string)), "CURRENT_TIMESTAMP") {
			defaultValue = "now"
		}
	} else if si != nil &&
		si.(string) != "" &&
		si.(string) != "NULL" && // null is automatically assigned as a default for null columns, and cannot be assigned as a default for non-null columns
		strings.Contains(strings.ToUpper(column.extra), "DEFAULT_GENERATED") {
		if extra == nil {
			extra = make(map[string]interface{})
		}
		extra["default"] = si.(string) // some kind of generated value that we should remember for recreating the column
	}

	if strings.Contains(strings.ToUpper(column.extra), "DEFAULT_GENERATED") {
		// The default value is generated by mysql, so we capture it for purposes of recreating the column,
		// but we need to ignore it for purposes of actually using it as a default in Go.
		// This is mysql only, and not in mariadb as of now

	} else if defaultValue == nil {
		defaultValue = column.defaultValue.UnpackDefaultValue(typ, int(maxLength))
	}

	return
}

func (m *DB) schemaFromRawTables(rawTables map[string]mysqlTable, options map[string]any) schema.Database {

	dd := schema.Database{
		ReferenceSuffix: options["reference_suffix"].(string),
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

		if strings2.EndsWith(tableName, dd.EnumTableSuffix) {
			if t, err := m.getEnumTableSchema(rawTable); err != nil {
				slog.Error("Enum rawTable skipped",
					slog.String(db.LogTable, tableName),
					slog.Any(db.LogError, err))
			} else {
				dd.EnumTables = append(dd.EnumTables, &t)
			}
		} else if strings2.EndsWith(tableName, dd.AssnTableSuffix) {
			if mm, err := m.getAssociationSchema(rawTable, dd.EnumTableSuffix); err != nil {
				slog.Error("Association rawTable skipped",
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

func (m *DB) getTableSchema(t mysqlTable, enumTableSuffix string) schema.Table {
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
			if idx.name == "PRIMARY" {
				level = schema.IndexLevelManualPrimaryKey
			} else if idx.nonUnique {
				level = schema.IndexLevelIndexed
			} else {
				level = schema.IndexLevelUnique
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

	schem := ""
	tableName := t.name
	parts := strings.Split(t.name, ".")
	if len(parts) == 2 {
		schem = parts[0]
		tableName = parts[1]
	}
	td := schema.Table{
		Name:    tableName,
		Schema:  schem,
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

func (m *DB) getEnumTableSchema(t mysqlTable) (ed schema.EnumTable, err error) {
	td := m.getTableSchema(t, "")

	var columnNames []string
	var receiverTypes []ReceiverType

	if len(td.Columns) < 2 {
		err = fmt.Errorf("error: An enum table must have at least 2 columns")
		return
	}

	ed.Name = td.Name
	ed.Fields = make(map[string]schema.EnumField)

	var hasConst bool
	var hasLabelOrIdentifier bool

	var rawColumns = make(map[string]mysqlColumn)
	for _, c := range t.columns {
		rawColumns[c.name] = c
	}

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
		recType := ReceiverTypeFromSchema(c.Type, c.Size)
		typ := c.Type
		if c.Name == schema.ConstKey && c.Type == schema.ColTypeAutoPrimaryKey {
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

	if !hasConst {
		err = fmt.Errorf(`error: An enum table must have a "const"" column`)
		return
	}
	if !hasLabelOrIdentifier {
		err = fmt.Errorf(`error: An enum table must have a "label" of "identifier" column`)
		return
	}

	var result *sql.Rows
	s := `
	SELECT ` +
		"`" + strings.Join(columnNames, "`,`") + "`" +
		`
	FROM ` +
		"`" + td.Name + "`" +
		` ORDER BY ` + "`" + columnNames[0] + "`"
	result, err = m.SqlDb().Query(s)
	if err != nil {
		panic(err)
	}

	var receiver []map[string]any
	receiver, err = sql2.ReceiveRows(result, receiverTypes, columnNames, nil, s, nil)
	if err != nil {
		panic(err)
	}
	for _, row := range receiver {
		values := make(map[string]any)
		for k := range ed.Fields {
			values[k] = row[k]
		}
		ed.Values = append(ed.Values, values)
	}
	ed.Comment = t.comment

	return
}

func (m *DB) getColumnSchema(table mysqlTable,
	column mysqlColumn,
	indexLevel schema.IndexLevel,
	enumTableSuffix string) schema.Column {

	cd := schema.Column{
		Name: column.name,
	}
	var err error
	var extra map[string]any
	cd.Type, cd.SubType, cd.Size, cd.DefaultValue, extra, err = m.processTypeInfo(column)
	if err != nil {
		slog.Warn(err.Error(),
			slog.String(db.LogComponent, "extract"),
			slog.String(db.LogTable, table.name),
			slog.String(db.LogColumn, column.name),
			slog.Any(db.LogError, err))
	}
	if extra != nil {
		if cd.DatabaseDefinition == nil {
			cd.DatabaseDefinition = make(map[string]map[string]interface{})
		}
		cd.DatabaseDefinition[db.DriverTypeMysql] = extra
	}

	isAuto := strings.Contains(column.extra, "auto_increment")
	if isAuto {
		// Note that unsigned auto increment primary keys is not supported
		cd.Type = schema.ColTypeAutoPrimaryKey
		cd.Size = 0 // hide details of auto id generation from schema file. If int size is needed, it should go in DatabaseDefinition
	} else {
		cd.IndexLevel = indexLevel
	}

	cd.IsNullable = column.isNullable == "YES"

	if enumTableSuffix != "" && strings.HasSuffix(column.name, enumTableSuffix) { // handle enum type
		if cd.Type == schema.ColTypeInt || cd.Type == schema.ColTypeUint {
			cd.Type = schema.ColTypeEnum
			if fk, ok := table.fkMap[column.name]; ok { // if it is a foreign key to the enum table, use that
				cd.Reference = &schema.Reference{
					Table: fk.referencedTableName.String,
				}
			} else {
				slog.Warn("Could not find foreign key to enum table.",
					slog.String(db.LogComponent, "extract"),
					slog.String(db.LogTable, table.name),
					slog.String(db.LogColumn, column.name),
					slog.Any(db.LogError, err))
			}
		} else if cd.Type == schema.ColTypeString {
			cd.Type = schema.ColTypeEnumArray
		}
		cd.Size = 0 // use default size
	} else if fk, ok2 := table.fkMap[column.name]; ok2 { // handle forward reference
		if fk.referencedColumnIndexName.String != "PRIMARY" {
			slog.Warn("Foregin key appears to not be pointing to a primary key. Only primary key foreign keys are supported.",
				slog.String(db.LogComponent, "extract"),
				slog.String(db.LogTable, table.name),
				slog.String(db.LogColumn, column.name),
				slog.Any(db.LogError, err))
		}
		cd.Reference = &schema.Reference{
			Table: fk.referencedTableName.String,
		}
		cd.Type = schema.ColTypeReference
		cd.Size = 0
	}

	cd.Comment = column.comment

	return cd
}

func (m *DB) getAssociationSchema(t mysqlTable, enumTableSuffix string) (mm schema.AssociationTable, err error) {
	td := m.getTableSchema(t, enumTableSuffix)
	if len(td.Columns) == 0 {
		err = fmt.Errorf("association table must have 2 columns")
		return
	}
	if td.Columns[0].Name == "_id" {
		// This was added because of a multi-column primary key, which is ok for an association table
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
	mm.Column1 = td.Columns[0].Name
	mm.Table2 = td.Columns[1].Reference.Table
	mm.Column2 = td.Columns[1].Name

	mm.Comment = t.comment
	return
}
