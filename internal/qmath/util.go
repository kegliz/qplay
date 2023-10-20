package qmath

import (
	"github.com/itsubaki/q"
)

type QRand struct {
	*q.Q
}

//var qrand = &QRand{q.New()}

func (qrand QRand) RandomBit() int64 {
	q0 := qrand.Zero()
	qrand.H(q0)
	m0 := qrand.Measure(q0)
	return m0.Int()
}
