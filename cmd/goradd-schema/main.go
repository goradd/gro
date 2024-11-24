package main

import (
	"flag"
	"fmt"
	"os"
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

	extract(configFile, outFile)
}
