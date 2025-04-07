package main

import (
	"flag"
	"fmt"
	"github.com/goradd/orm/pkg/codegen"
	"github.com/goradd/orm/pkg/schema"
	_ "github.com/goradd/orm/tmpl/template"
	"os"
	"path/filepath"
)

func main() {
	var schemaFile string
	var outdir string

	cwd, _ := os.Getwd()

	flag.StringVar(&schemaFile, "s", "", "Path to schema file")
	flag.StringVar(&outdir, "o", "", "Path to output directory")
	flag.Parse()

	if schemaFile == "" {
		_, _ = fmt.Fprintf(os.Stderr, "Path to schema file is required")
		os.Exit(1)
	} else {
		var err error
		schemaFile, err = filepath.Abs(schemaFile)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "cannot find schema file %s: %s", schemaFile, err)
			os.Exit(1)
		}
	}

	if outdir != "" {
		d, err := filepath.Abs(outdir)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "cannot find schema file %s: %s", schemaFile, err)
			os.Exit(1)
		}
		err = os.Chdir(d)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "cannot change directory to %s: %s", outdir, err)
			os.Exit(1)
		}
	}
	defer func() { _ = os.Chdir(cwd) }()

	var err error

	var schemas []*schema.Database
	schemas, err = schema.ReadJsonFile(schemaFile)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error opening or reading schema file %s: %s", schemaFile, err)
		os.Exit(1)
	}

	codegen.Generate(schemas)
}
