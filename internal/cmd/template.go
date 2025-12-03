package cmd

import (
	"io"

	model2 "github.com/goradd/gro/internal/model"
)

type Template interface {
	Overwrite() bool
}

type DatabaseGenerator interface {
	Template
	FileName(string) string
	GenerateDatabase(db *model2.Database, f io.Writer, importPath string) error
}

type TableGenerator interface {
	Template
	FileName(*model2.Table) string
	GenerateTable(table *model2.Table, f io.Writer, importPath string) error
}

type EnumGenerator interface {
	Template
	FileName(*model2.Enum) string
	GenerateEnum(table *model2.Enum, f io.Writer, importPath string) error
}

var templates []Template

func RegisterTemplate(t Template) {
	templates = append(templates, t)
}
