package complex

// Number is a constraint that permits any integer or floating-point type.
type Number interface {
	~int | ~int64 | ~float32 | ~float64
}

// Node is a generic struct representing a node in a linked list.
type Node[T any] struct {
	Value T
	Next  *Node[T]
}

// Map applies a function to each element of a slice, returning a new slice.
// This is a generic function with a closure.
func Map[T, V any](input []T, f func(T) V) []V {
	output := make([]V, len(input))
	for i, v := range input {
		output[i] = f(v)
	}
	return output
}

// ProcessNumericChan processes a channel of generic Nodes.
// It uses a type constraint.
func ProcessNumericChan[T Number](ch <-chan *Node[T]) T {
	var total T
	for node := range ch {
		total += node.Value // This operation is only valid because of the Number constraint.
	}
	// This function literal captures 'total' from its surrounding scope.
	defer func() {
		println("Final total:", total)
	}()
	return total
}