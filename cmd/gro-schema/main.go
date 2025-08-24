package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	var configFile string
	var outFile string
	var inFile string
	var dbKey string

	flag.StringVar(&configFile, "c", "", "Path to database configuration file")
	flag.StringVar(&outFile, "o", "", "Path to schema output file")
	flag.StringVar(&inFile, "i", "", "Path to schema input file")
	flag.StringVar(&dbKey, "k", "", "Key of database to use")

	flag.Parse()

	if configFile == "" {
		_, _ = fmt.Fprintf(os.Stderr, "Path to database configuration file is required")
		os.Exit(1)
	}
	if outFile == "" && inFile == "" {
		_, _ = fmt.Fprintf(os.Stderr, "Path to input file or output file is required")
		os.Exit(1)
	}
	if dbKey == "" {
		_, _ = fmt.Fprintf(os.Stderr, "Database key is required")
		os.Exit(1)
	}

	if outFile != "" {
		extract(configFile, outFile, dbKey)
	} else {
		build(configFile, inFile, dbKey)
	}
}
