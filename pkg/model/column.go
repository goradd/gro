package model

import (
	"fmt"
	"github.com/goradd/orm/pkg/db"
	. "github.com/goradd/orm/pkg/query"
	"github.com/goradd/orm/pkg/schema"
	strings2 "github.com/goradd/strings"
	"log/slog"
	"slices"
	"time"
)

const (
	CreatedTime  = "now"
	ModifiedTime = "update"
)

// Column describes a database column.
// The information is derived from the schema structure.
type Column struct {
	// QueryName is the name of the column in the database.
	QueryName string
	// Table is a pointer back to the table that this column is part of
	Table *Table
	// Identifier is the name of the accessor function in go code. e.g. "FirstName".
	Identifier string
	// Field is the name of the field in the Table struct that will hold the column's value.
	// It is also used as the name for a function parameter.
	Field string
	// FieldPlural is the plural form of Field.
	FieldPlural string
	// Label is the name that identifies the value to humans. e.g. "First Name".
	Label string
	// Type is the Go type that the column represents to the rest of the Go application.
	Type string
	// SchemaType is the type info specified in the schema description
	SchemaType schema.ColumnType
	// SchemaSubType further describes the schema type.
	SchemaSubType schema.ColumnSubType
	//  ReceiverType is the Go type that will be received from the database.
	ReceiverType ReceiverType
	// Size is the maximum length of runes to allow in the column if a string type column.
	// If a byte array, it is the number of bytes permitted.
	// If an attempt at entering more than this amount occurs, it is considered a programming bug
	// and we will panic.
	// If an integer or float type, it is the number of bits in the data type.
	Size uint64
	// DefaultValue is the default value as specified by the database. We will initialize new ORM objects
	// with this value.
	DefaultValue interface{}
	// IsNullable is true if the column can be given a NULL value.
	IsNullable bool
	// If this is an enum type, Enum will point to the Enum object.
	Enum *Enum
	// If this column is a reference, a pointer to the Reference object.
	Reference *Reference
	// Options are the options extracted from the comments string
	Options map[string]interface{}
}

func (c *Column) String() string {
	return c.Identifier
}

// DefaultConstantName returns the name of the default value constant that will be used to refer to the default value
func (c *Column) DefaultConstantName() string {
	return c.Table.Identifier + c.Identifier + "Default"
}

// DefaultValueAsValue returns the default value of the column as a GO value
func (c *Column) DefaultValueAsValue() string {
	if c.DefaultValue == nil {
		if c.SchemaType == schema.ColTypeAutoPrimaryKey {
			return `""`
		} else if c.IsEnum() {
			return "0"
		}
		return c.ReceiverType.DefaultValueString()
	}

	if c.ReceiverType == ColTypeTime {
		if c.DefaultValue == CreatedTime || c.DefaultValue == ModifiedTime {
			return "time.Time{}" // These times will be updated when the object is saved.
		} else {
			t := c.DefaultValue.(time.Time)
			if t.IsZero() {
				return "time.Time{}"
			}
			return fmt.Sprintf("time2.NewDateTime(%d, %d, %d, %d, %d, %d, %d)", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond())
		}
	}

	if c.IsEnum() {
		return fmt.Sprintf("%s(%d)", c.Enum.Identifier, c.DefaultValue) // should be casting an int to an enum type
	}

	return fmt.Sprintf("%#v", c.DefaultValue)
}

/*
// DefaultValueAsConstant returns the default value of the column as a Go constant
func (cd *Column) DefaultValueAsConstant() string {
	if cd.ReceiverType == ColTypeTime {
		if cd.DefaultValue == CreatedTime || cd.DefaultValue == ModifiedTime {
			return `time2.Current`
		} else if cd.DefaultValue == nil {
			return `time2.Zero`
		} else {
			d := cd.DefaultValue.(time.Time)
			if b, _ := d.MarshalText(); b == nil {
				return `time2.Zero`
			} else {
				s := string(b[:])
				return fmt.Sprintf("%#v", s)
			}
		}
	} else if cd.DefaultValue == nil || cd.IsAutoPk {
		v := cd.ReceiverType.DefaultValueString()
		if v == "nil" {
			return ""
		}
		return cd.ReceiverType.DefaultValueString()
	} else {
		return fmt.Sprintf("%#v", cd.DefaultValue)
	}
}
*/

// JsonKey returns the key used for the column when outputting JSON.
func (c *Column) JsonKey() string {
	return c.Field
}

// IsEnum returns true if the column contains a type defined by an enum table.
func (c *Column) IsEnum() bool {
	return c.SchemaType == schema.ColTypeEnum
}

// IsDecimal returns true if the column is a Decimal (Numeric) string value, meaning
// is a variable precision decimal number.
func (c *Column) IsDecimal() bool {
	return c.SchemaSubType == schema.ColSubTypeNumeric
}

// HasDefaultValue returns true if the column has a default value.
func (c *Column) HasDefaultValue() bool {
	return c.SchemaType == schema.ColTypeAutoPrimaryKey || c.DefaultValue != nil
}

// IsThePrimaryKey returns true if this is the single primary key of its table.
func (c *Column) IsThePrimaryKey() bool {
	return c.Table.PrimaryKeyColumn() == c
}

// IsAPrimaryKey returns true if this is one of the primary keys.
func (c *Column) IsAPrimaryKey() bool {
	return slices.Contains(c.Table.primaryKeyColumns, c)
}

// HasSetter returns true if the column should be allowed to be set by the programmer. Some columns should not be alterable,
// including AutoID columns, and time based columns that automatically set or update their times.
func (c *Column) HasSetter() bool {
	if c.ReceiverType == ColTypeTime {
		if c.DefaultValue == CreatedTime || c.DefaultValue == ModifiedTime {
			return false
		}
	}
	if c.SchemaSubType == schema.ColSubTypeTimestamp ||
		c.SchemaSubType == schema.ColSubTypeLock {
		return false
	}
	return true
}

// MaxInt returns the maximum integer that the column can hold if it is an integer type.
// Returns 0 if not.
func (c *Column) MaxInt() int64 {
	if c.ReceiverType == ColTypeInteger {
		switch c.Size {
		case 8:
			return 127
		case 16:
			return 32767
		case 24:
			return 8388607
		case 32:
			return 2147483647
		}
	} else if c.ReceiverType == ColTypeUnsigned {
		switch c.Size {
		case 8:
			return 255
		case 16:
			return 65535
		case 24:
			return 16777215
		case 32:
			return 4294967295
		}
	}
	return 0
}

// MinInt returns the minimum integer that the column can hold if it is an integer type.
// Returns 0 if not.
func (c *Column) MinInt() int64 {
	if c.ReceiverType == ColTypeInteger {
		switch c.Size {
		case 8:
			return -128
		case 16:
			return -32768
		case 24:
			return -8388608
		case 32:
			return -2147483648
		}
	}
	return 0
}

// MaxLength returns the maximum length of the column, which normally is Column.Size, but
// in certain situations might be something else.
func (c *Column) MaxLength() uint64 {
	if c.SchemaSubType == schema.ColSubTypeNumeric {
		return c.Size&0x0000ffff + 2 // allow for +/- and decimal point
	}
	return c.Size
}

// DecimalPrecision returns the precision value of a decimal number that is packed into the Size value.
func (c *Column) DecimalPrecision() uint64 {
	if c.SchemaSubType == schema.ColSubTypeNumeric {
		return c.Size & 0x0000ffff
	}
	return 0
}

// DecimalScale returns the scale value of a decimal number when it is packed into the Size value.
func (c *Column) DecimalScale() uint64 {
	if c.SchemaSubType == schema.ColSubTypeNumeric {
		return c.Size >> 16
	}
	return 0
}

// IsAutoPK returns true of the column is an auto-generated primary key.
func (c *Column) IsAutoPK() bool {
	return c.SchemaType == schema.ColTypeAutoPrimaryKey
}

func (m *Database) importColumn(schemaCol *schema.Column) *Column {
	col := &Column{
		QueryName:     schemaCol.Name,
		Identifier:    schemaCol.Identifier,
		Field:         strings2.Decap(schemaCol.Identifier),
		FieldPlural:   strings2.Plural(strings2.Decap(schemaCol.Identifier)),
		Label:         schemaCol.Label,
		SchemaType:    schemaCol.Type,
		SchemaSubType: schemaCol.SubType,
		ReceiverType:  ReceiverTypeFromSchema(schemaCol.Type, schemaCol.Size),
		Size:          schemaCol.Size,
		DefaultValue:  schemaCol.DefaultValue,
		IsNullable:    schemaCol.IsNullable,
	}

	if schemaCol.Type == schema.ColTypeEnum {
		col.Enum = m.Enum(schemaCol.EnumTable)
		col.Type = col.Enum.Identifier
	} else {
		col.Type = col.ReceiverType.GoType()
	}

	if (col.SchemaSubType == schema.ColSubTypeTimestamp || col.SchemaSubType == schema.ColSubTypeLock) &&
		col.IsNullable {
		slog.Warn("Column should not be nullable. Nullable status will be ignored.",
			slog.String(db.LogColumn, col.QueryName))
		col.IsNullable = false
	}

	return col
}
