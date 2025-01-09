package query

type PrimaryKeyer interface {
	PrimaryKeyNode() *ColumnNode
}
