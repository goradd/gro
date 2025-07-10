package ci

import (
	"context"
	"fmt"
	"github.com/goradd/orm/_test/config"
	"github.com/goradd/orm/_test/gen/orm/goradd"
	"os"
	"testing"
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
	f, err := os.Open("./../schema/data.json")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	goradd.ClearAll(ctx)
	err = goradd.JsonDecodeAll(ctx, f)
	if err != nil {
	}
}
