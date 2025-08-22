package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/goradd/orm/_test/gen/orm/goradd"
	"github.com/goradd/orm/pkg/config"
	_ "github.com/goradd/orm/tmpl/template"
)

func main() {
	var configFile string
	var outFile string

	flag.StringVar(&configFile, "c", "", "Path to database configuration file")
	flag.StringVar(&outFile, "o", "", "Path to output file")

	flag.Parse()

	if configFile == "" {
		_, _ = fmt.Fprintf(os.Stderr, "Path to database configuration file is required")
		os.Exit(1)
	} else if outFile == "" {
		_, _ = fmt.Fprintf(os.Stderr, "Path to output file is required")
		os.Exit(1)
	}

	encode(configFile, outFile)
}

func encode(dbConfigFile, outFile string) {
	if databaseConfigs, err := config.OpenConfigFile(dbConfigFile); err != nil {
		panic(err)
	} else if err := config.InitDatastore(databaseConfigs); err != nil {
		panic(err)
	}
	ctx := context.Background()

	f, err := os.Create(outFile)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	err = goradd.JsonEncodeAll(ctx, f)
	if err != nil {
		panic(err)
	}
}
