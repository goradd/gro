package model

import "github.com/goradd/anyutil"

// Index will create accessor functions related to Columns.
type Index struct {
	// IsUnique indicates whether the index is unique
	IsUnique bool
	// Columns are the columns that are part of the index
	Columns []*Column
}

func (i *Index) Name() string {
	return anyutil.Join(i.Columns, "_")
}
