package schema

import "encoding/json"

// ColumnType is a general specifier for the type of column in a database.
//
// Mapping data types in Go to data types in databases is tricky, and necessarily
// puts limitations on what kinds of data can be represented and how it is stored.
// The ORM provides some flexibility by allowing each column to specify its database
// type, while also specifying its Go type by using a combination of the DatabaseType,
// VariableType and Size fields as described below.
//
// # ColTypeUnknown
//
// This is a column type that is unknown to the orm and that will be provided to Go as a []byte slice.
// Many databases have non-standard or custom data types or types that simply don't fit well in Go.
// For example, the NUMERIC or DECIMAL type in many databases is an exact precision decimal
// number for which there is no equivalent in Go, since Go uses floating point numbers to represent
// decimal numbers. To provide data of this type to your application, you can create a custom
// type in Go, and then create accessor functions that translate between the database representation
// and your custom type.
//
// See Column.DatabaseDefinition for a way to specify the database specific type info for this column.
//
// # ColTypeBytes
//
// This is known as a BLOB type in many SQL databases, and is represented as a []byte slice in Go.
// If Column.Size is non-zero, the data definition will use this in the schema definition.
// Different databases treat this differently, and may or may not truncate the data internally.
//
// Attempting to insert a byte blob with a size bigger than Column.Size will panic in
// order to protect the integrity of the database. However, if a process outside the ORM
// sets the value to a size bigger than Column.Size, the value can be read by the ORM and it
// will not be truncated.
//
// # ColTypeString
//
// ColTypeString is always represented as a string in Go, but may have a variety of representations in
// the database depending on the Size value and the database.
//
// In order to protect data integrity, if a Column.Size is non-zero, the generated code will panic if an attempt is made to
// set the value to a string whose rune size is greater than Column.Size. Note that rune size is not the same as
// byte size. If a process outside the ORM sets the value larger than Column.Size runes, the value
// can be read by the ORM and will not be truncated.
//
// Postgres doc states the TEXT type is the fastest and most efficient text storage type,
// and by default this will always be the corresponding data type in Postgres.
//
// MySQL will use a VarChar(Size) if a Column.Size is present, and a Text if not. Mysql stores VarChar
// variables inside the table, and Text data outside of the table with a reference, so VarChar is more efficient.
//
// By default, collation will be by UTF8 case-sensitive rules. Set the Sort to case-insensitive to
// change this to case-insensitive. Other collations are outside the capabilities of the ORM and you
// should directly edit the database to set those up.
//
// # ColTypeInt and ColTypeUint
//
// These integer column types will be represented in Go as int or uint, and in the database as a
// 32-bit integer or unsigned integer by default. To specifically set the storage size, specify either
// 8, 16, 32 or 64 in the Size value, and the corresponding Go type will be used, as well as
// the corresponding type in the database. Not all databases support 8-bit integers. Check
// your database vendor to be sure.
//
// An int64 column may have the subtype of ColSubTypeTimestamp or ColSubTypeLock.
// ColSubTypeTimestamp indicates that the column will be filled in with time.Now().UnixMicro() when the record is saved.
// However, if a new value is computed, and the old value is greater than the new value, the new value will add one to the
// value. This prevents the scenario where separate systems with unsynchronized clocks might create a later
// value that would appear earlier, in case the timestamp is being used for UI synchronization. Combine with another
// column with ColSubTypeLock to make sure a race condition between systems will not possibly break this process,
// though if the app is not running scaled on multiple systems, this is unnecessary.
// ColSubTypeLock columns will store a version number and be used by the ORM to implement optimistic locking.
// Convention is to name these columns "gro_timestamp" and "gro_lock".
//
// # ColTypeTime
//
// This corresponds to a datetime in most databases. The value is stored in UTC time
// in the database, and when read back from the database, will initially be in UTC time.
// Usually this is what is wanted, since it allows times from different timezones to be correctly
// sorted, and javascript and other libraries are capable of converting UTC time to local
// time in the client locale for display purposes.
// MySQL will store the value as a DateTime and not a Timestamp, since Timestamps are assumed
// to be in server local time and not UTC time and get time shifted in transit. Also, some MySQLs have the
// yr2038 bug. Postgres uses Timestamp without timezone (which is the default).
// See the Column.DefaultValue doc for time specific behavior of default values.
//
// # ColTypeFloat
//
// By default, this will be a float64 value in Go, and double precision float in the database.
// Specify a MaxSize of 32 to make this a float32 and single precision float in the database (if supported).
// Note float32 only has 7 digits of precision, while float64 has 15.
//
// # ColTypeBool
//
// ColTypeBool is a bool in Go, and is database dependent in the database. MySQL uses a BIT column,
// and Postgres uses a boolean.
//
// # ColTypeAutoPrimaryKey
//
// These are auto-generated primary keys generated by the database or database driver,
// and are represented in Go as a string, even though databases can use a variety of types for
// a primary key internally. This is primarily for portability, so that if you change databases, you do not have to change
// your code. MySQL will use an auto incremented integer internally, and Postgres will use a serial type, which is an
// auto incremented integer. It is up to the database driver to ensure that the primary key is unique. The primary
// key is used in reference columns to create associations between tables.
//
// The ORM allows multi-column primary keys in the database, however such keys are not ColTypePrimaryKey, but rather
// whatever the type is they are assigned in the database. They are treated as multi-column unique indexes by the ORM,
// and cannot be auto-generated by the database driver, but rather must be set by the application.
//
// # ColTypeJSON
//
// This special type is supported by many databases and allows querying and data retrieval from
// within JSON documents stored in a field in a database. In Go, querying the entire field will
// result in a string type value.
//
// # ColTypeReference
//
// This is a column that contains the value of a primary key of a table, creating a reference to the
// object in that table. This is known as a foreign key in SQL databases or an edge in graph databases.
//
// # ColTypeEnum and ColTypeEnumArray
//
// Enum columns contain values of enumerated types that are described by Enum tables in the schema.
// ColTypeEnum is an integer value in the database, and ColTypeEnumArray is stored as a JSON array of integers.
// Note that filtering queries on ColTypeEnumArray columns may be limited based on your database type.
type ColumnType int

const (
	ColTypeUnknown ColumnType = iota
	ColTypeBytes
	ColTypeString
	ColTypeInt
	ColTypeUint
	ColTypeTime
	ColTypeFloat
	ColTypeBool
	ColTypeAutoPrimaryKey
	ColTypeJSON
	ColTypeReference
	ColTypeEnum
	ColTypeEnumArray
)

// GroTimestampColumnName is the convention for the name of a ColSubTypeTimestamp column
// that will automatically be updated with the UnixMicro time upon saving of the record.
const GroTimestampColumnName = "gro_timestamp"

// GroLockColumnName is the convention for the name of a ColSubTypeLock column
// that will also be used to perform optimistic locking by the ORM.
const GroLockColumnName = "gro_lock"

// String returns the string representation of a ColumnType.
func (ct ColumnType) String() string {
	switch ct {
	case ColTypeBytes:
		return "bytes"
	case ColTypeString:
		return "string"
	case ColTypeInt:
		return "int"
	case ColTypeUint:
		return "uint"
	case ColTypeTime:
		return "time"
	case ColTypeFloat:
		return "float"
	case ColTypeBool:
		return "bool"
	case ColTypeAutoPrimaryKey:
		return "auto_primary_key"
	case ColTypeJSON:
		return "json"
	case ColTypeReference:
		return "ref"
	case ColTypeEnum:
		return "enum"
	case ColTypeEnumArray:
		return "enum_array"
	default:
		return "unknown"
	}
}

// MarshalJSON customizes how ColumnType is serialized to JSON.
func (ct ColumnType) MarshalJSON() ([]byte, error) {
	// Return the string representation of the ReceiverType
	return json.Marshal(ct.String())
}

// UnmarshalJSON customizes how ColumnType is deserialized from JSON.
func (ct *ColumnType) UnmarshalJSON(data []byte) error {
	var ctStr string
	if err := json.Unmarshal(data, &ctStr); err != nil {
		return err
	}

	// Match the string representation and assign the corresponding ReceiverType value
	switch ctStr {
	case "bytes":
		*ct = ColTypeBytes
	case "string":
		*ct = ColTypeString
	case "int":
		*ct = ColTypeInt
	case "uint":
		*ct = ColTypeUint
	case "time":
		*ct = ColTypeTime
	case "float":
		*ct = ColTypeFloat
	case "bool":
		*ct = ColTypeBool
	case "auto_primary_key":
		*ct = ColTypeAutoPrimaryKey
	case "json":
		*ct = ColTypeJSON
	case "ref":
		*ct = ColTypeReference
	case "enum":
		*ct = ColTypeEnum
	case "enum_array":
		*ct = ColTypeEnumArray
	default:
		*ct = ColTypeUnknown
	}
	return nil
}

// ColumnSubType provides more description to a particular type
type ColumnSubType int

const (
	// ColSubTypeNone indicates no sub type and is the default
	ColSubTypeNone ColumnSubType = iota
	// ColSubTypeDateOnly is for time columns that only contain a date.
	// The time will be in UTC and will have zero values for hour, minute, seconds, and nanoseconds
	ColSubTypeDateOnly
	// ColSubTypeTimeOnly is for time columns that have no date component.
	// Date components will be 1-1-1 at timezone will be UTC.
	// Care must be taken when converting string times, since time.Parse will return a time with date component 1-1-0,
	// however, some database drivers will not accept a 1-1-0 time, even when setting a time-only column.
	ColSubTypeTimeOnly
	// ColSubTypeTimestamp is an int64 column that will automatically be given the UnixMicro value when a record is
	// successfully inserted or updated.
	ColSubTypeTimestamp
	// ColSubTypeLock is an int64 column that contains a record version number that will be used to optimistically
	// lock the record while saving. The value is automatically generated and will be unique even across multiple
	// instances of the application.
	ColSubTypeLock
	// ColSubTypeNumeric is a string that only accepts numeric values, as in positive or negative numbers with
	// a decimal point. SQL databases use a NUMERIC or DECIMAL column, and NoSQL generally just supports this natively.
	// This is not a floating point number, as those numbers can have precision limitations.
	// Because of this, numeric values are preferred when storing currency values.
	// See the math/big package for working with these types of values in Go.
	ColSubTypeNumeric
)

// String returns the string representation of a ColumnType.
func (ct ColumnSubType) String() string {
	switch ct {
	case ColSubTypeNone:
		return "none"
	case ColSubTypeDateOnly:
		return "date_only"
	case ColSubTypeTimeOnly:
		return "time_only"
	case ColSubTypeTimestamp:
		return "gro_timestamp"
	case ColSubTypeLock:
		return "gro_lock"
	case ColSubTypeNumeric:
		return "numeric"
	default:
		return "none"
	}
}

// MarshalJSON customizes how ColumnType is serialized to JSON.
func (cst ColumnSubType) MarshalJSON() ([]byte, error) {
	// Return the string representation of the ReceiverType
	return json.Marshal(cst.String())
}

// UnmarshalJSON customizes how ColumnType is deserialized from JSON.
func (cst *ColumnSubType) UnmarshalJSON(data []byte) error {
	var cstStr string
	if err := json.Unmarshal(data, &cstStr); err != nil {
		return err
	}

	// Match the string representation and assign the corresponding ReceiverType value
	switch cstStr {
	case "date_only":
		*cst = ColSubTypeDateOnly
	case "time_only":
		*cst = ColSubTypeTimeOnly
	case "gro_timestamp":
		*cst = ColSubTypeTimestamp
	case "gro_lock":
		*cst = ColSubTypeLock
	case "numeric":
		*cst = ColSubTypeNumeric
	default:
		*cst = ColSubTypeNone
	}
	return nil
}
