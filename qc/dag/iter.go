package dag

// Operations returns nodes in topological order ready for simulators/renderers.
func (d *DAG) Operations() []*Node {
	order := d.TopoSort()
	out := make([]*Node, len(order))
	for i, id := range order {
		out[i] = d.nodes[id]
	}
	return out
}
