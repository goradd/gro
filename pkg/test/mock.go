package test

import (
	"encoding/gob"
	"fmt"
)

// GobEncoder is a mock writer that will emit an error after Count number of encodings.
type GobEncoder struct {
	Count int
}

func (w *GobEncoder) Encode(v interface{}) error {
	if w.Count == 0 {
		return fmt.Errorf("failed write")
	}
	w.Count--
	return nil
}

// GobDecoder is a mock decoder that will emit an error after Count number of decodings.
// Before Count is reached, it will pass the decode on to the provided Decoder.
type GobDecoder struct {
	*gob.Decoder
	Count int
}

func (r *GobDecoder) Decode(v any) error {
	if r.Count == 0 {
		return fmt.Errorf("failed decode")
	}
	r.Count--
	return r.Decoder.Decode(v)
}
