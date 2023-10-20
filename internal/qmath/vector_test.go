package qmath

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExampleNew(t *testing.T) {
	assert := assert.New(t)

	ExampleNew()
	assert.True(true)
}
