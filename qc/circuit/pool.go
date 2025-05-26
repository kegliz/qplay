package circuit

import "sync"

var operationSlicePool = sync.Pool{
	New: func() any {
		return make([]Operation, 0, 25) // Pre-allocate with reasonable capacity
	},
}

func (c *circuit) OperationsFromPool() []Operation {
	result := operationSlicePool.Get().([]Operation)
	result = result[:len(c.ops)]
	copy(result, c.ops)
	return result
}

func ReturnOperationSlice(slice []Operation) {
	// No need to clear the slice, because we are returning it to the pool
	// and it will be reused with copy.
	operationSlicePool.Put(slice)
	// if cap(slice) <= 1024 { // Prevent memory leaks from very large slices
	// 	operationSlicePool.Put(slice[:0])
	// }
}
