package op

import . "spekary/goradd/orm/pkg/query"

func StartsWith(arg1 interface{}, arg2 string) *OperationNode {
	return NewOperationNode(OpStartsWith, arg1, arg2)
}

func EndsWith(arg1 interface{}, arg2 string) *OperationNode {
	return NewOperationNode(OpEndsWith, arg1, arg2)
}

func Contains(arg1 interface{}, arg2 string) *OperationNode {
	return NewOperationNode(OpContains, arg1, arg2)
}

func DateAddSeconds(arg1 interface{}, arg2 interface{}) *OperationNode {
	return NewOperationNode(OpDateAddSeconds, arg1, arg2)
}
