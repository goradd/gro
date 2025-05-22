package ci

import (
	"fmt"
	"github.com/goradd/orm/_test/config"
	"github.com/goradd/orm/_test/gen/orm/goradd"
	"github.com/goradd/orm/pkg/db"
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
	loadData()
}

func teardown() {
	// Cleanup logic here
	fmt.Println("Cleaning up after tests...")
}

func loadData() {
	f, err := os.Open("./../schema/data.json")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	ctx := db.NewContext(nil)
	goradd.ClearAll(ctx)
	err = goradd.JsonDecodeAll(ctx, f)
	if err != nil {
	}
}
