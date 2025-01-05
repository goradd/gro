package ci

import (
	"fmt"
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
	
	InitDB()
}

func teardown() {
	// Cleanup logic here
	fmt.Println("Cleaning up after tests...")
}
