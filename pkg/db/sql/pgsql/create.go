package pgsql

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

	quotedTableName := m.QuoteIdentifier(table.Name)
	if table.Schema != "" && table.Schema != "public" {
		quotedTableName = m.QuoteIdentifier(table.Schema) + "." + m.QuoteIdentifier(quotedTableName)
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

	sb.WriteString(fmt.Sprintf("CREATE TABLE %s (\n", quotedTableName))
	sb.WriteString(strings.Join(columnDefs, ",\n"))
	sb.WriteString("\n)")
	sb.WriteString(";\n")
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

	if def := col.DatabaseDefinition[db.DriverTypePostgres]; def != nil {
		if t, ok := def["type"].(string); ok {
			colType = t
		}
		if c, ok := def["collation"].(string); ok && c != "" {
			collation = fmt.Sprintf(`COLLATE "%s"`, c)
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
	// Note: Postgres does not automatically index foreign keys
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
		typ := intType(size, false)
		if !forReference {
			typ += " GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY"
		}
		return typ
	case schema.ColTypeString:
		if subType == schema.ColSubTypeNumeric {
			precision := size & 0x0000FFFF
			scale := size >> 16
			if precision != 0 && scale != 0 {
				return fmt.Sprintf("NUMERIC(%d, %d)", precision, scale)
			} else {
				return "NUMERIC" // max precision available in postgres
			}
		} else if size == 0 {
			return "TEXT"
		} else if size < 16383 { // this is arbitrary really. Postgres behaves differently when the whole row exceeds 2KB.
			return fmt.Sprintf("VARCHAR(%d)", size)
		} else {
			return "TEXT"
		}
	case schema.ColTypeBytes:
		return "BYTEA"
	case schema.ColTypeInt:
		return intType(size, false)
	case schema.ColTypeUint:
		return intType(size, true)
	case schema.ColTypeFloat:
		if size == 32 {
			return "FLOAT4"
		}
		return "FLOAT8"
	case schema.ColTypeBool:
		return "BOOLEAN"
	case schema.ColTypeTime:
		switch subType {
		case schema.ColSubTypeDateOnly:
			return "DATE"
		case schema.ColSubTypeTimeOnly:
			return "TIME"
		case schema.ColSubTypeNone:
			return "TIMESTAMP"
		default:
			slog.Warn("Wrong subtype for time column",
				slog.String("subtype", subType.String()))
		}
		return "TIMESTAMP"
	case schema.ColTypeReference:
		return intType(size, false)
	case schema.ColTypeEnum:
		return "INT"
	case schema.ColTypeJSON:
		fallthrough
	case schema.ColTypeEnumArray:
		return "JSONB"
	default:
		return "TEXT"
	}
}

func intType(size uint64, unsigned bool) string {
	var t string

	if unsigned {
		size += 1 // postgres does not support unsigned values, so we need to make sure we pick a type
		// that covers Go's possible matching unsigned values.
	}
	switch {
	case size == 0:
		t = "INT"
	case size == 1: // unsigned int
		t = "BIGINT"
	case size <= 16:
		t = "SMALLINT"
	case size <= 32:
		t = "INT"
	default:
		t = "BIGINT"
	}
	return t
}
