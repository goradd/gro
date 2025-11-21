package main

import "fmt"

type UUID struct {
	i1 uint64
	i2 uint64
}

type AutoId string

func main() {
	var bob AutoId

	bob = "me"

	var i interface{} = bob
	switch i.(type) {
	case string:
		fmt.Printf("String")
	case AutoId:
		fmt.Printf("Auto ID")

	}

}
