package schema

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
)

func WriteJsonFile(schema *Database, outFile string) error {
	// Serialize the struct to JSON with indentation for readability
	jsonData, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling JSON: %w", err)
	}

	// Create or open a file for writing
	file, err2 := os.Create(outFile)
	if err2 != nil {
		return fmt.Errorf("error creating file %s: %w", outFile, err)
	}
	defer file.Close()

	// Write the JSON data to the file
	_, err = file.Write(jsonData)
	if err != nil {
		return fmt.Errorf("error writing to file %s: %w", outFile, err)
	}
	return nil
}

func ReadJsonFile(infile string) (schema *Database, err error) {
	// Read the file's content
	var data []byte
	data, err = os.ReadFile(infile)
	if err != nil {
		return
	}

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()

	if err = decoder.Decode(&schema); err != nil {
		return
	}

	// Fix up interfaces
	for _, t := range schema.Tables {
		for _, c := range t.Columns {
			c.DefaultValue = fixVal(c.DefaultValue, c.Type, c.Size)
		}
	}
	for _, t := range schema.EnumTables {
		for rowidx, row := range t.Values {
			for idx, f := range row {
				t.Values[rowidx][idx] = fixVal(f, t.Fields[idx].Type, 32)
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
		}
	}
	return i

}
