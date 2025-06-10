package mysql

import (
	"context"
	"fmt"
	"github.com/google/go-cmp/cmp"
	"github.com/goradd/orm/pkg/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"
)

const mysqlConnectionString = "root:12345@tcp(127.0.0.1:3306)/goradd_test?parseTime=true&charset=utf8mb4&loc=Local"

func TestDB_CreateSchema(t *testing.T) {
	sampleSchemas := []struct {
		name        string
		schema      func() schema.Database // assume schema.Database is your top-level object
		zeroNonComp bool
	}{
		{
			name:        "SimpleSchema",
			schema:      sampleSchema, // your original function
			zeroNonComp: true,
		},
		{
			name:        "SchemaWithCollation",
			schema:      sampleSchemaWithCollation,
			zeroNonComp: false,
		},
		{
			name:        "SchemaTypes",
			schema:      sampleSchemaTypes,
			zeroNonComp: true,
		},
	}

	for _, tt := range sampleSchemas {
		t.Run(tt.name, func(t *testing.T) {
			d, err := NewDB("test", mysqlConnectionString, nil)
			require.NoError(t, err)

			ctx := context.Background()

			// prep
			s1 := tt.schema()
			d.DestroySchema(ctx, s1)

			err = d.CreateSchema(ctx, s1)

			defer func() {
				d.DestroySchema(ctx, s1)
			}()

			assert.NoError(t, err)

			options := map[string]any{
				"reference_suffix":  "_id",
				"enum_table_suffix": "_enum",
				"assn_table_suffix": "_assn",
			}

			s2 := d.ExtractSchema(options)
			err = s2.Clean()
			require.NoError(t, err)
			if tt.zeroNonComp {
				zeroNonCmp(&s1)
				zeroNonCmp(&s2)
			}

			v := reflect.DeepEqual(s1, s2)
			assert.True(t, v)

			if !v {
				fmt.Println("Mismatch:", cmp.Diff(s1, s2))
			}
		})
	}
}

// zero out items that we will not be comparing
func zeroNonCmp(db *schema.Database) {
	for _, t := range db.Tables {
		for _, c := range t.Columns {
			c.DatabaseDefinition = nil
		}
	}
}

func sampleSchema() schema.Database {
	db := schema.Database{
		Key:             "test",
		EnumTableSuffix: "_enum",
		AssnTableSuffix: "_assn",

		Tables: []*schema.Table{
			// User table
			{
				Name: "user",
				Columns: []*schema.Column{
					{
						Name: "id",
						Type: schema.ColTypeAutoPrimaryKey,
					},
					{
						Name:       "name",
						Type:       schema.ColTypeString,
						Size:       100,
						IsNullable: false,
					},
				},
			},

			// Post table, references user
			{
				Name: "post",
				Columns: []*schema.Column{
					{
						Name: "id",
						Type: schema.ColTypeAutoPrimaryKey,
					},
					{
						Name: "title",
						Type: schema.ColTypeString,
						Size: 200,
					},
					{
						Name:       "status_enum",
						Type:       schema.ColTypeEnum,
						IsNullable: false,
						IndexLevel: schema.IndexLevelIndexed, // foreign keys are always indexed
						EnumTable:  "post_status_enum",
					},
				},
				References: []*schema.Reference{
					{
						Table: "user",
					},
				},
			},
		},
		EnumTables: []*schema.EnumTable{
			// Enum table: post_status
			{
				Name: "post_status_enum",
				Fields: map[string]schema.EnumField{
					"const":      {Type: schema.ColTypeInt},
					"label":      {Type: schema.ColTypeString},
					"identifier": {Type: schema.ColTypeString},
				},
				Values: []map[string]any{
					{"const": 1, "label": "Open", "identifier": "open"},
					{"const": 2, "label": "Closed", "identifier": "closed"},
				},
			},
		},
		AssociationTables: []*schema.AssociationTable{
			{
				Table: "uer_post_assn",
				Ref1: schema.AssociationReference{
					Table:  "user",
					Column: "user_id",
				},
				Ref2: schema.AssociationReference{
					Table:  "post",
					Column: "post_id",
				},
			},
		},
	}

	if err := db.Clean(); err != nil {
		panic(err)
	}
	return db
}

func sampleSchemaWithCollation() schema.Database {
	db := schema.Database{
		Key:             "test",
		EnumTableSuffix: "_enum",
		AssnTableSuffix: "_assn",

		Tables: []*schema.Table{
			// User table
			{
				Name: "user",
				Columns: []*schema.Column{
					{
						Name: "id",
						Type: schema.ColTypeAutoPrimaryKey,
					},
					{
						Name:               "name",
						Type:               schema.ColTypeString,
						Size:               100,
						IsNullable:         false,
						DatabaseDefinition: map[string]map[string]interface{}{"mysql": {"collation": "utf8mb4_bin"}},
					},
				},
			},
		},
	}
	if err := db.Clean(); err != nil {
		panic(err)
	}
	return db
}

func sampleSchemaTypes() schema.Database {
	db := schema.Database{
		Key:             "test",
		EnumTableSuffix: "_enum",
		AssnTableSuffix: "_assn",

		Tables: []*schema.Table{
			{
				Name: "sample_types",
				Columns: []*schema.Column{
					{
						Name: "id",
						Type: schema.ColTypeAutoPrimaryKey,
					},
					{
						Name:       "username",
						Type:       schema.ColTypeString,
						Size:       100,
						IsNullable: false,
					},
					{
						Name:       "age",
						Size:       32,
						Type:       schema.ColTypeInt,
						IsNullable: false,
					},
					{
						Name:       "balance",
						Type:       schema.ColTypeFloat,
						Size:       32,
						IsNullable: false,
					},
					{
						Name:       "is_active",
						Type:       schema.ColTypeBool,
						IsNullable: false,
					},
					{
						Name:       "profile_picture",
						Type:       schema.ColTypeBytes,
						Size:       1024,
						IsNullable: true,
					},
					{
						Name:       "created_date",
						Type:       schema.ColTypeTime,
						IsNullable: true,
					},
				},
			},
		},
	}

	if err := db.Clean(); err != nil {
		panic(err)
	}
	return db
}
