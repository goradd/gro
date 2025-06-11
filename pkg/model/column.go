package model

import (
	"fmt"
	"github.com/goradd/orm/pkg/db"
	. "github.com/goradd/orm/pkg/query"
	"github.com/goradd/orm/pkg/schema"
	strings2 "github.com/goradd/strings"
	"log/slog"
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
	// Label is the name that identifies the value to humans. e.g. "First Name".
	Label string
	// SchemaType is the type info specified in the schema description
	SchemaType schema.ColumnType
	// SchemaSubType further describes the schema type.
	SchemaSubType schema.ColumnSubType
	//  ReceiverType indicates the Go type that matches the column.
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
	// If this is an enum type, EnumTable will point to that enum table
	EnumTable *Enum
	// Options are the options extracted from the comments string
	Options map[string]interface{}

	Reference *Reference // if a reference column, the containing reference

	// goType is the cached go type as a string for the column
	goType string
	// decapIdentifier is a cache for the lower case identifier for the column.
	decapIdentifier string
}

func (cd *Column) String() string {
	return cd.Identifier
}

// DefaultConstantName returns the name of the default value constant that will be used to refer to the default value
func (cd *Column) DefaultConstantName() string {
	return cd.Table.Identifier + cd.Identifier + "Default"
}

func (cd *Column) VariableIdentifier() string {
	return cd.decapIdentifier
}

func (cd *Column) VariableIdentifierPlural() string {
	return strings2.Plural(cd.decapIdentifier)
}

// DefaultValueAsValue returns the default value of the column as a GO value
func (cd *Column) DefaultValueAsValue() string {
	if cd.DefaultValue == nil {
		if cd.SchemaType == schema.ColTypeAutoPrimaryKey {
			return `""`
		} else if cd.IsEnum() {
			return "0"
		}
		return cd.ReceiverType.DefaultValueString()
	}

	if cd.ReceiverType == ColTypeTime {
		if cd.DefaultValue == CreatedTime || cd.DefaultValue == ModifiedTime {
			return "time.Time{}" // These times will be updated when the object is saved.
		} else {
			t := cd.DefaultValue.(time.Time)
			if t.IsZero() {
				return "time.Time{}"
			}
			return fmt.Sprintf("time2.NewDateTime(%d, %d, %d, %d, %d, %d, %d)", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond())
		}
	}

	if cd.IsEnum() {
		return fmt.Sprintf("%s(%d)", cd.EnumTable.Identifier, cd.DefaultValue) // should be casting an int to an enum type
	}

	return fmt.Sprintf("%#v", cd.DefaultValue)
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
func (cd *Column) JsonKey() string {
	return cd.decapIdentifier
}

// IsEnum returns true if the column contains a type defined by an enum table.
func (cd *Column) IsEnum() bool {
	return cd.SchemaType == schema.ColTypeEnum
}

// IsDecimal returns true if the column is a Decimal (Numeric) string value, meaning
// is a variable precision decimal number.
func (cd *Column) IsDecimal() bool {
	return cd.SchemaSubType == schema.ColSubTypeNumeric
}

// HasDefaultValue returns true if the column has a default value.
func (cd *Column) HasDefaultValue() bool {
	return cd.SchemaType == schema.ColTypeAutoPrimaryKey || cd.DefaultValue != nil
}

// GoType returns the Go type of the internal member variable corresponding to the column.
func (cd *Column) GoType() string {
	return cd.goType
}

// HasSetter returns true if the column should be allowed to be set by the programmer. Some columns should not be alterable,
// including AutoID columns, and time based columns that automatically set or update their times.
func (cd *Column) HasSetter() bool {
	if cd.ReceiverType == ColTypeTime {
		if cd.DefaultValue == CreatedTime || cd.DefaultValue == ModifiedTime {
			return false
		}
	}
	if cd.SchemaSubType == schema.ColSubTypeTimestamp ||
		cd.SchemaSubType == schema.ColSubTypeLock {
		return false
	}
	return true
}

// MaxInt returns the maximum integer that the column can hold if it is an integer type.
// Returns 0 if not.
func (cd *Column) MaxInt() int64 {
	if cd.ReceiverType == ColTypeInteger {
		switch cd.Size {
		case 8:
			return 127
		case 16:
			return 32767
		case 24:
			return 8388607
		case 32:
			return 2147483647
		}
	} else if cd.ReceiverType == ColTypeUnsigned {
		switch cd.Size {
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
func (cd *Column) MinInt() int64 {
	if cd.ReceiverType == ColTypeInteger {
		switch cd.Size {
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
func (cd *Column) MaxLength() uint64 {
	if cd.SchemaSubType == schema.ColSubTypeNumeric {
		return cd.Size&0x0000ffff + 2 // allow for +/- and decimal point
	}
	return cd.Size
}

// DecimalPrecision returns the precision value of a decimal number that is packed into the Size value.
func (cd *Column) DecimalPrecision() uint64 {
	if cd.SchemaSubType == schema.ColSubTypeNumeric {
		return cd.Size & 0x0000ffff
	}
	return 0
}

// DecimalScale returns the scale value of a decimal number when it is packed into the Size value.
func (cd *Column) DecimalScale() uint64 {
	if cd.SchemaSubType == schema.ColSubTypeNumeric {
		return cd.Size >> 16
	}
	return 0
}

func (m *Database) importColumn(schemaCol *schema.Column) *Column {
	col := &Column{
		QueryName:       schemaCol.Name,
		Identifier:      schemaCol.Identifier,
		Label:           schemaCol.Label,
		SchemaType:      schemaCol.Type,
		SchemaSubType:   schemaCol.SubType,
		ReceiverType:    ReceiverTypeFromSchema(schemaCol.Type, schemaCol.Size),
		Size:            schemaCol.Size,
		DefaultValue:    schemaCol.DefaultValue,
		IsNullable:      schemaCol.IsNullable,
		decapIdentifier: strings2.Decap(schemaCol.Identifier),
	}

	if schemaCol.Type == schema.ColTypeEnum {
		col.EnumTable = m.Enum(schemaCol.EnumTable)
		col.goType = col.EnumTable.Identifier
	} else {
		col.goType = col.ReceiverType.GoType()
	}

	if (col.SchemaSubType == schema.ColSubTypeTimestamp || col.SchemaSubType == schema.ColSubTypeLock) &&
		col.IsNullable {
		slog.Warn("Column should not be nullable. Nullable status will be ignored.",
			slog.String(db.LogColumn, col.QueryName))
		col.IsNullable = false
	}

	return col
}
