package dag

import (
	"fmt"

	"github.com/kegliz/qplay/qc/gate"
)

func checkGate(max int, g gate.Gate, qs []int) error {
	if len(qs) != g.QubitSpan() {
		return ErrSpan
	}
	for _, q := range qs {
		if q < 0 || q >= max {
			return ErrBadQubit
		}
	}
	return nil
}

// DFS cycle-check
func acyclic(d *DAG) error {
	state := make(map[NodeID]int) // 0 unvisited,1 visiting,2 done
	var dfs func(NodeID) error
	dfs = func(v NodeID) error {
		if state[v] == 1 {
			return fmt.Errorf("dag: cycle detected at %d", v)
		}
		if state[v] == 2 {
			return nil
		}
		// start visiting
		state[v] = 1
		for _, ch := range d.nodes[v].children {
			if err := dfs(ch); err != nil {
				return err
			}
		}
		state[v] = 2
		return nil
	}
	for id := range d.nodes {
		if err := dfs(id); err != nil {
			return err
		}
	}
	return nil
}
