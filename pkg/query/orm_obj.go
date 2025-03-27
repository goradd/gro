package query

// OrmObj is the interface describing the functions common to every record type that is returned by queries.
type OrmObj interface {
	String() string
	Key() string
	Label() string
	Initialize()
	Get(string) any
	MarshalBinary() ([]byte, error)
	UnmarshalBinary(data []byte) (err error)
	MarshalJSON() (data []byte, err error)
	MarshalStringMap() map[string]any
	UnmarshalJSON(data []byte) (err error)
	UnmarshalStringMap(m map[string]any) (err error)
}
