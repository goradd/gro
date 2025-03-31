package test

import (
	"fmt"
	"github.com/goradd/maps"
	"github.com/goradd/strings"
	"golang.org/x/exp/constraints"
	"strconv"
	"time"
)

func randomString(source string, n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = source[rng.Intn(len(source))]
	}
	return string(b)
}

// RandomValue provides to the generated tests random values for types corresponding to ReceiverType Go types.
// For strings and []byte types, if size is 0, a size of 10 will be used as a reasonable limit.
// size will indicate the number of bytes or characters generated.
// times do not generate fractional seconds, since the value might be truncated depending on the sql dialect and data type.
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
			b = append(b, uint8(rng.Intn(256)))
		}
		i = b
	case string:
		if size == 0 {
			size = 10
		}
		i = randomString(strings.AlphaNumeric, size)

	case int:
		var v int
		switch size {
		case 8:
			v = rng.Intn(256) - 128
		case 16:
			v = rng.Intn(0xffff) - 0x7fff
		case 32:
			v = int(rng.Uint32()) - 0x7fffffff
		case 64:
			v = int(rng.Int63() * int64(rng.Intn(2)*2-1))
		default:
			v = rng.Int()
		}
		i = v

	case uint:
		var v uint
		switch size {
		case 8:
			v = uint(rng.Intn(256))
		case 16:
			v = uint(rng.Intn(0xffff))
		case 32:
			v = uint(rng.Uint32())
		case 64:
			v = uint(rng.Uint64())
		default:
			if strconv.IntSize == 32 {
				v = uint(rng.Uint32())
			} else {
				v = uint(rng.Uint64())
			}
		}
		i = v
	case int64:
		i = rng.Int63() * int64(rng.Intn(2)*2-1)
	case uint64:
		i = rng.Uint64()
	case bool:
		i = rng.Intn(2) == 0
	case float64:
		i = rng.Float64()
	case float32:
		i = rng.Float32()
	case time.Time:
		i = time.Unix(int64(rng.Uint32()), 0).UTC()
	}
	return i.(T)
}

// RandomNum provides a random number in the given range.
func RandomNum[T constraints.Integer | constraints.Float](low int, high int) T {
	v := rng.Intn(high - low)
	return T(v + low)
}

func RandomNumberString() string {
	v := rng.Intn(10000) + 1
	return fmt.Sprint(v)
}

func RandomEnum[T ~int](valueList []T) T {
	v := rng.Intn(len(valueList))
	return valueList[v]
}

func RandomEnumArray[T ~int](valueList []T) *maps.OrderedSet[T] {
	values := maps.NewOrderedSet[T]()
	values.Add(valueList[0]) // at least 1 item
	for _, v := range valueList[1:] {
		if rng.Intn(2) == 0 {
			values.Add(T(v))
		}
	}
	return values
}

func RandomDecimal(precision int, scale int) string {
	l := precision - scale
	pc := rng.Intn(l + 1)
	sc := rng.Intn(scale + 1)
	var s string
	if pc > 0 {
		s = randomString(strings.Numbers, pc)
	}
	if sc > 0 {
		s += "." + randomString(strings.Numbers, sc)
	}
	if s == "" || s == "." {
		return "0"
	}
	s = randomString("+-", 1) + s
	return s
}
