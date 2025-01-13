package query

type conditioner interface {
	setCondition(condition NodeI)
	getCondition() NodeI
}

// Nodes that can have a condition can mix this in
type nodeCondition struct {
	condition NodeI
}

func (c *nodeCondition) setCondition(cond NodeI) {
	c.condition = cond
}

func (c *nodeCondition) getCondition() NodeI {
	return c.condition
}

// NodeSetCondition is used internally by the framework to set a condition on a node.
func NodeSetCondition(n NodeI, condition NodeI) {
	if condition != nil {
		if c, ok := n.(conditioner); !ok {
			panic("cannot set condition on this type of node")
		} else {
			c.setCondition(condition)
		}
	}
}

// NodeCondition is used internally by the framework to get a condition node.
func NodeCondition(n NodeI) NodeI {
	if cn, ok := n.(conditioner); ok {
		return cn.getCondition()
	}
	return nil
}
