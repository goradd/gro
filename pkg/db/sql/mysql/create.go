package mysql

import (
	"fmt"
	"github.com/goradd/orm/pkg/db"
	"github.com/goradd/orm/pkg/schema"
	"log/slog"
	"strings"
)

// TableDefinitionSql will return the sql needed to create the table.
// This can include clauses separated by semicolons that add additional capabilities to the table.
func (m *DB) TableDefinitionSql(d *schema.Database, table *schema.Table) string {
	var sb strings.Builder

	var pkCount int
	var columnDefs []string
	var tableClauses []string
	var extraClauses []string

	for _, col := range table.Columns {
		if col.IsPrimaryKey() {
			pkCount++
			if pkCount > 1 {
				slog.Error("Table skipped. Table has more than one primary key column",
					slog.String(db.LogTable, table.Name))
				return ""
			}
		}
		colDef, tc, xc := m.buildColumnDef(d, col)
		if colDef == "" {
			continue // error, already reported
		}
		columnDefs = append(columnDefs, "  "+colDef)
		tableClauses = append(tableClauses, tc...)
		extraClauses = append(extraClauses, xc...)
	}

	// Multi-column indexes
	for _, mci := range table.MultiColumnIndexes {
		cols := make([]string, len(mci.Columns))
		for i, name := range mci.Columns {
			cols[i] = m.QuoteIdentifier(name)
		}
		var idxType string
		switch mci.IndexLevel {
		case schema.IndexLevelManualPrimaryKey:
			idxType = "PRIMARY KEY"
		case schema.IndexLevelUnique:
			idxType = "UNIQUE INDEX"
		case schema.IndexLevelIndexed:
			idxType = "INDEX"
		default:
			slog.Error("Index skipped. Unknown index level in multi-column index",
				slog.String(db.LogTable, table.Name))
			continue
		}
		idxCols := strings.Join(cols, ", ")
		def := fmt.Sprintf("%s (%s)", idxType, idxCols)
		tableClauses = append(tableClauses, def)
	}
	columnDefs = append(columnDefs, tableClauses...)

	tableName := table.QualifiedName()
	sb.WriteString(fmt.Sprintf("CREATE TABLE %s (\n", m.QuoteIdentifier(tableName)))
	sb.WriteString(strings.Join(columnDefs, ",\n"))
	sb.WriteString("\n)")

	if table.Comment != "" {
		sb.WriteString(fmt.Sprintf(` COMMENT='%s'`, table.Comment))
	}
	sb.WriteString(";\n")
	sb.WriteString(strings.Join(extraClauses, ";\n"))
	return sb.String()
}

// buildColumnDef returns the sql that will create the column col.
// tableClauses will be included within the Create Table definition
// extraClauses will be executed outside the table definition
func (m *DB) buildColumnDef(d *schema.Database, col *schema.Column) (s string, tableClauses []string, extraClauses []string) {
	var colType string
	var collation string
	var defaultStr string

	if def := col.DatabaseDefinition[db.DriverTypeMysql]; def != nil {
		if t, ok := def["type"].(string); ok {
			colType = t
		}
		if c, ok := def["collation"].(string); ok && c != "" {
			collation = " COLLATE " + c
		}
		if d, ok := def["default"].(string); ok && d != "" {
			defaultStr = " DEFAULT " + d
		}
	}
	if col.Type == schema.ColTypeReference {
		if col.Reference == nil || col.Reference.Table == "" {
			slog.Error("Column skipped, Reference with a Table value is required",
				slog.String(db.LogColumn, col.Name))
			return
		}
		// match the referenced column's type
		t := d.FindTable(col.Reference.Table)
		if t == nil {
			slog.Error("Column skipped, ref table not found",
				slog.String(db.LogTable, col.Reference.Table),
				slog.String(db.LogColumn, col.Name))
		}
		c := t.PrimaryKeyColumn()
		if c == nil {
			slog.Error("Column skipped, reference table does not have a primary key",
				slog.String(db.LogTable, col.Reference.Table))
		}
		if colType == "" {
			colType = sqlType(c.Type, c.Size, c.SubType, true)
		}

		// foreign key will automatically index the column
		fk := fmt.Sprintf(" FOREIGN KEY (%s) REFERENCES %s(%s)", m.QuoteIdentifier(col.Name), m.QuoteIdentifier(t.Name), m.QuoteIdentifier(c.Name))
		tableClauses = append(tableClauses, fk)
	} else if col.Type == schema.ColTypeEnum {
		if col.Reference == nil || col.Reference.Table == "" {
			slog.Error("Column skipped, Reference with a Table value is required",
				slog.String(db.LogColumn, col.Name))
			return
		}

		// foreign key will automatically index the column
		fk := fmt.Sprintf(" FOREIGN KEY (%s) REFERENCES %s(%s)", m.QuoteIdentifier(col.Name), m.QuoteIdentifier(col.Reference.Table), m.QuoteIdentifier("const"))
		tableClauses = append(tableClauses, fk)
		colType = "INT"
	} else {
		colType = sqlType(col.Type, col.Size, col.SubType, false)

		switch col.IndexLevel {
		case schema.IndexLevelIndexed:
			tableClauses = append(tableClauses, fmt.Sprintf(" INDEX (%s)", m.QuoteIdentifier(col.Name)))
		case schema.IndexLevelUnique:
			tableClauses = append(tableClauses, fmt.Sprintf(" UNIQUE (%s)", m.QuoteIdentifier(col.Name)))
		case schema.IndexLevelManualPrimaryKey:
			tableClauses = append(tableClauses, fmt.Sprintf(" PRIMARY KEY (%s)", m.QuoteIdentifier(col.Name)))
		default:
			// do nothing
		}
	}

	if !col.IsNullable {
		colType += " NOT NULL "
	}

	var extraStr string
	if col.DefaultValue != nil && defaultStr == "" {
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

	commentStr := col.Comment
	if commentStr != "" {
		commentStr = fmt.Sprintf("COMMENT '%s'", commentStr)
	}

	s = fmt.Sprintf("%s %s %s %s %s %s", m.QuoteIdentifier(col.Name), colType, defaultStr, extraStr, collation, commentStr)
	return
}

// SqlType is used by the builder to return the SQL corresponding to the given colType that will create
// the column.
// If forReference is true, then it returns the SQL for creating a reference to the column.
func sqlType(colType schema.ColumnType, size uint64, subType schema.ColumnSubType, forReference bool) string {
	switch colType {
	case schema.ColTypeAutoPrimaryKey:
		typ := intType(size, false)
		if !forReference {
			typ += " AUTO_INCREMENT PRIMARY KEY"
		}
		return typ
	case schema.ColTypeString:
		if subType == schema.ColSubTypeNumeric {
			precision := size & 0x0000FFFF
			scale := size >> 16
			if precision != 0 && scale != 0 {
				return fmt.Sprintf("DECIMAL(%d, %d)", precision, scale)
			} else {
				return "DECIMAL(65, 30)" // max precision available in mysql
			}
		} else if size == 0 {
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
		switch subType {
		case schema.ColSubTypeDateOnly:
			return "DATE"
		case schema.ColSubTypeTimeOnly:
			return "TIME"
		case schema.ColSubTypeNone:
			return "DATETIME"
		default:
			slog.Warn("Wrong subtype for time column",
				slog.String("subtype", subType.String()))
		}
		return "DATETIME"
	case schema.ColTypeReference:
		return intType(size, false)
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
