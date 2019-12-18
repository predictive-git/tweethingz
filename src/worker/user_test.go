package worker

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestArrayDiff(t *testing.T) {

	a1 := []int64{1, 2, 3, 4, 5}
	a2 := []int64{6, 7, 3, 4, 5}
	a3 := getArrayDiff(a1, a2)
	assert.Len(t, a3, 2)

	a4 := getArrayDiff(a2, a1)
	assert.Len(t, a4, 2)
}
