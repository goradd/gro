package pgsql

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	schema2 "github.com/goradd/gro/internal/schema"
	"github.com/goradd/gro/query"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const postgresConnectionString = "host=127.0.0.1 port=5432 user=root password=12345 dbname=goradd_test sslmode=disable"

func TestDB_CreateSchema(t *testing.T) {
	sampleSchemas := []struct {
		name        string
		schema      func() schema2.Database // assume schema.Database is your top-level object
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

			ctx := context.Background()

			// prep
			s1 := tt.schema() // <== use dynamic schema generator
			d.DestroySchema(ctx, s1)

			err = d.CreateSchema(ctx, s1)

			defer d.DestroySchema(ctx, s1)

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

// TestDB_AutoGen tests the ability to reset the next value in an auto-generated sequence after manually
// entering a primary key that is auto generated. If we don't reset, then the nextval might be one of the
// manually entered values.
func TestDB_AutoGen(t *testing.T) {
	d, err := NewDB("test", postgresConnectionString, nil)
	require.NoError(t, err)

	ctx := context.Background()

	s1 := sampleSchemaWithSchemaName()
	d.DestroySchema(ctx, s1)
	err = d.CreateSchema(ctx, s1)
	require.NoError(t, err)
	defer d.DestroySchema(ctx, s1)

	// manual id
	fields := map[string]interface{}{
		"id":   1,
		"name": "Bob",
	}
	err = d.Insert(ctx, "test.user", fields, "id")
	assert.Equal(t, 1, fields["id"])
	assert.NoError(t, err)

	// auto id
	fields = map[string]interface{}{
		"name": "Sue",
	}
	err = d.Insert(ctx, "test.user", fields, "id")
	assert.NoError(t, err)
	assert.Equal(t, query.NewAutoPrimaryKey(int64(2)), fields["id"])
}

// zero out items that we will not be comparing
func zeroNonCmp(db *schema2.Database) {
	for _, t := range db.Tables {
		for _, c := range t.Columns {
			c.DatabaseDefinition = nil
		}
	}
}

func TestDB_CrudSampleSchema(t *testing.T) {
	d, err := NewDB("test", postgresConnectionString, nil)
	require.NoError(t, err)

	ctx := context.Background()

	// prep
	s1 := sampleSchema()
	err = d.CreateSchema(ctx, s1)

	// insert, update, delete
	fields := map[string]interface{}{"name": "Bob"}
	err = d.Insert(ctx, "user", fields, "id")
	assert.NotEmpty(t, fields["id"])
	assert.NoError(t, err)

	fields = map[string]interface{}{"title": "This", "user_id": fields["id"], "status_enum": 1}
	err = d.Insert(ctx, "post", fields, "id")
	require.NoError(t, err)

	err = d.Update(ctx, "post", map[string]any{"id": fields["id"]}, map[string]any{"title": "That"}, "", 0)
	require.NoError(t, err)

	var cursor query.CursorI
	cursor, err = d.Query(ctx,
		"post",
		map[string]query.ReceiverType{"title": query.ColTypeString},
		map[string]any{"id": fields["id"]},
		nil)
	require.NoError(t, err)

	var data map[string]interface{}
	data, err = cursor.Next()
	assert.NoError(t, err)
	assert.Equal(t, "That", data["title"].(string))
}

func sampleSchema() schema2.Database {
	db := schema2.Database{
		Key:             "test",
		EnumTableSuffix: "_enum",
		AssnTableSuffix: "_assn",

		Tables: []*schema2.Table{
			// User table
			{
				Name: "user",
				Columns: []*schema2.Column{
					{
						Name: "id",
						Type: schema2.ColTypeAutoPrimaryKey,
					},
					{
						Name:       "name",
						Type:       schema2.ColTypeString,
						Size:       100,
						IsNullable: false,
					},
				},
			},

			// Post table, references user
			{
				Name: "post",
				Columns: []*schema2.Column{
					{
						Name: "id",
						Type: schema2.ColTypeAutoPrimaryKey,
					},
					{
						Name: "title",
						Type: schema2.ColTypeString,
						Size: 200,
					},
					{
						Name:       "status_enum",
						Type:       schema2.ColTypeEnum,
						IsNullable: false,
						IndexLevel: schema2.IndexLevelIndexed, // foreign keys are always indexed
						EnumTable:  "post_status_enum",
					},
				},
				References: []*schema2.Reference{
					{
						Table: "user",
					},
				},
			},
		},
		EnumTables: []*schema2.EnumTable{
			// Enum table: post_status
			{
				Name: "post_status_enum",
				Values: []map[string]any{
					{"name": "Open"},
					{"name": "Closed"},
				},
			},
		},
		AssociationTables: []*schema2.AssociationTable{
			{
				Table: "user_post_assn",
				Ref1: schema2.AssociationReference{
					Table:  "user",
					Column: "user_id",
				},
				Ref2: schema2.AssociationReference{
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

func sampleSchemaWithCollation() schema2.Database {
	db := schema2.Database{
		Key:             "test",
		EnumTableSuffix: "_enum",
		AssnTableSuffix: "_assn",

		Tables: []*schema2.Table{
			// User table
			{
				Name: "user",
				Columns: []*schema2.Column{
					{
						Name: "id",
						Type: schema2.ColTypeAutoPrimaryKey,
					},
					{
						Name:               "name",
						Type:               schema2.ColTypeString,
						Size:               100,
						IsNullable:         false,
						DatabaseDefinition: map[string]map[string]interface{}{"postgres": {"collation": "en-US-x-icu"}},
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

func sampleSchemaTypes() schema2.Database {
	db := schema2.Database{
		Key:             "test",
		EnumTableSuffix: "_enum",
		AssnTableSuffix: "_assn",

		Tables: []*schema2.Table{
			{
				Name: "sample_types",
				Columns: []*schema2.Column{
					{
						Name: "id",
						Type: schema2.ColTypeAutoPrimaryKey,
					},
					{
						Name:       "username",
						Type:       schema2.ColTypeString,
						Size:       100,
						IsNullable: false,
					},
					{
						Name:       "age",
						Size:       32,
						Type:       schema2.ColTypeInt,
						IsNullable: false,
					},
					{
						Name:       "balance",
						Type:       schema2.ColTypeFloat,
						Size:       32,
						IsNullable: false,
					},
					{
						Name:       "is_active",
						Type:       schema2.ColTypeBool,
						IsNullable: false,
					},
					{
						Name:       "profile_picture",
						Type:       schema2.ColTypeBytes,
						IsNullable: true,
					},
					{
						Name:       "created_date",
						Type:       schema2.ColTypeTime,
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

func sampleSchemaWithSchemaName() schema2.Database {
	db := schema2.Database{
		Key:             "test",
		EnumTableSuffix: "_enum",
		AssnTableSuffix: "_assn",

		Tables: []*schema2.Table{
			// User table
			{
				Name:   "user",
				Schema: "test",
				Columns: []*schema2.Column{
					{
						Name: "id",
						Type: schema2.ColTypeAutoPrimaryKey,
					},
					{
						Name:       "name",
						Type:       schema2.ColTypeString,
						Size:       100,
						IsNullable: false,
					},
				},
			},
		},
	}
	//db.fillDefault()
	return db
}
