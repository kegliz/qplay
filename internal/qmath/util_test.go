package qmath

import (
	"fmt"
	"testing"

	"github.com/itsubaki/q"
	"github.com/stretchr/testify/assert"
)

func TestRandomBit(t *testing.T) {
	assert := assert.New(t)
	one := 0
	for i := 0; i < 100; i++ {
		qrand := &QRand{q.New()}
		if qrand.RandomBit() == 1 {
			one++
		}
	}
	assert.True(one > 45 && one < 55, "one=%d", one)
	fmt.Println(one)
}
