package sqlite

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/goradd/anyutil"
	"github.com/goradd/gro/db"
	schema2 "github.com/goradd/gro/internal/schema"
)

// TableDefinitionSql will return the sql needed to create the table.
// This can include clauses separated by semicolons that add additional capabilities to the table.
func (m *DB) TableDefinitionSql(d *schema2.Database, table *schema2.Table) (tableSql string, extraClauses []string) {
	var sb strings.Builder

	var columnDefs []string
	var tableClauses []string

	for _, col := range table.Columns {
		colDef, tc, xc := m.buildColumnDef(col)
		if colDef == "" {
			continue // error, already reported
		}
		columnDefs = append(columnDefs, "  "+colDef)
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
		tSql, extraSql := m.indexSql(table, mci)
		if tSql != "" {
			tableClauses = append(tableClauses, tSql)
		}
		if extraSql != "" {
			extraClauses = append(extraClauses, extraSql)
		}
	}

	tableName := table.QualifiedName()
	columnDefs = append(columnDefs, tableClauses...)
	if table.Comment != "" {
		cmt := fmt.Sprintf("COMMENT ON TABLE %s IS '%s'", m.QuoteIdentifier(tableName), table.Comment)
		extraClauses = append(extraClauses, cmt)
	}

	sb.WriteString(fmt.Sprintf("CREATE TABLE %s (\n", m.QuoteIdentifier(tableName)))
	sb.WriteString(strings.Join(columnDefs, ",\n"))
	sb.WriteString("\n)")
	return sb.String(), extraClauses
}

// ColumnDefinitionSql returns the sql that will create the column col.
// This will include single-column foreign key references.
// This will not include a primary key designation.
func (m *DB) buildColumnDef(col *schema2.Column) (s string, tableClauses []string, extraClauses []string) {
	var colType string
	var collation string
	var defaultStr string
	var extraStr string

	if def := col.DatabaseDefinition[db.DriverTypeSQLite]; def != nil {
		if c, ok := def["collation"].(string); ok && c != "" {
			collation = fmt.Sprintf(`COLLATE %s`, c) // BINARY (default), NOCASE, or RTRIM only
		}
		if d, ok := def["default"].(string); ok && d != "" {
			defaultStr = " DEFAULT " + d
		}

	}
	if col.Type == schema2.ColTypeEnum {
		if col.EnumTable == "" {
			slog.Error("Column skipped, Enum not specified for an enum value.",
				slog.String(db.LogColumn, col.Name))
			return
		}

		fk := fmt.Sprintf(" FOREIGN KEY (%s) REFERENCES %s(%s)",
			m.QuoteIdentifier(col.Name),
			m.QuoteIdentifier(col.EnumTable),
			m.QuoteIdentifier(schema2.ValueKey))
		tableClauses = append(tableClauses, fk)
		colType = "INTEGER" // NOT INT!
	} else {
		colType = sqlType(col.Type, col.Size, col.SubType)
		if col.Type == schema2.ColTypeAutoPrimaryKey {
			extraStr += "PRIMARY KEY AUTOINCREMENT"
		}
	}

	if !col.IsNullable {
		colType += " NOT NULL"
	}

	if col.DefaultValue != nil && defaultStr == "" {
		switch val := col.DefaultValue.(type) {
		case string:
			if col.Type == schema2.ColTypeTime {
				if val == "now" {
					defaultStr = "DEFAULT CURRENT_TIMESTAMP"
				} else if val == "update" {
					// The way to do this is through a trigger. Since we are providing the value programmatically, we will punt on it.
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

	s = fmt.Sprintf("%s %s %s %s %s %s", m.QuoteIdentifier(col.Name), colType, extraStr, defaultStr, collation, commentStr)
	return
}

// SqlType is used by the builder to return the SQL corresponding to the given colType that will create
// the column.
// If forReference is true, then it returns the SQL for creating a reference to the column.
func sqlType(colType schema2.ColumnType, size uint64, subType schema2.ColumnSubType) string {
	switch colType {
	case schema2.ColTypeAutoPrimaryKey:
		return "INTEGER"
	case schema2.ColTypeString:
		if subType == schema2.ColSubTypeNumeric {
			return "TEXT" // NUMERIC is not infinite precision, just allows sqlite to convert to int or real as needed.
		} else {
			return "TEXT"
		}
	case schema2.ColTypeBytes:
		return "BLOB"
	case schema2.ColTypeInt, schema2.ColTypeUint:
		return "INTEGER"
	case schema2.ColTypeFloat:
		return "REAL"
	case schema2.ColTypeBool:
		return "INTEGER"
	case schema2.ColTypeTime:
		switch subType {
		case schema2.ColSubTypeDateOnly:
			return "TEXT"
		case schema2.ColSubTypeTimeOnly:
			return "TEXT"
		case schema2.ColSubTypeNone:
			return "INTEGER"
		default:
			slog.Warn("Wrong subtype for time column",
				slog.String("subtype", subType.String()))
		}
		return "INTEGER"
	case schema2.ColTypeEnum:
		return "INTEGER"
	case schema2.ColTypeJSON:
		return "BLOB" // use new jsonb format
	default:
		return "TEXT"
	}
}

// indexSql returns sql to be included after a table definition that will create an
// index on the columns. table should NOT include the schema, since index names only need to
// be unique within the schema in postgres.
func (m *DB) indexSql(table *schema2.Table, i *schema2.Index) (tableSql string, extraSql string) {
	quotedCols := anyutil.MapSliceFunc(i.Columns, func(s string) string {
		return m.QuoteIdentifier(s)
	})

	var idxType string
	switch i.IndexLevel {
	case schema2.IndexLevelPrimaryKey:
		idxType = "PRIMARY KEY" // constraint with auto generated index
	case schema2.IndexLevelUnique:
		idxType = "UNIQUE" // constraint with auto generated index
	case schema2.IndexLevelIndexed:
		// regular indexes must be added after the table definition
		idx_name := i.Name
		extraSql = fmt.Sprintf("CREATE INDEX %s ON %s (%s)",
			m.QuoteIdentifier(idx_name),
			m.QuoteIdentifier(table.Name),
			strings.Join(quotedCols, ","))
		return
	default:
		return
	}
	// handle primary and unique
	// single column indexes are declared in the table definition
	if len(i.Columns) == 1 {
		col2 := table.FindColumn(i.Columns[0])
		if col2 != nil {
			if col2.Type == schema2.ColTypeAutoPrimaryKey {
				return // auto primary keys are dealt with inline in the table
			}
		}
	}

	tableSql = fmt.Sprintf("%s (%s)", idxType, strings.Join(quotedCols, ","))

	return
}

func (m *DB) buildReferenceDef(db *schema2.Database, table *schema2.Table, ref *schema2.Reference) (columnClause string, tableClauses, extraClauses []string) {
	fk, pk := ref.ReferenceColumns(db, table)

	if fk.Type == schema2.ColTypeAutoPrimaryKey {
		fk.Type = schema2.ColTypeInt // auto columns internally are integers
	}

	columnClause, tableClauses, extraClauses = m.buildColumnDef(fk)
	if columnClause == "" {
		return // error, already logged
	}

	// SQLite only supports foreign keys defined at the table level, which means
	// cyclic foreign keys are not possible, and the ref.Table must already exist.
	s := fmt.Sprintf("FOREIGN KEY (%s) REFERENCES %s(%s)",
		m.QuoteIdentifier(fk.Name),
		m.QuoteIdentifier(ref.Table),
		m.QuoteIdentifier(pk.Name))
	tableClauses = append(tableClauses, s)
	return
}
