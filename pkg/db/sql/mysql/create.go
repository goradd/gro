package mysql

import (
	"fmt"
	"github.com/goradd/anyutil"
	"github.com/goradd/orm/pkg/db"
	"github.com/goradd/orm/pkg/schema"
	"log/slog"
	"strings"
)

// TableDefinitionSql will return the sql needed to create the table.
// This can include clauses separated by semicolons.
// All the returned tableSql items will be executed for all tables before extraSql will be executed.
// This allows extraSql to refer to other tables in the schema that might not have been created yet,
// and is particularly useful for handling cyclic foreign key references.
func (m *DB) TableDefinitionSql(d *schema.Database, table *schema.Table) (tableSql string, extraClauses []string) {
	var sb strings.Builder

	var columnDefs []string
	var tableClauses []string

	for _, col := range table.Columns {
		cc, tc, xc := m.buildColumnDef(col)
		if cc == "" {
			continue // error, already reported
		}
		columnDefs = append(columnDefs, cc)
		tableClauses = append(tableClauses, tc...)
		extraClauses = append(extraClauses, xc...)
	}

	// build the foreign keys
	for _, ref := range table.References {
		cc, tc, xc := m.buildReferenceDef(d, table, ref)
		if cc == "" {
			continue // error, already reported
		}
		columnDefs = append(columnDefs, cc)
		tableClauses = append(tableClauses, tc...)
		extraClauses = append(extraClauses, xc...)
	}

	for _, mci := range table.Indexes {
		def := m.indexSql(mci.IndexLevel, mci.Columns...)
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
	return sb.String(), extraClauses
}

// buildColumnDef returns the sql that will create the column col, including related indexes and foreign keys.
// columnClause is the column definition.
// tableClauses will be included within the Create Table definition after the column clauses.
// extraClauses will be executed outside the table definition after all tables and their columns have been created.
func (m *DB) buildColumnDef(col *schema.Column) (columnClause string, tableClauses []string, extraClauses []string) {
	var colType string
	var collation string
	var defaultStr string
	var extraStr string

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
	if col.Type == schema.ColTypeEnum {
		if col.EnumTable == "" {
			slog.Error("Column skipped, EnumTable not specified for an enum value.",
				slog.String(db.LogColumn, col.Name))
			return
		}

		// foreign key will automatically index the column
		fk := fmt.Sprintf(" FOREIGN KEY (%s) REFERENCES %s(%s)",
			m.QuoteIdentifier(col.Name),
			m.QuoteIdentifier(col.EnumTable),
			m.QuoteIdentifier("const"))
		tableClauses = append(tableClauses, fk)
		colType = "INT"
	} else {
		colType = sqlType(col.Type, col.Size, col.SubType)
		if col.Type == schema.ColTypeAutoPrimaryKey {
			extraStr += "AUTO_INCREMENT"
		}
	}

	if !col.IsNullable {
		colType += " NOT NULL "
	}

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

	columnClause = fmt.Sprintf("%s %s %s %s %s %s", m.QuoteIdentifier(col.Name), colType, defaultStr, extraStr, collation, commentStr)
	return
}

func (m *DB) indexSql(level schema.IndexLevel, cols ...string) string {
	var idxType string
	switch level {
	case schema.IndexLevelPrimaryKey:
		idxType = "PRIMARY KEY"
	case schema.IndexLevelUnique:
		idxType = "UNIQUE INDEX"
	case schema.IndexLevelIndexed:
		idxType = "INDEX"
	default:
		return ""
	}
	cols = anyutil.MapSliceFunc(cols, func(s string) string {
		return m.QuoteIdentifier(s)
	})
	idxCols := strings.Join(cols, ", ")
	def := fmt.Sprintf("%s (%s)", idxType, idxCols)
	return def
}

func (m *DB) buildReferenceDef(db *schema.Database, table *schema.Table, ref *schema.Reference) (columnClause string, tableClauses, extraClauses []string) {
	fk, pk := ref.ReferenceColumns(db, table)

	if fk.Type == schema.ColTypeAutoPrimaryKey {
		fk.Type = schema.ColTypeInt // auto columns internally are integers
	}

	columnClause, tableClauses, extraClauses = m.buildColumnDef(fk)
	if columnClause == "" {
		return // error, already logged
	}

	// Make a constraint name that will be unique within the database and logically related to the relationship.
	constraintName := table.Name + "_" + ref.Column + "_fk"

	// We use alter table after all tables are created in case of cyclic foreign keys.
	s := fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s FOREIGN KEY (%s) REFERENCES %s(%s)",
		m.QuoteIdentifier(table.Name),
		m.QuoteIdentifier(constraintName),
		m.QuoteIdentifier(fk.Name),
		m.QuoteIdentifier(ref.Table),
		m.QuoteIdentifier(pk.Name))
	extraClauses = append(extraClauses, s)
	return
}

// SqlType is used by the builder to return the SQL corresponding to the given colType that will create
// the column.
// If forReference is true, then it returns the SQL for creating a reference to the column.
func sqlType(colType schema.ColumnType, size uint64, subType schema.ColumnSubType) string {
	switch colType {
	case schema.ColTypeAutoPrimaryKey:
		typ := intType(size, false)
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
	case schema.ColTypeEnum:
		return "INT"
	case schema.ColTypeJSON:
		fallthrough
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
