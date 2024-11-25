package schema

import (
	"bytes"
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
	var data []byte
	data, err = os.ReadFile(infile)
	if err != nil {
		return
	}

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()

	if err := decoder.Decode(&schemas); err != nil {
		panic(err)
	}

	// Fix up interfaces
	for _, s := range schemas {
		for _, t := range s.Tables {
			for _, c := range t.Columns {
				c.DefaultValue = fixVal(c.DefaultValue, c.Type, c.Size)
			}
		}
		for _, t := range s.EnumTables {
			for rowidx, row := range t.Values {
				for idx, f := range row {
					t.Values[rowidx][idx] = fixVal(f, t.Fields[idx].Type, 32)
				}
			}
		}
	}

	return
}

func fixVal(i interface{}, t ColumnType, size uint64) interface{} {
	if n, ok := i.(json.Number); ok {
		switch t {
		case ColTypeFloat:
			i, _ = n.Float64()
			return i
		case ColTypeAutoPrimaryKey:
			fallthrough
		case ColTypeInt:
			v, _ := n.Int64()
			if size < 64 {
				i = int(v)
			} else {
				i = v
			}
			return i
		case ColTypeUint:
			u, _ := n.Int64()
			if size < 64 {
				i = uint(u)
			} else {
				i = uint64(u)
			}

			return i
		}
	}
	return i

}
