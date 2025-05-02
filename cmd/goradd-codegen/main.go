package main

import (
	"flag"
	"fmt"
	"github.com/goradd/orm/pkg/codegen"
	"github.com/goradd/orm/pkg/schema"
	_ "github.com/goradd/orm/tmpl/template"
	"log/slog"
	"os"
	"path/filepath"
)

func main() {
	var schemaFile string
	var outdir string

	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	flag.StringVar(&schemaFile, "s", "", "Path to schema file")
	flag.StringVar(&outdir, "o", "", "Path to output directory")
	flag.Parse()

	if schemaFile == "" {
		_, _ = fmt.Fprintf(os.Stderr, "Path to schema file is required")
		os.Exit(1)
	} else {
		schemaFile, err = filepath.Abs(schemaFile)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "cannot find schema file %s: %s", schemaFile, err)
			os.Exit(1)
		}
	}

	if outdir != "" {
		d, err2 := filepath.Abs(outdir)
		if err2 != nil {
			_, _ = fmt.Fprintf(os.Stderr, "cannot find schema file %s: %s", schemaFile, err2)
			os.Exit(1)
		}
		err = os.Chdir(d)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "cannot change directory to %s: %s", outdir, err)
			os.Exit(1)
		}
	}
	defer func() { _ = os.Chdir(cwd) }()

	var schemaDB *schema.Database
	schemaDB, err = schema.ReadJsonFile(schemaFile)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error opening or reading schema file %s: %s", schemaFile, err)
		os.Exit(1)
	}

	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level:     slog.LevelInfo, // optional: control log level
		AddSource: true,
	})
	logger := slog.New(handler)
	slog.SetDefault(logger)

	base := filepath.Base(outdir)
	schemaDB.Package = base
	codegen.Generate(schemaDB)
}
