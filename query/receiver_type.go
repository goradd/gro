package query

import (
	"fmt"
	"github.com/goradd/gro/internal/schema"
	"time"
)

// ReceiverType represents the Go type that a query will be received as.
type ReceiverType int

const (
	ColTypeUnknown ReceiverType = iota
	ColTypeBytes
	ColTypeString
	ColTypeInteger
	ColTypeInteger64
	ColTypeTime
	ColTypeFloat32
	ColTypeFloat64
	ColTypeBool
	ColTypeAutoPrimaryKey
	ColTypeUUID
	ColTypeULID
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
	case ColTypeInteger64:
		return "ColTypeInteger64"
	case ColTypeTime:
		return "ColTypeTime"
	case ColTypeFloat32:
		return "ColTypeFloat32"
	case ColTypeFloat64:
		return "ColTypeFloat64"
	case ColTypeBool:
		return "ColTypeBool"
	case ColTypeAutoPrimaryKey:
		return "ColTypeAutoPrimaryKey"
	case ColTypeUUID:
		return "ColTypeUUID"
	case ColTypeULID:
		return "ColTypeULID"
	}
	return ""
}

// DefaultValue returns a zero Go value for the type
func (g ReceiverType) DefaultValue() any {
	switch g {
	case ColTypeUnknown:
		return []byte{}
	case ColTypeBytes:
		return []byte{}
	case ColTypeString:
		return ""
	case ColTypeInteger:
		return 0
	case ColTypeInteger64:
		return int64(0)
	case ColTypeTime:
		return time.Time{}
	case ColTypeFloat32:
		return float32(0.0)
	case ColTypeFloat64:
		return float64(0.0)
	case ColTypeBool:
		return false
	case ColTypeAutoPrimaryKey:
		// For references. Actual primary keys handled earlier.
		return ZeroAutoPrimaryKey()
	case ColTypeUUID:
		return NewUUID()
	case ColTypeULID:
		return NewULID()
	}
	return nil
}

// GoType returns the actual GO type as go code
func (g ReceiverType) GoType() string {
	if g == ColTypeUnknown || g == ColTypeBytes {
		return "[]byte" // otherwise we get a []unit8
	} else if g == ColTypeAutoPrimaryKey {
		return "query.AutoPrimaryKey"
	} else if g == ColTypeUUID {
		return "query.UUID"
	} else if g == ColTypeULID {
		return "query.ULID"
	}
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
	case ColTypeAutoPrimaryKey:
		return "query.AutoPrimaryKey{}"
	case ColTypeUUID:
		return "query.UUID{}"
	case ColTypeULID:
		return "query.ULID{}"
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
// If the column is a ReferenceType, columnType should instead be the type of the primary key in the referenced table.
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
		return ColTypeAutoPrimaryKey
	case schema.ColTypeJSON:
		return ColTypeString
	case schema.ColTypeEnum:
		return ColTypeInteger
	case schema.ColTypeUUID:
		return ColTypeUUID
	case schema.ColTypeULID:
		return ColTypeULID
	}
	return ColTypeUnknown
}
