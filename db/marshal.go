package db

// Decoder provides support to the codegenerated structures, allowing the decoder to be mocked.
type Decoder interface {
	Decode(v interface{}) error
}

// Encoder provides support to the codegenerated structures, allowing the encoder to be mocked.
type Encoder interface {
	Encode(v interface{}) error
}
