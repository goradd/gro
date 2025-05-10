package codegen

import (
	"errors"
	"github.com/goradd/gofile/pkg/sys"
	db2 "github.com/goradd/orm/pkg/db"
	"github.com/goradd/orm/pkg/model"
	"github.com/goradd/orm/pkg/schema"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func Generate(schemaDB *schema.Database) {
	m := model.FromSchema(schemaDB)
	gen(m)
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
		slog.Error("Error opening file",
			slog.String(db2.LogFilename, filename),
			slog.Any(db2.LogError, err))
		return
	}
	defer f.Close()
	var importPath string
	importPath, err = sys.ImportPath(filename)

	if err = g.GenerateDatabase(db, f, importPath); err != nil {
		slog.Error("Error generating database template file",
			slog.String(db2.LogFilename, filename),
			slog.String(db2.LogComponent, "codegen"),
			slog.Any(db2.LogError, err))
		return
	}
	runGoImports(filename)
}

func genTableTemplate(g TableGenerator, table *model.Table) {
	filename := g.FileName(table)
	if filename == "" {
		return
	}
	if !g.Overwrite() && fileExists(filename) {
		return
	}

	f, err := openFile(filename)
	if err != nil {
		slog.Error("Error opening file",
			slog.String(db2.LogFilename, filename),
			slog.Any(db2.LogError, err))
		return
	}
	defer f.Close()
	var importPath string
	importPath, err = sys.ImportPath(filename)

	if err = g.GenerateTable(table, f, importPath); err != nil {
		slog.Error("Error generating table template file",
			slog.String(db2.LogFilename, filename),
			slog.String(db2.LogComponent, "codegen"),
			slog.Any(db2.LogError, err))
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
		slog.Error("The enum type has no values in the schema",
			slog.String(db2.LogFilename, filename),
			slog.String(db2.LogComponent, "codegen"),
			slog.Any(db2.LogTable, table.QueryName))
		deleteFile(filename)
		return
	}

	f, err := openFile(filename)
	if err != nil {
		slog.Error("Error opening file",
			slog.String(db2.LogFilename, filename),
			slog.Any(db2.LogError, err))
		return
	}
	defer f.Close()
	var importPath string
	importPath, err = sys.ImportPath(filename)

	if err = g.GenerateEnum(table, f, importPath); err != nil {
		slog.Error("Error generating enum template file",
			slog.String(db2.LogFilename, filename),
			slog.String(db2.LogComponent, "codegen"),
			slog.Any(db2.LogError, err))
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
			slog.Error("Error getting current directory for running goimports",
				slog.Any(db2.LogError, err))
			return
		}
		_ = os.Chdir(filepath.Dir(fileName)) // run it from the file's directory to pick up the correct go.mod file if there is one
		_, err = sys.ExecuteShellCommand("goimports -w " + filepath.Base(fileName))
		_ = os.Chdir(curDir)
		if err != nil {
			var e *exec.Error
			if errors.As(err, &e) {
				slog.Error("Error running goimports",
					slog.Any(db2.LogError, err))
			}
		}
	}
}
