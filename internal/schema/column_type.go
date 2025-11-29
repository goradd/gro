package schema

import (
	"encoding/json"
	"fmt"
)

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
// This is a unique value generated by the database, if the mechanism is provided for by the database,
// or the database driver if not.
// It is up to the database driver to ensure that the value is unique.
// This column is also the single primary key for the table.
// In many SQL databases, this is a serialized integer.
// Be careful about exposing serialized private keys to the world. If this number may be exposed over the
// internet, consider a UUID or ULID primary key instead, or an additional UUID column that will be the public key
// of the record.
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
// The actual column type will mirror the type of the primary key in the referenced table.
// The column will automatically be indexed.
//
// # ColTypeEnum
//
// Enum columns contain values of enumerated types that are described by Enum tables in the schema.
// ColTypeEnum is an integer value in the database.
// There is no ColTypeEnumArray because it is difficult to model cross-platform, and is better modeled
// as a series of ColTypeBool values.
//
// # ColTypeUUID
//
// UUID columns contain a UUID value. It is up to the database driver to determine the best way to store
// this in the database. For example, in MySQL, it is a binary(16), whereas in Postrgres a dedicated UUID type.
// When a record is newly created, a UUID column will be assigned a new value by the framework.
// By default, this will be a v7 UUID, but use ColSubTypeRandom to generate a v4 UUID.
//
// # ColTypeULID
//
// ULID columns contain a ULID, which is similar to a v7 UUID, but is expressed as a base32 string rather
// than a hex string. Use ColSubTypeRandom to turn this into an RULID, which is still expressed as a base32
// string, but is similar to a v4 UUID in that it is completely random.
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
	ColTypeEnum
	ColTypeUUID
	ColTypeULID
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
		return "ColTypeBytes"
	case ColTypeString:
		return "ColTypeString"
	case ColTypeInt:
		return "ColTypeInt"
	case ColTypeUint:
		return "ColTypeUint"
	case ColTypeTime:
		return "ColTypeTime"
	case ColTypeFloat:
		return "ColTypeFloat"
	case ColTypeBool:
		return "ColTypeBool"
	case ColTypeAutoPrimaryKey:
		return "ColTypeAutoPrimaryKey"
	case ColTypeJSON:
		return "ColTypeJSON"
	case ColTypeEnum:
		return "ColTypeEnum"
	case ColTypeUUID:
		return "ColTypeUUID"
	case ColTypeULID:
		return "ColTypeULID"
	default:
		return "ColTypeUnknown"
	}
}

// MarshalJSON customizes how ColumnType is serialized to JSON.
func (ct ColumnType) MarshalJSON() ([]byte, error) {
	return json.Marshal(ct.jsonRep())
}

func (ct ColumnType) jsonRep() string {
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
	case ColTypeEnum:
		return "enum"
	case ColTypeUUID:
		return "uuid"
	case ColTypeULID:
		return "ulid"
	default:
		return "unknown"
	}
}

// UnmarshalJSON customizes how ColumnType is deserialized from JSON.
func (ct *ColumnType) UnmarshalJSON(data []byte) error {
	var ctStr string
	if err := json.Unmarshal(data, &ctStr); err != nil {
		return err
	}

	// Match the string representation and assign the corresponding ReceiverType value
	switch ctStr {
	case "unknown":
		*ct = ColTypeUnknown
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
	case "enum":
		*ct = ColTypeEnum
	case "uuid":
		*ct = ColTypeUUID
	case "ulid":
		*ct = ColTypeULID
	default:
		return fmt.Errorf(`unknown column type "%s"`, ctStr)
	}
	return nil
}

// ColumnSubType provides more description to a particular type
type ColumnSubType int

const (
	// ColSubTypeNone indicates no subtype and is the default
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
	// a decimal point, and represents an arbitrary precision value.
	// Some SQL databases use a NUMERIC or DECIMAL column to represent these,
	// and NoSQL generally just supports this natively.
	// This is not a floating point number, as those numbers can have precision limitations.
	// Because of this, numeric values are preferred when storing currency values.
	// See the math/big package for working with these types of values in Go.
	//
	// Note: Although SQLite has a NUMERIC type, it
	// is not actually an arbitrary precision value, but rather in some cases could be converted to a REAL with
	// a loss of precision, so the ORM will store these as a string in SQLite. One result of this is that database numeric
	// operations will panic if attempted on these in SQLite.
	ColSubTypeNumeric
	// ColSubTypeRandom initializes UUID or ULID values to random values.
	ColSubTypeRandom
)

// String returns the string representation of a ColumnType.
func (ct ColumnSubType) String() string {
	switch ct {
	case ColSubTypeNone:
		return "ColSubTypeNone"
	case ColSubTypeDateOnly:
		return "ColSubTypeDateOnly"
	case ColSubTypeTimeOnly:
		return "ColSubTypeTimeOnly"
	case ColSubTypeTimestamp:
		return "ColSubTypeTimestamp"
	case ColSubTypeLock:
		return "ColSubTypeLock"
	case ColSubTypeNumeric:
		return "ColSubTypeNumeric"
	case ColSubTypeRandom:
		return "ColSubTypeRandom"
	default:
		return "ColSubTypeNone"
	}
}

// MarshalJSON customizes how ColumnType is serialized to JSON.
func (cst ColumnSubType) MarshalJSON() ([]byte, error) {
	// Return the string representation of the ReceiverType
	return json.Marshal(cst.jsonRep())
}

func (cst ColumnSubType) jsonRep() string {
	switch cst {
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
	case ColSubTypeRandom:
		return "random"
	default:
		return "none"
	}
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
	case "random":
		*cst = ColSubTypeRandom
	default:
		*cst = ColSubTypeNone
	}
	return nil
}
