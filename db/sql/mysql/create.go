package mysql

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/goradd/anyutil"
	"github.com/goradd/gro/db"
	schema2 "github.com/goradd/gro/internal/schema"
)

// TableDefinitionSql will return the sql needed to create the table.
// This can include clauses separated by semicolons.
// All the returned tableSql items will be executed for all tables before extraSql will be executed.
// This allows extraSql to refer to other tables in the schema that might not have been created yet,
// and is particularly useful for handling cyclic foreign key references.
func (m *DB) TableDefinitionSql(d *schema2.Database, table *schema2.Table) (tableSql string, extraClauses []string) {
	var sb strings.Builder

	var columnDefs []string
	var tableClauses []string

	for _, col := range table.Columns {
		cc, tc, xc := m.buildColumnDef(col, false)
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
		def := m.indexSql(mci)
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
func (m *DB) buildColumnDef(col *schema2.Column, isFkToAuto bool) (columnClause string, tableClauses []string, extraClauses []string) {
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
	if isFkToAuto {
		colType = "BINARY(16)"
	} else if col.Type == schema2.ColTypeEnum {
		if col.EnumTable == "" {
			slog.Error("Column skipped, Enum not specified for an enum value.",
				slog.String(db.LogColumn, col.Name))
			return
		}

		// foreign key will automatically index the column
		fk := fmt.Sprintf(" FOREIGN KEY (%s) REFERENCES %s(%s)",
			m.QuoteIdentifier(col.Name),
			m.QuoteIdentifier(col.EnumTable),
			m.QuoteIdentifier(schema2.ValueKey))
		tableClauses = append(tableClauses, fk)
		colType = "INT"
	} else {
		colType = sqlType(col.Type, col.Size, col.SubType)
		if col.Type == schema2.ColTypeAutoPrimaryKey {
			extraStr += "AUTO_INCREMENT"
		}
	}

	if !col.IsNullable {
		colType += " NOT NULL "
	}

	if col.DefaultValue != nil && defaultStr == "" {
		switch val := col.DefaultValue.(type) {
		case string:
			if col.Type == schema2.ColTypeTime {
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

func (m *DB) indexSql(idx *schema2.Index) string {
	var idxType string
	switch idx.IndexLevel {
	case schema2.IndexLevelPrimaryKey:
		idxType = "PRIMARY KEY"
	case schema2.IndexLevelUnique:
		idxType = "UNIQUE"
	case schema2.IndexLevelIndexed:
		idxType = "INDEX"
	default:
		return ""
	}
	cols := anyutil.MapSliceFunc(idx.Columns, func(s string) string {
		return m.QuoteIdentifier(s)
	})
	idxCols := strings.Join(cols, ", ")
	def := fmt.Sprintf("%s %s (%s)", idxType, idx.Name, idxCols)
	return def
}

func (m *DB) buildReferenceDef(db *schema2.Database, table *schema2.Table, ref *schema2.Reference) (columnClause string, tableClauses, extraClauses []string) {
	fk, pk := ref.ReferenceColumns(db, table)

	if fk.Type == schema2.ColTypeAutoPrimaryKey {
		fk.Type = schema2.ColTypeInt // auto columns internally are integers
		fk.Size = 32
	}

	columnClause, tableClauses, extraClauses = m.buildColumnDef(fk, fk.Type == schema2.ColTypeAutoPrimaryKey)
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
func sqlType(colType schema2.ColumnType, size uint64, subType schema2.ColumnSubType) string {
	switch colType {
	case schema2.ColTypeAutoPrimaryKey:
		typ := intType(size, false)
		return typ
	case schema2.ColTypeString:
		if subType == schema2.ColSubTypeNumeric {
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
	case schema2.ColTypeBytes:
		if size == 0 {
			return "BLOB"
		} else if size < 65532 {
			return fmt.Sprintf("VARBINARY(%d)", size)
		} else if size < 16777215 {
			return "MEDIUMBLOB"
		} else {
			return "LONGBLOB"
		}
	case schema2.ColTypeInt:
		return intType(size, false)
	case schema2.ColTypeUint:
		return intType(size, true)
	case schema2.ColTypeFloat:
		if size == 32 {
			return "FLOAT"
		}
		return "DOUBLE"
	case schema2.ColTypeBool:
		return "BOOLEAN"
	case schema2.ColTypeTime:
		switch subType {
		case schema2.ColSubTypeDateOnly:
			return "DATE"
		case schema2.ColSubTypeTimeOnly:
			return "TIME"
		case schema2.ColSubTypeNone:
			return "DATETIME"
		default:
			slog.Warn("Wrong subtype for time column",
				slog.String("subtype", subType.String()))
		}
		return "DATETIME"
	case schema2.ColTypeEnum:
		return "INT"
	case schema2.ColTypeJSON:
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
