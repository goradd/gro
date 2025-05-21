package pgsql

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

const postgresConnectionString = "host=127.0.0.1 port=5432 user=root password=12345 dbname=goradd_test sslmode=disable"

func sampleSchema() schema.Database {
	db := schema.Database{
		Key:             "test",
		ReferenceSuffix: "_id",
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
						Name:       "user_id",
						Type:       schema.ColTypeReference,
						IsNullable: false,
						IndexLevel: schema.IndexLevelIndexed, // foreign keys should always be indexed
						Reference: &schema.Reference{
							Table: "user",
						},
					},
					{
						Name:       "status_enum",
						Type:       schema.ColTypeEnum,
						IsNullable: false,
						IndexLevel: schema.IndexLevelIndexed,
						Reference: &schema.Reference{
							Table: "post_status_enum",
						},
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
			{Name: "rel_assn", Table1: "user", Column1: "user_id", Table2: "post", Column2: "post_id"},
		},
	}
	//db.FillDefaults()
	return db
}

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
			d, err := NewDB("test", postgresConnectionString, nil)
			require.NoError(t, err)

			ctx := d.NewContext(context.Background())

			// prep
			s1 := tt.schema() // <== use dynamic schema generator

			_ = d.DestroySchema(ctx, s1)
			err = d.CreateSchema(ctx, s1)

			defer func() {
				err := d.DestroySchema(ctx, s1)
				assert.NoError(t, err)
			}()

			assert.NoError(t, err)

			options := map[string]any{
				"reference_suffix":  "_id",
				"enum_table_suffix": "_enum",
				"assn_table_suffix": "_assn",
			}

			s2 := d.ExtractSchema(options)
			s2.Clean()
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

// TestDB_AutoGen tests the ability to reset the next value in an auto-generated sequence after manually
// entering a primary key that is auto generated. If we don't reset, then the nextval might be one of the
// manually entered values.
func TestDB_AutoGen(t *testing.T) {
	d, err := NewDB("test", postgresConnectionString, nil)
	require.NoError(t, err)

	ctx := d.NewContext(context.Background())

	s1 := sampleSchemaWithSchemaName()
	_ = d.DestroySchema(ctx, s1)
	err = d.CreateSchema(ctx, s1)
	require.NoError(t, err)
	defer func() {
		err := d.DestroySchema(ctx, s1)
		assert.NoError(t, err)
	}()

	// manual id
	id, err := d.Insert(ctx, "test.user", "id", map[string]interface{}{
		"id":   1,
		"name": "Bob",
	})
	assert.Equal(t, "1", id)
	assert.NoError(t, err)

	// auto id
	id, err = d.Insert(ctx, "test.user", "id", map[string]interface{}{
		"name": "Sue",
	})
	assert.NoError(t, err)
	assert.Equal(t, "2", id)
}

// zero out items that we will not be comparing
func zeroNonCmp(db *schema.Database) {
	for _, t := range db.Tables {
		for _, c := range t.Columns {
			c.DatabaseDefinition = nil
		}
	}
}

func sampleSchemaWithCollation() schema.Database {
	db := schema.Database{
		Key:             "test",
		ReferenceSuffix: "_id",
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
						DatabaseDefinition: map[string]map[string]interface{}{"postgres": map[string]interface{}{"collation": "en-US-x-icu"}},
					},
				},
			},
		},
	}
	//db.FillDefaults()
	return db
}

func sampleSchemaTypes() schema.Database {
	db := schema.Database{
		Key:             "test",
		ReferenceSuffix: "_id",
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

	//db.FillDefaults()
	return db
}

func sampleSchemaWithSchemaName() schema.Database {
	db := schema.Database{
		Key:             "test",
		ReferenceSuffix: "_id",
		EnumTableSuffix: "_enum",
		AssnTableSuffix: "_assn",

		Tables: []*schema.Table{
			// User table
			{
				Name:   "user",
				Schema: "test",
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
		},
	}
	//db.FillDefaults()
	return db
}
