package schema

import (
	"encoding/json"
	"fmt"
)

// IndexLevel indicates the type of index the column should have. Not all databases have indexes,
// and some databases index everything, so this doesn't precisely translate to an eventual index
// in the database.
//
// # IndexLevelNone is the default indicating no special treatment for the column
//
// IndexLevelIndexed will result in a QueryByXXX function in the generated ORM that will
// return a group of objects containing the given value in the column's field.
//
// IndexLevelUnique will result in a LoadByXXX function in the ORM that will return a single object with
// the given value in the column's field. Uniqueness is up to the database or database driver to ensure.
// Note that some databases (aka MongoDB) do not allow unique constraints on nullable fields to have more
// than one null value in the database, in which case the application will need custom logic to enforce
// uniqueness, rather than relying on the database.
// See also the comment on uniqueness in MultiColumnIndex.
//
// IndexLevelManualPrimaryKey and Unique are basically equivalent.
// IndexLevelManualPrimaryKey gives a hint to the database to mark this field as the primary key for the table.
// Only one column should be marked this way.
type IndexLevel int

const (
	IndexLevelNone IndexLevel = iota
	IndexLevelIndexed
	IndexLevelUnique
	IndexLevelManualPrimaryKey
)

func (il IndexLevel) String() string {
	switch il {
	case IndexLevelNone:
		return "None"
	case IndexLevelIndexed:
		return "Indexed"
	case IndexLevelUnique:
		return "Unique"
	case IndexLevelManualPrimaryKey:
		return "Primary"
	default:
		return "Unknown"
	}
}

// MarshalJSON implements custom JSON serialization for IndexLevel
func (il IndexLevel) MarshalJSON() ([]byte, error) {
	return json.Marshal(il.String())
}

// UnmarshalJSON implements custom JSON deserialization for IndexLevel
func (il *IndexLevel) UnmarshalJSON(data []byte) error {
	var levelStr string
	if err := json.Unmarshal(data, &levelStr); err != nil {
		return err
	}

	switch levelStr {
	case "None":
		*il = IndexLevelNone
	case "Indexed":
		*il = IndexLevelIndexed
	case "Unique":
		*il = IndexLevelUnique
	case "Primary":
		*il = IndexLevelManualPrimaryKey
	default:
		return fmt.Errorf("invalid IndexLevel: %s", levelStr)
	}

	return nil
}
