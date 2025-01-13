package query

import (
	"fmt"
	"github.com/goradd/orm/pkg/schema"
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

// DefaultValue returns a zero Go value for the type
func (g ReceiverType) DefaultValue() any {
	switch g {
	case ColTypeUnknown:
		return []byte(nil)
	case ColTypeBytes:
		return []byte(nil)
	case ColTypeString:
		return ""
	case ColTypeInteger:
		return int(0)
	case ColTypeUnsigned:
		return uint(0)
	case ColTypeInteger64:
		return int64(0)
	case ColTypeUnsigned64:
		return uint64(0)
	case ColTypeTime:
		return time.Time{}
	case ColTypeFloat32:
		return float32(0.0)
	case ColTypeFloat64:
		return float64(0.0)
	case ColTypeBool:
		return false
	}
	return nil
}

// GoType returns the actual GO type as go code
func (g ReceiverType) GoType() string {
	t := g.DefaultValue()
	if t != nil {
		return fmt.Sprintf("%T", g.DefaultValue())
	} else {
		return "[]byte" // all unknown types are byte slices
	}
}

// DefaultValueString returns a string that represents the GO default value for the corresponding type
func (g ReceiverType) DefaultValueString() string {
	switch g {
	case ColTypeTime:
		return "time.Time{}"
	default:
		v := g.DefaultValue()
		if v == nil {
			return "nil"
		}
		return fmt.Sprintf("%#v", v)
	}
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
	case schema.ColTypeEnum:
		return ColTypeInteger
	case schema.ColTypeReference:
		// Note that in the case of references to manually entered foreign keys, they
		// will always get queried as strings.
		return ColTypeString
	}
	return ColTypeUnknown
}
