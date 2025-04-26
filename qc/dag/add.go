package dag

import (
	"github.com/kegliz/qplay/qc/gate"
)

func (d *DAG) AddGate(g gate.Gate, qs []int) error {
	if d.valid {
		return ErrValidated
	}
	if err := checkGate(d.qubits, g, qs); err != nil {
		return err
	}
	n := &Node{
		ID:     nextID(),
		G:      g,
		Qubits: append([]int(nil), qs...),
		Cbit:   -1,
	}
	d.nodes[n.ID] = n

	// Build edges: parent = last op on each incident qubit.
	for _, q := range qs {
		if prev := d.last[q]; prev != 0 {
			n.parents = append(n.parents, prev)
			d.nodes[prev].children = append(d.nodes[prev].children, n.ID)
		}
		d.last[q] = n.ID
		d.byQ[q] = append(d.byQ[q], n.ID)
	}
	return nil
}

func (d *DAG) AddMeasure(q, c int) error {
	if d.valid {
		return ErrValidated
	}
	if q < 0 || q >= d.qubits {
		return ErrBadQubit
	}
	if c < 0 || c >= d.clbits {
		return ErrBadClbit
	}
	n := &Node{
		ID:     nextID(),
		G:      gate.Measure(),
		Qubits: []int{q},
		Cbit:   c,
	}
	d.nodes[n.ID] = n
	if prev := d.last[q]; prev != 0 {
		n.parents = []NodeID{prev}
		d.nodes[prev].children = append(d.nodes[prev].children, n.ID)
	}
	d.last[q] = n.ID
	d.byQ[q] = append(d.byQ[q], n.ID)
	return nil
}

func (d *DAG) Validate() error {
	if d.valid {
		return nil
	}
	if err := acyclic(d); err != nil {
		return err
	}
	d.valid = true
	return nil
}
