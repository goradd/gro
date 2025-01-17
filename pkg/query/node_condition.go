package query

// Conditioner is an interface for nodes that can be given a join condition.
type Conditioner interface {
	Node
	// SetCondition sets the condition that will be used to join the node.
	SetCondition(condition Node)
	// Condition returns the condition that will be used to join the node.
	Condition() Node
}

// nodeCondition is a mixin for nodes that can be joined with a condition.
type nodeCondition struct {
	condition Node
}

// SetCondition sets the condition that will be used to join the node.
func (c *nodeCondition) SetCondition(cond Node) {
	c.condition = cond
}

// Condition returns the condition that will be used to join the node.
func (c *nodeCondition) Condition() Node {
	return c.condition
}

// NodeCondition is used by the ORM to get a condition node.
func NodeCondition(n Node) Node {
	if cn, ok := n.(Conditioner); ok {
		return cn.Condition()
	}
	return nil
}
