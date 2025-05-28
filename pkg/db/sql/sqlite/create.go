package sqlite

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
		colDef, tc, xc := m.buildColumnDef(d, table, col)
		if colDef == "" {
			continue // error, already reported
		}
		columnDefs = append(columnDefs, "  "+colDef)
		tableClauses = append(tableClauses, tc...)
		extraClauses = append(extraClauses, xc...)
	}

	var quotedTableName string
	if table.Schema != "" && table.Schema != "public" {
		quotedTableName = m.QuoteIdentifier(table.Schema) + "." + m.QuoteIdentifier(table.Name)
	} else {
		quotedTableName = m.QuoteIdentifier(table.Name)
	}

	// Multi-column indexes
	for _, mci := range table.MultiColumnIndexes {
		cols := make([]string, len(mci.Columns))
		for i, name := range mci.Columns {
			cols[i] = m.QuoteIdentifier(name)
		}
		quotedCols := strings.Join(cols, ", ")

		var idxType string
		switch mci.IndexLevel {
		case schema.IndexLevelManualPrimaryKey:
			def := fmt.Sprintf("PRIMARY KEY (%s)", quotedCols)
			tableClauses = append(tableClauses, def)

		case schema.IndexLevelUnique:
			idxType = "UNIQUE "
			fallthrough
		case schema.IndexLevelIndexed:
			idxType += "INDEX"
			idxName := "idx_" + strings.Join(mci.Columns, "_")
			def := fmt.Sprintf("CREATE %s %s ON %s (%s)", idxType, m.QuoteIdentifier(idxName), quotedTableName, quotedCols)
			extraClauses = append(extraClauses, def)
		default:
			slog.Error("Index skipped. Unknown index level in multi-column index",
				slog.String(db.LogTable, table.Name))
			continue
		}
	}

	columnDefs = append(columnDefs, tableClauses...)
	if table.Comment != "" {
		cmt := fmt.Sprintf("COMMENT ON TABLE %s IS '%s'", quotedTableName, table.Comment)
		extraClauses = append(extraClauses, cmt)
	}

	if table.Schema != "" {
		// Make sure a named schema exists
		sb.WriteString(`CREATE SCHEMA IF NOT EXISTS ` + m.QuoteIdentifier(table.Schema) + ";\n")
	}

	sb.WriteString(fmt.Sprintf("CREATE TABLE %s (\n", quotedTableName))
	sb.WriteString(strings.Join(columnDefs, ",\n"))
	sb.WriteString("\n);\n")
	sb.WriteString(strings.Join(extraClauses, ";\n"))
	return sb.String()
}

// ColumnDefinitionSql returns the sql that will create the column col.
// This will include single-column foreign key references.
// This will not include a primary key designation.
func (m *DB) buildColumnDef(d *schema.Database, table *schema.Table, col *schema.Column) (s string, tableClauses []string, extraClauses []string) {
	var colType string
	var collation string
	var defaultStr string

	if def := col.DatabaseDefinition[db.DriverTypeSQLite]; def != nil {
		if c, ok := def["collation"].(string); ok && c != "" {
			collation = fmt.Sprintf(`COLLATE %s`, c) // BINARY (default), NOCASE, or RTRIM only
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

		fk := fmt.Sprintf(" FOREIGN KEY (%s) REFERENCES %s(%s)", m.QuoteIdentifier(col.Name), m.QuoteIdentifier(t.Name), m.QuoteIdentifier(c.Name))
		tableClauses = append(tableClauses, fk)
		// Note: SQLite does not automatically index foreign keys
		if col.IndexLevel == schema.IndexLevelNone {
			slog.Warn("Reference column is not indexed. Indexing reference columns is HIGHLY recommended.",
				slog.String(db.LogTable, table.Name),
				slog.String(db.LogColumn, col.Name))
		}
	} else if col.Type == schema.ColTypeEnum {
		if col.Reference == nil || col.Reference.Table == "" {
			slog.Error("Column skipped, Reference with a Table value is required",
				slog.String(db.LogColumn, col.Name))
			return
		}

		fk := fmt.Sprintf(" FOREIGN KEY (%s) REFERENCES %s(%s)", m.QuoteIdentifier(col.Name), m.QuoteIdentifier(col.Reference.Table), m.QuoteIdentifier("const"))
		tableClauses = append(tableClauses, fk)
		colType = "INT"
	} else {
		colType = sqlType(col.Type, col.Size, col.SubType, false)
	}

	if !col.IsNullable {
		colType += " NOT NULL"
	}

	var idx string
	switch col.IndexLevel {
	case schema.IndexLevelIndexed:
		idx_name := "idx_" + table.Name + "_" + col.Name
		s := fmt.Sprintf("CREATE INDEX %s ON %s (%s)", m.QuoteIdentifier(idx_name), m.QuoteIdentifier(table.Name), m.QuoteIdentifier(col.Name))
		extraClauses = append(extraClauses, s)
	case schema.IndexLevelUnique:
		idx = "UNIQUE" // inline
	case schema.IndexLevelManualPrimaryKey:
		idx = "PRIMARY KEY" // inline
	default:
		// do nothing
	}

	var extraStr string
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

	s = fmt.Sprintf("%s %s %s %s %s %s %s", m.QuoteIdentifier(col.Name), colType, idx, defaultStr, extraStr, collation, commentStr)
	return
}

// SqlType is used by the builder to return the SQL corresponding to the given colType that will create
// the column.
// If forReference is true, then it returns the SQL for creating a reference to the column.
func sqlType(colType schema.ColumnType, size uint64, subType schema.ColumnSubType, forReference bool) string {
	switch colType {
	case schema.ColTypeAutoPrimaryKey:
		typ := "INTEGER"
		if !forReference {
			typ += " PRIMARY KEY AUTOINCREMENT"
		}
		return typ
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
	case schema.ColTypeReference:
		return "INTEGER"
	case schema.ColTypeEnum:
		return "INTEGER"
	case schema.ColTypeJSON:
		return "BLOB" // use new jsonb format
	default:
		return "TEXT"
	}
}
