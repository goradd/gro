package mysql

import (
	"context"
	"fmt"
	"github.com/goradd/orm/pkg/db"
	sql2 "github.com/goradd/orm/pkg/db/sql"
	"github.com/goradd/orm/pkg/schema"
	"log/slog"
	"strings"
)

// BuildSchema builds the database schema from s.
func (m *DB) BuildSchema(ctx context.Context, s schema.Database) error {
	for _, table := range s.EnumTables {
		if err := m.buildEnum(ctx, table); err != nil {
			return err
		}
	}
	for _, table := range s.Tables {
		if err := m.buildTable(ctx, &s, table); err != nil {
			return err
		}
	}
	for _, table := range s.AssociationTables {
		if err := m.buildAssociation(ctx, &s, table); err != nil {
			return err
		}
	}

	return nil
}

func (m *DB) buildTable(ctx context.Context, s *schema.Database, table *schema.Table) (err error) {
	sql := tableSql(s, table)
	if sql == "" {
		return fmt.Errorf("error in table `%s`", table.Name)
	}
	_, err = m.SqlExec(ctx, sql)
	if err != nil {
		slog.Error("SQL error",
			slog.String("sql", sql),
			slog.Any("error", err))
	}
	return err
}

func (m *DB) buildEnum(ctx context.Context, table *schema.EnumTable) (err error) {
	var args []any
	sql := enumTableSql(table)
	if sql == "" {
		return fmt.Errorf("error in table `%s`", table.Name)
	}
	if _, err = m.SqlExec(ctx, sql); err != nil {
		slog.Error("SQL error",
			slog.String("sql", sql),
			slog.Any("error", err))

		return
	}

	fieldKeys := table.FieldKeys()
	for _, v := range table.Values {
		sql, args = enumValueSql(table.Name, fieldKeys, table.Fields, v)
		if _, err = m.SqlExec(ctx, sql, args...); err != nil {
			slog.Error("SQL error",
				slog.String("sql", sql),
				slog.Any("error", err),
				slog.Any("args", args))

			return
		}
	}
	return err
}

func (m *DB) buildAssociation(ctx context.Context, s *schema.Database, table *schema.AssociationTable) (err error) {
	sql := associationSql(s, table)
	if sql == "" {
		return fmt.Errorf("error in table `%s`", table.Name)
	}
	_, err = m.SqlExec(ctx, sql)
	if err != nil {
		slog.Error("SQL error",
			slog.String("sql", sql),
			slog.Any("error", err),
		)
	}
	return err
}

func tableSql(s *schema.Database, table *schema.Table) string {
	var sb strings.Builder

	tableName := table.Name
	if table.Schema != "" {
		tableName = fmt.Sprintf("%s.%s", table.Schema, table.Name)
	}

	sb.WriteString(fmt.Sprintf("CREATE TABLE `%s` (\n", tableName))

	var pkCols []string
	var columnDefs []string
	var indexDefs []string
	var foreignKeys []string

	for _, col := range table.Columns {
		if (col.Type == schema.ColTypeReference ||
			col.Type == schema.ColTypeEnum ||
			col.Type == schema.ColTypeEnumArray) &&
			(col.Reference == nil || col.Reference.Table == "") {
			slog.Error("Column skipped, Reference with a Table value is required",
				slog.String("column", col.Name))
			continue
		}
		colDef := buildColumnDef(col)
		columnDefs = append(columnDefs, "  "+colDef)

		// Primary Key
		if col.Type == schema.ColTypeAutoPrimaryKey ||
			col.IndexLevel == schema.IndexLevelManualPrimaryKey {
			pkCols = append(pkCols, col.Name)
		}

		// Foreign Key
		if col.Type == schema.ColTypeReference {
			refParts := strings.Split(col.Reference.Table, ".")
			refTable := col.Reference.Table
			if len(refParts) == 2 {
				refTable = fmt.Sprintf("`%s`.`%s`", refParts[0], refParts[1])
			} else {
				refTable = fmt.Sprintf("`%s`", refParts[0])
			}
			t := s.FindTable(col.Reference.Table)
			if t == nil {
				slog.Error("Column skipped, reference table not found",
					slog.String("table", refTable))
				continue
			}
			pk := t.PrimaryKeyColumn()
			if pk == nil {
				slog.Error("Column skipped, reference pk column not found",
					slog.String("table", refTable))
				continue
			}
			foreignKeys = append(foreignKeys,
				fmt.Sprintf("  FOREIGN KEY (`%s`) REFERENCES %s(`%s`)", col.Name, refTable, pk.Name),
			)
		} else if col.Type == schema.ColTypeEnum {
			refParts := strings.Split(col.Reference.Table, ".")
			refTable := col.Reference.Table
			if len(refParts) == 2 {
				refTable = fmt.Sprintf("`%s`.`%s`", refParts[0], refParts[1])
			} else {
				refTable = fmt.Sprintf("`%s`", refParts[0])
			}
			t := s.FindEnumTable(col.Reference.Table)
			if t == nil {
				slog.Error("Column skipped, reference table not found",
					slog.String("table", refTable))
				continue
			}
			pk := t.FieldKeys()[0]
			foreignKeys = append(foreignKeys,
				fmt.Sprintf("  FOREIGN KEY (`%s`) REFERENCES %s(`%s`)", col.Name, refTable, pk),
			)
		}

		// Indexes
		switch col.IndexLevel {
		case schema.IndexLevelIndexed:
			indexDefs = append(indexDefs, fmt.Sprintf("  INDEX (`%s`)", col.Name))
		case schema.IndexLevelUnique:
			indexDefs = append(indexDefs, fmt.Sprintf("  UNIQUE INDEX (`%s`)", col.Name))
		default:
			// do nothing
		}
	}

	// Add primary key
	if len(pkCols) > 0 {
		columnDefs = append(columnDefs, fmt.Sprintf("  PRIMARY KEY (`%s`)", strings.Join(pkCols, "`, `")))
	}

	// Multi-column indexes
	for _, mci := range table.MultiColumnIndexes {
		cols := make([]string, len(mci.Columns))
		for i, name := range mci.Columns {
			cols[i] = fmt.Sprintf("`%s`", name)
		}
		idx := "INDEX"
		if mci.IsUnique {
			idx = "UNIQUE INDEX"
		}
		indexDefs = append(indexDefs, fmt.Sprintf("  %s (%s)", idx, strings.Join(cols, ", ")))
	}

	allDefs := append(columnDefs, foreignKeys...)
	allDefs = append(allDefs, indexDefs...)
	sb.WriteString(strings.Join(allDefs, ",\n"))
	sb.WriteString("\n)")

	commentStr := sql2.TableComment(table)
	if commentStr != "" {
		sb.WriteString(fmt.Sprintf(` COMMENT='%s'`, commentStr))
	}
	sb.WriteString("\n")

	return sb.String()
}

func enumTableSql(table *schema.EnumTable) (s string) {
	var sb strings.Builder

	tableName := table.Name
	if table.Schema != "" {
		tableName = fmt.Sprintf("%s.%s", table.Schema, table.Name)
	}

	// Build CREATE TABLE
	sb.WriteString(fmt.Sprintf("CREATE TABLE `%s` (\n", tableName))
	var columnDefs []string

	for _, k := range table.FieldKeys() {
		var size uint64
		for _, v := range table.Values {
			if table.Fields[k].Type == schema.ColTypeString ||
				table.Fields[k].Type == schema.ColTypeBytes {
				if s, ok := v[k].(string); ok {
					size = max(size, uint64(len(s)))
				}
			}
		}
		commentStr := " COMMENT '" + sql2.EnumFieldComment(table.Fields[k]) + "'"
		colDef := buildEnumFieldDef(table.Fields[k].Type, size, k, commentStr)
		columnDefs = append(columnDefs, "  "+colDef)
	}

	// Add primary key
	columnDefs = append(columnDefs, fmt.Sprintf("  PRIMARY KEY (`%s`)", table.FieldKeys()[0]))

	sb.WriteString(strings.Join(columnDefs, ",\n"))
	sb.WriteString("\n)")

	// Add table comment
	commentStr := sql2.EnumTableComment(table) // assume you aliased import to sqlPkg if conflicting
	if commentStr != "" {
		sb.WriteString(fmt.Sprintf(` COMMENT='%s'`, commentStr))
	}
	sb.WriteString("\n")

	return sb.String()
}

func enumValueSql(tableName string, fieldKeys []string, fields map[string]schema.EnumField, v map[string]any) (sql string, args []any) {
	// Now add INSERTs
	var columns []string
	var placeholders []string
	for _, k := range fieldKeys {
		columns = append(columns, fmt.Sprintf("`%s`", k))

		fieldType := fields[k].Type
		value := v[k]

		placeholders = append(placeholders, "?")

		switch fieldType {
		case schema.ColTypeString:
			args = append(args, value.(string))
		case schema.ColTypeInt:
			args = append(args, value)
		case schema.ColTypeFloat:
			args = append(args, value)
		default:
			args = append(args, value)
		}
	}

	return fmt.Sprintf(
		"INSERT INTO `%s` (%s) VALUES (%s)",
		tableName,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	), args
}

func associationSql(s *schema.Database, table *schema.AssociationTable) string {
	var sb strings.Builder

	tableName := table.Name
	if table.Schema != "" {
		tableName = fmt.Sprintf("%s.%s", table.Schema, table.Name)
	}

	sb.WriteString(fmt.Sprintf("CREATE TABLE `%s` (\n", tableName))
	var columnDefs []string
	table1 := s.FindTable(table.Table1)
	if table1 == nil {
		slog.Error("Association table skipped, Table1 not found",
			slog.String("table", table.Table1))
		return ""
	}
	pk1 := table1.PrimaryKeyColumn()
	if pk1 == nil {
		slog.Error("Association table skipped, PrimaryKeyColumn not found",
			slog.String("table", table.Table1))
		return ""
	}
	typ1 := pk1.Type
	if typ1 == schema.ColTypeAutoPrimaryKey {
		typ1 = schema.ColTypeInt
	}
	colType1 := sqlType(typ1, pk1.Size)
	columnDefs = append(columnDefs, fmt.Sprintf("`%s` %s NOT NULL", table.Column1, colType1))

	table2 := s.FindTable(table.Table1)
	if table2 == nil {
		slog.Error("Association table skipped, Table1 not found",
			slog.String("table", table.Table2))
		return ""
	}
	pk2 := table1.PrimaryKeyColumn()
	if pk2 == nil {
		slog.Error("Association table skipped, PrimaryKeyColumn not found",
			slog.String("table", table.Table2))
		return ""
	}
	typ2 := pk2.Type
	if typ2 == schema.ColTypeAutoPrimaryKey {
		typ2 = schema.ColTypeInt
	}
	colType2 := sqlType(typ2, pk2.Size)
	columnDefs = append(columnDefs, fmt.Sprintf(" `%s` %s NOT NULL", table.Column2, colType2))

	columnDefs = append(columnDefs, fmt.Sprintf("  FOREIGN KEY (`%s`) REFERENCES %s(`%s`)", table.Column1, table.Table1, pk1.Name))
	columnDefs = append(columnDefs, fmt.Sprintf("  FOREIGN KEY (`%s`) REFERENCES %s(`%s`)", table.Column2, table.Table2, pk2.Name))
	columnDefs = append(columnDefs, fmt.Sprintf("  INDEX (`%s`)", table.Column1))
	columnDefs = append(columnDefs, fmt.Sprintf("  INDEX (`%s`)", table.Column2))
	columnDefs = append(columnDefs, fmt.Sprintf("  UNIQUE INDEX (`%s`, `%s`)", table.Column1, table.Column2))

	sb.WriteString(strings.Join(columnDefs, ",\n"))
	sb.WriteString("\n)")

	commentStr := sql2.AssociationTableComment(table)
	if commentStr != "" {
		sb.WriteString(fmt.Sprintf(` COMMENT='%s'`, commentStr))
	}
	sb.WriteString("\n")

	return sb.String()
}

func buildColumnDef(col *schema.Column) string {
	var colType string
	var collation string
	if m := col.DatabaseDefinition[db.DriverTypeMysql]; m != nil {
		if t, ok := m["type"].(string); ok {
			colType = t
		}
		if c, ok := m["collation"].(string); ok && c != "" {
			collation = " COLLATE '" + c + "'"
		}
	}
	if colType == "" {
		colType = sqlType(col.Type, col.Size)
	}
	nullStr := "NOT NULL"
	if col.IsNullable {
		nullStr = "NULL"
	}

	var defaultStr string
	var extraStr string
	if col.DefaultValue != nil {
		switch val := col.DefaultValue.(type) {
		case string:
			if col.Type == schema.ColTypeTime {
				if val == "now" {
					defaultStr = " DEFAULT CURRENT_TIMESTAMP"
				} else if val == "update" {
					defaultStr = " DEFAULT CURRENT_TIMESTAMP"
					extraStr = " ON UPDATE CURRENT_TIMESTAMP"
				} else {
					defaultStr = fmt.Sprintf("DEFAULT '%s'", val)
				}
			} else {
				defaultStr = fmt.Sprintf("DEFAULT '%s'", val)
			}
		default:
			defaultStr = fmt.Sprintf("DEFAULT %v", val)
		}
	}

	commentStr := sql2.ColumnComment(col)
	if commentStr != "" {
		commentStr = fmt.Sprintf("COMMENT '%s'", commentStr)
	}

	return fmt.Sprintf("`%s` %s %s %s %s %s %s", col.Name, colType, nullStr, defaultStr, extraStr, collation, commentStr)
}

func sqlType(colType schema.ColumnType, size uint64) string {
	switch colType {
	case schema.ColTypeAutoPrimaryKey:
		return "INT AUTO_INCREMENT"
	case schema.ColTypeString:
		if size == 0 {
			return "TEXT"
		} else if size < 16383 {
			return fmt.Sprintf("VARCHAR(%d)", size)
		} else if size < 4194303 {
			return "MEDIUMTEXT"
		} else {
			return "LONGTEXT"
		}
	case schema.ColTypeBytes:
		if size == 0 {
			return "BLOB"
		} else if size < 65532 {
			return fmt.Sprintf("VARBINARY(%d)", size)
		} else if size < 16777215 {
			return "MEDIUMBLOB"
		} else {
			return "LONGBLOB"
		}
	case schema.ColTypeInt:
		return intType(size, false)
	case schema.ColTypeUint:
		return intType(size, true)
	case schema.ColTypeFloat:
		if size == 32 {
			return "FLOAT"
		}
		return "DOUBLE"
	case schema.ColTypeBool:
		return "BOOLEAN"
	case schema.ColTypeTime:
		return "DATETIME"
	case schema.ColTypeReference:
		return "INT"
	case schema.ColTypeEnum:
		return "INT"
	case schema.ColTypeJSON:
		fallthrough
	case schema.ColTypeEnumArray:
		return "JSON" // In MariaDB, this becomes a LONGTEXT, but MariaDB will store in like a varchar if its <8KB.
	default:
		return "TEXT"
	}
}

func intType(size uint64, unsigned bool) string {
	var t string
	switch size {
	case 8:
		t = "TINYINT"
	case 16:
		t = "SMALLINT"
	case 32:
		t = "INT"
	case 64:
		t = "BIGINT"
	default:
		t = "INT"
	}
	if unsigned {
		t += " UNSIGNED"
	}
	return t
}

func buildEnumFieldDef(typ schema.ColumnType, size uint64, name string, comment string) string {
	colType := sqlType(typ, size)
	return fmt.Sprintf("`%s` %s NOT NULL %s", name, colType, comment)
}
