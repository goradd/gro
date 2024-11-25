package codegen

import (
	"io"
	"spekary/goradd/orm/pkg/model"
)

type Template interface {
	FileName(string) string
	Overwrite() bool
}

type DatabaseGenerator interface {
	GenerateDatabase(*model.Database, io.Writer) error
}

type TableGenerator interface {
	GenerateTable()
}

type EnumGenerator interface {
	GenerateEnum()
}

type AssociationGenerator interface {
	GenerateAssociation()
}

var templates []Template

func RegisterTemplate(t Template) {
	templates = append(templates, t)
}
