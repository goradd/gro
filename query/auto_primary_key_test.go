package query

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAutoPk_Int(t *testing.T) {
	tests := []struct {
		name   string
		val    interface{}
		want   int
		panics bool
	}{
		{"string", "1", 1, false},
		{"int", 1, 1, false},
		{"int64", int64(1), 1, false},
		{"float64", float64(1), 1, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.panics {
				assert.Panics(t, func() {
					a := AutoPrimaryKey{
						val: tt.val,
					}
					a.Int()
				})
			} else {
				a := AutoPrimaryKey{
					val: tt.val,
				}
				assert.Equal(t, tt.want, a.Int())
			}
		})
	}
}
