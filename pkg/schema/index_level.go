package schema

import (
	"encoding/json"
	"fmt"
)

// IndexLevel indicates the type of index the column should have. Not all databases have indexes,
// and some databases index everything, so this doesn't precisely translate to an eventual index
// in the database. What this DOES do is determine the primary key and LoadByXXX functions that will
// be generated in the API.
//
// # IndexLevelNone is the default indicating no special treatment for the column
//
// IndexLevelIndexed will result in a LoadByXXX function in the generated ORM that will
// return a group of objects containing the given value in the column's field.
// This will be for a single column. To create an index on multiple columns, use an Index.
//
// IndexLevelUnique will result in a LoadByXXX function in the ORM that will return a single object with
// the given value in the column's field. Uniqueness is up to the database or database driver to ensure.
// Note that some databases (aka MongoDB) do not allow unique constraints on nullable fields to have more
// than one null value in the database, in which case the application will need custom logic to enforce
// uniqueness, rather than relying on the database.
// See also the comment on uniqueness in Index.
//
// IndexLevelPrimaryKey indicates that this is a private key.
// Only one column or one index can be specified with this index level, in a table.
// Use an Index to specify a composite primary key.
type IndexLevel int

const (
	IndexLevelNone IndexLevel = iota
	IndexLevelIndexed
	IndexLevelUnique
	IndexLevelPrimaryKey
)

func (il IndexLevel) String() string {
	switch il {
	case IndexLevelNone:
		return "None"
	case IndexLevelIndexed:
		return "Indexed"
	case IndexLevelUnique:
		return "Unique"
	case IndexLevelPrimaryKey:
		return "Primary"
	default:
		return "Unknown"
	}
}

// MarshalJSON implements custom JSON serialization for IndexLevel
func (il IndexLevel) MarshalJSON() ([]byte, error) {
	return json.Marshal(il.String())
}

func (il IndexLevel) jsonRep() string {
	switch il {
	case IndexLevelNone:
		return "none"
	case IndexLevelIndexed:
		return "indexed"
	case IndexLevelUnique:
		return "unique"
	case IndexLevelPrimaryKey:
		return "primary"
	default:
		return "unknown"
	}
}

// UnmarshalJSON implements custom JSON deserialization for IndexLevel
func (il *IndexLevel) UnmarshalJSON(data []byte) error {
	var levelStr string
	if err := json.Unmarshal(data, &levelStr); err != nil {
		return err
	}

	switch levelStr {
	case "none":
		*il = IndexLevelNone
	case "indexed":
		*il = IndexLevelIndexed
	case "unique":
		*il = IndexLevelUnique
	case "primary":
		*il = IndexLevelPrimaryKey
	default:
		return fmt.Errorf("invalid IndexLevel: %s", levelStr)
	}

	return nil
}
