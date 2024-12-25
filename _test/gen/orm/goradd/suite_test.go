package goradd

import (
	"flag"
	"fmt"
	"github.com/goradd/orm/pkg/config"
	"os"
	"path/filepath"
	"testing"
)

func TestMain(m *testing.M) {
	setup(m)
	defer teardown()
	code := m.Run()
	os.Exit(code)
}

func setup(m *testing.M) {
	var dbConfigFile string
	flag.StringVar(&dbConfigFile, "c", "", "Path to database configuration file")
	flag.Parse()

	dir, _ := os.Getwd()
	fmt.Println("Current working directory: ", dir)

	if dbConfigFile == "" {
		panic("config file not specified. Use -args -c <filepath> to specify a database config file to the go test command.")
	}
	dbConfigFile, _ = filepath.Abs(dbConfigFile)
	fmt.Println("Initializing databases using config file: ", dbConfigFile)

	if databaseConfigs, err := config.OpenConfigFile(dbConfigFile); err != nil {
		panic(err)
	} else if err := config.InitDatastore(databaseConfigs); err != nil {
		panic(err)
	}
}

func teardown() {
	// Cleanup logic here
	fmt.Println("Cleaning up after tests...")
}
