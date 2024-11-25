package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"spekary/goradd/orm/pkg/codegen"
	"spekary/goradd/orm/pkg/schema"
	_ "spekary/goradd/orm/tmpl/template"
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
			log.Panicf("cannot find schema file %s: %s", schemaFile, err)
		}
	}

	if outdir != "" {
		d, err := filepath.Abs(outdir)
		if err != nil {
			log.Panicf("cannot find directory %s: %s", outdir, err)
		}
		err = os.Chdir(d)
		if err != nil {
			log.Panicf("cannot change directory to %s: %s", outdir, err)
		}
	}
	defer func() { _ = os.Chdir(cwd) }()

	var err error

	var schemas []*schema.Database
	schemas, err = schema.ReadJsonFile(schemaFile)
	if err != nil {
		log.Panic("Error opening schema file:", err)
	}

	codegen.Generate(schemas)
}
