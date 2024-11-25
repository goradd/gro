package codegen

import (
	"github.com/goradd/gofile/pkg/sys"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"spekary/goradd/orm/pkg/model"
	"spekary/goradd/orm/pkg/schema"
	"strings"
)

func Generate(schemas []*schema.Database) {
	m := model.FromSchemas(schemas)
	for _, db := range m {
		genDB(db)
	}
}

func genDB(db *model.Database) {
	for _, t := range templates {
		filename := t.FileName(db.Key)
		if !t.Overwrite() && fileExists(filename) {
			continue
		}
		filename, _ = filepath.Abs(filename)
		fp := filepath.Dir(filename)
		if err := os.MkdirAll(fp, 0777); err != nil {
			log.Print(err)
			continue
		}
		f, err := os.Create(filename)
		if err != nil {
			log.Print(err)
			continue
		}

		if d, ok := t.(DatabaseGenerator); ok {
			if err := d.GenerateDatabase(db, f); err != nil {
				log.Print(err)
			}
		}
		err = f.Close()
		if err != nil {
			log.Print(err)
			continue
		}
		runGoImports(filename)
	}
}

func fileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return false
	}
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
