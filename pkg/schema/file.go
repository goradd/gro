package schema

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

func WriteJsonFile(schemas []*Database, outFile string) {
	// Serialize the struct to JSON with indentation for readability
	jsonData, err := json.MarshalIndent(schemas, "", "  ")
	if err != nil {
		log.Fatal("Error marshaling JSON:", err)
		return
	}

	// Create or open a file for writing
	file, err := os.Create(outFile)
	if err != nil {
		log.Fatal("Error creating file:", err)
		return
	}
	defer file.Close()

	// Write the JSON data to the file
	_, err = file.Write(jsonData)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}
}

func ReadJsonFile(infile string) (schemas []*Database, err error) {
	// Read the file's content
	var bytes []byte
	bytes, err = os.ReadFile(infile)
	if err != nil {
		return
	}

	err = json.Unmarshal(bytes, &schemas)
	return
}
