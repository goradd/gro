package pgsql

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

const connectionString = "postgres://root:12345@127.0.0.1:5432/goradd_test?sslmode=disable"

func TestDB_DestroySchema(t *testing.T) {
	d, err := NewDB("test", connectionString, nil)
	require.NoError(t, err)

	ctx := d.NewContext(context.Background())
	err2 := d.DestroySchema(ctx)
	assert.NoError(t, err2)
}
