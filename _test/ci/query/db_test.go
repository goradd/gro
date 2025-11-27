package query

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/goradd/gro/_test/config"
	"github.com/goradd/gro/_test/gen/orm/goradd"
)

func TestMain(m *testing.M) {
	os.Exit(runTests(m))
}

func runTests(m *testing.M) int {
	setup(m)
	defer teardown()
	return m.Run()
}

func setup(m *testing.M) {
	fmt.Println("Setting up tests...")

	config.InitDB()
	ctx := context.Background()
	goradd.ClearAll(ctx)
	loadData(ctx)
}

func teardown() {
	// Cleanup logic here
	fmt.Println("Cleaning up after tests...")
}

func loadData(ctx context.Context) {
	f, err := os.Open("./../../schema/data.json")
	if err != nil {
		panic(err)
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			panic(err)
		}
	}(f)
	err = goradd.JsonDecodeAll(ctx, f)
	if err != nil {
		panic(fmt.Errorf("error loading data: %w", err))
	}
}
