package query

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"golang.org/x/exp/rand"
)

type UUID [16]byte

func NewUUID() UUID {
	return UUID{}
}

// RandomUUID returns a randomized UUID, usable as a v4 UUID
func RandomUUID() UUID {
	var b UUID
	_, err := rand.Read(b[:])
	if err != nil {
		panic(err) // the system RNG generator is down.
	}
	return b
}

// NewV4UUID returns a random UUID
func NewV4UUID() UUID {
	return RandomUUID()
}

// NewV7UUID generates a UUID using v7 rules of 48-bit timestamp
// plus 74-bits of randomness, plus version info
func NewV7UUID() UUID {
	var u UUID
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
	// Overlay the version info
	u[6] = 0x78 | (u[6] & 0x03)
	return u
}

const sz = "00000000000000000000000000000000"

func UUIDFromString(src string) (u UUID, err error) {
	// remove dashes
	s := strings.Map(func(r rune) rune {
		if r == '-' {
			return -1
		}
		return r
	}, src)
	if len(s) > 32 {
		err = fmt.Errorf("invalid UUID: %s", s)
		return
	}
	// pad with zeros
	s = sz[:32-len(s)] + s
	_, err = hex.Decode(u[:], []byte(s))
	return
}

// UUIDFromBytes returns a UUID copied from the upper 16 bytes of the given
// byte slice. Will panic if the byte slice is too small.
func UUIDFromBytes(src []byte) (u UUID, err error) {
	if len(src) < 16 {
		err = fmt.Errorf("byte slice too short")
		return
	}
	for i := 0; i < 16; i++ {
		u[i] = src[i]
	}
	return
}

func (u UUID) String() string {
	s := hex.EncodeToString(u[:])
	return fmt.Sprintf("%s-%s-%s-%s-%s", s[:8], s[8:12], s[12:16], s[16:20], s[20:])
}

func (u UUID) MarshalBinary() ([]byte, error) {
	return u[:], nil
}

func (u *UUID) UnmarshalBinary(b []byte) error {
	var err error
	*u, err = UUIDFromBytes(b)
	return err
}

func (u UUID) MarshalText() ([]byte, error) {
	return []byte(u.String()), nil
}

func (u *UUID) UnmarshalText(b []byte) error {
	var err error
	*u, err = UUIDFromString(string(b))
	return err
}

func (u UUID) Less(b UUID) bool {
	return bytes.Compare(u[:], b[:]) < 0
}

func (u *UUID) Compare(b UUID) int {
	return bytes.Compare(u[:], b[:])
}
