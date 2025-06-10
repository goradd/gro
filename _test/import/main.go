package main

import (
	"flag"
	"fmt"
	"github.com/goradd/orm/_test/gen/orm/goradd"
	"github.com/goradd/orm/pkg/config"
	"github.com/goradd/orm/pkg/db"
	"os"
)

func main() {
	var configFile string
	var inFile string

	flag.StringVar(&configFile, "c", "", "Path to database configuration file")
	flag.StringVar(&inFile, "i", "", "Path to input file")

	flag.Parse()

	if configFile == "" {
		_, _ = fmt.Fprintf(os.Stderr, "Path to database configuration file is required")
		os.Exit(1)
	} else if inFile == "" {
		_, _ = fmt.Fprintf(os.Stderr, "Path to input file is required")
		os.Exit(1)
	}

	decode(configFile, inFile)
}

func decode(dbConfigFile, inFile string) {
	if databaseConfigs, err := config.OpenConfigFile(dbConfigFile); err != nil {
		panic(err)
	} else if err := config.InitDatastore(databaseConfigs); err != nil {
		panic(err)
	}
	ctx := db.NewContext(nil)

	f, err := os.Open(inFile)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	goradd.ClearAll(ctx)
	err = goradd.JsonDecodeAll(ctx, f)
	if err != nil {
		panic(err)
	}
}
