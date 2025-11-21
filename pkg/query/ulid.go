package query

import (
	"bytes"
	"fmt"
	"time"

	"golang.org/x/exp/rand"
)

type ULID [16]byte

// RULID returns a randomized ULID that has no timestamp information, suitable for
// use as an exposed ULID.
func RULID() ULID {
	var b ULID
	_, err := rand.Read(b[:])
	if err != nil {
		panic(err) // the system RNG generator is down.
	}
	return b
}

// NewULID generates a ULID using standard ULID rules of 48-bit timestamp
// plus 80-bits of randomness
func NewULID() ULID {
	var u ULID
	tm := time.Now().UnixMilli()
	// quick copy lower bytes into upper range of UUID
	for i := 5; i >= 0; i-- {
		var b byte
		b = byte(tm & 0xff)
		tm >>= 8
		u[i] = b
	}
	_, err := rand.Read(u[6:])
	if err != nil {
		panic(err) // random number generator is offline. Very unusual.
	}
	return u
}

var crockfordAlphabet = []byte("0123456789ABCDEFGHJKMNPQRSTVWXYZ")

func encodeCrockford(src [16]byte) string {
	dst := make([]byte, 26) // 128 bits => 26 chars
	var bitBuffer uint
	var bitCount uint

	for i, j := 0, 0; i < len(src) && j < len(dst); {
		bitBuffer = (bitBuffer << 8) | uint(src[i])
		bitCount += 8
		if bitCount >= 5 {
			bitCount -= 5
			dst[j] = crockfordAlphabet[(bitBuffer>>bitCount)&31]
			j++
		}
		if bitCount < 5 && i == len(src)-1 && j < len(dst) {
			dst[j] = crockfordAlphabet[(bitBuffer<<(5-bitCount))&31]
			j++
		}
		i++
	}
	return string(dst)
}

func (u ULID) String() string {
	s := encodeCrockford(u)
	return s
}

func ULIDFromString(src string) (u ULID, err error) {
	u, err = parseCrockford16(src)
	return
}

// ULIDFromBytes returns a ULID copied from the upper 16 bytes of the given
// byte slice. Will panic if the byte slice is too small.
func ULIDFromBytes(src []byte) (u ULID, err error) {
	if len(src) < 16 {
		err = fmt.Errorf("byte slice too short")
		return
	}
	for i := 0; i < 16; i++ {
		u[i] = src[i]
	}
	return
}

func (u ULID) MarshalBinary() ([]byte, error) {
	return u[:], nil
}

func (u *ULID) UnmarshalBinary(b []byte) error {
	var err error
	*u, err = ULIDFromBytes(b)
	return err
}

func (u ULID) MarshalText() ([]byte, error) {
	return []byte(u.String()), nil
}

func (u *ULID) UnmarshalText(b []byte) error {
	var err error
	*u, err = ULIDFromString(string(b))
	return err
}

func (u ULID) Less(b UUID) bool {
	return bytes.Compare(u[:], b[:]) < 0
}

func (u *ULID) Compare(b UUID) int {
	return bytes.Compare(u[:], b[:])
}

// decodeMap maps ASCII chars (case-insensitive) to 5-bit values.
var decodeMap [256]byte

func init() {
	for i := range decodeMap {
		decodeMap[i] = 0xFF // invalid
	}
	for i, c := range crockfordAlphabet {
		decodeMap[c] = byte(i)
		decodeMap[lower(rune(c))] = byte(i)
	}
	// Add common lookalikes
	decodeMap['O'], decodeMap['o'] = decodeMap['0'], decodeMap['0']
	decodeMap['I'], decodeMap['i'] = decodeMap['1'], decodeMap['1']
	decodeMap['L'], decodeMap['l'] = decodeMap['1'], decodeMap['1']
}

func lower(r rune) byte {
	if 'A' <= r && r <= 'Z' {
		return byte(r + 32)
	}
	return byte(r)
}

func parseCrockford16(s string) ([16]byte, error) {
	var out [16]byte

	if len(s) != 26 {
		return out, fmt.Errorf("crockford: input must be 26 characters for 16 bytes")
	}

	var buf uint64
	bits := 0
	oi := 0

	for i := 0; i < 26; i++ {
		val := decodeMap[s[i]]
		if val == 0xFF {
			return out, fmt.Errorf("crockford: invalid character")
		}

		buf = (buf << 5) | uint64(val)
		bits += 5

		for bits >= 8 && oi < 16 {
			bits -= 8
			out[oi] = byte(buf >> bits)
			buf &= (1<<bits - 1)
			oi++
		}
	}

	if oi != 16 {
		return out, fmt.Errorf("crockford: wrong number of output bytes")
	}
	// 26×5 = 130 bits, so we expect exactly 2 leftover bits = 0
	if bits != 2 || buf != 0 {
		return out, fmt.Errorf("crockford: non-zero leftover bits (invalid encoding)")
	}

	return out, nil
}
