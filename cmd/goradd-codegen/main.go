package main

import (
	"flag"
	"fmt"
	"github.com/goradd/gofile/pkg/sys"
	"github.com/goradd/orm/pkg/codegen"
	"github.com/goradd/orm/pkg/schema"
	_ "github.com/goradd/orm/tmpl/template"
	"log/slog"
	"os"
	"path/filepath"
)

func main() {
	/*
		f, err := os.Create("cpu.prof")
		if err != nil {
			panic(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	*/
	/*
		f, _ := os.Create("trace.out")
		trace.Start(f)
		defer trace.Stop()
	*/
	var schemaFile string
	var outdir string

	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	flag.StringVar(&schemaFile, "c", "", "Path to database configuration file")
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
		var d string
		d, err = filepath.Abs(outdir)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "error with output directory path: %s", err)
			os.Exit(1)
		}
		if err = os.MkdirAll(d, 0777); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "could not create output directory: %s", err)
			os.Exit(1)
		}
		err = os.Chdir(d)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "cannot change directory to %s: %s", d, err)
			os.Exit(1)
		}
		defer func() { _ = os.Chdir(cwd) }()
	}

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
	if schemaDB.Package == "" {
		schemaDB.Package = base
	}
	if schemaDB.ImportPath == "" {
		ip, err := sys.ImportPath(".") // current dir is the outdir
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Could not determine import path: %s", err)
			os.Exit(1)
		}
		schemaDB.ImportPath = ip
	}
	codegen.Generate(schemaDB)
}
