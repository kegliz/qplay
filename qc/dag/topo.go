package dag

// Topological order (Kahn) and depth calculation.
func (d *DAG) TopoSort() []NodeID {
	inDeg := make(map[NodeID]int, len(d.nodes))
	for id := range d.nodes {
		inDeg[id] = len(d.nodes[id].parents)
	}
	var q []NodeID
	for id, deg := range inDeg {
		if deg == 0 {
			q = append(q, id)
		}
	}
	var order []NodeID
	for len(q) > 0 {
		v := q[0]
		q = q[1:]
		order = append(order, v)
		for _, ch := range d.nodes[v].children {
			inDeg[ch]--
			if inDeg[ch] == 0 {
				q = append(q, ch)
			}
		}
	}
	return order
}

// Depth = longest path (len = #layers across qubits)
func (d *DAG) Depth() int {
	depth := make(map[NodeID]int)
	max := 0
	for _, v := range d.TopoSort() {
		for _, p := range d.nodes[v].parents {
			if depth[p]+1 > depth[v] {
				depth[v] = depth[p] + 1
			}
		}
		if depth[v] > max {
			max = depth[v]
		}
	}
	return max + 1 // layers are 0-based
}
