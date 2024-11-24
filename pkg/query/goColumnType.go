package query

import (
	"spekary/goradd/orm/pkg/schema"
	"strconv"
	"time"
)

// ReceiverType represents the Go type that a query will be received as.
type ReceiverType int

const (
	ColTypeUnknown ReceiverType = iota
	ColTypeBytes
	ColTypeString
	ColTypeInteger
	ColTypeUnsigned
	ColTypeInteger64
	ColTypeUnsigned64
	ColTypeTime
	ColTypeFloat32
	ColTypeFloat64
	ColTypeBool
)

// String returns the constant type name as a string
func (g ReceiverType) String() string {
	switch g {
	case ColTypeUnknown:
		return "ColTypeUnknown"
	case ColTypeBytes:
		return "ColTypeBytes"
	case ColTypeString:
		return "ColTypeString"
	case ColTypeInteger:
		return "ColTypeInteger"
	case ColTypeUnsigned:
		return "ColTypeUnsigned"
	case ColTypeInteger64:
		return "ColTypeInteger64"
	case ColTypeUnsigned64:
		return "ColTypeUnsigned64"
	case ColTypeTime:
		return "ColTypeTime"
	case ColTypeFloat32:
		return "ColTypeFloat32"
	case ColTypeFloat64:
		return "ColTypeFloat64"
	case ColTypeBool:
		return "ColTypeBool"
	}
	return ""
}

// GoType returns the actual GO type as go code
func (g ReceiverType) GoType() string {
	switch g {
	case ColTypeUnknown:
		return "Unknown"
	case ColTypeBytes:
		return "[]byte"
	case ColTypeString:
		return "string"
	case ColTypeInteger:
		return "int"
	case ColTypeUnsigned:
		return "uint"
	case ColTypeInteger64:
		return "int64"
	case ColTypeUnsigned64:
		return "uint64"
	case ColTypeTime:
		return "time.Time"
	case ColTypeFloat32:
		return "float32" // always internally represent with max bits
	case ColTypeFloat64:
		return "float64" // always internally represent with max bits
	case ColTypeBool:
		return "bool"
	}
	return ""
}

// DefaultValue returns a string that represents the GO default value for the corresponding type
func (g ReceiverType) DefaultValue() string {
	switch g {
	case ColTypeUnknown:
		return ""
	case ColTypeBytes:
		return ""
	case ColTypeString:
		return "\"\""
	case ColTypeInteger:
		return "0"
	case ColTypeUnsigned:
		return "0"
	case ColTypeInteger64:
		return "0"
	case ColTypeUnsigned64:
		return "0"
	case ColTypeTime:
		return "time.Time{}"
	case ColTypeFloat32:
		return "0.0" // always internally represent with max bits
	case ColTypeFloat64:
		return "0.0" // always internally represent with max bits
	case ColTypeBool:
		return "false"
	}
	return ""
}

func ColTypeFromGoTypeString(name string) ReceiverType {
	switch name {
	case "Unknown":
		return ColTypeUnknown
	case "[]byte":
		return ColTypeBytes
	case "string":
		return ColTypeString
	case "int":
		return ColTypeInteger
	case "uint":
		return ColTypeUnsigned
	case "int64":
		return ColTypeInteger64
	case "uint64":
		return ColTypeUnsigned64
	case "time.Time":
		return ColTypeTime
	case "float32":
		return ColTypeFloat32
	case "float64":
		return ColTypeFloat64
	case "bool":
		return ColTypeBool
	default:
		panic("unknown column go type " + name)
	}
}

// FromString will convert from a string to the correct Go type.
// Time strings must be in RFC3339 format.
func (g ReceiverType) FromString(s string) any {
	switch g {
	case ColTypeUnknown:
		return nil
	case ColTypeBytes:
		return nil
	case ColTypeString:
		return s
	case ColTypeInteger:
		if s == "" {
			return int(0)
		}
		i, _ := strconv.Atoi(s)
		return i
	case ColTypeUnsigned:
		if s == "" {
			return uint(0)
		}
		i, _ := strconv.ParseUint(s, 10, 64)
		return uint(i)
	case ColTypeInteger64:
		if s == "" {
			return int64(0)
		}
		i, _ := strconv.ParseInt(s, 10, 64)
		return i
	case ColTypeUnsigned64:
		if s == "" {
			return uint64(0)
		}
		i, _ := strconv.ParseUint(s, 10, 64)
		return i
	case ColTypeTime:
		if s == "" {
			return time.Time{}
		}
		d, _ := time.Parse(time.RFC3339, s)
		return d
	case ColTypeFloat32:
		if s == "" {
			return float32(0)
		}
		f, _ := strconv.ParseFloat(s, 32)
		return float32(f)
	case ColTypeFloat64:
		if s == "" {
			return float64(0)
		}
		f, _ := strconv.ParseFloat(s, 32)
		return f
	case ColTypeBool:
		return s == "true"
	}
	return ""
}

// ReceiverTypeFromSchema converts a schema column type to a Go language type.
// If maxLength is zero, the default will be chosen.
func ReceiverTypeFromSchema(columnType schema.ColumnType, maxLength uint64) ReceiverType {
	switch columnType {
	case schema.ColTypeUnknown:
		return ColTypeUnknown
	case schema.ColTypeBytes:
		return ColTypeBytes
	case schema.ColTypeString:
		return ColTypeString
	case schema.ColTypeInt:
		if maxLength == 64 {
			return ColTypeInteger64
		}
		return ColTypeInteger
	case schema.ColTypeUint:
		if maxLength == 64 {
			return ColTypeUnsigned64
		}
		return ColTypeUnsigned
	case schema.ColTypeTime:
		return ColTypeTime
	case schema.ColTypeFloat:
		if maxLength == 32 {
			return ColTypeFloat32
		}
		return ColTypeFloat64
	case schema.ColTypeBool:
		return ColTypeBool
	case schema.ColTypeAutoPrimaryKey:
		return ColTypeString
	case schema.ColTypeJSON:
		return ColTypeString
	case schema.ColTypeReference:
		// Note that in the case of references to manually entered foreign keys, they
		// will always get queried as strings. Also, if this is pointing to an enum table,
		// the caller will need to change it to a ColTypeInt.
		return ColTypeString
	}
	return ColTypeUnknown
}
