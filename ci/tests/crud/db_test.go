package crud

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/goradd/gro/ci/config"
	"github.com/goradd/gro/ci/tests/gen/goradd_unit"
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
	goradd_unit.ClearAll(ctx)
}

func teardown() {
	// Cleanup logic here
	fmt.Println("Cleaning up after tests...")
}
