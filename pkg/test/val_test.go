package test

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestRandomValue(t *testing.T) {
	seed := NewSeed()
	fmt.Println("Seed=", seed)

	i := RandomValue[int](0)
	fmt.Println(i)

	u := RandomValue[uint](0)
	fmt.Println(u)

	i64 := RandomValue[int64](0)
	fmt.Println(i64)

	u64 := RandomValue[uint64](0)
	fmt.Println(u64)

	j := RandomValue[[]byte](0)
	fmt.Println(j)

	t2 := RandomValue[time.Time](0)
	fmt.Println(t2)

	s := RandomValue[string](0)
	fmt.Println(s)

	f32 := RandomValue[float32](0)
	fmt.Println(f32)

	f64 := RandomValue[float64](0)
	fmt.Println(f64)

	i8 := RandomValue[int](8)
	fmt.Println(i8)
	assert.Less(t, i8, 128)

	u8 := RandomValue[uint](8)
	fmt.Println(u8)
	assert.Less(t, u8, uint(256))

	i16 := RandomValue[int](16)
	fmt.Println(i16)
	assert.Less(t, i16, 0x7fff)

	u16 := RandomValue[uint](16)
	fmt.Println(u16)
	assert.Less(t, u16, uint(0xffff)+1)

	i32 := RandomValue[int](32)
	fmt.Println(i32)
	assert.Less(t, i32, 0x7fffffff)

	u32 := RandomValue[uint](32)
	fmt.Println(u32)
	assert.Less(t, u32, uint(0xffffffff)+1)

	// verify reusing a seed will produce same value
	UseSeed(seed)

	i2 := RandomValue[int](0)
	assert.Equal(t, i2, i)

}
