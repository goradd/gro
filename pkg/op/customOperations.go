package op

import . "github.com/goradd/gro/pkg/query"

func StartsWith(arg1 any, arg2 string) *OperationNode {
	return NewOperationNode(OpStartsWith, arg1, arg2)
}

func EndsWith(arg1 any, arg2 string) *OperationNode {
	return NewOperationNode(OpEndsWith, arg1, arg2)
}

// Contains returns on operation node that will return true for the following situations:
//   - A text type field contains the given string value
//   - An enum array node contains the given value
func Contains(arg1, arg2 any) *OperationNode {
	return NewOperationNode(OpContains, arg1, arg2)
}

func DateAddSeconds(arg1 interface{}, arg2 interface{}) *OperationNode {
	return NewOperationNode(OpDateAddSeconds, arg1, arg2)
}
