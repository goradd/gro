package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"spekary/goradd/orm/pkg/codegen"
	"spekary/goradd/orm/pkg/schema"
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
	}

	if outdir != "" {
		err := os.Chdir(outdir)
		if err != nil {
			log.Panic("cannot change directory to ", outdir, err)
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
