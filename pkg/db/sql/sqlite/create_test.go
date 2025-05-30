package sqlite

import (
	"context"
	"github.com/goradd/orm/pkg/query"
	"github.com/goradd/orm/pkg/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestDB_CreateSampleSchema(t *testing.T) {
	d, err := NewDB("test", "") // memory only database will automatically be destroyed after test
	require.NoError(t, err)

	ctx := d.NewContext(context.Background())

	// prep
	s1 := sampleSchema()
	err = d.CreateSchema(ctx, s1)

	// insert, update, delete
	var userId string
	userId, err = d.Insert(ctx, "user", "id", map[string]interface{}{"name": "Bob"})
	assert.NotEmpty(t, userId)
	assert.NoError(t, err)

	var postId string
	postId, err = d.Insert(ctx, "post", "id", map[string]interface{}{"title": "This", "user_id": userId, "status_enum": 1})
	assert.NotEmpty(t, postId)
	require.NoError(t, err)

	_, err = d.Update(ctx, "post", "id", postId, map[string]any{"title": "That"}, "", 0)
	require.NoError(t, err)

	var cursor query.CursorI
	cursor, err = d.Query(ctx,
		"post",
		map[string]query.ReceiverType{"title": query.ColTypeString},
		map[string]any{"id": postId},
		nil)
	require.NoError(t, err)

	var data map[string]interface{}
	data, err = cursor.Next()
	assert.NoError(t, err)
	assert.Equal(t, "That", data["title"].(string))
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
			{Name: "rel_assn", Table1: "user", Name1: "user_id", Table2: "post", Name2: "post_id"},
		},
	}
	//db.fillDefault()
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
						DatabaseDefinition: map[string]map[string]interface{}{"postgres": map[string]interface{}{"collation": "en-US-x-icu"}},
					},
				},
			},
		},
	}
	//db.fillDefault()
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

	//db.fillDefault()
	return db
}

func sampleSchemaWithSchemaName() schema.Database {
	db := schema.Database{
		Key:             "test",
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
	//db.fillDefault()
	return db
}
