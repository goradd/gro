package codegen

import (
	"github.com/goradd/orm/pkg/model"
	"io"
)

type Template interface {
	Overwrite() bool
}

type DatabaseGenerator interface {
	Template
	FileName(string) string
	GenerateDatabase(*model.Database, io.Writer) error
}

type TableGenerator interface {
	Template
	FileName(*model.Table) string
	GenerateTable(table *model.Table, f io.Writer, importPath string) error
}

type EnumGenerator interface {
	Template
	FileName(*model.Enum) string
	GenerateEnum(table *model.Enum, f io.Writer, importPath string) error
}

var templates []Template

func RegisterTemplate(t Template) {
	templates = append(templates, t)
}
