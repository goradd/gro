package codegen

import (
	"github.com/goradd/gofile/pkg/sys"
	"github.com/goradd/orm/pkg/model"
	"github.com/goradd/orm/pkg/schema"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func Generate(schemas []*schema.Database) {
	m := model.FromSchemas(schemas)
	for _, db := range m {
		gen(db)
	}
}

func gen(db *model.Database) {
	for _, tmpl := range templates {
		if g, ok := tmpl.(DatabaseGenerator); ok {
			genDatabaseTemplate(g, db)
		} else if g, ok := tmpl.(TableGenerator); ok {
			for _, tbl := range db.Tables {
				genTableTemplate(g, tbl)
			}
		} else if g, ok := tmpl.(EnumGenerator); ok {
			for _, tbl := range db.Enums {
				genEnumTemplate(g, tbl)
			}
		}
	}
}

func genDatabaseTemplate(g DatabaseGenerator, db *model.Database) {
	filename := g.FileName(db.Key)
	if !g.Overwrite() && fileExists(filename) {
		return
	}
	f, err := openFile(filename)
	if err != nil {
		log.Print(err)
		return
	}
	defer f.Close()

	if err = g.GenerateDatabase(db, f); err != nil {
		log.Print(err)
		return
	}
	runGoImports(filename)
}

func genTableTemplate(g TableGenerator, table *model.Table) {
	filename := g.FileName(table)
	if !g.Overwrite() && fileExists(filename) {
		return
	}

	f, err := openFile(filename)
	if err != nil {
		log.Print(err)
		return
	}
	defer f.Close()
	var importPath string
	importPath, err = sys.ImportPath(filename)

	if err = g.GenerateTable(table, f, importPath); err != nil {
		log.Print(err)
		return
	}
	runGoImports(filename)
}

func genEnumTemplate(g EnumGenerator, table *model.Enum) {
	filename := g.FileName(table)
	if !g.Overwrite() && fileExists(filename) {
		return
	}
	if g.Overwrite() && len(table.Values) == 0 {
		// an error occurred with the schema
		deleteFile(filename)
		return
	}

	f, err := openFile(filename)
	if err != nil {
		log.Print(err)
		return
	}
	defer f.Close()
	var importPath string
	importPath, err = sys.ImportPath(filename)

	if err = g.GenerateEnum(table, f, importPath); err != nil {
		log.Print(err)
		return
	}
	runGoImports(filename)
}

func openFile(filename string) (f *os.File, err error) {
	filename, _ = filepath.Abs(filename)
	fp := filepath.Dir(filename)
	if err = os.MkdirAll(fp, 0777); err != nil {
		return
	}
	f, err = os.Create(filename)
	return
}

func fileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return false
	}
	return err == nil
}

func deleteFile(filename string) bool {
	filename, _ = filepath.Abs(filename)
	err := os.Remove(filename)
	return err == nil
}

func runGoImports(fileName string) {
	if strings.HasSuffix(fileName, ".go") {
		curDir, err := os.Getwd()
		if err != nil {
			log.Print(err)
			return
		}
		_ = os.Chdir(filepath.Dir(fileName)) // run it from the file's directory to pick up the correct go.mod file if there is one
		_, err = sys.ExecuteShellCommand("goimports -w " + filepath.Base(fileName))
		_ = os.Chdir(curDir)
		if err != nil {
			if e, ok := err.(*exec.Error); ok {
				panic("error running goimports: " + e.Error()) // perhaps goimports is not installed?
			} else if e2, ok2 := err.(*exec.ExitError); ok2 {
				// Likely a syntax error in the resulting file
				log.Print(string(e2.Stderr))
			}
		}
	}
}
