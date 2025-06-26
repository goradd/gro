package sqlite

import (
	"fmt"
	"github.com/goradd/anyutil"
	"github.com/goradd/orm/pkg/db"
	"github.com/goradd/orm/pkg/schema"
	"log/slog"
	"strings"
)

// TableDefinitionSql will return the sql needed to create the table.
// This can include clauses separated by semicolons that add additional capabilities to the table.
func (m *DB) TableDefinitionSql(d *schema.Database, table *schema.Table) (tableSql string, extraClauses []string) {
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
		tSql, extraSql := m.indexSql(table, mci.IndexLevel, mci.Columns...)
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
func (m *DB) buildColumnDef(col *schema.Column) (s string, tableClauses []string, extraClauses []string) {
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
	if col.Type == schema.ColTypeEnum {
		if col.EnumTable == "" {
			slog.Error("Column skipped, EnumTable not specified for an enum value.",
				slog.String(db.LogColumn, col.Name))
			return
		}

		fk := fmt.Sprintf(" FOREIGN KEY (%s) REFERENCES %s(%s)",
			m.QuoteIdentifier(col.Name),
			m.QuoteIdentifier(col.EnumTable),
			m.QuoteIdentifier("const"))
		tableClauses = append(tableClauses, fk)
		colType = "INT"
	} else {
		colType = sqlType(col.Type, col.Size, col.SubType)
		if col.Type == schema.ColTypeAutoPrimaryKey {
			extraStr += "PRIMARY KEY AUTOINCREMENT"
		}
	}

	if !col.IsNullable {
		colType += " NOT NULL"
	}

	if col.DefaultValue != nil && defaultStr == "" {
		switch val := col.DefaultValue.(type) {
		case string:
			if col.Type == schema.ColTypeTime {
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
func sqlType(colType schema.ColumnType, size uint64, subType schema.ColumnSubType) string {
	switch colType {
	case schema.ColTypeAutoPrimaryKey:
		return "INTEGER"
	case schema.ColTypeString:
		if subType == schema.ColSubTypeNumeric {
			return "TEXT" // NUMERIC is not infinite precision, just allows sqlite to convert to int or real as needed.
		} else {
			return "TEXT"
		}
	case schema.ColTypeBytes:
		return "BLOB"
	case schema.ColTypeInt, schema.ColTypeUint:
		return "INTEGER"
	case schema.ColTypeFloat:
		return "REAL"
	case schema.ColTypeBool:
		return "INTEGER"
	case schema.ColTypeTime:
		switch subType {
		case schema.ColSubTypeDateOnly:
			return "TEXT"
		case schema.ColSubTypeTimeOnly:
			return "TEXT"
		case schema.ColSubTypeNone:
			return "INTEGER"
		default:
			slog.Warn("Wrong subtype for time column",
				slog.String("subtype", subType.String()))
		}
		return "INTEGER"
	case schema.ColTypeEnum:
		return "INTEGER"
	case schema.ColTypeJSON:
		return "BLOB" // use new jsonb format
	default:
		return "TEXT"
	}
}

// indexSql returns sql to be included after a table definition that will create an
// index on the columns. table should NOT include the schema, since index names only need to
// be unique within the schema in postgres.
func (m *DB) indexSql(table *schema.Table, level schema.IndexLevel, cols ...string) (tableSql string, extraSql string) {
	quotedCols := anyutil.MapSliceFunc(cols, func(s string) string {
		return m.QuoteIdentifier(s)
	})

	var idxType string
	switch level {
	case schema.IndexLevelPrimaryKey:
		idxType = "PRIMARY KEY"
	case schema.IndexLevelUnique:
		idxType = "UNIQUE INDEX"
	case schema.IndexLevelIndexed:
		// regular indexes must be added after the table definition
		idx_name := "idx_" + table.Name + "_" + strings.Join(cols, "_")
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
	if len(cols) == 1 {
		col2 := table.FindColumn(cols[0])
		if col2 != nil {
			if col2.Type == schema.ColTypeAutoPrimaryKey {
				return // auto primary keys are dealt with inline in the table
			}
		}
	}

	tableSql = fmt.Sprintf("%s (%s)", idxType, strings.Join(quotedCols, ","))

	return
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

	// SQLite only supports foreign keys defined at the table level, which means
	// cyclic foreign keys are not possible, and the ref.Table must already exist.
	s := fmt.Sprintf("FOREIGN KEY (%s) REFERENCES %s(%s)",
		m.QuoteIdentifier(fk.Name),
		m.QuoteIdentifier(ref.Table),
		m.QuoteIdentifier(pk.Name))
	tableClauses = append(tableClauses, s)
	return
}
