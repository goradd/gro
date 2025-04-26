package sql

import (
	"encoding/json"
	"github.com/goradd/orm/pkg/db"
	"github.com/goradd/orm/pkg/schema"
	"log/slog"
	"strings"
)

// tableCommentItems holds fields that will be included in the table comment.
type tableCommentItems struct {
	Label            string `json:"label,omitempty"`
	LabelPlural      string `json:"label_plural,omitempty"`
	Identifier       string `json:"id,omitempty"`
	IdentifierPlural string `json:"id_plural,omitempty"`
	Key              string `json:"key"`
	NoOrm            bool   `json:"no_orm,omitempty"`
}

// TableComment returns extra schema fields to be stored as a JSON object in the table comment.
func TableComment(t *schema.Table) string {
	ti := tableCommentItems{
		Label:            t.Label,
		LabelPlural:      t.LabelPlural,
		Identifier:       t.Identifier,
		IdentifierPlural: t.IdentifierPlural,
		Key:              t.Key,
		NoOrm:            t.NoOrm,
	}

	data, _ := json.Marshal(ti)
	return string(data)
}

// FillTableCommentFields extracts and parses a JSON object from the comment string
// and fills the relevant fields in the schema.Table.
func FillTableCommentFields(t *schema.Table, comment string) {
	start := strings.Index(comment, "{")
	end := strings.LastIndex(comment, "}")
	if start == -1 || end == -1 || end < start {
		// No valid JSON object found
		return
	}

	jsonPart := comment[start : end+1]

	var ti tableCommentItems
	if err := json.Unmarshal([]byte(jsonPart), &ti); err != nil {
		// JSON is malformed or doesn't match expected fields
		return
	}

	// Fill in the fields
	t.Label = ti.Label
	t.LabelPlural = ti.LabelPlural
	t.Identifier = ti.Identifier
	t.IdentifierPlural = ti.IdentifierPlural
	t.Key = ti.Key
	t.NoOrm = ti.NoOrm
}

// columnCommentItems holds fields that will be included in the column comments.
type columnCommentItems struct {
	Label                   string `json:"label,omitempty"`
	Identifier              string `json:"id,omitempty"`
	Key                     string `json:"key"`
	ReferenceIdentifier     string `json:"ref_id,omitempty"`
	ReferenceLabel          string `json:"ref_label,omitempty"`
	ReverseLabel            string `json:"rev_label,omitempty"`
	ReverseLabelPlural      string `json:"rev_label_plural,omitempty"`
	ReverseIdentifier       string `json:"rev_id,omitempty"`
	ReverseIdentifierPlural string `json:"rev_id_plural,omitempty"`
	EnumTable               string `json:"enum_table,omitempty"`
	Type                    string `json:"type,omitempty"` // override calculated type, for special situations
}

// ColumnComment returns extra schema fields to be stored as a JSON object in the column comment.
func ColumnComment(c *schema.Column) string {
	ti := columnCommentItems{
		Label:      c.Label,
		Identifier: c.Identifier,
		Key:        c.Key,
	}
	if c.Reference != nil {
		ti.ReferenceIdentifier = c.Reference.Identifier
		ti.ReferenceLabel = c.Reference.Label
		ti.ReverseLabel = c.Reference.ReverseLabel
		ti.ReverseLabelPlural = c.Reference.ReverseLabelPlural
		ti.ReverseIdentifier = c.Reference.ReverseIdentifier
		ti.ReverseIdentifierPlural = c.Reference.ReverseIdentifierPlural
		if c.Type == schema.ColTypeEnumArray {
			ti.EnumTable = c.Reference.Table
		}
	}

	if c.Type == schema.ColTypeJSON {
		ti.Type = "JSON" // for MariadB, since it changes JSON into LONGTEXT
	}

	data, _ := json.Marshal(ti)
	return string(data)
}

// FillColumnCommentFields extracts and decodes the JSON metadata in a column comment.
func FillColumnCommentFields(c *schema.Column, comment string) {
	start := strings.Index(comment, "{")
	end := strings.LastIndex(comment, "}")
	if start == -1 || end == -1 || end < start {
		// No valid JSON object found
		return
	}

	jsonPart := comment[start : end+1]

	var ci columnCommentItems
	if err := json.Unmarshal([]byte(jsonPart), &ci); err != nil {
		// Invalid or unrecognized JSON, ignore
		return
	}

	// Fill basic fields
	c.Label = ci.Label
	c.Identifier = ci.Identifier
	c.Key = ci.Key

	if c.Reference != nil {
		c.Reference.Identifier = ci.ReferenceIdentifier
		c.Reference.Label = ci.ReferenceLabel
		c.Reference.ReverseLabel = ci.ReverseLabel
		c.Reference.ReverseLabelPlural = ci.ReverseLabelPlural
		c.Reference.ReverseIdentifier = ci.ReverseIdentifier
		c.Reference.ReverseIdentifierPlural = ci.ReverseIdentifierPlural
	}

	if ci.EnumTable != "" {
		if c.Reference != nil {
			slog.Error("Reference has an enum table settting.",
				slog.String(db.LogColumn, c.Name))
		} else {
			// An enum array. Add a reference to the enum table since there is no implicit way to do it.
			c.Reference = &schema.Reference{
				Table: ci.EnumTable,
			}
		}
	}

	if ci.Type != "" {
		if ci.Type == "JSON" {
			c.Type = schema.ColTypeJSON
			c.Size = 0
		}
	}
}

// enumTableCommentItems holds fields that will be included in the enum table comment.
type enumTableCommentItems struct {
	Label            string `json:"label,omitempty"`
	LabelPlural      string `json:"label_plural,omitempty"`
	Identifier       string `json:"id,omitempty"`
	IdentifierPlural string `json:"id_plural,omitempty"`
	Key              string `json:"key"`
}

// EnumTableComment returns extra schema fields to be stored as a JSON object in the table comment.
func EnumTableComment(t *schema.EnumTable) string {
	ti := enumTableCommentItems{
		Label:            t.Label,
		LabelPlural:      t.LabelPlural,
		Identifier:       t.Identifier,
		IdentifierPlural: t.IdentifierPlural,
		Key:              t.Key,
	}

	data, _ := json.Marshal(ti)
	return string(data)
}

// FillEnumCommentFields extracts and fills fields from a JSON object embedded in a comment.
func FillEnumCommentFields(t *schema.EnumTable, comment string) {
	start := strings.Index(comment, "{")
	end := strings.LastIndex(comment, "}")
	if start == -1 || end == -1 || end < start {
		// No valid JSON found
		return
	}

	jsonPart := comment[start : end+1]

	var ei enumTableCommentItems
	if err := json.Unmarshal([]byte(jsonPart), &ei); err != nil {
		// Malformed or incompatible JSON
		return
	}

	t.Label = ei.Label
	t.LabelPlural = ei.LabelPlural
	t.Identifier = ei.Identifier
	t.IdentifierPlural = ei.IdentifierPlural
	t.Key = ei.Key
}

// enumFieldCommentItems holds fields that will be included in the enum field column comment.
type enumFieldCommentItems struct {
	Identifier       string `json:"id,omitempty"`
	IdentifierPlural string `json:"id_plural,omitempty"`
}

// EnumFieldComment returns extra schema fields to be stored as a JSON object in the field column comment.
func EnumFieldComment(t schema.EnumField) string {
	ti := enumFieldCommentItems{
		Identifier:       t.Identifier,
		IdentifierPlural: t.IdentifierPlural,
	}

	data, _ := json.Marshal(ti)
	return string(data)
}

// FillEnumFieldCommentFields extracts and fills fields from a JSON object embedded in a comment.
func FillEnumFieldCommentFields(t *schema.EnumField, comment string) {
	start := strings.Index(comment, "{")
	end := strings.LastIndex(comment, "}")
	if start == -1 || end == -1 || end < start {
		// No valid JSON found
		return
	}

	jsonPart := comment[start : end+1]

	var ei enumFieldCommentItems
	if err := json.Unmarshal([]byte(jsonPart), &ei); err != nil {
		// Malformed or incompatible JSON
		return
	}

	if ei.Identifier != "" {
		t.Identifier = ei.Identifier
	}
	if ei.IdentifierPlural != "" {
		t.IdentifierPlural = ei.IdentifierPlural
	}
}

// associationTableCommentItems holds fields that will be included in the enum table comment.
type associationTableCommentItems struct {
	Label1            string `json:"label1,omitempty"`
	Label1Plural      string `json:"label1_plural,omitempty"`
	Identifier1       string `json:"id1,omitempty"`
	Identifier1Plural string `json:"id1_plural,omitempty"`
	Label2            string `json:"label2,omitempty"`
	Label2Plural      string `json:"label2_plural,omitempty"`
	Identifier2       string `json:"id2,omitempty"`
	Identifier2Plural string `json:"id2_plural,omitempty"`
	Key               string `json:"key"`
}

// AssociationTableComment returns extra schema fields to be stored as a JSON object in the table comment.
func AssociationTableComment(t *schema.AssociationTable) string {
	ti := associationTableCommentItems{
		Label1:            t.Label1,
		Label1Plural:      t.Label1Plural,
		Identifier1:       t.Identifier1,
		Identifier1Plural: t.Identifier1Plural,
		Label2:            t.Label2,
		Label2Plural:      t.Label2Plural,
		Identifier2:       t.Identifier2,
		Identifier2Plural: t.Identifier2Plural,
		Key:               t.Key,
	}

	data, _ := json.Marshal(ti)
	return string(data)
}

// FillAssociationCommentFields extracts and sets fields on the AssociationTable from JSON in the comment.
func FillAssociationCommentFields(t *schema.AssociationTable, comment string) {
	start := strings.Index(comment, "{")
	end := strings.LastIndex(comment, "}")
	if start == -1 || end == -1 || end < start {
		// No valid JSON object found
		return
	}

	jsonPart := comment[start : end+1]

	var ai associationTableCommentItems
	if err := json.Unmarshal([]byte(jsonPart), &ai); err != nil {
		// Ignore malformed JSON
		return
	}

	t.Label1 = ai.Label1
	t.Label1Plural = ai.Label1Plural
	t.Identifier1 = ai.Identifier1
	t.Identifier1Plural = ai.Identifier1Plural
	t.Label2 = ai.Label2
	t.Label2Plural = ai.Label2Plural
	t.Identifier2 = ai.Identifier2
	t.Identifier2Plural = ai.Identifier2Plural
	t.Key = ai.Key
}
