package query

type PrimaryKeyer interface {
	PrimaryKeyNode() *ColumnNode
}

type TableNodeI interface {
	ColumnNodes_() []NodeI
	Columns_() []string
}
