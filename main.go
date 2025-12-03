package main

import (
	"fmt"
	"log/slog"
	"os"

	cmdpkg "github.com/goradd/gro/internal/cmd"
	"github.com/spf13/cobra"

	_ "github.com/goradd/gro/internal/tmpl/template" // register templates through init calls
)

var (
	cfgPath    string
	schemaPath string
	key        string
	outputPath string

	rootCmd = &cobra.Command{
		Use:   "gro",
		Short: "Gro — code generator and database utilities",
	}
)

func main() {
	initGen()
	initGet()
	initPut()

	// Shared flags for all subcommands
	rootCmd.PersistentFlags().StringVarP(&schemaPath, "schema", "s", "", "Path to schema file (required)")

	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level:     slog.LevelInfo,
		AddSource: true,
	})
	logger := slog.New(handler)
	slog.SetDefault(logger)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func initGen() {
	genCmd := &cobra.Command{
		Use:   "gen",
		Short: "Generate ORM code from the schema",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Required for gen: config, schema
			if outputPath == "" {
				return fmt.Errorf("missing required flag: -o/--output")
			}
			if schemaPath == "" {
				return fmt.Errorf("missing required flag: -s/--schema")
			}
			return cmdpkg.Generate(schemaPath, outputPath)
		},
	}

	genCmd.Flags().StringVarP(&outputPath, "output", "o", "", "Output directory for generated code (required)")

	rootCmd.AddCommand(genCmd)
}

func initGet() {
	getCmd := &cobra.Command{
		Use:   "get",
		Short: "Build a schema file from a database (mysql and postgres only)",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Required for get: config, schema, key
			if cfgPath == "" {
				return fmt.Errorf("missing required flag: -c/--config")
			}
			if schemaPath == "" {
				return fmt.Errorf("missing required flag: -s/--schema")
			}
			if key == "" {
				return fmt.Errorf("missing required flag: -k/--key")
			}

			return cmdpkg.Extract(cfgPath, schemaPath, key)
		},
	}

	getCmd.Flags().StringVarP(&cfgPath, "config", "c", "", "Path to configuration file (required)")
	getCmd.Flags().StringVarP(&cfgPath, "key", "k", "", "Database key of database to get (required)")

	rootCmd.AddCommand(getCmd)
}

func initPut() {
	putCmd := &cobra.Command{
		Use:   "put",
		Short: "Insert or update data in the database",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Required for put: config, schema, key
			if cfgPath == "" {
				return fmt.Errorf("missing required flag: -c/--config")
			}
			if schemaPath == "" {
				return fmt.Errorf("missing required flag: -s/--schema")
			}
			if key == "" {
				return fmt.Errorf("missing required flag: -k/--key")
			}

			return cmdpkg.Rebuild(cfgPath, schemaPath, key)
		},
	}

	putCmd.Flags().StringVarP(&cfgPath, "config", "c", "", "Path to configuration file (required)")
	putCmd.Flags().StringVarP(&cfgPath, "key", "k", "", "Database key of database to put (required)")

	rootCmd.AddCommand(putCmd)
}
