package test

import (
	"github.com/goradd/strings"
	"math/rand"
	"strconv"
	"time"
)

// RandomValue provides to the generated tests random values for types corresponding to ReceiverType Go types.
// For strings and []byte types, if size is 0, a size of 10 will be used as a reasonable limit.
func RandomValue[T any](size int) T {
	var v T
	var i any
	i = v
	switch i.(type) {
	case []byte:
		var b []byte
		if size == 0 {
			size = 10 // some reasonable value
		}
		for j := 0; j < size; j++ {
			b = append(b, uint8(rand.Intn(256)))
		}
		i = b
	case string:
		if size == 0 {
			size = 10
		}
		i = strings.RandomString(strings.AlphaAll, size)
	case int:
		var v int
		switch size {
		case 8:
			v = rand.Intn(256) - 128
		case 16:
			v = rand.Intn(0xffff) - 0x7fff
		case 32:
			v = int(rand.Uint32()) - 0x7fffffff
		case 64:
			v = int(rand.Int63() * int64(rand.Intn(2)*2-1))
		default:
			v = rand.Int()
		}
		i = v

	case uint:
		var v uint
		switch size {
		case 8:
			v = uint(rand.Intn(256))
		case 16:
			v = uint(rand.Intn(0xffff))
		case 32:
			v = uint(rand.Uint32())
		case 64:
			v = uint(rand.Uint64())
		default:
			if strconv.IntSize == 32 {
				v = uint(rand.Uint32())
			} else {
				v = uint(rand.Uint64())
			}
		}
		i = v
	case int64:
		i = rand.Int63() * int64(rand.Intn(2)*2-1)
	case uint64:
		i = rand.Uint64()
	case bool:
		i = rand.Intn(2) == 0
	case float64:
		i = rand.Float64()
	case float32:
		i = rand.Float32()
	case time.Time:
		i = time.Unix(int64(rand.Uint32()), 0)
	}
	return i.(T)
}
